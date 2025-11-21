# Calibre API 代码结构说明

## 概览

为了提高代码可维护性和可读性，`internal/calibre` 模块已按功能拆分为多个文件。

## 文件结构

### 核心文件

#### `route.go`
- **职责**: 路由配置和 API 初始化
- **主要内容**:
  - `Api` 结构体定义
  - `SetupRouter()` - 路由注册
  - `NewClient()` - API 客户端初始化

### 处理器文件 (Handlers)

#### `book_handler.go`
- **职责**: 书籍基本操作
- **主要功能**:
  - `getBook()` - 获取书籍信息
  - `deleteBook()` - 删除书籍
  - `updateMetadata()` - 更新元数据
  - `recently()` - 最近更新的书籍
  - `random()` - 随机书籍
  - `getAllBooks()` - 获取所有书籍（支持游标分页）
  - `listPublisher()` - 获取出版社列表

#### `book_content_handler.go`
- **职责**: 书籍内容相关操作
- **主要功能**:
  - `getCover()` - 获取封面
  - `proxyCover()` - 代理封面图片
  - `getBookFile()` - 下载书籍文件
  - `getBookToc()` - 获取目录
  - `getBookContent()` - 获取书籍内容（路径参数）
  - `getBookContentByQuery()` - 获取书籍内容（查询参数）
  - `getFileOrCache()` - 文件缓存管理
  - `expansionTree()` - 目录树展开

#### `search_handler.go`
- **职责**: 搜索功能
- **主要功能**:
  - `search()` - 通用搜索（支持关键词、语义、混合搜索）
  - `semanticSearch()` - 语义搜索

#### `metadata_handler.go`
- **职责**: 元数据操作
- **主要功能**:
  - `getIsbn()` - 通过 ISBN 获取元数据
  - `queryMetadata()` - 查询元数据

#### `task_handler.go`
- **职责**: 任务管理
- **主要功能**:
  - `listTasks()` - 获取任务列表
  - `startTask()` - 启动任务
  - `stopTask()` - 停止任务

#### `mcp_handler.go`
- **职责**: MCP (Model Context Protocol) 增强工具端点
- **主要功能**:
  - `getEnhancedTools()` - 获取增强工具列表
  - `executeEnhancedTool()` - 执行增强工具

### 工具文件

#### `converter.go`
- **职责**: 数据类型转换
- **主要功能**:
  - `convertSemanticToBook()` - semantic.Book → calibre.Book
  - `convertSemanticToBooks()` - 批量转换
  - `convertContentBooks()` - content.Book → calibre.Book
  - `convertContentToBooks()` - content.Content → calibre.Book

#### `errors.go`
- **职责**: 错误定义
- **定义的错误**:
  - `ErrSearchServiceNotAvailable` - 搜索服务不可用
  - `ErrBookNotFound` - 书籍未找到
  - `ErrInvalidParameters` - 无效参数

### 其他文件（未修改）

- `types.go` - 数据类型定义
- `schemas.go` - Schema 定义
- `utils.go` - 工具函数
- `zip.go` - ZIP 文件处理
- `mcp_enhanced_tools.go` - MCP 增强工具实现
- `mcp_prompts.go` - MCP 提示管理
- `mcp_resources.go` - MCP 资源管理

## 路由映射

### 书籍基本操作
```
GET    /api/book/:id              → book_handler.go: getBook()
POST   /api/book/:id/delete       → book_handler.go: deleteBook()
POST   /api/book/:id/update       → book_handler.go: updateMetadata()
GET    /api/recently              → book_handler.go: recently()
GET    /api/random                → book_handler.go: random()
GET    /api/books/all             → book_handler.go: getAllBooks()
GET    /api/publisher             → book_handler.go: listPublisher()
```

### 书籍内容
```
GET    /api/get/cover/:id         → book_content_handler.go: getCover()
GET    /api/proxy/cover/*path     → book_content_handler.go: proxyCover()
GET    /api/download/book/:id     → book_content_handler.go: getBookFile()
GET    /api/read/:id/toc          → book_content_handler.go: getBookToc()
GET    /api/read/:id/file/*path   → book_content_handler.go: getBookContent()
GET    /api/book/content          → book_content_handler.go: getBookContentByQuery()
```

### 搜索
```
GET    /api/search                → search_handler.go: search()
POST   /api/search                → search_handler.go: search()
GET    /api/search/semantic       → search_handler.go: semanticSearch()
```

### 元数据
```
GET    /api/metadata/isbn/:isbn   → metadata_handler.go: getIsbn()
GET    /api/metadata/search       → metadata_handler.go: queryMetadata()
```

### 任务管理
```
GET    /api/tasks                 → task_handler.go: listTasks()
POST   /api/tasks/start           → task_handler.go: startTask()
POST   /api/tasks/:id/stop        → task_handler.go: stopTask()
```

### MCP 增强工具
```
GET    /api/mcp/tools/enhanced        → mcp_handler.go: getEnhancedTools()
POST   /api/mcp/tools/enhanced/:tool  → mcp_handler.go: executeEnhancedTool()
```

## 设计原则

1. **单一职责**: 每个文件负责一类相关功能
2. **清晰命名**: 文件名直接反映其职责
3. **易于维护**: 修改某个功能时只需关注对应文件
4. **便于扩展**: 新增功能时可以轻松添加新的 handler 文件
5. **低耦合**: 各模块之间通过 `Api` 结构体共享必要的依赖

## 依赖关系

```
route.go (核心)
    ↓
各种 handler 文件
    ↓
converter.go, errors.go (工具)
    ↓
types.go (数据定义)
```

## 未来优化建议

1. 考虑将 `Api` 结构体拆分为更小的服务组件
2. 引入依赖注入容器统一管理依赖
3. 为每个 handler 添加单元测试
4. 考虑使用接口进行解耦，便于 mock 和测试
5. 添加中间件层处理通用逻辑（如认证、日志、错误处理）

