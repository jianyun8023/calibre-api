package tasks

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	// MaxHistorySize 任务历史记录的最大数量
	MaxHistorySize = 50
	// TaskStateRunning 任务运行中状态
	TaskStateRunning = "running"
	// TaskStateCompleted 任务完成状态
	TaskStateCompleted = "completed"
	// TaskStateError 任务错误状态
	TaskStateError = "error"
	// ProgressComplete 任务完成的进度值
	ProgressComplete = 100
)

// Manager 任务管理器，负责任务的创建、执行和状态跟踪
type Manager struct {
	tasks      map[string]Task
	history    []TaskStatus
	mu         sync.RWMutex
	sseManager *SSEManager // SSE 管理器，用于实时推送任务更新
}

var (
	instance *Manager
	once     sync.Once
)

// GetManager 获取任务管理器的单例实例
func GetManager() *Manager {
	once.Do(func() {
		instance = &Manager{
			tasks:      make(map[string]Task),
			history:    make([]TaskStatus, 0),
			sseManager: nil, // 将在 SetSSEManager 中设置
		}
	})
	return instance
}

// SetSSEManager 设置 SSE 管理器
func (m *Manager) SetSSEManager(sseManager *SSEManager) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sseManager = sseManager
}

// broadcastTaskUpdate 广播任务更新（内部方法）
func (m *Manager) broadcastTaskUpdate(status TaskStatus) {
	if m.sseManager != nil {
		m.sseManager.BroadcastTaskUpdate(status)
	}
}

// BroadcastTaskProgress 广播任务进度更新（供任务实现调用）
func (m *Manager) BroadcastTaskProgress(taskID string) {
	m.mu.RLock()
	task, exists := m.tasks[taskID]
	m.mu.RUnlock()

	if exists {
		status := task.GetStatus()
		m.broadcastTaskUpdate(status)
	}
}

// StartTask 启动一个新任务
// t: 任务类型
// mode: 任务模式（全量/增量）
// factory: 任务工厂函数，用于创建具体的任务实例
// 返回任务 ID 和可能的错误
func (m *Manager) StartTask(t TaskType, mode TaskMode, factory func(string) Task) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if task of same type is running
	// Exception: DeleteBook and UpdateMetadata tasks can run concurrently
	if t != TaskTypeDeleteBook && t != TaskTypeUpdateMetadata && t != TaskTypeCheckMissing {
		for _, task := range m.tasks {
			status := task.GetStatus()
			if status.Type == t && status.State == TaskStateRunning {
				return "", fmt.Errorf("task of type %s is already running", t)
			}
		}
	}

	id := uuid.New().String()
	task := factory(id)
	m.tasks[id] = task

	// 广播任务启动
	initialStatus := task.GetStatus()
	m.broadcastTaskUpdate(initialStatus)

	go func() {
		err := task.Run()
		m.mu.Lock()
		defer m.mu.Unlock()

		status := task.GetStatus()
		status.EndTime = time.Now()
		if err != nil {
			status.State = TaskStateError
			status.Error = err.Error()
		} else if status.State == TaskStateRunning {
			status.State = TaskStateCompleted
			status.Progress = ProgressComplete
		}

		// 广播任务完成/失败状态
		m.broadcastTaskUpdate(status)

		// Move to history
		m.history = append([]TaskStatus{status}, m.history...)
		if len(m.history) > MaxHistorySize {
			m.history = m.history[:MaxHistorySize]
		}
		delete(m.tasks, id)
	}()

	return id, nil
}

// StopTask 停止指定的任务
func (m *Manager) StopTask(id string) error {
	m.mu.RLock()
	task, exists := m.tasks[id]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("task not found")
	}

	task.Stop()
	return nil
}

// GetTasks 获取所有任务状态（包括活动任务和历史任务）
func (m *Manager) GetTasks() []TaskStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []TaskStatus

	// Active tasks
	for _, task := range m.tasks {
		result = append(result, task.GetStatus())
	}

	// History
	result = append(result, m.history...)

	return result
}
