package tasks

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SSEMessageType 定义 SSE 消息类型
type SSEMessageType string

const (
	// SSEMessageTypeTaskList 任务列表消息
	SSEMessageTypeTaskList SSEMessageType = "task_list"
	// SSEMessageTypeTaskUpdate 任务更新消息
	SSEMessageTypeTaskUpdate SSEMessageType = "task_update"
	// SSEMessageTypeHeartbeat 心跳消息
	SSEMessageTypeHeartbeat SSEMessageType = "heartbeat"
)

// SSEMessage SSE 消息结构
type SSEMessage struct {
	Type      SSEMessageType `json:"type"`
	Tasks     []TaskStatus   `json:"tasks,omitempty"`
	Task      *TaskStatus    `json:"task,omitempty"`
	Timestamp string         `json:"timestamp"`
}

// SSEClient 表示一个 SSE 客户端连接
type SSEClient struct {
	ID         string
	Channel    chan SSEMessage
	LastSeen   time.Time
	mu         sync.RWMutex
	closed     bool
	closedChan chan struct{}
}

// NewSSEClient 创建新的 SSE 客户端
func NewSSEClient(id string) *SSEClient {
	return &SSEClient{
		ID:         id,
		Channel:    make(chan SSEMessage, 10), // 缓冲 10 条消息
		LastSeen:   time.Now(),
		closed:     false,
		closedChan: make(chan struct{}),
	}
}

// Send 发送消息到客户端
func (c *SSEClient) Send(msg SSEMessage) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return fmt.Errorf("client connection closed")
	}
	c.mu.RUnlock()

	select {
	case c.Channel <- msg:
		c.mu.Lock()
		c.LastSeen = time.Now()
		c.mu.Unlock()
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("send timeout")
	}
}

// Close 关闭客户端连接
func (c *SSEClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.closed = true
		close(c.closedChan)
		close(c.Channel)
	}
}

// IsClosed 检查连接是否已关闭
func (c *SSEClient) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// SSEManager 管理所有 SSE 客户端连接
type SSEManager struct {
	clients       map[string]*SSEClient
	mu            sync.RWMutex
	maxClients    int
	heartbeatTick time.Duration
	staleTimeout  time.Duration
	stopChan      chan struct{}
	taskManager   *Manager
}

// NewSSEManager 创建新的 SSE 管理器
func NewSSEManager(taskManager *Manager, maxClients int) *SSEManager {
	if maxClients <= 0 {
		maxClients = 100 // 默认最大 100 个连接
	}

	manager := &SSEManager{
		clients:       make(map[string]*SSEClient),
		maxClients:    maxClients,
		heartbeatTick: 30 * time.Second,
		staleTimeout:  60 * time.Second,
		stopChan:      make(chan struct{}),
		taskManager:   taskManager,
	}

	// 启动心跳和清理 goroutine
	go manager.heartbeatLoop()
	go manager.cleanupLoop()

	return manager
}

// RegisterClient 注册新的 SSE 客户端
func (m *SSEManager) RegisterClient(client *SSEClient) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查连接数限制
	if len(m.clients) >= m.maxClients {
		return fmt.Errorf("maximum client connections reached")
	}

	m.clients[client.ID] = client
	fmt.Printf("[SSE DEBUG] Client %s registered, total clients: %d\n", client.ID, len(m.clients))

	// 发送初始任务列表
	go func() {
		tasks := m.taskManager.GetTasks()
		fmt.Printf("[SSE DEBUG] Sending initial task list to client %s, task count: %d\n", client.ID, len(tasks))
		for i, t := range tasks {
			fmt.Printf("[SSE DEBUG]   Task[%d]: ID=%s, Type=%s, State=%s\n", i, t.ID, t.Type, t.State)
		}
		msg := SSEMessage{
			Type:      SSEMessageTypeTaskList,
			Tasks:     tasks,
			Timestamp: time.Now().Format(time.RFC3339),
		}
		if err := client.Send(msg); err != nil {
			fmt.Printf("[SSE DEBUG] Failed to send initial task list to client %s: %v\n", client.ID, err)
		} else {
			fmt.Printf("[SSE DEBUG] Initial task list sent successfully to client %s\n", client.ID)
		}
	}()

	return nil
}

// UnregisterClient 注销 SSE 客户端
func (m *SSEManager) UnregisterClient(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, exists := m.clients[clientID]; exists {
		client.Close()
		delete(m.clients, clientID)
	}
}

// BroadcastTaskUpdate 广播任务更新到所有客户端
func (m *SSEManager) BroadcastTaskUpdate(task TaskStatus) {
	m.mu.RLock()
	clients := make([]*SSEClient, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.mu.RUnlock()

	msg := SSEMessage{
		Type:      SSEMessageTypeTaskUpdate,
		Task:      &task,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// 并发发送到所有客户端
	for _, client := range clients {
		go func(c *SSEClient) {
			if err := c.Send(msg); err != nil {
				// 发送失败，标记客户端为待清理
				m.UnregisterClient(c.ID)
			}
		}(client)
	}
}

// heartbeatLoop 定期发送心跳消息
func (m *SSEManager) heartbeatLoop() {
	ticker := time.NewTicker(m.heartbeatTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.sendHeartbeat()
		case <-m.stopChan:
			return
		}
	}
}

// sendHeartbeat 发送心跳消息到所有客户端
func (m *SSEManager) sendHeartbeat() {
	m.mu.RLock()
	clients := make([]*SSEClient, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.mu.RUnlock()

	msg := SSEMessage{
		Type:      SSEMessageTypeHeartbeat,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	for _, client := range clients {
		go func(c *SSEClient) {
			if err := c.Send(msg); err != nil {
				m.UnregisterClient(c.ID)
			}
		}(client)
	}
}

// cleanupLoop 定期清理过期连接
func (m *SSEManager) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupStaleConnections()
		case <-m.stopChan:
			return
		}
	}
}

// cleanupStaleConnections 清理过期的连接
func (m *SSEManager) cleanupStaleConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, client := range m.clients {
		client.mu.RLock()
		lastSeen := client.LastSeen
		closed := client.closed
		client.mu.RUnlock()

		if closed || now.Sub(lastSeen) > m.staleTimeout {
			client.Close()
			delete(m.clients, id)
		}
	}
}

// Stop 停止 SSE 管理器
func (m *SSEManager) Stop() {
	close(m.stopChan)

	m.mu.Lock()
	defer m.mu.Unlock()

	// 关闭所有客户端连接
	for _, client := range m.clients {
		client.Close()
	}
	m.clients = make(map[string]*SSEClient)
}

// GetClientCount 获取当前连接的客户端数量
func (m *SSEManager) GetClientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// FormatSSEMessage 格式化 SSE 消息为标准格式
func FormatSSEMessage(msg SSEMessage) (string, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("data: %s\n\n", string(data)), nil
}
