---
status: complete
created: '2025-12-08'
tags: []
priority: medium
created_at: '2025-12-08T13:03:54.361Z'
updated_at: '2025-12-08T13:06:00.334Z'
completed_at: '2025-12-08T13:06:00.087Z'
completed: '2025-12-08'
transitions:
  - status: complete
    at: '2025-12-08T13:06:00.087Z'
depends_on:
  - 001-book-management
---

# search-functionality

> **Status**: ✅ Complete · **Priority**: Medium · **Created**: 2025-12-08

## Overview

**问题**: 关键词搜索无法理解语义，语义搜索可能遗漏精确匹配。单一策略无法满足所有场景。

**解决方案**: 实现混合搜索策略，并发执行关键词和语义搜索，合并去重后返回最优结果。

**为什么现在**: 搜索是书籍发现的核心入口，直接影响用户体验和 AI 问答质量。

## Design

### 搜索策略

**三种模式**:
1. `keyword`: 精确匹配标题、作者、标签（Calibre Content Server API）
2. `semantic`: 语义理解查询意图（Qdrant 向量搜索）
3. `hybrid`: 并发执行两者，语义结果优先，关键词补充

**混合搜索工作流**:
```go
// 1. 并发执行
ch1 := searchKeyword(query)
ch2 := searchSemantic(query)

// 2. 按 book_id 去重
results := map[int64]*Book{}
for book := range ch2 { results[book.ID] = book }  // 语义优先
for book := range ch1 { 
    if _, exists := results[book.ID]; !exists { 
        results[book.ID] = book  // 补充缺失
    }
}

// 3. 按相关性排序
sort.Slice(list, func(i, j int) bool {
    return list[i].Score > list[j].Score
})
```

**权衡**: 串行 vs 并发
- 选择并发：减少响应时间（从 ~500ms 降至 ~250ms）
- 代价：增加服务器负载，需要 goroutine 管理

**去重策略**: 为何语义优先？
- 语义搜索结果更符合用户意图
- 关键词补充避免遗漏精确匹配
- Score 字段记录相关性用于排序

### API 设计

```
POST /api/search
{
  "query": "人工智能",
  "strategy": "hybrid",  // keyword | semantic | hybrid
  "limit": 20
}

Response:
{
  "books": [...],
  "total": 42,
  "strategy_used": "hybrid"
}
```

## Plan

- [x] 实现关键词搜索（对接 Calibre Content Server）
- [x] 实现语义搜索（对接 Qdrant）
- [x] 实现混合策略（并发 + 去重）
- [x] 添加 strategy 参数支持三种模式
- [x] 添加搜索性能监控日志
- [x] 前端集成搜索 UI

## Test

- [x] 关键词搜索精确匹配书名
- [x] 语义搜索理解同义词（如 "AI" vs "人工智能"）
- [x] 混合搜索结果无重复
- [x] 并发执行响应时间 < 300ms
- [x] 空查询返回合理错误

## Notes

**性能指标**:
- 关键词搜索: ~100ms
- 语义搜索: ~200ms
- 混合搜索: ~250ms（并发优化）

**已知问题**:
- 语义搜索依赖 Embedding 服务（Ollama/SiliconFlow）
- 向量数据库需预先同步（见 spec 005）

**未来优化**:
- 添加搜索结果缓存（Redis）
- 支持高级过滤（出版社、年份、标签）
- 实现搜索历史和热门查询

**依赖项**:
- `001-book-management`: 书籍数据结构
- `005-qdrant-vector-search`: 语义搜索能力
