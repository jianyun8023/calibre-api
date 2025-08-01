package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/calibre"
)

// SSEMCPServer represents the SSE-based MCP server
type SSEMCPServer struct {
	tools       *MCPTools
	initialized bool
	clients     sync.Map // map[string]*SSEClient
	mu          sync.RWMutex
}

// SSEClient represents a connected SSE client
type SSEClient struct {
	ID       string
	Writer   gin.ResponseWriter
	Flusher  http.Flusher
	Context  context.Context
	Cancel   context.CancelFunc
	LastPing time.Time
}

// NewSSEMCPServer creates a new SSE-based MCP server instance
func NewSSEMCPServer(calibreAPI *calibre.Api) *SSEMCPServer {
	return &SSEMCPServer{
		tools: NewMCPTools(calibreAPI),
	}
}

// NewSSEMCPServerWithIntegration creates a new SSE-based MCP server instance with API integration
func NewSSEMCPServerWithIntegration(calibreAPI *calibre.Api, baseURL string) *SSEMCPServer {
	return &SSEMCPServer{
		tools: NewMCPToolsWithIntegration(calibreAPI, baseURL),
	}
}

// SetupRoutes sets up the SSE MCP routes
func (s *SSEMCPServer) SetupRoutes(r *gin.RouterGroup) {
	mcp := r.Group("/mcp")
	{
		// SSE endpoint for MCP communication
		mcp.GET("/connect", s.handleSSEConnect)

		// HTTP endpoints for MCP operations
		mcp.POST("/initialize", s.handleHTTPInitialize)
		mcp.GET("/tools/list", s.handleHTTPToolsList)
		mcp.POST("/tools/call", s.handleHTTPToolsCall)

		// WebSocket alternative (optional)
		// mcp.GET("/ws", s.handleWebSocket)
	}
}

// handleSSEConnect handles SSE connection establishment
func (s *SSEMCPServer) handleSSEConnect(c *gin.Context) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Get flusher
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	// Create client context
	ctx, cancel := context.WithCancel(c.Request.Context())

	// Generate client ID
	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())

	// Create SSE client
	client := &SSEClient{
		ID:       clientID,
		Writer:   c.Writer,
		Flusher:  flusher,
		Context:  ctx,
		Cancel:   cancel,
		LastPing: time.Now(),
	}

	// Store client
	s.clients.Store(clientID, client)
	defer func() {
		s.clients.Delete(clientID)
		cancel()
	}()

	// Send connection established message
	s.sendSSEMessage(client, "connected", map[string]interface{}{
		"client_id": clientID,
		"server": map[string]interface{}{
			"name":    ServerName,
			"version": ServerVersion,
		},
	})

	// Keep connection alive with ping
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sendSSEMessage(client, "ping", map[string]interface{}{
				"timestamp": time.Now().Unix(),
			})
		}
	}
}

// sendSSEMessage sends an SSE message to a client
func (s *SSEMCPServer) sendSSEMessage(client *SSEClient, event string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Format SSE message
	message := fmt.Sprintf("event: %s\ndata: %s\n\n", event, jsonData)

	// Write to client
	_, err = fmt.Fprint(client.Writer, message)
	if err != nil {
		return err
	}

	client.Flusher.Flush()
	return nil
}

// handleHTTPInitialize handles HTTP-based initialize request
func (s *SSEMCPServer) handleHTTPInitialize(c *gin.Context) {
	var req InitializeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid initialize request parameters",
			"details": err.Error(),
		})
		return
	}

	result := InitializeResult{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    ServerName,
			Version: ServerVersion,
		},
	}

	s.initialized = true

	c.JSON(http.StatusOK, gin.H{
		"result": result,
		"status": "initialized",
	})
}

// handleHTTPToolsList handles HTTP-based tools/list request
func (s *SSEMCPServer) handleHTTPToolsList(c *gin.Context) {
	if !s.initialized {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Server not initialized",
			"code":  -32002,
		})
		return
	}

	tools := s.tools.GetTools()
	c.JSON(http.StatusOK, gin.H{
		"tools": tools,
		"count": len(tools),
	})
}

// handleHTTPToolsCall handles HTTP-based tools/call request
func (s *SSEMCPServer) handleHTTPToolsCall(c *gin.Context) {
	if !s.initialized {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Server not initialized",
			"code":  -32002,
		})
		return
	}

	var req ToolCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid tool call request parameters",
			"details": err.Error(),
		})
		return
	}

	// Execute tool
	result, err := s.tools.CallTool(req.Name, req.Arguments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Tool execution error",
			"details": err.Error(),
			"tool":    req.Name,
		})
		return
	}

	// Send result to all connected SSE clients (optional)
	s.broadcastToSSEClients("tool_result", map[string]interface{}{
		"tool":      req.Name,
		"arguments": req.Arguments,
		"result":    result,
		"timestamp": time.Now().Unix(),
	})

	c.JSON(http.StatusOK, gin.H{
		"result": result,
		"tool":   req.Name,
	})
}

// broadcastToSSEClients broadcasts a message to all connected SSE clients
func (s *SSEMCPServer) broadcastToSSEClients(event string, data interface{}) {
	s.clients.Range(func(key, value interface{}) bool {
		client := value.(*SSEClient)
		if err := s.sendSSEMessage(client, event, data); err != nil {
			log.Printf("Failed to send SSE message to client %s: %v", client.ID, err)
			// Remove failed client
			s.clients.Delete(key)
		}
		return true
	})
}

// HandleSSEMessage processes SSE messages (for bidirectional communication)
func (s *SSEMCPServer) HandleSSEMessage(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id required"})
		return
	}

	// Read message from request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Parse message
	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON message"})
		return
	}

	// Handle message using existing logic
	response := s.handleMessage(&msg)

	if response != nil {
		// Send response via SSE if client is connected
		if clientValue, ok := s.clients.Load(clientID); ok {
			client := clientValue.(*SSEClient)
			s.sendSSEMessage(client, "response", response)
		}

		c.JSON(http.StatusOK, gin.H{"status": "message_processed"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "no_response_needed"})
	}
}

// handleMessage processes incoming MCP messages (reuse existing logic)
func (s *SSEMCPServer) handleMessage(msg *Message) *Message {
	switch msg.Method {
	case "initialize":
		return s.handleInitializeMessage(msg)
	case "initialized":
		s.initialized = true
		return nil // No response needed for notification
	case "tools/list":
		return s.handleToolsListMessage(msg)
	case "tools/call":
		return s.handleToolsCallMessage(msg)
	default:
		return &Message{
			ID:   msg.ID,
			Type: "response",
			Error: &ErrorInfo{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", msg.Method),
			},
		}
	}
}

// handleInitializeMessage handles the initialize message
func (s *SSEMCPServer) handleInitializeMessage(msg *Message) *Message {
	var req InitializeRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return &Message{
			ID:   msg.ID,
			Type: "response",
			Error: &ErrorInfo{
				Code:    -32602,
				Message: "Invalid initialize request parameters",
				Data:    err.Error(),
			},
		}
	}

	result := InitializeResult{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    ServerName,
			Version: ServerVersion,
		},
	}

	resultBytes, _ := json.Marshal(result)
	return &Message{
		ID:     msg.ID,
		Type:   "response",
		Result: resultBytes,
	}
}

// handleToolsListMessage handles the tools/list message
func (s *SSEMCPServer) handleToolsListMessage(msg *Message) *Message {
	if !s.initialized {
		return &Message{
			ID:   msg.ID,
			Type: "response",
			Error: &ErrorInfo{
				Code:    -32002,
				Message: "Server not initialized",
			},
		}
	}

	tools := s.tools.GetTools()
	result := struct {
		Tools []Tool `json:"tools"`
	}{
		Tools: tools,
	}

	resultBytes, _ := json.Marshal(result)
	return &Message{
		ID:     msg.ID,
		Type:   "response",
		Result: resultBytes,
	}
}

// handleToolsCallMessage handles the tools/call message
func (s *SSEMCPServer) handleToolsCallMessage(msg *Message) *Message {
	if !s.initialized {
		return &Message{
			ID:   msg.ID,
			Type: "response",
			Error: &ErrorInfo{
				Code:    -32002,
				Message: "Server not initialized",
			},
		}
	}

	var req ToolCallRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return &Message{
			ID:   msg.ID,
			Type: "response",
			Error: &ErrorInfo{
				Code:    -32602,
				Message: "Invalid tool call request parameters",
				Data:    err.Error(),
			},
		}
	}

	result, err := s.tools.CallTool(req.Name, req.Arguments)
	if err != nil {
		return &Message{
			ID:   msg.ID,
			Type: "response",
			Error: &ErrorInfo{
				Code:    -32603,
				Message: "Tool execution error",
				Data:    err.Error(),
			},
		}
	}

	resultBytes, _ := json.Marshal(result)
	return &Message{
		ID:     msg.ID,
		Type:   "response",
		Result: resultBytes,
	}
}

// GetConnectedClients returns the number of connected SSE clients
func (s *SSEMCPServer) GetConnectedClients() int {
	count := 0
	s.clients.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// LogError logs an error message
func (s *SSEMCPServer) LogError(msg string, err error) {
	log.Printf("SSE MCP Server Error: %s - %v", msg, err)
}
