# Design Document

## Overview

本文档定义了优化 Chat Function Call 返回数据的设计方案。通过分析当前实现，我们发现 `internal/chat/tools.go` 已经实现了良好的优化，而 `internal/calibre/mcp_tools.go` 中的 MCP 工具返回了大量冗余字段。本设计将重点优化 MCP 工具的返回数据，预期减少 40-60% 的 token 消耗。

## Current State Analysis

### Chat Tools (internal/chat/tools.go) ✅

**已优化，无需修改：**

1. **semantic_search** - 只返回 4 个字段：
   - id, title, authors, score
   - 估计每本书约 50-80 tokens

2. **recommend_books** - 只返回 4 个字段：
   - id, title, authors, reason
   - 估计每本书约 50-80 tokens

3. **get_book_toc** - 返回 TOC 摘要
   - 已经通过 fetcher 函数优化
   - 估计约 100-200 tokens

### MCP Tools (internal/calibre/mcp_tools.go) ⚠️

**需要优化的工具：**

1. **search_books** - 返回完整 Book 对象
   - 当前：约 15-20 个字段，每本书 200-300 tokens
   - 目标：精简到 6 个字段，每本书 80-120 tokens
   - 减少：约 60% tokens

2. **get_book** - 返回完整 Book 对象 + TOC
   - 当前：约 20 个字段 + 完整 TOC，500-1000 tokens
   - 目标：精简到 8 个字段 + TOC 摘要，200-400 tokens
   - 减少：约 50-60% tokens

3. **random_books** - 返回完整 Book 对象
   - 当前：约 15-20 个字段，每本书 200-300 tokens
   - 目标：精简到 6 个字段，每本书 80-120 tokens
   - 减少：约 60% tokens

4. **recent_books** - 返回完整 Book 对象
   - 当前：约 15-20 个字段，每本书 200-300 tokens
   - 目标：精简到 6 个字段，每本书 80-120 tokens
   - 减少：约 60% tokens

## Design Decisions

### Decision 1: 创建精简的响应结构体

**Context:** 当前 MCP 工具直接返回完整的 Book 对象，包含大量 LLM 不需要的字段。

**Decision:** 创建专门的响应结构体用于 MCP 工具返回。

**Rationale:**
- 分离数据模型和 API 响应，遵循单一职责原则
- 便于维护和扩展，不影响现有 Book 结构
- 可以为不同场景定义不同的响应格式

**Alternatives Considered:**
- 修改 Book 结构体：会影响其他模块，风险高
- 使用 JSON tag 控制序列化：不够灵活，难以处理复杂逻辑

**Implementation:**
```go
// CompactBook 精简的书籍信息（用于列表场景）
type CompactBook struct {
    ID       string   `json:"id"`
    Title    string   `json:"title"`
    Authors  []string `json:"authors"`
    Score    float64  `json:"score,omitempty"`    // 仅搜索结果包含
    Rating   float64  `json:"rating,omitempty"`   // 仅推荐场景包含
}

// DetailedBook 详细的书籍信息（用于单本书场景）
type DetailedBook struct {
    ID          string   `json:"id"`
    Title       string   `json:"title"`
    Authors     []string `json:"authors"`
    Publisher   string   `json:"publisher,omitempty"`
    ISBN        string   `json:"isbn,omitempty"`
    Comments    string   `json:"comments,omitempty"`    // 限制 500 字符
    TocSummary  string   `json:"toc_summary,omitempty"` // TOC 摘要
    Rating      float64  `json:"rating,omitempty"`
}
```

### Decision 2: TOC 信息优化策略

**Context:** 完整的 TOC 可能包含数百个章节，消耗大量 tokens。

**Decision:** 提供 TOC 摘要而不是完整结构。

**Rationale:**
- LLM 通常只需要了解书籍的大致结构
- 摘要信息足以回答"这本书讲什么"类问题
- 如需完整 TOC，用户可以通过 Web UI 查看

**Implementation:**
```go
func generateTocSummary(toc interface{}) string {
    // 提取章节数量和前 3 章标题
    // 格式：共 X 章，包括：1. xxx, 2. xxx, 3. xxx...
}
```

### Decision 3: Comments 字段截断策略

**Context:** 书籍评论可能很长，但 LLM 只需要摘要信息。

**Decision:** 限制 comments 字段为 500 字符，超出部分截断并添加省略号。

**Rationale:**
- 500 字符足以提供书籍概要信息
- 减少 token 消耗，提高响应速度
- 保持信息完整性和可读性

**Implementation:**
```go
func truncateComments(comments string, maxLen int) string {
    if len(comments) <= maxLen {
        return comments
    }
    return comments[:maxLen] + "..."
}
```

### Decision 4: 统一的转换函数

**Context:** 需要在多个地方将 Book 转换为精简格式。

**Decision:** 创建统一的转换函数，确保一致性。

**Rationale:**
- DRY 原则，避免重复代码
- 便于维护和测试
- 确保所有工具使用相同的优化策略

**Implementation:**
```go
func toCompactBook(book Book, score float64) CompactBook {
    return CompactBook{
        ID:      book.ID,
        Title:   book.Title,
        Authors: book.Authors,
        Score:   score,
        Rating:  book.Rating,
    }
}

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
```

## Data Flow

### Before Optimization

```
User Query → LLM → Function Call → MCP Tool
                                      ↓
                                  Full Book Object (20 fields)
                                      ↓
                                  JSON Serialization
                                      ↓
                                  ~300 tokens per book
                                      ↓
                                  LLM Context
```

### After Optimization

```
User Query → LLM → Function Call → MCP Tool
                                      ↓
                                  Compact/Detailed Book (6-8 fields)
                                      ↓
                                  JSON Serialization
                                      ↓
                                  ~100 tokens per book
                                      ↓
                                  LLM Context (60% reduction)
```

## Token Estimation

### Search Results (20 books)

| Scenario | Before | After | Reduction |
|----------|--------|-------|-----------|
| search_books | 6,000 tokens | 2,400 tokens | 60% |
| random_books | 3,000 tokens | 1,200 tokens | 60% |
| recent_books | 6,000 tokens | 2,400 tokens | 60% |

### Single Book Details

| Scenario | Before | After | Reduction |
|----------|--------|-------|-----------|
| get_book | 800 tokens | 350 tokens | 56% |
| get_book (with TOC) | 1,500 tokens | 500 tokens | 67% |

### Overall Impact

- **Average reduction:** 60%
- **Context window saved:** 可以多容纳 2-3 轮对话
- **Cost reduction:** API 调用成本降低 60%
- **Response speed:** 更快的序列化和传输

## Testing Strategy

### Unit Tests

1. **转换函数测试**
   - 测试 toCompactBook 正确性
   - 测试 toDetailedBook 正确性
   - 测试 truncateComments 边界情况
   - 测试 generateTocSummary 格式

2. **MCP 工具测试**
   - 验证返回字段数量
   - 验证字段值正确性
   - 验证 JSON 序列化格式

### Integration Tests

1. **Token 计数测试**
   - 测量优化前后的 token 数量
   - 验证减少比例达到目标（≥40%）

2. **功能验证测试**
   - 验证 LLM 仍能准确回答问题
   - 验证搜索结果质量不受影响
   - 验证推荐功能正常工作

### Performance Tests

1. **响应时间测试**
   - 测量序列化时间
   - 测量网络传输时间
   - 对比优化前后性能

## Monitoring and Metrics

### Key Metrics

1. **Token Consumption**
   - 每个工具的平均 token 数
   - 总体 token 减少百分比
   - 按时间趋势分析

2. **Response Quality**
   - LLM 回答准确率
   - 用户满意度
   - 错误率

3. **Performance**
   - 平均响应时间
   - P95/P99 延迟
   - 吞吐量

### Logging

```go
log.Debugf("MCP Tool %s: returned %d books, estimated %d tokens", 
    toolName, len(books), estimateTokens(books))
```

## Migration Plan

### Phase 1: 实现新的响应结构体
- 创建 CompactBook 和 DetailedBook 结构体
- 实现转换函数
- 添加单元测试

### Phase 2: 更新 MCP 工具
- 更新 search_books
- 更新 get_book
- 更新 random_books
- 更新 recent_books

### Phase 3: 验证和监控
- 部署到测试环境
- 测量 token 减少效果
- 验证功能正确性
- 收集性能数据

### Phase 4: 生产部署
- 部署到生产环境
- 持续监控指标
- 收集用户反馈
- 必要时调整优化策略

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| 信息不足导致 LLM 回答不准确 | High | 保留核心字段，通过测试验证 |
| 破坏现有功能 | High | 只修改 MCP 工具，不影响其他模块 |
| 性能回退 | Medium | 性能测试验证，监控关键指标 |
| 维护成本增加 | Low | 统一转换函数，减少重复代码 |

## Future Enhancements

1. **动态字段选择**
   - 根据查询类型返回不同字段
   - 支持客户端指定需要的字段

2. **智能摘要生成**
   - 使用 LLM 生成更好的 TOC 摘要
   - 智能截断 comments 保持语义完整

3. **缓存优化**
   - 缓存精简后的响应数据
   - 减少重复转换开销

4. **A/B 测试**
   - 对比不同优化策略的效果
   - 持续优化字段选择
