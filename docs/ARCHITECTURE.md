# 系统架构

> Calibre API 的详细架构设计和模块说明

## 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      前端层 (Next.js)                        │
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

## 模块职责

### 1. 路由层 (`main.go`, `route.go`)
- **职责**: 初始化配置、注册路由、启动服务
- **关键组件**:
  - `initConfig()`: 加载配置文件和环境变量
  - `SetupRouter()`: 注册 RESTful API 路由
  - MCP SSE 端点注册 (`/mcp/sse`, `/mcp/message`)

### 2. 处理器层 (Handlers)

#### `book_handler.go` - 书籍操作
- `getBook()` - 获取单本书籍详情
- `getAllBooks()` - 游标分页获取书籍列表
- `deleteBook()` - 删除书籍
- `updateMetadata()` - 更新元数据
- `recently()` - 最近更新书籍
- `random()` - 随机推荐
- `listPublisher()` - 出版社列表

#### `book_content_handler.go` - 内容处理
- `getCover()` - 获取封面图片
- `getBookFile()` - 下载书籍文件
- `getBookToc()` - 获取目录结构 (TOC)
- `getBookContent()` - 读取书籍内容
- `getFileOrCache()` - 缓存管理
- `expansionTree()` - 目录树展开

#### `search_handler.go` - 搜索功能
- `search()` - 统一搜索入口（支持多种策略）
  - 关键词搜索: Calibre Content Server API
  - 语义搜索: Qdrant 向量搜索
  - 混合搜索: 先召回后精排
- `semanticSearch()` - 纯语义搜索端点

**混合搜索优化策略** (v1.2+):
1. **召回阶段**: 语义搜索获取 top 100 候选（保证召回率）
2. **精排阶段**: 关键词匹配计算得分（提高准确率）
3. **综合评分**: `finalScore = semanticScore * 0.4 + keywordScore * 0.6`
4. **返回结果**: 按综合得分排序返回 top N

#### `chat_handler.go` - 智能问答
- `CreateConversation()` - 创建会话
- `ListConversations()` - 列出会话
- `GetConversation()` - 获取会话详情
- `GetConversationMessages()` - 获取消息历史
- `DeleteConversation()` - 删除会话
- `SendMessage()` - 发送消息（智能应答）

**ChatAgent 工作流程**:
1. 接收用户问题
2. 语义搜索相关书籍 (Qdrant)
3. 提取相关书籍目录 (TOC)
4. 构建上下文（书籍信息 + 目录）
5. 调用 LLM 生成回答
6. 支持工具调用（搜索、获取详情等）

#### `metadata_handler.go` - 元数据
- `getIsbn()` - ISBN 查询元数据
- `queryMetadata()` - 在线元数据搜索

#### `task_handler.go` - 任务管理
- `listTasks()` - 任务列表
- `startTask()` - 启动任务（同步 Qdrant、提取 TOC 等）
- `stopTask()` - 停止任务

### 3. MCP 层 (`mcp_server.go`, `mcp_tools.go`)

#### MCP Server 架构
```go
type MCPServer struct {
    server      *server.MCPServer  // mcp-go 核心服务器
    sseServer   *server.SSEServer  // SSE 传输层
    api         *Api               // 业务逻辑访问
    toolMgr     *EnhancedToolManager  // 工具管理
    resourceMgr *ResourceManager      // 资源管理
}
```

#### 注册的 MCP Tools (v1.2.0+)
**只读安全工具** (6 个):
- `search_books`: 语义搜索书籍（使用向量相似度）
- `get_book`: 获取书籍详情（包含 TOC 目录结构）
- `random_books`: 随机推荐
- `recent_books`: 最近更新书籍
- `get_isbn_metadata`: ISBN 元数据查询（豆瓣）
- `search_metadata`: 在线元数据搜索（豆瓣）

**已移除的危险工具** (出于安全考虑):
- ~~`update_book_metadata`~~: 更新元数据（应通过 Web UI 操作）
- ~~`delete_book`~~: 删除书籍（不可逆操作，应通过 Web UI 确认）

#### 注册的 MCP Resources
- `book://{id}`: 单本书籍资源（JSON 格式，含元数据、TOC、文件列表）
- `search_results://{query}`: 搜索结果缓存

#### 注册的 MCP Prompts
- `book_search`: 书籍搜索提示模板
- `book_recommendation`: 推荐提示模板

### 4. 服务层 (Services)

#### ContentAPI (`pkg/content/api.go`)
- 与 Calibre Content Server 交互
- 书籍元数据获取
- 文件下载和代理

#### QdrantClient (`internal/semantic/qdrant/client.go`)
- 向量数据库操作
- 点 (Points) 增删改查
- 批量操作支持

#### Searcher (`internal/semantic/qdrant/searcher.go`)
- 语义搜索实现
- 向量化查询文本
- 结果过滤和排序

#### ChatAgent (`internal/chat/agent.go`)
```go
type Agent struct {
    llm         LLM                // LLM 客户端
    searcher    SemanticSearcher   // 语义搜索器
    tocFetcher  TocFetcher         // 目录获取函数
}

// 核心方法
- Chat(ctx, conversationID, userMessage) // 智能对话
```

#### EmbeddingProvider (`internal/semantic/embedding/provider.go`)
- 支持 Ollama (本地)
- 支持 SiliconFlow (云端)
- 统一的向量化接口

#### CacheManager (`internal/cache/manager.go`)
- EPUB 文件缓存
- LRU 策略
- 自动清理
- 容量管理

#### TaskManager (`internal/tasks/manager.go`)
- 异步任务调度
- 进度跟踪和状态管理
- 支持的任务类型:
  - `qdrant_sync`: 同步书籍到 Qdrant (~10min/1000本)
  - `toc_extract`: 提取书籍目录 (~2-3min/1000本，优化后提升 5-7x)
  - `check_missing`: 检查缺失书籍 (~30s)

**TOC 提取性能优化** (v1.1.0):
- 缓存管理器锁优化：无锁快速路径（缓存命中时 10-20x 提升）
- 批量 Qdrant 更新：20 本/批次（网络延迟减少 95%）
- Worker 并发数提升：10 个 worker（处理速度提升 80-100%）

## 部署架构

### Docker 部署
```yaml
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
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    volumes:
      - ./config.yaml:/app/config.yaml
      - ./cache:/app/.cache
      - chat_db:/app/chat.db
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
  chat_db:
```

### 性能指标
- **并发**: 100+ 并发请求
- **响应时间**: P95 < 500ms
- **关键词搜索**: ~100ms
- **语义搜索**: ~200ms
- **混合搜索**: ~250ms
- **MCP Tool 调用**: ~100ms

## 技术选型

### 后端技术
- **Go 1.24.4**: 高性能、并发友好
- **Gin**: 轻量级 Web 框架
- **SQLite**: 嵌入式数据库，易部署
- **Qdrant**: 高性能向量数据库

### 前端技术
- **Next.js 15**: React 框架，App Router
- **Shadcn/UI**: 现代化 UI 组件库
- **Tailwind CSS 4**: 原子化 CSS

### AI/LLM
- **OpenAI API**: GPT-4 等模型
- **Ollama**: 本地 LLM 部署
- **MCP v1.2.0**: AI 助手协议

## 安全设计

### MCP 安全
- 只暴露只读操作工具
- 写操作通过 Web UI 进行
- 支持 SSE 长连接

### 数据安全
- 参数化查询防止 SQL 注入
- API 密钥使用环境变量
- 日志脱敏处理

### 并发安全
- `sync.RWMutex` 保护共享资源
- Channel 通信避免竞态
- Context 控制超时和取消

