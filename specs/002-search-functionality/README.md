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
3. `hybrid`: 先召回后精排（推荐）

### 混合搜索优化策略

**核心思想**: 先用向量搜索召回语义相关候选集，再用关键词精确匹配重排序

**工作流程**:
    用户查询 "深度学习入门"
        ↓
    1. 语义搜索（召回阶段）
       - 使用向量相似度搜索
       - 获取 top 100 个语义相关书籍
       - 确保召回率高，不遗漏相关内容
        ↓
    2. 关键词重排序（精排阶段）
       - 在上述 100 个候选中进行关键词匹配
       - 计算关键词匹配得分
       - 结合语义得分和关键词得分
        ↓
    3. 返回最终结果
       - 返回 top 20 个综合得分最高的书籍

**评分机制**:

语义得分（40% 权重）:
- 由向量相似度计算，范围: 0-1
- 优势: 理解语义，召回相关内容

关键词得分（60% 权重）:
- 多维度匹配计算，范围: 0-1
- 优势: 精确匹配，提高准确率

关键词匹配规则:
- 完整查询匹配（权重最高）:
  - 标题完全匹配: 10 分
  - 作者完全匹配: 8 分
  - 出版社完全匹配: 6 分
- 单个关键词匹配（每个关键词最多 5 分）:
  - 标题包含: 3.0 分
  - 作者包含: 1.5 分
  - 简介包含: 0.3 分

最终得分: finalScore = semanticScore × 0.4 + keywordScore × 0.6

**实现代码**:
    // 1. 语义搜索召回候选集
    semanticLimit := max(limit * 3, 100)
    candidates := searcher.Search(query, semanticLimit)
    
    // 2. 关键词重排序
    for _, book := range candidates {
        keywordScore := calculateKeywordScore(book, query)
        book.FinalScore = book.SemanticScore*0.4 + keywordScore*0.6
    }
    
    // 3. 按综合得分排序并返回 top N
    sort.Slice(candidates, cmp FinalScore desc)
    return candidates[:limit]

**优势对比**:
- ✅ 语义搜索保证召回率，不遗漏相关内容
- ✅ 关键词匹配提高精确度，优先推荐精确匹配
- ✅ 两阶段策略清晰，各司其职
- ✅ 计算效率高（只在候选集中匹配）
- ✅ 更符合人类搜索心理模型

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
- 混合搜索: ~250ms（先召回后精排优化）

**参数调优**:

候选集大小:
    semanticLimit := max(limit * 3, 100)
- 小结果集 (limit ≤ 10): 使用 100 个候选
- 中等结果集 (10 < limit ≤ 50): 使用 3 倍候选
- 大结果集 (limit > 50): 使用 2 倍候选

权重调整（当前配置适用于大多数场景）:
- 语义得分: 40%
- 关键词得分: 60%

可调整场景:
- 更重视精确匹配: semanticScore × 0.3 + keywordScore × 0.7
- 更重视语义理解: semanticScore × 0.5 + keywordScore × 0.5
- 纯关键词模式: semanticScore × 0.2 + keywordScore × 0.8

**已知限制**:
- 语义搜索依赖 Embedding 服务（Ollama/SiliconFlow）
- 向量数据库需预先同步（见 spec 005）
- 混合搜索需要两个服务都可用

**未来优化**:
- 动态权重调整（根据用户点击行为）
- 查询理解（识别查询意图，动态选择策略）
- 搜索结果缓存（Redis）
- 多字段权重配置
- A/B 测试对比不同策略

**依赖项**:
- `001-book-management`: 书籍数据结构
- `005-qdrant-vector-search`: 语义搜索能力
