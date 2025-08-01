package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jianyun8023/calibre-api/internal/calibre"
)

const (
	MCPProtocolVersion = "2024-11-05"
	ServerName         = "calibre-mcp-server"
	ServerVersion      = "1.0.0"
)

// MCPServer represents the MCP server
type MCPServer struct {
	tools       *MCPTools
	initialized bool
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(calibreAPI *calibre.Api) *MCPServer {
	return &MCPServer{
		tools: NewMCPTools(calibreAPI),
	}
}

// NewMCPServerWithIntegration creates a new MCP server instance with API integration
func NewMCPServerWithIntegration(calibreAPI *calibre.Api, baseURL string) *MCPServer {
	return &MCPServer{
		tools: NewMCPToolsWithIntegration(calibreAPI, baseURL),
	}
}

// Start starts the MCP server
func (s *MCPServer) Start() error {
	reader := bufio.NewReader(os.Stdin)
	writer := os.Stdout

	for {
		// Read message
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading input: %w", err)
		}

		// Parse message
		var msg Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Handle message
		response := s.handleMessage(&msg)
		if response != nil {
			// Send response
			responseBytes, err := json.Marshal(response)
			if err != nil {
				log.Printf("Error marshaling response: %v", err)
				continue
			}

			fmt.Fprintf(writer, "%s\n", responseBytes)
		}
	}

	return nil
}

// handleMessage processes incoming MCP messages
func (s *MCPServer) handleMessage(msg *Message) *Message {
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "initialized":
		s.initialized = true
		return nil // No response needed for notification
	case "tools/list":
		return s.handleToolsList(msg)
	case "tools/call":
		return s.handleToolsCall(msg)
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

// handleInitialize handles the initialize request
func (s *MCPServer) handleInitialize(msg *Message) *Message {
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

// handleToolsList handles the tools/list request
func (s *MCPServer) handleToolsList(msg *Message) *Message {
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

// handleToolsCall handles the tools/call request
func (s *MCPServer) handleToolsCall(msg *Message) *Message {
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

// LogError logs an error message
func (s *MCPServer) LogError(msg string, err error) {
	log.Printf("MCP Server Error: %s - %v", msg, err)
}
