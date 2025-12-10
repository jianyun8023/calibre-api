---
status: complete
created: '2025-12-08'
tags:
  - architecture
  - overview
priority: high
created_at: '2025-12-08T13:12:00.000Z'
updated_at: '2025-12-08T13:12:00.000Z'
completed_at: '2025-12-08T13:12:00.000Z'
completed: '2025-12-08'
transitions:
  - status: complete
    at: '2025-12-08T13:12:00.000Z'
---

# Calibre API - 项目总览

> **Status**: ✅ Complete · **Priority**: High · **Created**: 2025-12-08

## Overview

**愿景**: 将 Calibre 从传统电子书管理工具升级为 AI 原生的智能书库系统。

**核心价值**:
- 📚 **现代化 API**: RESTful + MCP 双协议支持
- 🔍 **智能搜索**: 关键词 + 语义混合搜索
- 🤖 **AI 问答**: 基于书库内容的智能推荐
- 🚀 **开箱即用**: Docker 一键部署，跨平台支持

**为什么现在**: AI 时代需要重新定义人与知识的交互方式，书库应该"懂你"。

## Design

### 系统架构分层

**5 层架构**:
1. **Client**: Web UI (Vue.js) / AI Assistant (MCP) / HTTP API
2. **Protocol**: REST (Gin) / MCP SSE
3. **Business**: 6 个核心模块（见模块依赖关系）
4. **Service**: Calibre / Qdrant / LLM / Embedding
5. **Data**: Calibre DB / Chat DB / Qdrant / Cache

### 模块依赖关系

```
001-book-management (基础层)
    ↓
    ├─→ 002-search-functionality
    │       ↓
    │       └─→ 003-mcp-integration
    │       ↓
    │       └─→ 004-chat-agent
    │
    └─→ 005-qdrant-vector-search
            ↓
            ├─→ 002-search-functionality
            ├─→ 004-chat-agent
            └─→ 006-task-management
```

**设计原则**:
1. **分层隔离**: 每层只依赖下层，避免循环依赖
2. **模块自治**: 每个 spec 可独立理解和测试
3. **渐进增强**: 核心功能先行，高级功能按需启用

### 技术栈选型

**关键决策**:
- **Go + Gin**: 高性能、部署简单、并发支持强
- **Vue 3 + TS**: 组合式 API、类型安全
- **Qdrant**: 轻量级向量库，RESTful API
- **双 LLM 模式**: OpenAI (质量) + Ollama (成本)

### 数据流示例

**AI 问答流程**:
用户提问 → 语义搜索 → 获取目录 → LLM 生成回答 → 保存历史

## Plan

核心模块开发顺序（已完成）:

- [x] **Phase 1 - 基础层** (001)
  - 书籍管理 CRUD
  - 文件访问和缓存
  - 目录解析

- [x] **Phase 2 - 搜索层** (002, 005)
  - 关键词搜索
  - 向量化和语义搜索
  - 混合搜索策略

- [x] **Phase 3 - AI 层** (003, 004)
  - MCP 协议集成
  - ChatAgent 智能问答
  - Tool Calling 框架

- [x] **Phase 4 - 运维层** (006)
  - 异步任务管理
  - 数据同步和迁移
  - 进度追踪

## Test

**集成测试覆盖**:

- [x] **端到端流程**:
  - 用户浏览书籍 → 搜索 → 查看详情 → 下载
  - AI 提问 → 搜索 → 推荐 → 获取详情

- [x] **MCP 集成**:
  - Cursor 连接 MCP Server
  - AI 调用 Tools 搜索书籍
  - AI 读取 Resources 获取详情

- [x] **性能基准**:
  - 1000 本书籍加载 < 5s
  - 混合搜索响应 < 300ms
  - AI 问答首次 < 3s

- [x] **跨平台**:
  - macOS/Linux/Windows 构建通过
  - Docker 镜像运行正常

## Notes

### 技术指标

- 代码: ~13K 行 (Go 8K + 前端 5K)
- 性能: 100+ 并发, P95 < 500ms
- Spec Token: ~9.9K (7 specs, 最优范围)

### 部署模式

**个人**: 本地 + Ollama (0 元/月，需 GPU)  
**团队**: 云服务 + OpenAI (~50 元/月)

### 未来路线图

**v1.3** - 章节级 RAG（精确段落定位）  
**v1.4** - 多模态（封面理解、语音问答）  
**v1.5** - 社交化（书评、进度同步）  
**v2.0** - 智能体（主动推荐、知识图谱）

### 贡献指南

**添加新功能**: `lean-spec create` → 填写 spec → 建立依赖 → 实现 → 更新 CHANGELOG

**代码规范**: 遵循 `AGENTS.md`，每个 PR 更新 CHANGELOG.md，测试覆盖率 > 60%

### 相关文档

- `README.md` / `AGENTS.md`: 快速开始和开发指南
- `docs/*.md`: 架构、API、功能文档
- `CHANGELOG.md`: 版本变更记录
- `specs/`: 功能 specs（本目录）

### Spec 清单

| Spec | 模块 | 状态 | Token | 依赖 |
|------|------|------|-------|------|
| 000-project-overview | 项目总览 | ✅ | ~1400 | - |
| 001-book-management | 书籍管理 | ✅ | 1086 | - |
| 002-search-functionality | 搜索功能 | ✅ | 1117 | 001 |
| 003-mcp-integration | MCP 集成 | ✅ | 1512 | 001, 002 |
| 004-chat-agent | 聊天代理 | ✅ | 1835 | 002, 005 |
| 005-qdrant-vector-search | 向量搜索 | ✅ | 1311 | 001 |
| 006-task-management | 任务管理 | ✅ | 1618 | 005 |

**Total**: 7 specs, ~9,900 tokens, 100% complete
