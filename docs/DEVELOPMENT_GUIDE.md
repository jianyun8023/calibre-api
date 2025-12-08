# 开发指南

> 详细的代码规范、开发流程和最佳实践

## 📐 代码规范

### Go 代码规范

#### 1. 命名规范
```go
// 包名: 全小写，简短有意义
package calibre

// 导出类型: 首字母大写，驼峰命名
type BookHandler struct {}

// 未导出类型: 首字母小写
type cacheEntry struct {}

// 接口: 以 -er 结尾或动词短语
type Searcher interface {}
type ContentProvider interface {}

// 方法: 驼峰命名，动词开头
func (h *BookHandler) GetBook(id string) (*Book, error) {}

// 常量: 全大写或驼峰
const MaxRetryCount = 3
const defaultTimeout = 30 * time.Second

// 变量: 驼峰命名，见名知意
var ErrBookNotFound = errors.New("book not found")
```

#### 2. 错误处理
```go
// ✅ 推荐: 明确的错误类型
var (
    ErrBookNotFound              = errors.New("book not found")
    ErrSearchServiceNotAvailable = errors.New("search service not available")
    ErrInvalidParameters         = errors.New("invalid parameters")
)

// ✅ 推荐: 包装错误提供上下文
if err != nil {
    return nil, fmt.Errorf("failed to fetch book %d: %w", id, err)
}

// ✅ 推荐: 早返回
if book == nil {
    return nil, ErrBookNotFound
}

// ✅ 推荐: 处理或记录错误
if err := saveToCache(book); err != nil {
    log.Warnf("Failed to cache book %d: %v", book.ID, err)
}
```

#### 3. 结构体设计
```go
// ✅ 推荐: 字段按逻辑分组，添加标签
type Book struct {
    // 基本信息
    ID           int64     `json:"id"`
    Title        string    `json:"title"`
    Authors      []string  `json:"authors"`
    
    // 元数据
    Publisher    string    `json:"publisher"`
    PubDate      time.Time `json:"pubdate"`
    ISBN         string    `json:"isbn"`
    
    // 系统字段
    LastModified time.Time `json:"last_modified"`
}

// ✅ 推荐: 使用嵌入减少重复
type ApiBase struct {
    config *Config
    http   *client.Client
}

type Api struct {
    ApiBase
    contentApi *content.Api
    // ... 其他字段
}
```

#### 4. 函数设计
```go
// ✅ 推荐: 参数不超过 3-4 个，使用结构体
type SearchOptions struct {
    Query  string
    Limit  int
    Offset int
    Filter map[string]string
}

func (s *Searcher) Search(ctx context.Context, opts SearchOptions) ([]Book, error) {
    // ...
}

// ✅ 推荐: 使用上下文控制超时和取消
func (c *Client) FetchBook(ctx context.Context, id int64) (*Book, error) {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    // ...
}

// ✅ 推荐: 返回具体错误，而非 panic
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}
```

#### 5. 并发安全
```go
// ✅ 推荐: 使用 sync.RWMutex 保护共享资源
type Cache struct {
    mu    sync.RWMutex
    items map[string]*Entry
}

func (c *Cache) Get(key string) (*Entry, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    entry, ok := c.items[key]
    return entry, ok
}

func (c *Cache) Set(key string, entry *Entry) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = entry
}

// ✅ 推荐: 使用 channel 进行 goroutine 通信
func (m *Manager) processJobs(ctx context.Context) {
    for {
        select {
        case job := <-m.jobChan:
            m.handleJob(job)
        case <-ctx.Done():
            return
        }
    }
}
```

#### 6. 日志规范
```go
// 使用统一的日志包
import "github.com/jianyun8023/calibre-api/pkg/log"

// 日志级别
log.Debugf("Debug info: %+v", obj)        // 调试信息
log.Infof("Server started on %s", addr)  // 重要事件
log.Warnf("Cache miss for key %s", key)  // 警告但不影响运行
log.Errorf("Failed to connect: %v", err) // 错误需要关注
log.Fatal(err)                            // 致命错误，程序退出

// ✅ 推荐: 结构化日志
log.Infof("Book updated: id=%d, title=%s, user=%s", 
    book.ID, book.Title, user)

// ❌ 避免: 敏感信息
log.Infof("User logged in: password=%s", pwd) // 不要记录密码
```

### API 设计规范

#### 1. RESTful 路由
```go
// ✅ 推荐: 遵循 RESTful 约定
GET    /api/books          // 列表
GET    /api/books/:id      // 详情
POST   /api/books          // 创建
PUT    /api/books/:id      // 完整更新
PATCH  /api/books/:id      // 部分更新
DELETE /api/books/:id      // 删除

// ✅ 推荐: 资源嵌套
GET    /api/books/:id/chapters      // 书籍的章节
GET    /api/books/:id/comments      // 书籍的评论

// ✅ 推荐: 动作用动词
POST   /api/books/:id/publish       // 发布书籍
POST   /api/search                  // 搜索
POST   /api/tasks/start             // 启动任务
```

#### 2. 响应格式
```go
// ✅ 推荐: 统一的成功响应
{
    "data": { ... },           // 数据
    "message": "success",      // 可选消息
    "timestamp": "2024-01-01T00:00:00Z"
}

// ✅ 推荐: 统一的错误响应
{
    "error": {
        "code": "BOOK_NOT_FOUND",
        "message": "Book with ID 123 not found",
        "details": { ... }      // 可选详情
    },
    "timestamp": "2024-01-01T00:00:00Z"
}

// ✅ 推荐: 分页响应
{
    "data": [ ... ],
    "pagination": {
        "total": 1000,
        "limit": 20,
        "offset": 0,
        "has_more": true
    }
}
```

#### 3. 状态码使用
```go
// 2xx 成功
200 OK              // 成功（GET, PUT, PATCH）
201 Created         // 创建成功（POST）
204 No Content      // 成功但无返回内容（DELETE）

// 4xx 客户端错误
400 Bad Request     // 请求参数错误
401 Unauthorized    // 未认证
403 Forbidden       // 无权限
404 Not Found       // 资源不存在
409 Conflict        // 冲突（如重复创建）
422 Unprocessable   // 语义错误

// 5xx 服务器错误
500 Internal Error  // 服务器内部错误
502 Bad Gateway     // 上游服务错误
503 Unavailable     // 服务不可用
```

### 数据库规范

#### 1. Calibre SQLite 查询
```go
// ✅ 推荐: 使用参数化查询防止 SQL 注入
query := "SELECT * FROM books WHERE id = ?"
row := db.QueryRow(query, bookID)

// ✅ 推荐: 处理 NULL 值
var publisher sql.NullString
if err := row.Scan(&publisher); err != nil {
    return err
}
if publisher.Valid {
    book.Publisher = publisher.String
}

// ✅ 推荐: 使用事务保证一致性
tx, err := db.Begin()
if err != nil {
    return err
}
defer tx.Rollback() // 失败时回滚

// 执行多个操作
if err := tx.Exec(...); err != nil {
    return err
}

return tx.Commit() // 成功时提交
```

#### 2. Qdrant 操作
```go
// ✅ 推荐: 批量操作提高性能
points := make([]*qdrant.PointStruct, 0, batchSize)
for _, book := range books {
    point := createPoint(book)
    points = append(points, point)
}
err := client.UpsertPoints(ctx, collection, points)

// ✅ 推荐: 使用过滤器精确查询
filter := &qdrant.Filter{
    Must: []*qdrant.Condition{
        {
            Field: "publisher",
            Match: &qdrant.Match{Value: "O'Reilly"},
        },
    },
}
results, err := client.Search(ctx, query, filter)
```

## 🧪 测试规范

### 单元测试
```go
// 测试文件命名: *_test.go
// 测试函数命名: Test<FunctionName>

func TestConvertSemanticToBook(t *testing.T) {
    // Arrange - 准备测试数据
    semanticBook := &semantic.Book{
        ID:      123,
        Title:   "Test Book",
        Authors: "Author A, Author B",
    }
    
    // Act - 执行测试
    result := convertSemanticToBook(semanticBook)
    
    // Assert - 验证结果
    if result.ID != 123 {
        t.Errorf("Expected ID 123, got %d", result.ID)
    }
    if len(result.Authors) != 2 {
        t.Errorf("Expected 2 authors, got %d", len(result.Authors))
    }
}

// ✅ 推荐: 使用表驱动测试
func TestSearchBooks(t *testing.T) {
    tests := []struct {
        name     string
        query    string
        expected int
        wantErr  bool
    }{
        {"empty query", "", 0, true},
        {"valid query", "golang", 10, false},
        {"special chars", "C++", 5, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            results, err := searchBooks(tt.query)
            if (err != nil) != tt.wantErr {
                t.Errorf("wantErr %v, got %v", tt.wantErr, err)
            }
            if len(results) != tt.expected {
                t.Errorf("expected %d results, got %d", tt.expected, len(results))
            }
        })
    }
}
```

### 集成测试
```bash
# 测试脚本: test_*.sh
# 测试 MCP 端点
./examples/test_mcp_full.sh

# 测试 Chat API
./test_chat_api.sh

# 测试 Schema
./test_mcp_schemas.sh
```

## 🚀 开发工作流

### 1. 环境设置
```bash
# 克隆项目
git clone https://github.com/jianyun8023/calibre-api.git
cd calibre-api

# 安装 Go 依赖
go mod download

# 配置环境
cp config.yaml.example config.yaml
# 编辑 config.yaml 配置数据库等

# 启动 Qdrant（Docker）
docker run -d -p 6333:6333 qdrant/qdrant

# 构建并运行
make build
./calibre-api
```

### 2. 开发流程
```bash
# 1. 创建功能分支
git checkout -b feature/new-search-algorithm

# 2. 开发功能（遵循代码规范）

# 3. 运行测试
go test ./...
make test

# 4. 更新 CHANGELOG.md（重要！）
# 在 [Unreleased] 部分记录本次变更
# - Added: 新功能
# - Changed: 功能变更
# - Fixed: Bug 修复
# - Removed: 移除的功能

# 5. 提交代码
git add .
git commit -m "feat: implement new search algorithm"

# 6. 推送并创建 PR
git push origin feature/new-search-algorithm
```

### 3. Makefile 命令
```bash
# 构建
make build          # 构建主程序
make build-all      # 构建所有平台

# 测试
make test           # 运行测试

# 清理
make clean          # 清理构建文件

# Docker
make docker-build   # 构建 Docker 镜像
make docker-run     # 运行 Docker 容器
```

## 🔍 常见开发任务

### 添加新的 API 端点
1. 在 `internal/calibre/` 创建或更新 handler 文件
2. 在 `SetupRouter()` 中注册路由
3. 在 `schemas.go` 中定义参数结构（如需要）
4. 更新 API 文档
5. 在 `CHANGELOG.md` 记录变更

### 添加新的 MCP Tool
1. 在 `mcp_tools.go` 中添加工具定义和描述
2. 在工具处理函数中实现业务逻辑（调用现有 Handler）
3. 使用 MCP Inspector 测试工具调用
4. 更新 `docs/MCP_README.md` 工具列表
5. 在 `CHANGELOG.md` 记录变更

**注意**：只添加只读操作工具，写操作（更新、删除）应通过 Web UI 进行

### 添加新的搜索策略
1. 在 `search_handler.go` 中实现搜索逻辑
2. 更新 `search()` 函数支持新策略
3. 测试搜索结果质量
4. 在 `CHANGELOG.md` 记录变更

### 添加新的任务类型
1. 在 `internal/tasks/` 中实现任务逻辑
2. 在 `TaskManager` 中注册任务类型
3. 在 `task_handler.go` 中添加启动入口
4. 在 `CHANGELOG.md` 记录变更

### 维护 CHANGELOG.md

#### 📋 变更日志规范

所有代码变更必须在 `CHANGELOG.md` 中记录，遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/) 规范。

#### 变更类型分类
- **Added**: 新的 API 端点、MCP Tool、配置选项、依赖包
- **Changed**: API 行为调整、配置结构变更、依赖版本升级、性能优化
- **Fixed**: Bug 修复、相关 Issue 编号
- **Removed**: 废弃的 API、移除的依赖、删除的配置项
- **Security**: 安全漏洞修复、权限控制改进
- **Breaking Changes**: ⚠️ 标记会破坏向后兼容性的变更

#### 最佳实践
1. ✅ **及时更新**: 每次代码变更时立即更新 CHANGELOG
2. ✅ **用户视角**: 从用户角度描述变更影响，而非技术实现
3. ✅ **具体明确**: 提供具体的 API、配置名称，避免模糊描述
4. ✅ **关联 Issue**: 引用相关的 Issue 或 PR 编号
5. ❌ **避免技术细节**: 不要记录代码重构、内部重命名等不影响用户的变更
6. ❌ **避免 Git 日志**: CHANGELOG 是面向用户的，不是 Git Commit 日志的复制

## 🐛 调试技巧

### 启用调试模式
```yaml
# config.yaml
debug: true
```

### 查看日志
```bash
# 实时查看日志
tail -f app.log

# 过滤错误
grep "ERROR" app.log

# 查看特定模块
grep "MCP" app.log
```

### 使用 MCP Inspector
```bash
# 测试 MCP 端点
./examples/test_mcp_full.sh

# 使用浏览器访问
open examples/mcp_sse_client.html
```

### 性能分析
```go
// 添加性能分析
import _ "net/http/pprof"

// main.go
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

```bash
# 查看 CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile

# 查看内存 profile
go tool pprof http://localhost:6060/debug/pprof/heap
```

## 📝 注意事项

### 重要约定
1. **配置管理**: 敏感信息使用环境变量，不要提交到代码库
2. **数据库迁移**: Chat DB 使用 SQL 迁移文件，不要直接修改表结构
3. **向量维度**: Embedding 维度必须与 Qdrant 集合一致（默认 4096）
4. **缓存清理**: CacheManager 自动清理，注意磁盘空间配额
5. **MCP 端点**: 使用 SSE 传输（`/mcp/sse`），MCP v1.2.0+ 标准
6. **并发安全**: 共享状态必须加锁或使用 channel
7. **变更日志**: 任何代码变更必须同步更新 `CHANGELOG.md`，保持版本历史可追溯
8. **MCP 安全**: 只暴露只读工具，写操作通过 Web UI 执行

### 性能考虑
1. 使用 `context.Context` 控制超时
2. 大量数据使用流式处理或分页
3. 频繁访问的数据使用缓存
4. Qdrant 批量操作提高性能
5. 避免在循环中进行网络请求

### 安全注意
1. 验证所有用户输入
2. 使用参数化查询防止 SQL 注入
3. 限制文件上传大小和类型
4. 日志中不要记录敏感信息
5. API 密钥使用环境变量

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'feat: Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### Commit 规范
```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式（不影响功能）
refactor: 重构
test: 测试相关
chore: 构建/工具相关
```

