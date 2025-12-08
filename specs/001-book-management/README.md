---
status: complete
created: '2025-12-08'
tags: []
priority: medium
created_at: '2025-12-08T13:03:51.822Z'
updated_at: '2025-12-08T13:05:06.079Z'
completed_at: '2025-12-08T13:05:06.079Z'
completed: '2025-12-08'
transitions:
  - status: complete
    at: '2025-12-08T13:05:06.079Z'
---

# book-management

> **Status**: ✅ Complete · **Priority**: Medium · **Created**: 2025-12-08

## Overview

**问题**: Calibre 原生 Content Server 仅提供基础的书籍浏览，缺乏现代化 API 和扩展能力。

**解决方案**: 构建 RESTful API 层包装 Calibre 数据库，提供完整 CRUD 操作、元数据管理和文件访问。

**为什么现在**: 这是所有高级功能（搜索、AI 问答、MCP 集成）的基础。

## Design

### 架构决策

**Handler 层分离**:
- `book_handler.go`: 元数据操作（列表、详情、更新、删除）
- `book_content_handler.go`: 文件操作（封面、下载、目录、内容读取）

**权衡**: 单一 handler vs 职责分离
- 选择分离：更清晰的职责边界，避免单个文件过大
- 代价：需要在两个文件间共享 `Api` 实例

**缓存策略**:
- EPUB 文件缓存避免重复解压
- LRU 策略控制磁盘占用
- 跨平台支持（Darwin/Linux/Windows）

**游标分页**:
```go
// 使用 last_modified + id 作为游标
cursor := fmt.Sprintf("%d_%d", book.LastModified.Unix(), book.ID)
```

**权衡**: offset vs cursor
- 选择 cursor：避免深分页性能问题，支持实时数据变更
- 代价：不支持跳页，只能顺序翻页

### 核心接口

**书籍操作**:
- `GET /api/books` - 游标分页列表
- `GET /api/books/:id` - 单本详情
- `POST /api/books/:id/update` - 更新元数据
- `DELETE /api/books/:id` - 删除书籍
- `GET /api/recently` - 最近更新
- `GET /api/random` - 随机推荐

**内容访问**:
- `GET /api/:id/cover.jpg` - 封面图
- `GET /api/:id/:file` - 下载文件
- `GET /api/:id/toc` - 目录结构
- `GET /api/:id/content` - 章节内容

## Plan

- [x] 设计 RESTful 路由结构
- [x] 实现 Calibre DB 查询逻辑
- [x] 实现缓存管理器（跨平台）
- [x] 实现 EPUB TOC 解析
- [x] 实现文件代理和下载
- [x] 添加元数据更新支持
- [x] 添加书籍删除（需调用 Content Server API）

## Test

- [x] 分页浏览 1000+ 书籍库
- [x] 封面图正确加载（包括中文书名）
- [x] EPUB 目录正确解析多层结构
- [x] 元数据更新同步到 Calibre DB
- [x] 删除操作不破坏数据库完整性
- [x] 缓存 LRU 策略生效

## Notes

**已知限制**:
- 仅支持 SQLite Calibre 数据库
- 删除操作依赖 Content Server API（需配置凭证）
- EPUB 目录解析不支持非标准格式

**未来优化**:
- 支持批量元数据更新
- 添加书籍封面上传
- 支持更多电子书格式（MOBI, AZW3）

**依赖项**:
- `pkg/content/api.go`: Calibre Content Server 客户端
- `internal/cache/manager.go`: 文件缓存管理
