# Design Document

## Overview

本设计文档描述 Go 后端代码优化与架构改进的技术方案，重点关注代码清理、架构分层、依赖注入和性能优化。

## Architecture

### 目标架构（三层架构）

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Layer (Handler)                    │
│  - 请求验证                                                   │
│  - 参数解析                                                   │
│  - 响应构建                                                   │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer (Business Logic)           │
│  - 业务逻辑                                                   │
│  - 数据转换                                                   │
│  - 事务管理                                                   │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   Repository Layer (Data Access)             │
│  - 数据库操作                                                 │
│  - 缓存管理                                                   │
│  - 外部 API 调用                                              │
└─────────────────────────────────────────────────────────────┘
```

### 当前问题

1. **代码组织混乱**: Handler 中包含业务逻辑和数据访问
2. **依赖管理**: 使用全局变量和单例，难以测试
3. **文档残留**: MeiliSearch 相关文档和配置未清理
4. **错误处理**: 不统一，缺少上下文信息
5. **日志规范**: 格式不统一，缺少结构化日志

## Components and Interfaces

### 1. 依赖注入容器

```go
type Container struct {
    Config           *Config
    ContentAPI       *content.Api
    QdrantClient     *qdrant.Client
    QdrantSearcher   *qdrant.Searcher
    CacheManager     *cache.Manager
    ChatDB           *chat.DB
    ChatAgent        *chat.Agent
    SSEManager       *tasks.SSEManager
    TaskManager      *tasks.Manager
}

func NewContainer(config *Config) (*Container, error) {
    // 初始化所有依赖
    // 返回容器实例
}
```

### 2. Service 层接口

```go
type BookService interface {
    GetBook(ctx context.Context, id int64) (*Book, error)
    ListBooks(ctx context.Context, opts ListOptions) ([]Book, int64, error)
    UpdateBook(ctx context.Context, id int64, updates *BookUpdate) error
    DeleteBook(ctx context.Context, id int64) error
}

type SearchService interface {
    Search(ctx context.Context, query string, opts SearchOptions) ([]Book, int64, error)
    SemanticSearch(ctx context.Context, query string, opts SearchOptions) ([]Book, error)
}
```

### 3. Repository 层接口

```go
type BookRepository interface {
    FindByID(ctx context.Context, id int64) (*Book, error)
    FindAll(ctx context.Context, opts QueryOptions) ([]Book, int64, error)
    Update(ctx context.Context, book *Book) error
    Delete(ctx context.Context, id int64) error
}
```

## Data Models

### 统一错误类型

```go
type AppError struct {
    Code    string
    Message string
    Err     error
    Context map[string]interface{}
}

func (e *AppError) Error() string {
    return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
}
```

### 统一响应格式

```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

type PaginatedResponse struct {
    Records    interface{} `json:"records"`
    Total      int64       `json:"total"`
    Limit      int         `json:"limit"`
    Offset     int         `json:"offset"`
    NextCursor string      `json:"next_cursor,omitempty"`
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: 分层隔离性

*For any* 业务逻辑变更，Handler 层和 Repository 层代码应该不需要修改

**Validates: Requirements 1.3**

### Property 2: 依赖可测试性

*For any* Service 或 Repository，应该能够通过 mock 接口进行单元测试

**Validates: Requirements 2.2**

### Property 3: 错误信息完整性

*For any* 错误响应，应该包含错误码、错误消息和足够的上下文信息

**Validates: Requirements 3.3**

### Property 4: 文档一致性

*For any* 文档中提到的搜索功能，应该只引用 Qdrant，不应该包含 MeiliSearch

**Validates: Requirements 13.1, 13.4**

### Property 5: 配置完整性

*For any* 必需的配置项，应该在启动时验证，缺失时应该返回明确的错误

**Validates: Requirements 6.3**

## Error Handling

### 错误分类

1. **业务错误**: 用户输入错误、资源不存在等
2. **系统错误**: 数据库连接失败、外部 API 超时等
3. **验证错误**: 参数验证失败

### 错误处理流程

```go
// Service 层返回 AppError
func (s *BookService) GetBook(ctx context.Context, id int64) (*Book, error) {
    book, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, &AppError{
            Code: "BOOK_NOT_FOUND",
            Message: "Book not found",
            Err: err,
            Context: map[string]interface{}{"book_id": id},
        }
    }
    return book, nil
}

// Handler 层统一处理错误
func (h *BookHandler) GetBook(c *gin.Context) {
    book, err := h.service.GetBook(c.Request.Context(), id)
    if err != nil {
        handleError(c, err)
        return
    }
    c.JSON(200, Response{Code: 200, Data: book})
}
```

## Testing Strategy

### 单元测试

1. **Service 层测试**: Mock Repository，测试业务逻辑
2. **Repository 层测试**: 使用测试数据库或 mock
3. **Handler 层测试**: Mock Service，测试 HTTP 处理

### 集成测试

1. **API 端到端测试**: 测试完整的请求响应流程
2. **数据库集成测试**: 测试实际的数据库操作

### 测试覆盖率目标

- Service 层: 80%+
- Repository 层: 70%+
- Handler 层: 60%+
- 总体: 70%+

## Implementation Notes

### 优先级排序

**P0 (紧急 - 本周完成)**:
1. 清理 MeiliSearch 文档残留
2. 移除 MeiliSearch 配置项
3. 更新 API 文档说明

**P1 (高优先级 - 2 周内)**:
1. 实现依赖注入容器
2. 重构 BookHandler 分离业务逻辑
3. 统一错误处理

**P2 (中优先级 - 1 个月内)**:
1. 实现 Service 层
2. 实现 Repository 层
3. 添加单元测试

**P3 (低优先级 - 后续迭代)**:
1. 性能优化
2. 完善文档
3. 提高测试覆盖率

### 重构策略

采用**渐进式重构**，避免大规模改动：

1. **第一阶段**: 清理文档和配置（不影响代码）
2. **第二阶段**: 引入依赖注入（保持接口兼容）
3. **第三阶段**: 分离 Service 层（逐个模块重构）
4. **第四阶段**: 引入 Repository 层（完善测试）

### MeiliSearch 清理清单

**文档清理**:
- `docs/QUICK_START.md` - 移除 MeiliSearch 启动说明
- `docs/MCP_README.md` - 移除 MeiliSearch 配置
- `docs/API_DOCUMENTATION.md` - 更新搜索说明
- `qdrant.md` - 移除 MeiliSearch 对比内容

**配置清理**:
- `config.yaml` - 移除 search.host 等配置
- 环境变量文档 - 移除 CALIBRE_SEARCH_HOST

**代码验证**:
- 确认没有 meilisearch 包导入
- 确认没有相关函数调用

## Performance Considerations

### 缓存策略

1. **书籍元数据缓存**: 使用 Redis 或内存缓存
2. **搜索结果缓存**: 缓存热门查询结果
3. **缓存失效**: 更新/删除时主动失效

### 数据库优化

1. **连接池**: 合理配置连接池大小
2. **索引优化**: 为常用查询字段添加索引
3. **批量操作**: 使用批量插入/更新

### 并发控制

1. **Goroutine 池**: 限制并发数避免资源耗尽
2. **Context 超时**: 为所有操作设置超时
3. **优雅关闭**: 确保服务关闭时完成所有请求

## Security Considerations

1. **输入验证**: 所有用户输入必须验证
2. **SQL 注入防护**: 使用预编译语句
3. **敏感信息**: 日志中脱敏处理
4. **CORS 配置**: 生产环境限制允许的域名
5. **认证授权**: 为敏感操作添加认证（后续）
