---
status: in-progress
phase: IMPLEMENTATION
phase_history:
  - phase: REQUIREMENTS
    status: APPROVED
    date: '2025-12-12'
    notes: 需求文档已完成，13个主要需求定义
  - phase: DESIGN
    status: APPROVED
    date: '2025-12-12'
    notes: 技术设计完成，包含10个正确性属性
  - phase: IMPLEMENTATION
    status: IN_PROGRESS
    date: '2025-12-19'
    notes: 核心架构已实现，完成度约80%
complexity: medium
created: '2025-12-12'
tags:
  - frontend
  - api
  - cleanup
  - migration
  - typescript
  - nextjs
priority: high
created_at: '2025-12-12T12:28:23.743Z'
depends_on:
  - 026-backend-code-optimization
updated_at: '2025-12-19T07:18:07.716Z'
current_action: 完成剩余组件迁移和测试覆盖
transitions:
  - status: planned
    at: '2025-12-12T12:28:23.743Z'
  - status: in-progress
    at: '2025-12-12T13:00:00.000Z'
---

# Frontend API Interface Cleanup and Migration

> **Status**: ⏳ In progress · **Priority**: High · **Created**: 2025-12-12 · **Tags**: frontend, api, cleanup, migration, typescript, nextjs

## 概述 (Overview)

在 Spec 026 后端代码优化完成后，后端已实现统一的 API 响应格式和错误处理机制。前端需要相应地更新 API 接口调用逻辑，清理冗余代码，统一错误处理，并适配新的响应格式，确保前后端接口的一致性和可维护性。

**解决的核心问题**:
- 前后端接口格式不一致
- 错误处理分散且不统一
- 存在冗余的 API 调用代码
- 缺少统一的类型定义

## 文档索引 (Documents)

- [requirements.md](requirements.md) - 需求定义和用户故事 (13 个主要需求)
- [design.md](design.md) - 技术设计和架构 (统一 API 客户端、错误处理、10 个正确性属性)
- [IMPLEMENTATION.md](IMPLEMENTATION.md) - 实施进度和详细状态

## 需求总结 (Requirements Summary)

### 核心用户故事 (Core User Stories)

1. **作为前端开发者**，我希望有统一的 API 响应类型定义，以便在整个应用中保持类型安全和一致性
2. **作为前端开发者**，我希望有统一的 API 客户端，以便简化 API 调用和错误处理逻辑
3. **作为前端开发者**，我希望统一错误处理逻辑，以便为用户提供一致的错误体验

### 验收标准 (Acceptance Criteria)

- [x] 创建符合后端 V2 格式的 TypeScript 接口定义
- [x] 实现统一的 API 客户端（UnifiedApiClient）
- [x] 实现统一的错误处理器（ErrorHandler）
- [x] 创建 API 服务层（BookService, ChatService, MetadataService, TaskService）
- [x] 提供向后兼容层支持渐进式迁移
- [ ] 迁移所有组件使用新的 API 客户端（当前 60%）
- [ ] 清理旧的 API 客户端代码
- [ ] 测试覆盖率达到 80%+（当前 50%）
- [ ] 启用缓存策略优化性能
- [ ] 编写迁移指南和最佳实践文档

## 设计总结 (Design Summary)

### 架构概述 (Architecture Overview)

采用**分层架构**设计:
```
组件层 (React Components)
    ↓
API 服务层 (BookService, ChatService, etc.)
    ↓
统一 API 客户端 (UnifiedApiClient)
    ↓
HTTP / REST API
```

### 关键技术决策 (Key Technical Decisions)

1. **统一 API 客户端** - 使用单一 `UnifiedApiClient` 处理所有 HTTP 请求，提供拦截器、重试、超时机制
2. **类型安全优先** - 完整的 TypeScript 类型定义，避免使用 `any` 类型
3. **渐进式迁移** - 提供向后兼容层，允许新旧代码共存
4. **服务层抽象** - 按业务领域划分服务（BookService, ChatService 等），封装 API 调用细节
5. **错误处理统一** - 统一的 ErrorHandler 提供用户友好的错误提示和恢复机制

## 阶段状态 (Phase Status)

| 阶段 (Phase) | 状态 (Status) | 完成日期 (Completed) | 备注 (Notes) |
|-------------|--------------|---------------------|-------------|
| REQUIREMENTS | ✅ 已完成 (Approved) | 2025-12-12 | 13个需求定义完成 |
| DESIGN | ✅ 已完成 (Approved) | 2025-12-12 | 架构设计和10个正确性属性定义 |
| IMPLEMENTATION | ⏳ 进行中 (In Progress) | - | 核心架构完成，组件迁移中（80%） |

## 实施状态 (Implementation Status)

**当前状态**: ⏳ **实施阶段 (Implementation Phase)** - 核心架构已完成，正在进行组件迁移

**完成度**: 80%

**已完成**:
- ✅ 统一 API 客户端 (UnifiedApiClient)
- ✅ API 响应类型定义 (ApiResponse, PaginatedResponse 等)
- ✅ 错误处理器 (ErrorHandler)
- ✅ API 服务层 (BookService, ChatService, MetadataService, TaskService)
- ✅ 向后兼容层 (api/books.ts)
- ✅ 基础单元测试和属性测试

**进行中**:
- ⏳ 组件层迁移（60% 完成）
- ⏳ 测试覆盖补充

**待完成**:
- 📋 Chat 页面迁移（使用 chatService 替换 fetch）
- 📋 旧 API 客户端清理（api-client.ts）
- 📋 分页参数统一（改用 page/page_size）
- 📋 性能优化（启用缓存）
- 📋 迁移文档编写

**详细实施记录**: 参见 [IMPLEMENTATION.md](IMPLEMENTATION.md)

## 相关链接 (Related Links)

- 相关规格 (Related Specs): 
  - [026-backend-code-optimization](../026-backend-code-optimization/) - 后端 API 优化（依赖）
  - [023-i18n-internationalization](../023-i18n-internationalization/) - 国际化（错误消息翻译）
- 依赖项 (Dependencies): 
  - Spec 026 必须完成（✅ 已完成）
  - 后端 V2 API 响应格式
- 代码位置:
  - `web-next/src/lib/api-client-v2.ts` - 统一 API 客户端
  - `web-next/src/lib/services/` - API 服务层
  - `web-next/src/types/api-v2.ts` - 类型定义
