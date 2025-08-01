package sse_mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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
func NewSSEMCPServer(baseURL string) *SSEMCPServer {
	return &SSEMCPServer{
		tools: NewMCPTools(baseURL),
	}
}

// NewSSEMCPServerWithIntegration creates a new SSE-based MCP server instance with API integration
func NewSSEMCPServerWithIntegration(calibreAPI interface{}, baseURL string) *SSEMCPServer {
	return &SSEMCPServer{
		tools: NewMCPTools(baseURL),
	}
}

// SetupRoutes sets up the SSE MCP routes
func (s *SSEMCPServer) SetupRoutes(r *gin.RouterGroup) {
	log.Printf("Setting up SSE MCP routes on group: %+v", r)
	mcpGroup := r.Group("/mcp")
	{
		// SSE endpoint for MCP communication
		mcpGroup.GET("/connect", s.handleSSEConnect)

		// HTTP endpoints for MCP operations
		mcpGroup.POST("/initialize", s.handleHTTPInitialize)
		mcpGroup.GET("/tools/list", s.handleHTTPToolsList)
		mcpGroup.POST("/tools/call", s.handleHTTPToolsCall)

		// Status endpoint
		mcpGroup.GET("/status", s.handleStatus)
	}
	log.Printf("SSE MCP routes setup completed")
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

// handleStatus handles status request
func (s *SSEMCPServer) handleStatus(c *gin.Context) {
	log.Printf("SSE MCP Status endpoint called")
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{
		"status":            "running",
		"initialized":       s.initialized,
		"connected_clients": s.GetConnectedClients(),
		"server_info": map[string]interface{}{
			"name":    ServerName,
			"version": ServerVersion,
		},
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
