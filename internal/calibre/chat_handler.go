package calibre

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/chat"
)

// CreateConversation 创建新对话
func (c *Api) CreateConversation(r *gin.Context) {
	if c.chatDB == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat database not initialized"})
		return
	}

	var req struct {
		Title string `json:"title" binding:"required"`
	}

	if err := r.BindJSON(&req); err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	conv, err := c.chatDB.CreateConversation(req.Title)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	r.JSON(http.StatusOK, conv)
}

// ListConversations 列出对话历史
func (c *Api) ListConversations(r *gin.Context) {
	if c.chatDB == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat database not initialized"})
		return
	}

	conversations, err := c.chatDB.ListConversations(50)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	r.JSON(http.StatusOK, conversations)
}

// GetConversation 获取对话详情
func (c *Api) GetConversation(r *gin.Context) {
	if c.chatDB == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat database not initialized"})
		return
	}

	conversationID := r.Param("id")

	conv, err := c.chatDB.GetConversation(conversationID)
	if err != nil {
		r.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	r.JSON(http.StatusOK, conv)
}

// GetConversationMessages 获取对话消息
func (c *Api) GetConversationMessages(r *gin.Context) {
	if c.chatDB == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat database not initialized"})
		return
	}

	conversationID := r.Param("id")

	messages, err := c.chatDB.GetConversationMessages(conversationID)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	r.JSON(http.StatusOK, messages)
}

// DeleteConversation 删除对话
func (c *Api) DeleteConversation(r *gin.Context) {
	if c.chatDB == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat database not initialized"})
		return
	}

	conversationID := r.Param("id")

	if err := c.chatDB.DeleteConversation(conversationID); err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	r.JSON(http.StatusOK, gin.H{"message": "conversation deleted"})
}

// DeleteMessage 删除消息
func (c *Api) DeleteMessage(r *gin.Context) {
	if c.chatDB == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat database not initialized"})
		return
	}

	messageID := r.Param("id")

	if err := c.chatDB.DeleteMessage(messageID); err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	r.JSON(http.StatusOK, gin.H{"message": "message deleted"})
}

// SendMessage 发送消息（流式响应）
func (c *Api) SendMessage(r *gin.Context) {
	if c.chatDB == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat database not initialized"})
		return
	}
	if c.chatAgent == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{"error": "Chat agent not initialized"})
		return
	}

	conversationID := r.Param("id")

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := r.BindJSON(&req); err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 检查对话是否存在
	conv, err := c.chatDB.GetConversation(conversationID)
	if err != nil {
		r.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	// 保存用户消息
	userMsg := &chat.Message{
		ConversationID: conversationID,
		Role:           "user",
		Content:        req.Content,
	}
	if err := c.chatDB.SaveMessage(userMsg); err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 设置 SSE 响应头
	r.Header("Content-Type", "text/event-stream")
	r.Header("Cache-Control", "no-cache")
	r.Header("Connection", "keep-alive")
	r.Header("X-Accel-Buffering", "no")

	// 获取对话历史
	messages, _ := c.chatDB.GetConversationMessages(conversationID)

	// 创建上下文
	ctx, cancel := context.WithTimeout(r.Request.Context(), 60*time.Second)
	defer cancel()

	// 检查是否需要使用搜索
	var fullResponse string
	var responseBooks []map[string]interface{}

	if c.chatAgent.ShouldUseSearch(req.Content) {
		// 使用搜索并生成回复
		response, books, err := c.chatAgent.SearchAndRespond(ctx, req.Content)
		if err != nil {
			fmt.Fprintf(r.Writer, "data: {\"error\": \"%s\"}\n\n", err.Error())
			r.Writer.Flush()
			return
		}
		fullResponse = response
		responseBooks = books

		// 发送元数据
		if len(books) > 0 {
			metadata := map[string]interface{}{
				"type":  "metadata",
				"books": books,
			}
			if metadataJSON, err := json.Marshal(metadata); err == nil {
				fmt.Fprintf(r.Writer, "data: %s\n\n", string(metadataJSON))
				r.Writer.Flush()
			}
		}

		// 流式发送响应
		for _, char := range response {
			fmt.Fprintf(r.Writer, "data: %s\n\n", string(char))
			r.Writer.Flush()
			time.Sleep(10 * time.Millisecond) // 模拟流式效果
		}
	} else {
		// 普通对话
		err := c.chatAgent.ChatStream(ctx, req.Content, messages, func(chunk string) error {
			fullResponse += chunk
			fmt.Fprintf(r.Writer, "data: %s\n\n", chunk)
			r.Writer.Flush()
			return nil
		})

		if err != nil {
			fmt.Fprintf(r.Writer, "data: {\"error\": \"%s\"}\n\n", err.Error())
			r.Writer.Flush()
			return
		}
	}

	// 解析思考和答案
	parsed := chat.ParseThinkingResponse(fullResponse)

	// 保存 AI 响应
	assistantMsg := &chat.Message{
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        parsed.Answer,
		Thinking:       parsed.Thinking,
	}

	// 如果有书籍推荐，保存到 metadata
	if len(responseBooks) > 0 {
		metadata := map[string]interface{}{
			"books": responseBooks,
		}
		if metadataJSON, err := json.Marshal(metadata); err == nil {
			assistantMsg.Metadata = string(metadataJSON)
		}
	}

	c.chatDB.SaveMessage(assistantMsg)

	// 自动生成标题（如果是新对话）
	if conv.Title == "新对话" {
		go func() {
			// 创建新的上下文，因为请求上下文可能已取消
			bgCtx := context.Background()
			newTitle, err := c.chatAgent.GenerateTitle(bgCtx, req.Content, parsed.Answer)
			if err == nil && newTitle != "" {
				c.chatDB.UpdateConversationTitle(conversationID, newTitle)
			}
		}()
	}

	// 发送完成信号
	fmt.Fprintf(r.Writer, "data: [DONE]\n\n")
	r.Writer.Flush()
}
