package calibre

import (
	"fmt"
	"net/http"

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

	// 检查 SSE 管理器是否已初始化
	if c.sseManager == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    503,
			"message": "SSE service is not available",
		})
		return
	}

	// 创建新的 SSE 客户端
	clientID := uuid.New().String()
	client := tasks.NewSSEClient(clientID)

	// 注册客户端
	if err := c.sseManager.RegisterClient(client); err != nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    503,
			"message": fmt.Sprintf("Failed to register client: %v", err),
		})
		return
	}

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
				return
			}

			// 格式化并发送 SSE 消息
			formatted, err := tasks.FormatSSEMessage(msg)
			if err != nil {
				// 记录错误但继续服务其他连接
				fmt.Printf("Error formatting SSE message: %v\n", err)
				continue
			}

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
