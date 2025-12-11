# Tasks Document

## Task Breakdown

### Task 1: 创建精简的响应结构体

**Description:** 在 `internal/calibre/mcp_tools.go` 中定义新的响应结构体，用于返回精简的书籍信息。

**Dependencies:** None

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] 创建 `CompactBook` 结构体（6 个字段）
- [ ] 创建 `DetailedBook` 结构体（8 个字段）
- [ ] 添加 JSON tags 和 omitempty 标记
- [ ] 添加结构体注释说明用途

**Implementation Notes:**
```go
// CompactBook 精简的书籍信息，用于列表场景（搜索、推荐等）
// 只包含 LLM 回答问题所需的核心字段，减少 token 消耗
type CompactBook struct {
    ID      string   `json:"id"`
    Title   string   `json:"title"`
    Authors []string `json:"authors"`
    Score   float64  `json:"score,omitempty"`  // 仅搜索结果包含
    Rating  float64  `json:"rating,omitempty"` // 书籍评分
}

// DetailedBook 详细的书籍信息，用于单本书详情场景
// 包含更多元数据，但仍然精简不必要的字段
type DetailedBook struct {
    ID         string   `json:"id"`
    Title      string   `json:"title"`
    Authors    []string `json:"authors"`
    Publisher  string   `json:"publisher,omitempty"`
    ISBN       string   `json:"isbn,omitempty"`
    Comments   string   `json:"comments,omitempty"`    // 限制 500 字符
    TocSummary string   `json:"toc_summary,omitempty"` // TOC 摘要
    Rating     float64  `json:"rating,omitempty"`
}
```

---

### Task 2: 实现辅助函数

**Description:** 创建转换和处理函数，用于将 Book 对象转换为精简格式。

**Dependencies:** Task 1

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] 实现 `toCompactBook()` 函数
- [ ] 实现 `toDetailedBook()` 函数
- [ ] 实现 `truncateComments()` 函数
- [ ] 实现 `generateTocSummary()` 函数
- [ ] 添加函数注释和示例

**Implementation Notes:**
```go
// toCompactBook 将 Book 转换为 CompactBook
func toCompactBook(book Book, score float64) CompactBook {
    return CompactBook{
        ID:      book.ID,
        Title:   book.Title,
        Authors: book.Authors,
        Score:   score,
        Rating:  book.Rating,
    }
}

// toDetailedBook 将 Book 转换为 DetailedBook
func toDetailedBook(book Book, toc interface{}) DetailedBook {
    return DetailedBook{
        ID:         book.ID,
        Title:      book.Title,
        Authors:    book.Authors,
        Publisher:  book.Publisher,
        ISBN:       book.Isbn,
        Comments:   truncateComments(book.Comments, 500),
        TocSummary: generateTocSummary(toc),
        Rating:     book.Rating,
    }
}

// truncateComments 截断评论到指定长度
func truncateComments(comments string, maxLen int) string {
    if len(comments) <= maxLen {
        return comments
    }
    // 使用 rune 处理多字节字符
    runes := []rune(comments)
    if len(runes) <= maxLen {
        return comments
    }
    return string(runes[:maxLen]) + "..."
}

// generateTocSummary 生成 TOC 摘要
func generateTocSummary(toc interface{}) string {
    if toc == nil {
        return ""
    }
    
    // 根据 TOC 结构提取信息
    // 格式：共 X 章，包括：1. xxx, 2. xxx, 3. xxx...
    // 实现细节根据实际 TOC 数据结构调整
    
    return "TOC summary placeholder"
}
```

---

### Task 3: 优化 search_books 工具

**Description:** 更新 `handleSearchBooks` 函数，使用 `CompactBook` 返回精简数据。

**Dependencies:** Task 1, Task 2

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] 修改 `performSemanticSearch` 返回 `[]CompactBook`
- [ ] 更新 `handleSearchBooks` 使用新的返回格式
- [ ] 保留 score 字段用于相关性排序
- [ ] 验证返回字段数量为 5 个（id, title, authors, score, rating）

**Implementation Notes:**
```go
func (m *MCPServer) performSemanticSearch(query string, limit int) ([]CompactBook, error) {
    searcher, ok := m.api.semanticSearcher.(*qdrant.Searcher)
    if !ok || searcher == nil {
        return nil, fmt.Errorf("search service not available")
    }

    results, err := searcher.Search(query, limit)
    if err != nil {
        return nil, err
    }

    books := make([]CompactBook, 0, len(results))
    for _, result := range results {
        book := convertSemanticToBook(result.Book)
        books = append(books, toCompactBook(book, result.Score))
    }

    return books, nil
}
```

---

### Task 4: 优化 get_book 工具

**Description:** 更新 `handleGetBook` 函数，使用 `DetailedBook` 返回精简数据。

**Dependencies:** Task 1, Task 2

**Estimated Effort:** 1.5 hours

**Acceptance Criteria:**
- [ ] 修改返回格式使用 `DetailedBook`
- [ ] 实现 TOC 摘要生成
- [ ] 限制 comments 字段为 500 字符
- [ ] 移除不必要的字段（path, has_cover, timestamp 等）
- [ ] 验证返回字段数量为 8 个

**Implementation Notes:**
```go
func (m *MCPServer) handleGetBook(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // ... 参数解析 ...

    books, _, err := searcher.SearchByKeyword(id, "id", 1, 0)
    if err != nil || len(books) == 0 {
        return mcp.NewToolResultError("book not found"), nil
    }

    book := convertSemanticToBook(books[0])

    // 获取 TOC
    toc, tocErr := m.api.GetBookTocData(id)
    if tocErr != nil {
        log.Debugf("Failed to get TOC for book %s: %v", id, tocErr)
    }

    // 使用 DetailedBook 格式
    detailedBook := toDetailedBook(book, toc)

    return mcp.NewToolResultText(formatToolResult(detailedBook)), nil
}
```

---

### Task 5: 优化 random_books 工具

**Description:** 更新 `handleRandomBooks` 函数，使用 `CompactBook` 返回精简数据。

**Dependencies:** Task 1, Task 2

**Estimated Effort:** 0.5 hours

**Acceptance Criteria:**
- [ ] 修改返回格式使用 `CompactBook`
- [ ] 移除 score 字段（随机推荐不需要）
- [ ] 验证返回字段数量为 4-5 个

**Implementation Notes:**
```go
func (m *MCPServer) handleRandomBooks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // ... 参数解析 ...

    semanticBooks, err := searcher.GetRandom(limit)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("random search failed: %v", err)), nil
    }

    books := make([]CompactBook, 0, len(semanticBooks))
    for _, sb := range semanticBooks {
        book := convertSemanticToBook(sb)
        books = append(books, toCompactBook(book, 0)) // score = 0 for random
    }

    result := map[string]interface{}{
        "books": books,
        "count": len(books),
    }

    return mcp.NewToolResultText(formatToolResult(result)), nil
}
```

---

### Task 6: 优化 recent_books 工具

**Description:** 更新 `handleRecentBooks` 函数，使用 `CompactBook` 返回精简数据。

**Dependencies:** Task 1, Task 2

**Estimated Effort:** 0.5 hours

**Acceptance Criteria:**
- [ ] 修改返回格式使用 `CompactBook`
- [ ] 移除 score 字段（最近更新不需要）
- [ ] 验证返回字段数量为 4-5 个

**Implementation Notes:**
```go
func (m *MCPServer) handleRecentBooks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // ... 参数解析 ...

    semanticBooks, total, err := searcher.GetRecent(limit, offset)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("get recent books failed: %v", err)), nil
    }

    books := make([]CompactBook, 0, len(semanticBooks))
    for _, sb := range semanticBooks {
        book := convertSemanticToBook(sb)
        books = append(books, toCompactBook(book, 0)) // score = 0 for recent
    }

    result := map[string]interface{}{
        "books":  books,
        "total":  total,
        "limit":  limit,
        "offset": offset,
    }

    return mcp.NewToolResultText(formatToolResult(result)), nil
}
```

---

### Task 7: 添加单元测试

**Description:** 为新增的结构体和函数添加单元测试。

**Dependencies:** Task 1, Task 2

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] 测试 `toCompactBook` 转换正确性
- [ ] 测试 `toDetailedBook` 转换正确性
- [ ] 测试 `truncateComments` 边界情况（空字符串、短字符串、长字符串、多字节字符）
- [ ] 测试 `generateTocSummary` 格式正确性
- [ ] 测试覆盖率 > 80%

**Implementation Notes:**
```go
func TestToCompactBook(t *testing.T) {
    book := Book{
        ID:      "123",
        Title:   "Test Book",
        Authors: []string{"Author 1", "Author 2"},
        Rating:  4.5,
    }
    
    compact := toCompactBook(book, 0.95)
    
    assert.Equal(t, "123", compact.ID)
    assert.Equal(t, "Test Book", compact.Title)
    assert.Equal(t, 2, len(compact.Authors))
    assert.Equal(t, 0.95, compact.Score)
    assert.Equal(t, 4.5, compact.Rating)
}

func TestTruncateComments(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        maxLen   int
        expected string
    }{
        {"empty", "", 100, ""},
        {"short", "short text", 100, "short text"},
        {"exact", "exact", 5, "exact"},
        {"long", "this is a very long comment", 10, "this is a ..."},
        {"multibyte", "这是一个很长的中文评论", 5, "这是一个很..."},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := truncateComments(tt.input, tt.maxLen)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

---

### Task 8: 添加 Token 计数和监控

**Description:** 添加 token 计数功能，用于测量优化效果。

**Dependencies:** Task 3, Task 4, Task 5, Task 6

**Estimated Effort:** 1.5 hours

**Acceptance Criteria:**
- [ ] 实现简单的 token 估算函数
- [ ] 在每个工具调用时记录 token 数量
- [ ] 添加日志输出 token 统计信息
- [ ] 创建性能对比报告

**Implementation Notes:**
```go
// estimateTokens 估算 JSON 数据的 token 数量
// 简单估算：1 token ≈ 4 字符
func estimateTokens(data interface{}) int {
    jsonBytes, err := json.Marshal(data)
    if err != nil {
        return 0
    }
    return len(jsonBytes) / 4
}

// 在每个工具中添加日志
log.Debugf("MCP Tool %s: returned %d books, estimated %d tokens (before: ~%d)", 
    "search_books", len(books), estimateTokens(books), len(books)*75)
```

---

### Task 9: 集成测试和验证

**Description:** 进行端到端测试，验证优化效果和功能正确性。

**Dependencies:** Task 3, Task 4, Task 5, Task 6, Task 7, Task 8

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] 测试所有 MCP 工具返回正确的数据格式
- [ ] 验证 token 减少达到 40% 以上
- [ ] 验证 LLM 仍能准确回答问题
- [ ] 测试边界情况（空结果、大量结果等）
- [ ] 性能测试（响应时间、吞吐量）

**Test Cases:**
1. 搜索 20 本书，验证返回格式和 token 数量
2. 获取单本书详情，验证 TOC 摘要和 comments 截断
3. 随机推荐 10 本书，验证返回格式
4. 获取最近 20 本书，验证分页和格式
5. 对比优化前后的 token 消耗

---

### Task 10: 文档更新

**Description:** 更新相关文档，说明优化内容和使用方法。

**Dependencies:** Task 9

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] 更新 MCP 工具文档，说明返回字段
- [ ] 添加优化效果说明（token 减少比例）
- [ ] 更新 API 文档
- [ ] 添加迁移指南（如果需要）

---

## Task Summary

| Task | Effort | Priority | Status |
|------|--------|----------|--------|
| Task 1: 创建响应结构体 | 1h | High | Pending |
| Task 2: 实现辅助函数 | 2h | High | Pending |
| Task 3: 优化 search_books | 1h | High | Pending |
| Task 4: 优化 get_book | 1.5h | High | Pending |
| Task 5: 优化 random_books | 0.5h | Medium | Pending |
| Task 6: 优化 recent_books | 0.5h | Medium | Pending |
| Task 7: 添加单元测试 | 2h | High | Pending |
| Task 8: Token 计数监控 | 1.5h | Medium | Pending |
| Task 9: 集成测试验证 | 2h | High | Pending |
| Task 10: 文档更新 | 1h | Low | Pending |

**Total Estimated Effort:** 13 hours

## Implementation Order

1. **Phase 1: 基础设施** (3 hours)
   - Task 1: 创建响应结构体
   - Task 2: 实现辅助函数

2. **Phase 2: 核心优化** (3.5 hours)
   - Task 3: 优化 search_books
   - Task 4: 优化 get_book
   - Task 5: 优化 random_books
   - Task 6: 优化 recent_books

3. **Phase 3: 测试和验证** (4 hours)
   - Task 7: 添加单元测试
   - Task 8: Token 计数监控
   - Task 9: 集成测试验证

4. **Phase 4: 收尾** (1 hour)
   - Task 10: 文档更新

## Success Criteria

- ✅ 所有 MCP 工具返回精简的数据格式
- ✅ Token 消耗减少至少 40%
- ✅ 所有测试通过，覆盖率 > 80%
- ✅ LLM 功能正常，回答准确率不下降
- ✅ 性能指标正常或改善
- ✅ 文档完整更新
