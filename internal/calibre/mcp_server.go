package calibre

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/pkg/log"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer MCP 服务器封装
type MCPServer struct {
	mcpServer            *server.MCPServer
	sseServer            *server.SSEServer
	streamableHTTPServer *server.StreamableHTTPServer
	api                  *Api
	config               MCPConfig
}

// NewMCPServer 创建 MCP 服务器实例
func NewMCPServer(api *Api, config MCPConfig) *MCPServer {
	// 创建 mcp-go 服务器
	s := server.NewMCPServer(config.ServerName, config.Version)

	log.Infof("Creating MCP Server: name=%s, version=%s, transport=%s",
		config.ServerName, config.Version, config.Transport)

	m := &MCPServer{
		mcpServer: s,
		api:       api,
		config:    config,
	}

	// 初始化传输层（根据配置）
	if config.Transport == "sse" || config.Transport == "" {
		// 使用配置的端点创建 SSEServer
		m.sseServer = server.NewSSEServer(s,
			server.WithSSEEndpoint(config.SSEEndpoint),
			server.WithMessageEndpoint(config.MessageEndpoint))
		log.Infof("SSE transport initialized with endpoints: sse=%s, message=%s",
			config.SSEEndpoint, config.MessageEndpoint)
	} else if config.Transport == "http" {
		m.streamableHTTPServer = server.NewStreamableHTTPServer(s)
		log.Info("StreamableHTTP transport initialized")
	} else {
		log.Warnf("Unknown transport mode: %s, defaulting to SSE", config.Transport)
		m.sseServer = server.NewSSEServer(s)
	}

	// 注册所有功能
	m.registerTools() // 阶段二：工具注册
	// m.registerResources() // 阶段三实现
	// m.registerPrompts()   // 阶段三实现

	return m
}

// Mount 将 MCP 服务器挂载到 Gin 路由
func (m *MCPServer) Mount(r *gin.Engine) {
	if m.sseServer != nil {
		// SSE 模式
		sseEndpoint := m.config.SSEEndpoint
		messageEndpoint := m.config.MessageEndpoint
		if sseEndpoint == "" {
			sseEndpoint = "/mcp/sse"
		}
		if messageEndpoint == "" {
			messageEndpoint = "/mcp/message"
		}

		log.Infof("Mounting MCP SSE endpoints: %s, %s", sseEndpoint, messageEndpoint)

		// 使用 SSEServer 的专用 Handler 方法
		r.GET(sseEndpoint, gin.WrapH(m.sseServer.SSEHandler()))
		r.POST(messageEndpoint, gin.WrapH(m.sseServer.MessageHandler()))

	} else if m.streamableHTTPServer != nil {
		// StreamableHTTP 模式
		endpoint := "/mcp"
		log.Infof("Mounting MCP StreamableHTTP endpoint: %s", endpoint)

		// StreamableHTTPServer 也实现了 http.Handler
		r.Any(endpoint, gin.WrapH(m.streamableHTTPServer))
	}
}

// registerTools 声明（实现在 mcp_tools.go）

// registerResources 注册 MCP 资源（阶段三实现）
func (m *MCPServer) registerResources() {
	log.Info("Registering MCP resources...")
	// 将在阶段三实现
}

// registerPrompts 注册 MCP 提示（阶段三实现）
func (m *MCPServer) registerPrompts() {
	log.Info("Registering MCP prompts...")
	// 将在阶段三实现
}

// formatToolResult 格式化工具执行结果为文本
func formatToolResult(result interface{}) string {
	// 简单的 JSON 格式化，后续可以优化
	switch v := result.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		// 使用 JSON 序列化
		bytes, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Sprintf("%+v", v)
		}
		return string(bytes)
	}
}
