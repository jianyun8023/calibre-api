package calibre

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jianyun8023/calibre-api/internal/tasks"
)

// streamTasks 处理 SSE 任务流请求
// 使用 Gin 的 c.Stream 方法实现 SSE，确保正确刷新
func (c *Api) streamTasks(r *gin.Context) {
	fmt.Println("[STREAM DEBUG] streamTasks handler called")

	// 检查 SSE 管理器是否已初始化
	if c.sseManager == nil {
		fmt.Println("[STREAM DEBUG] sseManager is nil!")
		r.JSON(503, gin.H{
			"code":    503,
			"message": "SSE service is not available",
		})
		return
	}

	fmt.Println("[STREAM DEBUG] sseManager is available")

	// 创建新的 SSE 客户端
	clientID := uuid.New().String()
	client := tasks.NewSSEClient(clientID)
	fmt.Printf("[STREAM DEBUG] Created client: %s\n", clientID)

	// 注册客户端
	if err := c.sseManager.RegisterClient(client); err != nil {
		fmt.Printf("[STREAM DEBUG] Failed to register client: %v\n", err)
		r.JSON(503, gin.H{
			"code":    503,
			"message": fmt.Sprintf("Failed to register client: %v", err),
		})
		return
	}

	fmt.Printf("[STREAM DEBUG] Client %s registered successfully\n", clientID)

	// 确保在函数退出时注销客户端
	defer c.sseManager.UnregisterClient(clientID)

	// 使用 Gin 的 Stream 方法，它会自动处理 SSE 的刷新
	r.Stream(func(w io.Writer) bool {
		select {
		case <-r.Request.Context().Done():
			// 客户端断开连接
			fmt.Printf("[STREAM DEBUG] Client %s disconnected\n", clientID)
			return false

		case msg, ok := <-client.Channel:
			if !ok {
				// 通道已关闭
				fmt.Printf("[STREAM DEBUG] Client %s channel closed\n", clientID)
				return false
			}

			fmt.Printf("[STREAM DEBUG] Received message for client %s: type=%s\n", clientID, msg.Type)

			// 格式化 SSE 消息
			data, err := json.Marshal(msg)
			if err != nil {
				fmt.Printf("Error marshaling SSE message: %v\n", err)
				return true // 继续等待下一条消息
			}

			// 写入 SSE 格式的数据
			// SSE 格式: "data: <json>\n\n"
			sseData := fmt.Sprintf("data: %s\n\n", string(data))
			_, err = w.Write([]byte(sseData))
			if err != nil {
				fmt.Printf("[STREAM DEBUG] Write error for client %s: %v\n", clientID, err)
				return false // 写入失败，关闭连接
			}

			fmt.Printf("[STREAM DEBUG] Message sent to client %s\n", clientID)
			return true // 继续等待下一条消息

		case <-time.After(100 * time.Millisecond):
			// 短超时，让 Stream 有机会检查连接状态并刷新
			return true
		}
	})
}

// sendInitialTaskList 发送初始任务列表到客户端
// 注意：这个函数现在不再使用，初始列表通过 RegisterClient 中的 goroutine 发送
func sendInitialTaskList(client *tasks.SSEClient) {
	initialTasks := tasks.GetManager().GetTasks()
	fmt.Printf("[STREAM DEBUG] Preparing initial task list: %d tasks\n", len(initialTasks))

	msg := tasks.SSEMessage{
		Type:      tasks.SSEMessageTypeTaskList,
		Tasks:     initialTasks,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := client.Send(msg); err != nil {
		fmt.Printf("[STREAM DEBUG] Failed to send initial task list: %v\n", err)
	} else {
		fmt.Printf("[STREAM DEBUG] Initial task list queued for sending\n")
	}
}
