package mcp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jianyun8023/calibre-api/internal/calibre"
)

// MCPTools wraps the Calibre API to provide MCP tools
type MCPTools struct {
	calibreAPI     CalibreAPI
	apiIntegration *APIIntegrationTools
}

func NewMCPTools(calibreAPI interface{}) *MCPTools {
	// 使用类型断言来获取实际的 API 对象
	if api, ok := calibreAPI.(CalibreAPI); ok {
		return &MCPTools{
			calibreAPI: api,
		}
	}
	// 如果类型断言失败，返回 nil
	return &MCPTools{
		calibreAPI: nil,
	}
}

func NewMCPToolsWithIntegration(calibreAPI interface{}, baseURL string) *MCPTools {
	// 使用类型断言来获取实际的 API 对象
	if api, ok := calibreAPI.(*calibre.Api); ok {
		return &MCPTools{
			calibreAPI:     api,
			apiIntegration: NewAPIIntegrationTools(baseURL),
		}
	}
	// 如果类型断言失败，返回没有 API 集成的版本
	return &MCPTools{
		calibreAPI:     nil,
		apiIntegration: NewAPIIntegrationTools(baseURL),
	}
}

// GetTools returns all available MCP tools
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
			Description: "获取最近更新的书籍列表",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
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
				},
			},
		},
		{
			Name:        "get_random_books",
			Description: "获取随机书籍列表",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"limit": {
						Type:        "integer",
						Description: "返回结果数量限制",
						Default:     10,
					},
				},
			},
		},
		{
			Name:        "update_book_metadata",
			Description: "更新书籍元数据信息",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"id": {
						Type:        "string",
						Description: "书籍ID",
					},
					"title": {
						Type:        "string",
						Description: "书籍标题",
					},
					"authors": {
						Type:        "array",
						Description: "作者列表",
					},
					"publisher": {
						Type:        "string",
						Description: "出版社",
					},
					"isbn": {
						Type:        "string",
						Description: "ISBN号码",
					},
					"comments": {
						Type:        "string",
						Description: "书籍简介或评论",
					},
					"tags": {
						Type:        "array",
						Description: "标签列表",
					},
					"rating": {
						Type:        "number",
						Description: "评分（0-10分）",
					},
					"pubdate": {
						Type:        "string",
						Description: "出版日期（ISO 8601格式）",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "delete_book",
			Description: "删除指定的书籍",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"id": {
						Type:        "string",
						Description: "要删除的书籍ID",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "search_metadata",
			Description: "在线搜索书籍元数据信息（豆瓣等）",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"query": {
						Type:        "string",
						Description: "搜索查询词，通常是书名",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "get_metadata_by_isbn",
			Description: "根据ISBN获取书籍元数据信息",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"isbn": {
						Type:        "string",
						Description: "ISBN号码",
					},
				},
				Required: []string{"isbn"},
			},
		},
		{
			Name:        "get_publishers",
			Description: "获取所有出版社列表",
			InputSchema: ToolSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "update_search_index",
			Description: "更新搜索索引，同步最新的书籍数据",
			InputSchema: ToolSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
	}
}

// CallTool executes a tool call
func (t *MCPTools) CallTool(name string, arguments json.RawMessage) (*ToolCallResult, error) {
	switch name {
	case "search_books":
		return t.searchBooks(arguments)
	case "get_book":
		return t.getBook(arguments)
	case "get_recent_books":
		return t.getRecentBooks(arguments)
	case "get_random_books":
		return t.getRandomBooks(arguments)
	case "update_book_metadata":
		return t.updateBookMetadata(arguments)
	case "delete_book":
		return t.deleteBook(arguments)
	case "search_metadata":
		return t.searchMetadata(arguments)
	case "get_metadata_by_isbn":
		return t.getMetadataByISBN(arguments)
	case "get_publishers":
		return t.getPublishers(arguments)
	case "update_search_index":
		return t.updateSearchIndex(arguments)
	default:
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("未知的工具: %s", name),
			}},
			IsError: true,
		}, fmt.Errorf("unknown tool: %s", name)
	}
}

func (t *MCPTools) searchBooks(arguments json.RawMessage) (*ToolCallResult, error) {
	var args SearchBooksArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("参数解析错误: %v", err),
			}},
			IsError: true,
		}, err
	}

	// 设置默认值
	if args.Limit == 0 {
		args.Limit = 10
	}

	// 如果有 API 集成，使用实际的 API 调用
	if t.apiIntegration != nil {
		return t.apiIntegration.SearchBooksAPI(args)
	}

	// 否则返回占位符信息
	result := fmt.Sprintf("搜索书籍: %s (限制: %d, 偏移: %d, 排序: %s)",
		args.Query, args.Limit, args.Offset, args.Sort)

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: result,
		}},
	}, nil
}

func (t *MCPTools) getBook(arguments json.RawMessage) (*ToolCallResult, error) {
	var args GetBookArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("参数解析错误: %v", err),
			}},
			IsError: true,
		}, err
	}

	// 如果有 API 集成，使用实际的 API 调用
	if t.apiIntegration != nil {
		return t.apiIntegration.GetBookAPI(args)
	}

	result := fmt.Sprintf("获取书籍详情: ID=%s", args.ID)

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: result,
		}},
	}, nil
}

func (t *MCPTools) getRecentBooks(arguments json.RawMessage) (*ToolCallResult, error) {
	var args GetRecentBooksArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("参数解析错误: %v", err),
			}},
			IsError: true,
		}, err
	}

	if args.Limit == 0 {
		args.Limit = 10
	}

	result := fmt.Sprintf("获取最近更新的书籍: 限制=%d, 偏移=%d", args.Limit, args.Offset)

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: result,
		}},
	}, nil
}

func (t *MCPTools) getRandomBooks(arguments json.RawMessage) (*ToolCallResult, error) {
	var args GetRandomBooksArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("参数解析错误: %v", err),
			}},
			IsError: true,
		}, err
	}

	if args.Limit == 0 {
		args.Limit = 10
	}

	result := fmt.Sprintf("获取随机书籍: 限制=%d", args.Limit)

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: result,
		}},
	}, nil
}

func (t *MCPTools) updateBookMetadata(arguments json.RawMessage) (*ToolCallResult, error) {
	var args UpdateBookArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("参数解析错误: %v", err),
			}},
			IsError: true,
		}, err
	}

	// 如果有 API 集成，使用实际的 API 调用
	if t.apiIntegration != nil {
		return t.apiIntegration.UpdateBookMetadataAPI(args)
	}

	result := fmt.Sprintf("更新书籍元数据: ID=%s", args.ID)
	if args.Title != "" {
		result += fmt.Sprintf(", 标题=%s", args.Title)
	}
	if len(args.Authors) > 0 {
		result += fmt.Sprintf(", 作者=%v", args.Authors)
	}
	if args.Publisher != "" {
		result += fmt.Sprintf(", 出版社=%s", args.Publisher)
	}

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: result,
		}},
	}, nil
}

func (t *MCPTools) deleteBook(arguments json.RawMessage) (*ToolCallResult, error) {
	var args DeleteBookArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("参数解析错误: %v", err),
			}},
			IsError: true,
		}, err
	}

	// 如果有 API 集成，使用实际的 API 调用
	if t.apiIntegration != nil {
		return t.apiIntegration.DeleteBookAPI(args)
	}

	result := fmt.Sprintf("删除书籍: ID=%s", args.ID)

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: result,
		}},
	}, nil
}

func (t *MCPTools) searchMetadata(arguments json.RawMessage) (*ToolCallResult, error) {
	var args SearchMetadataArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("参数解析错误: %v", err),
			}},
			IsError: true,
		}, err
	}

	result := fmt.Sprintf("搜索书籍元数据: %s", args.Query)

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: result,
		}},
	}, nil
}

func (t *MCPTools) getMetadataByISBN(arguments json.RawMessage) (*ToolCallResult, error) {
	var args GetMetadataByISBNArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("参数解析错误: %v", err),
			}},
			IsError: true,
		}, err
	}

	result := fmt.Sprintf("根据ISBN获取元数据: %s", args.ISBN)

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: result,
		}},
	}, nil
}

func (t *MCPTools) getPublishers(arguments json.RawMessage) (*ToolCallResult, error) {
	result := "获取所有出版社列表"

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: result,
		}},
	}, nil
}

func (t *MCPTools) updateSearchIndex(arguments json.RawMessage) (*ToolCallResult, error) {
	result := "更新搜索索引"

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: result,
		}},
	}, nil
}

// Helper function to convert time string to time.Time
func parseTimeString(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, nil
	}

	// Try different time formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}
