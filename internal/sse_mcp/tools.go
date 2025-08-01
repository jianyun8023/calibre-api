package sse_mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// MCPTools 提供 MCP 工具功能
type MCPTools struct {
	baseURL string
	client  *http.Client
}

// NewMCPTools 创建新的 MCP 工具实例
func NewMCPTools(baseURL string) *MCPTools {
	return &MCPTools{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// GetTools 返回所有可用的 MCP 工具
func (t *MCPTools) GetTools() []Tool {
	return []Tool{
		{
			Name:        "search_books",
			Description: "搜索书籍。可以按标题、作者、ISBN等搜索，支持分页和排序。",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"query": {
						Type:        "string",
						Description: "搜索查询词，可以是书名、作者名、ISBN等",
					},
					"limit": {
						Type:        "integer",
						Description: "返回结果数量限制",
						Default:     10,
					},
					"offset": {
						Type:        "integer",
						Description: "分页偏移量",
						Default:     0,
					},
					"sort": {
						Type:        "string",
						Description: "排序方式",
						Enum:        []string{"id:asc", "id:desc", "title:asc", "title:desc", "pubdate:asc", "pubdate:desc"},
						Default:     "id:desc",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "get_book",
			Description: "根据ID获取书籍详细信息",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"id": {
						Type:        "string",
						Description: "书籍ID",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "get_recent_books",
			Description: "获取最近添加的书籍列表",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"limit": {
						Type:        "integer",
						Description: "返回结果数量限制",
						Default:     20,
					},
				},
			},
		},
		{
			Name:        "update_book_metadata",
			Description: "更新书籍元数据",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"id": {
						Type:        "string",
						Description: "书籍ID",
					},
					"metadata": {
						Type:        "object",
						Description: "要更新的元数据",
					},
				},
				Required: []string{"id", "metadata"},
			},
		},
		{
			Name:        "delete_book",
			Description: "删除书籍",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"id": {
						Type:        "string",
						Description: "书籍ID",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "search_metadata",
			Description: "搜索书籍元数据",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"query": {
						Type:        "string",
						Description: "搜索查询词",
					},
					"source": {
						Type:        "string",
						Description: "元数据源",
						Default:     "douban",
					},
				},
				Required: []string{"query"},
			},
		},
	}
}

// CallTool 调用指定的工具
func (t *MCPTools) CallTool(name string, args map[string]interface{}) (*ToolCallResult, error) {
	switch name {
	case "search_books":
		return t.searchBooks(args)
	case "get_book":
		return t.getBook(args)
	case "get_recent_books":
		return t.getRecentBooks(args)
	case "update_book_metadata":
		return t.updateBookMetadata(args)
	case "delete_book":
		return t.deleteBook(args)
	case "search_metadata":
		return t.searchMetadata(args)
	default:
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("未知工具: %s", name),
			}},
			IsError: true,
		}, fmt.Errorf("unknown tool: %s", name)
	}
}

// searchBooks 搜索书籍
func (t *MCPTools) searchBooks(args map[string]interface{}) (*ToolCallResult, error) {
	query, ok := args["query"].(string)
	if !ok {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: "缺少必需参数: query",
			}},
			IsError: true,
		}, fmt.Errorf("missing required parameter: query")
	}

	// 构建查询参数
	params := url.Values{}
	params.Set("q", query)

	if limit, ok := args["limit"]; ok {
		if limitInt, ok := limit.(float64); ok {
			params.Set("limit", strconv.Itoa(int(limitInt)))
		}
	}

	if offset, ok := args["offset"]; ok {
		if offsetInt, ok := offset.(float64); ok {
			params.Set("offset", strconv.Itoa(int(offsetInt)))
		}
	}

	if sort, ok := args["sort"].(string); ok {
		params.Set("sort", sort)
	}

	// 发起 HTTP 请求
	apiURL := fmt.Sprintf("%s/api/search?%s", t.baseURL, params.Encode())
	resp, err := t.client.Get(apiURL)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("API 请求失败: %v", err),
			}},
			IsError: true,
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("读取响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	if resp.StatusCode != http.StatusOK {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("API 请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body)),
			}},
			IsError: true,
		}, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// 解析响应
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("解析响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	// 格式化响应
	resultText, err := json.MarshalIndent(apiResp, "", "  ")
	if err != nil {
		resultText = []byte(fmt.Sprintf("搜索完成，但格式化结果失败: %v", err))
	}

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: string(resultText),
		}},
		IsError: false,
	}, nil
}

// getBook 获取书籍详情
func (t *MCPTools) getBook(args map[string]interface{}) (*ToolCallResult, error) {
	id, ok := args["id"].(string)
	if !ok {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: "缺少必需参数: id",
			}},
			IsError: true,
		}, fmt.Errorf("missing required parameter: id")
	}

	apiURL := fmt.Sprintf("%s/api/book/%s", t.baseURL, id)
	resp, err := t.client.Get(apiURL)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("API 请求失败: %v", err),
			}},
			IsError: true,
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("读取响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: string(body),
		}},
		IsError: resp.StatusCode != http.StatusOK,
	}, nil
}

// getRecentBooks 获取最近书籍
func (t *MCPTools) getRecentBooks(args map[string]interface{}) (*ToolCallResult, error) {
	params := url.Values{}
	if limit, ok := args["limit"]; ok {
		if limitInt, ok := limit.(float64); ok {
			params.Set("limit", strconv.Itoa(int(limitInt)))
		}
	}

	apiURL := fmt.Sprintf("%s/api/recently?%s", t.baseURL, params.Encode())
	resp, err := t.client.Get(apiURL)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("API 请求失败: %v", err),
			}},
			IsError: true,
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("读取响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: string(body),
		}},
		IsError: resp.StatusCode != http.StatusOK,
	}, nil
}

// updateBookMetadata 更新书籍元数据
func (t *MCPTools) updateBookMetadata(args map[string]interface{}) (*ToolCallResult, error) {
	id, ok := args["id"].(string)
	if !ok {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: "缺少必需参数: id",
			}},
			IsError: true,
		}, fmt.Errorf("missing required parameter: id")
	}

	metadata, ok := args["metadata"]
	if !ok {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: "缺少必需参数: metadata",
			}},
			IsError: true,
		}, fmt.Errorf("missing required parameter: metadata")
	}

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("序列化元数据失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	apiURL := fmt.Sprintf("%s/api/book/%s/update", t.baseURL, id)
	resp, err := t.client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("API 请求失败: %v", err),
			}},
			IsError: true,
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("读取响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: string(body),
		}},
		IsError: resp.StatusCode != http.StatusOK,
	}, nil
}

// deleteBook 删除书籍
func (t *MCPTools) deleteBook(args map[string]interface{}) (*ToolCallResult, error) {
	id, ok := args["id"].(string)
	if !ok {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: "缺少必需参数: id",
			}},
			IsError: true,
		}, fmt.Errorf("missing required parameter: id")
	}

	apiURL := fmt.Sprintf("%s/api/book/%s/delete", t.baseURL, id)
	resp, err := t.client.Post(apiURL, "application/json", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("API 请求失败: %v", err),
			}},
			IsError: true,
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("读取响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: string(body),
		}},
		IsError: resp.StatusCode != http.StatusOK,
	}, nil
}

// searchMetadata 搜索元数据
func (t *MCPTools) searchMetadata(args map[string]interface{}) (*ToolCallResult, error) {
	query, ok := args["query"].(string)
	if !ok {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: "缺少必需参数: query",
			}},
			IsError: true,
		}, fmt.Errorf("missing required parameter: query")
	}

	params := url.Values{}
	params.Set("query", query)

	apiURL := fmt.Sprintf("%s/api/metadata/search?%s", t.baseURL, params.Encode())
	resp, err := t.client.Get(apiURL)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("API 请求失败: %v", err),
			}},
			IsError: true,
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("读取响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: string(body),
		}},
		IsError: resp.StatusCode != http.StatusOK,
	}, nil
}
