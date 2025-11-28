package calibre

import (
	"context"
	"fmt"

	"github.com/jianyun8023/calibre-api/internal/semantic/qdrant"
	"github.com/jianyun8023/calibre-api/pkg/log"
	"github.com/mark3labs/mcp-go/mcp"
)

// registerTools 注册所有 MCP 工具
func (m *MCPServer) registerTools() {
	log.Info("Registering MCP tools...")

	m.registerSearchTools()
	m.registerBookTools()
	m.registerRecommendationTools()
	m.registerMetadataTools()

	log.Info("MCP tools registered successfully")
}

// ============================================================================
// 搜索工具
// ============================================================================

func (m *MCPServer) registerSearchTools() {
	// search_books - 混合搜索（关键词 + 语义）
	m.mcpServer.AddTool(
		mcp.Tool{
			Name:        "search_books",
			Description: "搜索书籍，支持关键词搜索和语义搜索。可以按标题、作者、出版社、ISBN、标签等搜索。",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "搜索关键词或问题",
					},
					"filter": map[string]interface{}{
						"type":        "string",
						"description": "搜索过滤器类型：title（标题）、author（作者）、publisher（出版社）、isbn、tags（标签）",
						"enum":        []string{"title", "author", "publisher", "isbn", "tags"},
						"default":     "title",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "返回结果数量（默认20）",
						"default":     20,
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "分页偏移量（默认0）",
						"default":     0,
					},
				},
				Required: []string{"query"},
			},
		},
		m.handleSearchBooks,
	)
}

func (m *MCPServer) handleSearchBooks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments type"), nil
	}

	// 提取参数
	query, _ := args["query"].(string)
	if query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	filterType := "title"
	if f, ok := args["filter"].(string); ok {
		filterType = f
	}

	limit := 20
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	offset := 0
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}

	log.Debugf("MCP Tool search_books: query=%s, filter=%s, limit=%d, offset=%d", query, filterType, limit, offset)

	// 执行搜索（调用现有逻辑）
	books, total, err := m.performSearch(query, filterType, limit, offset)
	if err != nil {
		log.Warnf("Search failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	// 格式化结果
	result := map[string]interface{}{
		"books":  books,
		"total":  total,
		"query":  query,
		"limit":  limit,
		"offset": offset,
	}

	return mcp.NewToolResultText(formatToolResult(result)), nil
}

// performSearch 执行搜索（复用现有逻辑）
func (m *MCPServer) performSearch(query, filterType string, limit, offset int) ([]Book, int64, error) {
	searcher, ok := m.api.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		return nil, 0, fmt.Errorf("search service not available")
	}

	// 使用关键词搜索
	semanticBooks, total, err := searcher.SearchByKeyword(query, filterType, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// 转换结果
	books := make([]Book, 0, len(semanticBooks))
	for _, sb := range semanticBooks {
		books = append(books, convertSemanticToBook(sb))
	}

	return books, total, nil
}

// ============================================================================
// 书籍管理工具
// ============================================================================

func (m *MCPServer) registerBookTools() {
	// get_book - 获取书籍详情（只读，安全）
	m.mcpServer.AddTool(
		mcp.Tool{
			Name:        "get_book",
			Description: "根据书籍 ID 获取详细信息，包括标题、作者、出版社、ISBN、摘要等元数据。此操作为只读操作，不会修改任何数据。",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "书籍 ID",
					},
				},
				Required: []string{"id"},
			},
		},
		m.handleGetBook,
	)

	// 注意：update_book_metadata 和 delete_book 工具已被移除
	// 这些危险操作不应该通过 MCP 暴露给 AI，应该由用户通过 Web UI 手动操作
}

func (m *MCPServer) handleGetBook(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments type"), nil
	}

	id, _ := args["id"].(string)
	if id == "" {
		return mcp.NewToolResultError("id parameter is required"), nil
	}

	log.Debugf("MCP Tool get_book: id=%s", id)

	searcher, ok := m.api.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		return mcp.NewToolResultError("search service not available"), nil
	}

	// 通过 ID 搜索（只读操作）
	books, _, err := searcher.SearchByKeyword(id, "id", 1, 0)
	if err != nil || len(books) == 0 {
		return mcp.NewToolResultError("book not found"), nil
	}

	book := convertSemanticToBook(books[0])
	return mcp.NewToolResultText(formatToolResult(book)), nil
}

// handleUpdateBookMetadata 和 handleDeleteBook 已移除
// 危险操作不应该通过 MCP 暴露

// ============================================================================
// 推荐工具
// ============================================================================

func (m *MCPServer) registerRecommendationTools() {
	// random_books - 随机推荐书籍
	m.mcpServer.AddTool(
		mcp.Tool{
			Name:        "random_books",
			Description: "随机推荐书籍，用于发现新书或获取阅读灵感。",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "返回书籍数量（默认10）",
						"default":     10,
					},
				},
			},
		},
		m.handleRandomBooks,
	)

	// recent_books - 最近更新的书籍
	m.mcpServer.AddTool(
		mcp.Tool{
			Name:        "recent_books",
			Description: "获取最近添加或更新的书籍列表。",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "返回书籍数量（默认20）",
						"default":     20,
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "分页偏移量（默认0）",
						"default":     0,
					},
				},
			},
		},
		m.handleRecentBooks,
	)
}

func (m *MCPServer) handleRandomBooks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments type"), nil
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	log.Debugf("MCP Tool random_books: limit=%d", limit)

	searcher, ok := m.api.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		return mcp.NewToolResultError("search service not available"), nil
	}

	// 使用 Qdrant 随机搜索
	semanticBooks, err := searcher.GetRandom(limit)
	if err != nil {
		log.Warnf("Random search failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("random search failed: %v", err)), nil
	}

	books := make([]Book, 0, len(semanticBooks))
	for _, sb := range semanticBooks {
		books = append(books, convertSemanticToBook(sb))
	}

	result := map[string]interface{}{
		"books": books,
		"count": len(books),
	}

	return mcp.NewToolResultText(formatToolResult(result)), nil
}

func (m *MCPServer) handleRecentBooks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments type"), nil
	}

	limit := 20
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	offset := 0
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}

	log.Debugf("MCP Tool recent_books: limit=%d, offset=%d", limit, offset)

	searcher, ok := m.api.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		return mcp.NewToolResultError("search service not available"), nil
	}

	// 获取最近更新的书籍
	semanticBooks, total, err := searcher.GetRecent(limit, offset)
	if err != nil {
		log.Warnf("Get recent books failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("get recent books failed: %v", err)), nil
	}

	books := make([]Book, 0, len(semanticBooks))
	for _, sb := range semanticBooks {
		books = append(books, convertSemanticToBook(sb))
	}

	result := map[string]interface{}{
		"books":  books,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	return mcp.NewToolResultText(formatToolResult(result)), nil
}

// ============================================================================
// 元数据工具
// ============================================================================

func (m *MCPServer) registerMetadataTools() {
	// get_isbn_metadata - 根据 ISBN 查询元数据
	m.mcpServer.AddTool(
		mcp.Tool{
			Name:        "get_isbn_metadata",
			Description: "根据 ISBN 号从在线数据库查询书籍元数据，包括标题、作者、出版社、封面等信息。",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"isbn": map[string]interface{}{
						"type":        "string",
						"description": "ISBN-10 或 ISBN-13 号码",
					},
				},
				Required: []string{"isbn"},
			},
		},
		m.handleGetISBNMetadata,
	)

	// search_metadata - 在线搜索书籍元数据
	m.mcpServer.AddTool(
		mcp.Tool{
			Name:        "search_metadata",
			Description: "在在线数据库搜索书籍元数据，支持按标题、作者等搜索。",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "书籍标题",
					},
					"author": map[string]interface{}{
						"type":        "string",
						"description": "作者姓名（可选）",
					},
				},
				Required: []string{"title"},
			},
		},
		m.handleSearchMetadata,
	)
}

func (m *MCPServer) handleGetISBNMetadata(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments type"), nil
	}

	isbn, _ := args["isbn"].(string)
	if isbn == "" {
		return mcp.NewToolResultError("isbn parameter is required"), nil
	}

	log.Debugf("MCP Tool get_isbn_metadata: isbn=%s", isbn)

	// 调用豆瓣 API 获取元数据
	var jsonData map[string]interface{}
	resp, err := m.api.http.R().SetResult(&jsonData).Get(m.api.config.Metadata.DoubanUrl + "/v2/book/isbn/" + isbn)
	if err != nil || resp.StatusCode() != 200 {
		log.Warnf("Failed to get ISBN metadata: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to get ISBN metadata from douban: %v", err)), nil
	}

	return mcp.NewToolResultText(formatToolResult(jsonData)), nil
}

func (m *MCPServer) handleSearchMetadata(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments type"), nil
	}

	title, _ := args["title"].(string)
	if title == "" {
		return mcp.NewToolResultError("title parameter is required"), nil
	}

	author, _ := args["author"].(string)

	// 构建查询字符串
	query := title
	if author != "" {
		query = title + " " + author
	}

	log.Debugf("MCP Tool search_metadata: query=%s", query)

	// 调用豆瓣 API 搜索元数据
	var jsonData map[string]interface{}
	resp, err := m.api.http.R().SetResult(&jsonData).SetQueryParam("q", query).Get(m.api.config.Metadata.DoubanUrl + "/v2/book/search")
	if err != nil || resp.StatusCode() != 200 {
		log.Warnf("Failed to search metadata: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to search metadata from douban: %v", err)), nil
	}

	return mcp.NewToolResultText(formatToolResult(jsonData)), nil
}
