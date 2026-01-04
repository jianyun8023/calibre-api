package calibre

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jianyun8023/calibre-api/internal/tasks"
)

// streamTasks 处理 SSE 任务流请求
func (c *Api) streamTasks(r *gin.Context) {
	// 设置 SSE 响应头
	r.Header("Content-Type", "text/event-stream")
	r.Header("Cache-Control", "no-cache")
	r.Header("Connection", "keep-alive")
	r.Header("Access-Control-Allow-Origin", "*")
	r.Header("X-Accel-Buffering", "no") // 禁用 nginx 缓冲

	fmt.Println("[STREAM DEBUG] streamTasks handler called")

	// 检查 SSE 管理器是否已初始化
	if c.sseManager == nil {
		fmt.Println("[STREAM DEBUG] sseManager is nil!")
		r.JSON(http.StatusServiceUnavailable, gin.H{
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
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    503,
			"message": fmt.Sprintf("Failed to register client: %v", err),
		})
		return
	}

	fmt.Printf("[STREAM DEBUG] Client %s registered successfully\n", clientID)

	// 确保在函数退出时注销客户端
	defer c.sseManager.UnregisterClient(clientID)

	// 获取响应写入器
	w := r.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		r.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Streaming not supported",
		})
		return
	}

	// 立即发送初始任务列表（在进入循环之前）
	initialTasks := tasks.GetManager().GetTasks()
	fmt.Printf("[STREAM DEBUG] Sending initial task list: %d tasks\n", len(initialTasks))
	initialMsg := tasks.SSEMessage{
		Type:      tasks.SSEMessageTypeTaskList,
		Tasks:     initialTasks,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	if formatted, err := tasks.FormatSSEMessage(initialMsg); err == nil {
		fmt.Fprint(w, formatted)
		flusher.Flush()
		fmt.Printf("[STREAM DEBUG] Initial task list sent successfully\n")
	}

	// 监听客户端断开连接
	notify := r.Request.Context().Done()

	// 持续发送消息
	for {
		select {
		case <-notify:
			// 客户端断开连接
			return

		case msg, ok := <-client.Channel:
			if !ok {
				// 通道已关闭
				fmt.Printf("[STREAM DEBUG] Client %s channel closed\n", clientID)
				return
			}

			fmt.Printf("[STREAM DEBUG] Received message for client %s: type=%s, tasks=%d\n", clientID, msg.Type, len(msg.Tasks))

			// 格式化并发送 SSE 消息
			formatted, err := tasks.FormatSSEMessage(msg)
			if err != nil {
				// 记录错误但继续服务其他连接
				fmt.Printf("Error formatting SSE message: %v\n", err)
				continue
			}

			fmt.Printf("[STREAM DEBUG] Sending formatted message to client %s: %s\n", clientID, formatted)

			// 写入响应
			_, err = fmt.Fprint(w, formatted)
			if err != nil {
				// 写入失败，客户端可能已断开
				return
			}

			// 立即刷新缓冲区
			flusher.Flush()
		}
	}
}
