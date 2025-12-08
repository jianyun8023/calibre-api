---
status: complete
created: '2025-12-08'
tags: []
priority: medium
created_at: '2025-12-08T13:03:58.996Z'
updated_at: '2025-12-08T13:09:00.000Z'
completed_at: '2025-12-08T13:09:00.000Z'
completed: '2025-12-08'
transitions:
  - status: complete
    at: '2025-12-08T13:09:00.000Z'
depends_on:
  - 002-search-functionality
  - 005-qdrant-vector-search
---

# AI 智能问答代理

> **Status**: ✅ Complete · **Priority**: Medium · **Created**: 2025-12-08

## Overview

**问题**: 用户需要在海量书库中找到相关书籍，传统搜索无法理解复杂意图（如 "有没有适合初学者的 Python 书？"）。

**解决方案**: 构建 ChatAgent，结合语义搜索和 LLM，提供上下文感知的智能推荐和问答。

**为什么现在**: 搜索和向量化基础已具备，用户需要更自然的交互方式。

## Design

### Agent 工作流

**问答流程**:
```
用户提问 → 语义搜索相关书籍 → 获取书籍 TOC
→ 构建上下文 Prompt → LLM 生成回答 → 保存历史
```

**上下文构建策略**:
```go
context := fmt.Sprintf(`
相关书籍:
%s

书籍目录:
%s

用户问题: %s
`, booksSummary, tocSummary, userQuery)
```

**为何需要 TOC**?
- 书名不足以判断内容深度（如 "Python 编程" 可能是入门或高级）
- 目录反映知识结构，帮助 LLM 准确推荐
- 代价：增加 Token 消耗（~500 tokens/book）

### LLM 集成

**双模式支持**:
1. **OpenAI API**: GPT-4, GPT-3.5
2. **Ollama**: 本地模型（llama3, qwen）

**权衡**: API vs 本地
- API: 质量高、零部署，但有成本
- 本地: 无成本、数据私密，但需 GPU/质量稍低

**配置示例**:
```yaml
chat:
  llm:
    provider: openai  # or ollama
    openai:
      api_key: ${OPENAI_API_KEY}
      model: gpt-4
    ollama:
      url: http://localhost:11434
      model: llama3
```

### Tool Calling 支持

**注册的 Tools**:
- `search_books`: 搜索更多相关书籍
- `get_book_details`: 获取完整书籍信息
- `get_book_toc`: 获取详细目录

**工作原理**:
```
LLM 判断需要更多信息 → 返回 Tool Call
→ Agent 执行 Tool → 将结果返回 LLM
→ LLM 基于新信息生成最终回答
```

**权衡**: 自动 Tool Call vs 用户确认
- 选择自动执行：提升交互流畅度
- 代价：可能产生非预期的搜索（通过 Prompt 约束）

### 会话管理

**SQLite 存储**:
```sql
CREATE TABLE conversations (
  id TEXT PRIMARY KEY,
  title TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE TABLE messages (
  id INTEGER PRIMARY KEY,
  conversation_id TEXT,
  role TEXT,  -- user | assistant | tool
  content TEXT,
  created_at TIMESTAMP
);
```

**为何不用 Redis**?
- 会话不需要高并发写入
- SQLite 迁移简单（文件拷贝）
- 保持 Calibre 一致性（同样用 SQLite）

### Thinking 功能

**可视化推理过程**:
```json
{
  "thinking": "用户想找 Python 入门书，先搜索相关书籍...",
  "tool_calls": [{"name": "search_books", "args": {...}}],
  "answer": "我为你找到了 3 本适合初学者的书..."
}
```

**作用**:
- 提升 AI 可信度（展示推理过程）
- 调试 Agent 行为（观察决策链路）
- 用户可中断不合理推理

## Plan

- [x] 设计 Agent 架构（LLM + Searcher + TocFetcher）
- [x] 实现 Chat DB 和 SQL 迁移
- [x] 实现 LLM 客户端抽象层
- [x] 集成 OpenAI API
- [x] 集成 Ollama 本地模型
- [x] 实现 Tool Calling 框架
- [x] 实现会话管理 API（CRUD）
- [x] 实现消息发送和历史查询
- [x] 添加 Thinking 可视化
- [x] 前端聊天界面集成

## Test

- [x] 创建会话并发送消息
- [x] LLM 正确理解用户意图
- [x] 语义搜索返回相关书籍（前 5 本）
- [x] TOC 正确附加到上下文
- [x] Tool Call 自动执行并返回结果
- [x] 多轮对话保持上下文连贯性
- [x] Thinking 过程正确显示
- [x] 会话历史持久化

## Notes

**上下文窗口管理**:
- 保留最近 10 轮对话
- 超过限制时自动截断（保留系统 Prompt）
- 书籍信息和 TOC 约 2000 tokens/轮

**Token 成本优化**:
- 语义搜索限制返回 5 本书（而非 20）
- TOC 仅提取前两级结构（~300 tokens/book）
- 使用 GPT-3.5 而非 GPT-4（成本降低 90%）

**已知限制**:
- 不支持多模态（图片、音频）
- 不支持实时流式输出（待 SSE 改造）
- Tool Call 最多执行 3 次（防止死循环）

**性能指标**:
- 首次回答延迟: ~3s（含搜索 + LLM）
- 后续追问: ~1s（无需搜索）
- Tool Call 额外开销: +1s/call

**未来优化**:
- 添加 RAG（检索增强生成）支持章节级搜索
- 支持流式输出（SSE）
- 添加对话摘要（长会话压缩）
- 支持多模态（书籍封面理解）

**相关文档**:
- `docs/LLM_CHAT_COMPLETE.md`: 聊天功能完整文档
- `docs/LLM_CHAT_TEST_REPORT.md`: 测试报告

**依赖项**:
- `002-search-functionality`: 语义搜索能力
- `005-qdrant-vector-search`: 向量化和相似度计算
- `001-book-management`: 书籍元数据和 TOC
