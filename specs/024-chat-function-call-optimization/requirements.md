# Requirements Document

## Introduction

本文档定义了优化 Chat 功能中 Function Call 返回数据的需求。当前实现中，function call 返回完整的书籍对象，包含大量不必要的字段，导致 token 消耗过高，降低了上下文的有效性。本优化旨在精简返回数据，只保留 LLM 回答用户问题所需的关键信息。

## Glossary

- **Function Call**: LLM 调用工具函数获取外部数据的机制
- **Token**: LLM 处理文本的基本单位，影响成本和上下文窗口
- **Context Window**: LLM 可以处理的最大 token 数量
- **Tool**: 提供给 LLM 的函数工具，如搜索、推荐等
- **Semantic Search**: 基于语义相似度的搜索功能
- **MCP Tools**: Model Context Protocol 工具，用于 AI 交互

## Requirements

### Requirement 1

**User Story:** 作为系统管理员，我希望减少 function call 返回的数据量，以便降低 API 成本并提高响应速度。

#### Acceptance Criteria

1. WHEN semantic_search tool 返回搜索结果 THEN 系统应只包含 id、title、authors 和 score 字段
2. WHEN recommend_books tool 返回推荐结果 THEN 系统应只包含 id、title、authors 和 reason 字段
3. WHEN get_book tool 返回书籍详情 THEN 系统应只包含 id、title、authors、publisher、isbn、comments 和 toc_summary 字段
4. WHEN search_books MCP tool 返回结果 THEN 系统应移除 path、has_cover、timestamp 等冗余字段
5. WHEN get_book MCP tool 返回结果 THEN 系统应将完整 TOC 替换为摘要信息

### Requirement 2

**User Story:** 作为 AI 助手，我希望获得精简但足够的信息，以便能够准确回答用户问题而不浪费上下文空间。

#### Acceptance Criteria

1. WHEN 返回书籍列表 THEN 系统应限制每本书的字段数量不超过 6 个
2. WHEN 返回 TOC 信息 THEN 系统应提供章节数量和前 3 章标题的摘要
3. WHEN 返回搜索结果 THEN 系统应按相关性排序并包含 score 字段
4. WHEN 返回书籍详情 THEN 系统应将 comments 字段限制在 500 字符以内
5. WHEN 返回作者信息 THEN 系统应使用逗号分隔的字符串而不是数组

### Requirement 3

**User Story:** 作为开发者，我希望有统一的数据格式规范，以便维护和扩展工具函数。

#### Acceptance Criteria

1. WHEN 定义新的 tool THEN 系统应遵循统一的字段命名规范（snake_case）
2. WHEN 返回错误信息 THEN 系统应使用简洁的错误描述而不是完整堆栈
3. WHEN 格式化数据 THEN 系统应使用辅助函数确保一致性
4. WHEN 添加新字段 THEN 系统应评估其对 token 消耗的影响
5. WHEN 返回 JSON 数据 THEN 系统应移除所有空值和零值字段

### Requirement 4

**User Story:** 作为性能监控人员，我希望能够测量优化效果，以便验证改进是否达到预期目标。

#### Acceptance Criteria

1. WHEN 执行 function call THEN 系统应记录返回数据的 token 数量
2. WHEN 比较优化前后 THEN 系统应显示 token 减少的百分比
3. WHEN 分析工具使用 THEN 系统应统计各工具的平均 token 消耗
4. WHEN 评估效果 THEN 系统应确保 token 减少至少 40%
5. WHEN 验证功能 THEN 系统应确保 LLM 仍能准确回答用户问题
