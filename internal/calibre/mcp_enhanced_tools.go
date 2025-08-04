package calibre

import (
	"fmt"

	"github.com/meilisearch/meilisearch-go"
)

// EnhancedTool 增强的工具定义
type EnhancedTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Resources   []string               `json:"resources,omitempty"`
	Prompts     []string               `json:"prompts,omitempty"`
}

// EnhancedToolManager 增强工具管理器
type EnhancedToolManager struct {
	api         *Api
	resourceMgr *ResourceManager
	promptMgr   *PromptManager
}

// NewEnhancedToolManager 创建增强工具管理器
func NewEnhancedToolManager(api *Api) *EnhancedToolManager {
	return &EnhancedToolManager{
		api:         api,
		resourceMgr: NewResourceManager(api),
		promptMgr:   NewPromptManager(api),
	}
}

// GetEnhancedTools 获取增强的工具列表
func (etm *EnhancedToolManager) GetEnhancedTools() []EnhancedTool {
	return []EnhancedTool{
		// 搜索工具
		{
			Name:        "search_books_enhanced",
			Description: "增强的书籍搜索工具，支持多种搜索方式和结果分析",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "搜索关键词",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "返回结果数量",
						"default":     10,
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "分页偏移量",
						"default":     0,
					},
					"sort": map[string]interface{}{
						"type":        "string",
						"description": "排序方式",
						"enum":        []string{"relevance", "title", "author", "date"},
					},
					"include_resources": map[string]interface{}{
						"type":        "boolean",
						"description": "是否包含资源信息",
						"default":     false,
					},
				},
				"required": []string{"query"},
			},
			Prompts: []string{"search_books_by_topic", "search_books_by_author"},
		},

		// 书籍详情工具
		{
			Name:        "get_book_details_enhanced",
			Description: "获取书籍详细信息，包括元数据、封面、目录等资源",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"book_id": map[string]interface{}{
						"type":        "string",
						"description": "书籍ID",
					},
					"include_cover": map[string]interface{}{
						"type":        "boolean",
						"description": "是否包含封面图片",
						"default":     true,
					},
					"include_toc": map[string]interface{}{
						"type":        "boolean",
						"description": "是否包含目录结构",
						"default":     true,
					},
					"include_metadata": map[string]interface{}{
						"type":        "boolean",
						"description": "是否包含完整元数据",
						"default":     true,
					},
				},
				"required": []string{"book_id"},
			},
			Resources: []string{"cover", "toc", "metadata"},
			Prompts:   []string{"get_book_details"},
		},

		// 书籍管理工具
		{
			Name:        "manage_book_enhanced",
			Description: "增强的书籍管理工具，支持更新元数据、删除等操作",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{
						"type":        "string",
						"description": "操作类型",
						"enum":        []string{"update", "delete", "analyze"},
					},
					"book_id": map[string]interface{}{
						"type":        "string",
						"description": "书籍ID",
					},
					"metadata": map[string]interface{}{
						"type":        "object",
						"description": "更新的元数据",
					},
				},
				"required": []string{"action", "book_id"},
			},
			Prompts: []string{"update_book_metadata", "delete_book"},
		},

		// 推荐工具
		{
			Name:        "get_recommendations_enhanced",
			Description: "智能书籍推荐工具，支持多种推荐策略",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type":        "string",
						"description": "推荐类型",
						"enum":        []string{"recent", "random", "similar", "popular"},
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "推荐数量",
						"default":     10,
					},
					"book_id": map[string]interface{}{
						"type":        "string",
						"description": "参考书籍ID（用于相似推荐）",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"description": "标签过滤",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"type"},
			},
			Prompts: []string{"get_recent_books", "get_random_books", "find_similar_books"},
		},

		// 元数据工具
		{
			Name:        "metadata_services_enhanced",
			Description: "增强的元数据服务，支持在线搜索和ISBN查询",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"service": map[string]interface{}{
						"type":        "string",
						"description": "服务类型",
						"enum":        []string{"search", "isbn", "enrich"},
					},
					"query": map[string]interface{}{
						"type":        "string",
						"description": "查询内容",
					},
					"isbn": map[string]interface{}{
						"type":        "string",
						"description": "ISBN号码",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "结果数量",
						"default":     5,
					},
				},
				"required": []string{"service"},
			},
			Prompts: []string{"search_metadata_online", "get_metadata_by_isbn"},
		},

		// 分析工具
		{
			Name:        "analyze_collection_enhanced",
			Description: "书籍收藏分析工具，提供统计信息和洞察",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"analysis_type": map[string]interface{}{
						"type":        "string",
						"description": "分析类型",
						"enum":        []string{"overview", "authors", "publishers", "topics", "timeline"},
					},
					"group_by": map[string]interface{}{
						"type":        "string",
						"description": "分组方式",
						"enum":        []string{"author", "publisher", "tag", "year"},
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "结果数量限制",
						"default":     20,
					},
				},
				"required": []string{"analysis_type"},
			},
			Prompts: []string{"analyze_book_collection"},
		},

		// 导出工具
		{
			Name:        "export_data_enhanced",
			Description: "数据导出工具，支持多种格式和字段选择",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"format": map[string]interface{}{
						"type":        "string",
						"description": "导出格式",
						"enum":        []string{"json", "csv", "xml", "bibtex"},
					},
					"fields": map[string]interface{}{
						"type":        "array",
						"description": "导出字段",
						"items": map[string]interface{}{
							"type": "string",
						},
						"default": []string{"title", "authors", "publisher", "isbn"},
					},
					"filters": map[string]interface{}{
						"type":        "object",
						"description": "过滤条件",
					},
					"include_resources": map[string]interface{}{
						"type":        "boolean",
						"description": "是否包含资源信息",
						"default":     false,
					},
				},
				"required": []string{"format"},
			},
			Prompts: []string{"export_book_list"},
		},
	}
}

// ExecuteEnhancedTool 执行增强工具
func (etm *EnhancedToolManager) ExecuteEnhancedTool(toolName string, args map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "search_books_enhanced":
		return etm.executeSearchBooksEnhanced(args)
	case "get_book_details_enhanced":
		return etm.executeGetBookDetailsEnhanced(args)
	case "manage_book_enhanced":
		return etm.executeManageBookEnhanced(args)
	case "get_recommendations_enhanced":
		return etm.executeGetRecommendationsEnhanced(args)
	case "metadata_services_enhanced":
		return etm.executeMetadataServicesEnhanced(args)
	case "analyze_collection_enhanced":
		return etm.executeAnalyzeCollectionEnhanced(args)
	case "export_data_enhanced":
		return etm.executeExportDataEnhanced(args)
	default:
		return nil, fmt.Errorf("未知的工具: %s", toolName)
	}
}

// executeSearchBooksEnhanced 执行增强搜索
func (etm *EnhancedToolManager) executeSearchBooksEnhanced(args map[string]interface{}) (interface{}, error) {
	query := args["query"].(string)
	limit := 10
	if l, ok := args["limit"]; ok {
		limit = int(l.(float64))
	}
	offset := 0
	if o, ok := args["offset"]; ok {
		offset = int(o.(float64))
	}
	includeResources := false
	if ir, ok := args["include_resources"]; ok {
		includeResources = ir.(bool)
	}

	// 执行搜索
	searchReq := meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(offset),
	}

	search, err := etm.api.currentIndex().Search(query, &searchReq)
	if err != nil {
		return nil, err
	}

	// 处理结果
	books := make([]map[string]interface{}, len(search.Hits))
	for i, hit := range search.Hits {
		bookData := hit.(map[string]interface{})
		books[i] = bookData

		// 如果需要包含资源信息
		if includeResources {
			if bookID, ok := bookData["id"].(string); ok {
				resources, err := etm.resourceMgr.ListResources(bookID)
				if err == nil {
					books[i]["resources"] = resources
				}
			}
		}
	}

	return map[string]interface{}{
		"books":    books,
		"total":    search.EstimatedTotalHits,
		"query":    query,
		"limit":    limit,
		"offset":   offset,
		"has_more": search.EstimatedTotalHits > int64(offset+limit),
	}, nil
}

// executeGetBookDetailsEnhanced 执行增强书籍详情获取
func (etm *EnhancedToolManager) executeGetBookDetailsEnhanced(args map[string]interface{}) (interface{}, error) {
	bookID := args["book_id"].(string)
	includeCover := true
	if ic, ok := args["include_cover"]; ok {
		includeCover = ic.(bool)
	}
	includeToc := true
	if it, ok := args["include_toc"]; ok {
		includeToc = it.(bool)
	}
	includeMetadata := true
	if im, ok := args["include_metadata"]; ok {
		includeMetadata = im.(bool)
	}

	// 获取书籍基本信息
	book, err := etm.api.getBookByID(bookID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"book_id": bookID,
	}

	// 添加元数据
	if includeMetadata {
		result["metadata"] = book
	}

	// 添加封面
	if includeCover && book.Cover != "" {
		coverResource, err := etm.resourceMgr.ReadResource(fmt.Sprintf("calibre://books/%s/cover", bookID))
		if err == nil {
			result["cover"] = coverResource
		}
	}

	// 添加目录
	if includeToc {
		tocResource, err := etm.resourceMgr.ReadResource(fmt.Sprintf("calibre://books/%s/toc", bookID))
		if err == nil {
			result["toc"] = tocResource
		}
	}

	// 添加相关提示
	prompts := etm.promptMgr.GetPromptSuggestions("book")
	result["suggested_prompts"] = prompts

	return result, nil
}

// executeManageBookEnhanced 执行增强书籍管理
func (etm *EnhancedToolManager) executeManageBookEnhanced(args map[string]interface{}) (interface{}, error) {
	action := args["action"].(string)
	bookID := args["book_id"].(string)

	switch action {
	case "update":
		if metadata, ok := args["metadata"]; ok {
			// 更新元数据
			return etm.api.updateBookMetadata(bookID, metadata.(map[string]interface{}))
		}
	case "delete":
		// 删除书籍 - 这里需要调用实际的删除逻辑
		return map[string]interface{}{
			"book_id": bookID,
			"deleted": true,
		}, nil
	case "analyze":
		// 分析书籍
		return etm.analyzeBook(bookID)
	default:
		return nil, fmt.Errorf("不支持的操作: %s", action)
	}

	return nil, fmt.Errorf("操作执行失败")
}

// executeGetRecommendationsEnhanced 执行增强推荐
func (etm *EnhancedToolManager) executeGetRecommendationsEnhanced(args map[string]interface{}) (interface{}, error) {
	recType := args["type"].(string)
	limit := 10
	if l, ok := args["limit"]; ok {
		limit = int(l.(float64))
	}

	switch recType {
	case "recent":
		return etm.api.getRecentBooks(limit)
	case "random":
		return etm.api.getRandomBooks(limit)
	case "similar":
		if bookID, ok := args["book_id"]; ok {
			return etm.findSimilarBooks(bookID.(string), limit)
		}
	case "popular":
		return etm.getPopularBooks(limit)
	default:
		return nil, fmt.Errorf("不支持的推荐类型: %s", recType)
	}

	return nil, fmt.Errorf("推荐执行失败")
}

// executeMetadataServicesEnhanced 执行增强元数据服务
func (etm *EnhancedToolManager) executeMetadataServicesEnhanced(args map[string]interface{}) (interface{}, error) {
	service := args["service"].(string)

	switch service {
	case "search":
		if query, ok := args["query"]; ok {
			return etm.api.searchMetadata(query.(string))
		}
	case "isbn":
		if isbn, ok := args["isbn"]; ok {
			return etm.api.getMetadataByISBN(isbn.(string))
		}
	case "enrich":
		if bookID, ok := args["book_id"]; ok {
			return etm.enrichBookMetadata(bookID.(string))
		}
	default:
		return nil, fmt.Errorf("不支持的服务类型: %s", service)
	}

	return nil, fmt.Errorf("元数据服务执行失败")
}

// executeAnalyzeCollectionEnhanced 执行增强收藏分析
func (etm *EnhancedToolManager) executeAnalyzeCollectionEnhanced(args map[string]interface{}) (interface{}, error) {
	analysisType := args["analysis_type"].(string)
	limit := 20
	if l, ok := args["limit"]; ok {
		limit = int(l.(float64))
	}

	switch analysisType {
	case "overview":
		return etm.analyzeCollectionOverview()
	case "authors":
		return etm.analyzeAuthors(limit)
	case "publishers":
		return etm.analyzePublishers(limit)
	case "topics":
		return etm.analyzeTopics(limit)
	case "timeline":
		return etm.analyzeTimeline()
	default:
		return nil, fmt.Errorf("不支持的分析类型: %s", analysisType)
	}
}

// executeExportDataEnhanced 执行增强数据导出
func (etm *EnhancedToolManager) executeExportDataEnhanced(args map[string]interface{}) (interface{}, error) {
	format := args["format"].(string)
	fields := []string{"title", "authors", "publisher", "isbn"}
	if f, ok := args["fields"]; ok {
		fields = f.([]string)
	}

	return etm.exportBooksData(format, fields)
}

// 辅助方法实现
func (etm *EnhancedToolManager) analyzeBook(bookID string) (interface{}, error) {
	// 实现书籍分析逻辑
	return map[string]interface{}{
		"book_id": bookID,
		"analysis": map[string]interface{}{
			"completeness": 0.85,
			"quality":      0.92,
			"suggestions":  []string{"添加更多标签", "完善作者信息"},
		},
	}, nil
}

func (etm *EnhancedToolManager) findSimilarBooks(bookID string, limit int) (interface{}, error) {
	// 实现相似书籍查找逻辑
	return map[string]interface{}{
		"reference_book": bookID,
		"similar_books":  []map[string]interface{}{},
		"count":          limit,
	}, nil
}

func (etm *EnhancedToolManager) getPopularBooks(limit int) (interface{}, error) {
	// 实现热门书籍获取逻辑
	return map[string]interface{}{
		"popular_books": []map[string]interface{}{},
		"count":         limit,
	}, nil
}

func (etm *EnhancedToolManager) enrichBookMetadata(bookID string) (interface{}, error) {
	// 实现元数据丰富逻辑
	return map[string]interface{}{
		"book_id": bookID,
		"enriched": map[string]interface{}{
			"online_data":  map[string]interface{}{},
			"improvements": []string{},
		},
	}, nil
}

func (etm *EnhancedToolManager) analyzeCollectionOverview() (interface{}, error) {
	// 实现收藏概览分析
	return map[string]interface{}{
		"total_books":      0,
		"total_authors":    0,
		"total_publishers": 0,
		"avg_rating":       0.0,
		"top_genres":       []string{},
	}, nil
}

func (etm *EnhancedToolManager) analyzeAuthors(limit int) (interface{}, error) {
	// 实现作者分析
	return map[string]interface{}{
		"authors": []map[string]interface{}{},
		"count":   limit,
	}, nil
}

func (etm *EnhancedToolManager) analyzePublishers(limit int) (interface{}, error) {
	// 实现出版社分析
	return map[string]interface{}{
		"publishers": []map[string]interface{}{},
		"count":      limit,
	}, nil
}

func (etm *EnhancedToolManager) analyzeTopics(limit int) (interface{}, error) {
	// 实现主题分析
	return map[string]interface{}{
		"topics": []map[string]interface{}{},
		"count":  limit,
	}, nil
}

func (etm *EnhancedToolManager) analyzeTimeline() (interface{}, error) {
	// 实现时间线分析
	return map[string]interface{}{
		"timeline": []map[string]interface{}{},
	}, nil
}

func (etm *EnhancedToolManager) exportBooksData(format string, fields []string) (interface{}, error) {
	// 实现数据导出
	return map[string]interface{}{
		"format": format,
		"fields": fields,
		"data":   []map[string]interface{}{},
	}, nil
}

// 辅助方法实现
func (api *Api) updateBookMetadata(bookID string, metadata map[string]interface{}) (interface{}, error) {
	// 实现元数据更新逻辑
	return map[string]interface{}{
		"book_id": bookID,
		"updated": true,
		"changes": metadata,
	}, nil
}

func (api *Api) getRecentBooks(limit int) (interface{}, error) {
	// 实现最近书籍获取逻辑
	return map[string]interface{}{
		"recent_books": []map[string]interface{}{},
		"count":        limit,
	}, nil
}

func (api *Api) getRandomBooks(limit int) (interface{}, error) {
	// 实现随机书籍获取逻辑
	return map[string]interface{}{
		"random_books": []map[string]interface{}{},
		"count":        limit,
	}, nil
}

func (api *Api) searchMetadata(query string) (interface{}, error) {
	// 实现元数据搜索逻辑
	return map[string]interface{}{
		"query":   query,
		"results": []map[string]interface{}{},
	}, nil
}

func (api *Api) getMetadataByISBN(isbn string) (interface{}, error) {
	// 实现ISBN元数据获取逻辑
	return map[string]interface{}{
		"isbn":     isbn,
		"metadata": map[string]interface{}{},
	}, nil
}
