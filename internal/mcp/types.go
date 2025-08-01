package mcp

import (
	"encoding/json"
	"time"
)

// MCP Protocol Types

type Message struct {
	ID     int             `json:"id,omitempty"`
	Type   string          `json:"type"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *ErrorInfo      `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type InitializeRequest struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ClientInfo      ClientInfo   `json:"clientInfo"`
}

type Capabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool Types

type Tool struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema ToolSchema `json:"inputSchema"`
}

type ToolSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Default     any      `json:"default,omitempty"`
}

type ToolCallRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ToolCallResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

type Content struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// Book Search Types

type SearchBooksArgs struct {
	Query  string `json:"query"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
	Sort   string `json:"sort,omitempty"`
}

type GetBookArgs struct {
	ID string `json:"id"`
}

type UpdateBookArgs struct {
	ID        string    `json:"id"`
	Title     string    `json:"title,omitempty"`
	Authors   []string  `json:"authors,omitempty"`
	Publisher string    `json:"publisher,omitempty"`
	ISBN      string    `json:"isbn,omitempty"`
	Comments  string    `json:"comments,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	Rating    float64   `json:"rating,omitempty"`
	PubDate   time.Time `json:"pubdate,omitempty"`
}

type DeleteBookArgs struct {
	ID string `json:"id"`
}

type GetRecentBooksArgs struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

type GetRandomBooksArgs struct {
	Limit int `json:"limit,omitempty"`
}

type SearchMetadataArgs struct {
	Query string `json:"query"`
}

type GetMetadataByISBNArgs struct {
	ISBN string `json:"isbn"`
}

// Response wrapper for consistent API responses
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
