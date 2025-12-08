# Calibre API - AI 助手开发指南

> 本文档为 AI 助手提供项目架构、代码规范和开发指南

## 📋 项目概述

**Calibre API** 是一个基于 Go 语言开发的 Calibre 书籍管理系统，集成了语义搜索、智能问答和 MCP (Model Context Protocol) 支持，提供强大的书籍管理和 AI 交互能力。

### 核心特性
- 📚 **书籍管理**: 完整的 CRUD 操作、元数据管理、文件处理
- 🔍 **智能搜索**: 关键词搜索、语义搜索、混合搜索策略
- 🤖 **AI 集成**: MCP 协议支持、LLM 智能问答、上下文感知对话
- 🚀 **高性能**: Qdrant 向量数据库、缓存管理、并发处理
- 📦 **多平台**: Docker 部署、跨平台二进制、CI/CD 自动化

### 技术栈
- **后端**: Go 1.24.4, Gin Web Framework
- **前端**: Vue.js 3 (详见 [前端开发指南](app/AGENTS.md))
- **数据库**: SQLite (Calibre), SQLite (Chat History)
- **向量数据库**: Qdrant (语义搜索)
- **AI/LLM**: OpenAI API, Ollama (本地部署)
- **协议**: MCP (Model Context Protocol) with SSE
- **构建工具**: Make, Docker, GitHub Actions

## 🏗️ 架构设计

### 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        前端层 (Vue.js)                       │
│  ┌────────────┬───────────────┬──────────────┬────────────┐ │
│  │  书籍浏览  │   搜索界面    │  阅读器界面  │  聊天界面  │ │
│  └────────────┴───────────────┴──────────────┴────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │ HTTP/REST API
┌─────────────────────────────┼─────────────────────────────────┐
│                    API 网关层 (Gin Router)                    │
│  ┌────────────┬────────────┬────────────┬──────────────────┐ │
│  │  书籍路由  │  搜索路由  │  聊天路由  │  MCP SSE 端点    │ │
│  └────────────┴────────────┴────────────┴──────────────────┘ │
└───────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────┼─────────────────────────────────┐
│                       业务逻辑层 (Handlers)                    │
│  ┌─────────────────────┬──────────────────┬─────────────────┐│
│  │  BookHandler        │  SearchHandler   │  ChatHandler    ││
│  │  - CRUD 操作        │  - 关键词搜索    │  - 会话管理     ││
│  │  - 元数据管理       │  - 语义搜索      │  - 消息处理     ││
│  │  - 文件处理         │  - 混合策略      │  - 工具调用     ││
│  └─────────────────────┴──────────────────┴─────────────────┘│
│  ┌─────────────────────┬──────────────────┬─────────────────┐│
│  │  ContentHandler     │  MetadataHandler │  TaskHandler    ││
│  │  - 目录获取         │  - ISBN 查询     │  - 任务管理     ││
│  │  - 文件读取         │  - 在线搜索      │  - 进度跟踪     ││
│  │  - 封面处理         │  - 数据补全      │  - 异步执行     ││
│  └─────────────────────┴──────────────────┴─────────────────┘│
│  ┌─────────────────────────────────────────────────────────┐ │
│  │  MCPServer (MCP Protocol Handler)                      │ │
│  │  - Tool Registration & Execution                        │ │
│  │  - Resource Management (books, search results)         │ │
│  │  - Prompt Templates                                     │ │
│  └─────────────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────┼─────────────────────────────────┐
│                     服务层 (Services & Clients)                │
│  ┌──────────────┬───────────────┬──────────────┬───────────┐ │
│  │ ContentAPI   │  QdrantClient │  ChatAgent   │ LLMClient │ │
│  │ (Calibre)    │  (向量搜索)    │  (智能代理)  │ (AI 模型) │ │
│  └──────────────┴───────────────┴──────────────┴───────────┘ │
│  ┌──────────────┬───────────────┬──────────────────────────┐ │
│  │ CacheManager │ TaskManager   │ EmbeddingProvider        │ │
│  │ (文件缓存)    │ (任务调度)     │ (向量化服务)             │ │
│  └──────────────┴───────────────┴──────────────────────────┘ │
└───────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────┼─────────────────────────────────┐
│                        数据层 (Data Storage)                   │
│  ┌──────────────┬───────────────┬──────────────┬───────────┐ │
│  │  Calibre DB  │  Qdrant DB    │  Chat DB     │ File Cache│ │
│  │  (SQLite)    │  (向量)        │  (SQLite)    │ (Disk)    │ │
│  └──────────────┴───────────────┴──────────────┴───────────┘ │
└───────────────────────────────────────────────────────────────┘
```

### 模块职责

#### 1. 路由层 (`main.go`, `route.go`)
- **职责**: 初始化配置、注册路由、启动服务
- **关键组件**:
  - `initConfig()`: 加载配置文件和环境变量
  - `SetupRouter()`: 注册 RESTful API 路由
  - MCP SSE 端点注册 (`/mcp/sse`, `/mcp/message`)

#### 2. 处理器层 (Handlers)

##### `book_handler.go` - 书籍操作
```go
// 核心功能
- getBook()        // 获取单本书籍详情
- getAllBooks()    // 游标分页获取书籍列表
- deleteBook()     // 删除书籍
- updateMetadata() // 更新元数据
- recently()       // 最近更新书籍
- random()         // 随机推荐
- listPublisher()  // 出版社列表
```

##### `book_content_handler.go` - 内容处理
```go
// 核心功能
- getCover()            // 获取封面图片
- getBookFile()         // 下载书籍文件
- getBookToc()          // 获取目录结构 (TOC)
- getBookContent()      // 读取书籍内容
- getFileOrCache()      // 缓存管理
- expansionTree()       // 目录树展开
```

##### `search_handler.go` - 搜索功能
```go
// 搜索策略
- search()          // 统一搜索入口（支持多种策略）
  ├─ 关键词搜索     // Calibre Content Server API
  ├─ 语义搜索       // Qdrant 向量搜索
  └─ 混合搜索       // 结合两者并去重
- semanticSearch()  // 纯语义搜索端点
```

**混合搜索策略**:
1. 并发执行关键词搜索和语义搜索
2. 使用 `map[int64]*Book` 按 `book_id` 去重
3. 语义搜索结果优先（更智能）
4. 关键词搜索补充缺失结果
5. 按相关性排序返回

##### `chat_handler.go` - 智能问答
```go
// 会话管理
- CreateConversation()      // 创建会话
- ListConversations()        // 列出会话
- GetConversation()          // 获取会话详情
- GetConversationMessages()  // 获取消息历史
- DeleteConversation()       // 删除会话
- SendMessage()              // 发送消息（智能应答）
```

**ChatAgent 工作流程**:
1. 接收用户问题
2. 语义搜索相关书籍 (Qdrant)
3. 提取相关书籍目录 (TOC)
4. 构建上下文（书籍信息 + 目录）
5. 调用 LLM 生成回答
6. 支持工具调用（搜索、获取详情等）

##### `metadata_handler.go` - 元数据
```go
- getIsbn()        // ISBN 查询元数据
- queryMetadata()  // 在线元数据搜索
```

##### `task_handler.go` - 任务管理
```go
- listTasks()   // 任务列表
- startTask()   // 启动任务（同步 Qdrant、提取 TOC 等）
- stopTask()    // 停止任务
```

#### 3. MCP 层 (`mcp_server.go`, `mcp_enhanced_tools.go`)

##### MCP Server 架构
```go
type MCPServer struct {
    server      *server.MCPServer  // mcp-go 核心服务器
    sseServer   *server.SSEServer  // SSE 传输层
    api         *Api               // 业务逻辑访问
    toolMgr     *EnhancedToolManager  // 工具管理
    resourceMgr *ResourceManager      // 资源管理
}
```

##### 注册的 MCP Tools
- `search_books`: 搜索书籍（支持混合策略）
- `get_book`: 获取书籍详情
- `random_books`: 随机推荐
- `update_book_metadata`: 更新元数据
- `delete_book`: 删除书籍
- `get_isbn_metadata`: ISBN 元数据查询
- `search_metadata`: 在线元数据搜索

##### 注册的 MCP Resources
- `book://{id}`: 单本书籍资源
- `search_results://{query}`: 搜索结果资源

##### 注册的 MCP Prompts
- `book_search`: 书籍搜索提示模板
- `book_recommendation`: 推荐提示模板

#### 4. 服务层 (Services)

##### ContentAPI (`pkg/content/api.go`)
- 与 Calibre Content Server 交互
- 书籍元数据获取
- 文件下载和代理

##### QdrantClient (`internal/semantic/qdrant/client.go`)
- 向量数据库操作
- 点 (Points) 增删改查
- 批量操作支持

##### Searcher (`internal/semantic/qdrant/searcher.go`)
- 语义搜索实现
- 向量化查询文本
- 结果过滤和排序

##### ChatAgent (`internal/chat/agent.go`)
```go
type Agent struct {
    llm         LLM                // LLM 客户端
    searcher    SemanticSearcher   // 语义搜索器
    tocFetcher  TocFetcher         // 目录获取函数
}

// 核心方法
- Chat(ctx, conversationID, userMessage) // 智能对话
```

##### EmbeddingProvider (`internal/semantic/embedding/provider.go`)
- 支持 Ollama (本地)
- 支持 SiliconFlow (云端)
- 统一的向量化接口

##### CacheManager (`internal/cache/manager.go`)
- EPUB 文件缓存
- LRU 策略
- 自动清理
- 容量管理

##### TaskManager (`internal/tasks/manager.go`)
- 异步任务调度
- 进度跟踪
- 支持的任务类型:
  - `qdrant_sync`: 同步书籍到 Qdrant
  - `toc_extract`: 提取书籍目录
  - `check_missing`: 检查缺失书籍

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

// ❌ 避免: 忽略错误
_ = saveToCache(book) // 不好

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

### 前端开发规范

前端代码规范和开发指南请参考：[前端开发指南](app/AGENTS.md)

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

## 📦 部署指南

### Docker 部署
```yaml
# docker-compose.yaml
version: '3.8'
services:
  calibre-api:
    image: calibre-api:latest
    ports:
      - "8080:8080"
    environment:
      - CALIBRE_CONTENT_SERVER=https://your-calibre-server.com
      - CALIBRE_QDRANT_URL=http://qdrant:6333
      - CALIBRE_MCP_ENABLED=true
    volumes:
      - ./config.yaml:/app/config.yaml
      - ./cache:/app/.cache
    depends_on:
      - qdrant
  
  qdrant:
    image: qdrant/qdrant:latest
    ports:
      - "6333:6333"
    volumes:
      - qdrant_data:/qdrant/storage

volumes:
  qdrant_data:
```

### 二进制部署
```bash
# 下载发布版本
wget https://github.com/jianyun8023/calibre-api/releases/download/v1.0.0/calibre-api-linux-amd64

# 添加执行权限
chmod +x calibre-api-linux-amd64

# 配置
cp config.yaml.example config.yaml
vim config.yaml

# 运行
./calibre-api-linux-amd64
```

### Systemd 服务
```ini
# /etc/systemd/system/calibre-api.service
[Unit]
Description=Calibre API Service
After=network.target

[Service]
Type=simple
User=calibre
WorkingDirectory=/opt/calibre-api
ExecStart=/opt/calibre-api/calibre-api
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# 启用服务
sudo systemctl enable calibre-api
sudo systemctl start calibre-api
sudo systemctl status calibre-api
```

## 🔍 常见开发任务

### 添加新的 API 端点
1. 在 `internal/calibre/` 创建或更新 handler 文件
2. 在 `SetupRouter()` 中注册路由
3. 在 `schemas.go` 中定义参数结构（如需要）
4. 更新 API 文档
5. 在 `CHANGELOG.md` 记录变更

### 添加新的 MCP Tool
1. 在 `mcp_enhanced_tools.go` 的 `GetEnhancedTools()` 中添加工具定义
2. 在 `ExecuteEnhancedTool()` 中实现工具逻辑
3. 测试工具调用
4. 在 `CHANGELOG.md` 记录变更

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
```markdown
### Added - 新增功能
- 新的 API 端点
- 新的 MCP Tool
- 新的配置选项
- 新的依赖包

### Changed - 功能变更
- API 行为调整
- 配置结构变更
- 依赖版本升级
- 性能优化

### Fixed - Bug 修复
- 修复的问题描述
- 相关 Issue 编号

### Removed - 移除功能
- 废弃的 API
- 移除的依赖
- 删除的配置项

### Security - 安全修复
- 安全漏洞修复
- 权限控制改进

### Breaking Changes - 不兼容变更
⚠️ 标记会破坏向后兼容性的变更
```

#### 编写示例
```markdown
## [Unreleased]

### Added
- **语义搜索增强**: 新增混合搜索策略，结合关键词和语义搜索
- **MCP Tool**: 新增 `get_book_toc` 工具获取书籍目录
- **配置项**: 新增 `search.hybrid_mode` 配置启用混合搜索

### Changed
- **搜索 API**: `/api/search` 默认使用混合搜索策略
- **缓存策略**: LRU 缓存大小从 100MB 增加到 500MB

### Fixed
- 修复语义搜索结果重复的问题 (#123)
- 修复 EPUB 文件解析中文目录乱码 (#124)

### Breaking Changes
⚠️ **搜索 API 参数变更**: `strategy` 参数从 `keyword|semantic` 改为 `keyword|semantic|hybrid`
```

#### 发布时更新
发布新版本时：
1. 将 `[Unreleased]` 改为版本号和日期
```markdown
## [1.3.0] - 2024-12-01
```
2. 在顶部添加新的 `[Unreleased]` 部分
3. 在底部添加版本比较链接
```markdown
[1.3.0]: https://github.com/jianyun8023/calibre-api/compare/v1.2.0...v1.3.0
```

#### 最佳实践
1. ✅ **及时更新**: 每次代码变更时立即更新 CHANGELOG
2. ✅ **用户视角**: 从用户角度描述变更影响，而非技术实现
3. ✅ **具体明确**: 提供具体的 API、配置名称，避免模糊描述
4. ✅ **关联 Issue**: 引用相关的 Issue 或 PR 编号
5. ❌ **避免技术细节**: 不要记录代码重构、内部重命名等不影响用户的变更
6. ❌ **避免 Git 日志**: CHANGELOG 是面向用户的，不是 Git Commit 日志的复制

#### 检查清单
提交 PR 前确认：
- [ ] 变更已记录在 `CHANGELOG.md` 的 `[Unreleased]` 部分
- [ ] 变更分类正确（Added/Changed/Fixed/Removed）
- [ ] 描述清晰，用户可理解
- [ ] 不兼容变更已标记 `⚠️ Breaking Changes`
- [ ] 相关 Issue 已引用

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

## 📚 相关文档

- [API 文档](docs/API_DOCUMENTATION.md)
- [代码结构](docs/CODE_STRUCTURE.md)
- [MCP 集成指南](docs/MCP_README.md)
- [混合搜索策略](docs/HYBRID_SEARCH_STRATEGY.md)
- [快速开始](docs/QUICK_START.md)
- [性能优化](docs/PERFORMANCE_OPTIMIZATION.md)
- [变更日志](CHANGELOG.md)

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

## 📝 注意事项

### 重要约定
1. **配置管理**: 敏感信息使用环境变量，不要提交到代码库
2. **数据库迁移**: Chat DB 使用 SQL 迁移文件，不要直接修改表结构
3. **向量维度**: 确保 Embedding 提供商返回的维度与 Qdrant 集合配置一致（4096）
4. **缓存清理**: CacheManager 会自动清理，但注意磁盘空间
5. **MCP 路由**: MCP 路由必须在 NoRoute 之前注册
6. **并发安全**: 共享状态必须加锁或使用 channel
7. **变更日志**: 任何代码变更必须同步更新 `CHANGELOG.md`，保持版本历史可追溯

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

---

**版本**: 1.1.0  
**最后更新**: 2024-11-28  
**维护者**: jianyun8023

