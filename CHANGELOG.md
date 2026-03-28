# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **草稿取消接口** (2026-03-28):
  - 新增 `POST /api/drafts/cancel` 接口，支持通过书籍 ID 取消待处理草稿
  - 取消操作直接删除数据库记录，不保留历史
  - 支持批量取消多本书籍的草稿
  - 完善的错误处理，支持部分成功场景（部分书籍取消成功，部分失败）
  - 区分三种响应：全部成功、部分成功、完全失败（均返回 200 状态码）
  - 详细的响应数据：`cancelled_books`（取消的书籍数）、`cancelled_drafts_count`（删除的草稿总数）
  - 完整的单元测试和集成测试覆盖

- **草稿管理系统** (2026-03-27):
  - 新增外部书籍变更草稿管理功能
  - 支持书籍删除和更新变更的草稿审核流程
  - 新增草稿历史记录查询 API (`GET /api/drafts/history`)
  - 自动清理过期草稿（可配置过期天数）
  - 后台定时任务重置卡住的草稿状态
  - 支持分页查询待审核草稿和历史记录

### Changed
- **草稿管理性能优化** (2026-03-27):
  - 批量草稿应用支持并发处理（最大并发数: 5）
  - 新增批量查询优化，避免 N+1 查询问题
  - 重复草稿自动合并或更新，避免重复创建
  - 添加 SQLite WAL 模式提升并发性能
  - 新增数据库索引优化查询性能：
    - `idx_book_drafts_status`: 状态+创建时间索引
    - `idx_book_drafts_lookup`: 书籍ID+操作+状态复合索引
    - `idx_book_draft_history_time`: 历史记录时间倒序索引

- **草稿管理原子性增强** (2026-03-27):
  - 使用数据库事务确保状态更新和历史记录的原子性
  - 新增 `processing` 中间状态，防止并发重复处理
  - 草稿应用失败时自动回滚到 `pending` 状态
  - 新增 `expired` 状态，自动标记过期草稿

- **草稿管理 API 响应优化** (2026-03-27):
  - 批量操作返回部分成功结果和错误列表
  - 新增分页响应，包含总记录数和页码信息
  - 前端显示敏感字段时自动遮掩（token, password, secret, key, auth）

### Fixed
- **搜索索引同步修复** (2026-03-28):
  - 修复草稿应用后 MeiliSearch 索引未正确更新的问题
  - 问题原因：`UpdateMetadata` 在更新 Calibre 后从 MeiliSearch 读取旧数据，导致旧数据覆盖新数据
  - 解决方案：更新后直接从 Calibre Content Server 获取最新数据，确保索引正确同步
  - 新增 `ContentAPI.GetBookDetail` 方法用于从 Calibre 读取最新书籍元数据
  - 影响范围：所有通过草稿管理系统更新的书籍元数据

- **草稿管理稳定性修复** (2026-03-27):
  - 修复应用崩溃导致草稿永久处于 processing 状态的问题
  - 启动时自动重置超过 1 小时的卡住草稿
  - 修复并发应用同一草稿的竞态条件问题
  - 修复草稿历史记录未正确创建的问题
  - **修复 Goroutine 泄露风险**: 后台清理任务支持优雅退出
    - 添加 Context 控制和信号处理机制（SIGINT, SIGTERM）
    - HTTP 服务器支持优雅关闭（5 秒超时）
    - 后台任务在应用关闭时正确清理，防止资源泄露

- **草稿更新字段规则优化** (2026-03-28):
  - **更新规则**：保护书籍元数据完整性，只有 Tags 允许清空，其他字段只能补全
    - **Tags 字段**：允许清空（`[]` = 清空，`null` = 不更新）
    - **Authors 字段**：只允许非空数组，不允许清空
    - **字符串字段**（Title, Publisher, Comments, Isbn）：只允许非空值，不允许清空
  - **数据结构优化**：BookUpdate 字段改为指针类型，支持明确区分"不更新"和"设置值"
    - `nil` = 不更新该字段
    - 空字符串 `""` = 被拒绝（不允许清空）
    - 非空值 = 更新为新值
  - 前端显示优化：空数组现在会正确显示在草稿变更列表中
  - 典型应用场景：批量清理书籍标签中的垃圾推广信息（如 "公众号：xxx" 等）

### Security
- **草稿管理数据安全** (2026-03-27):
  - 前端数据展示自动遮掩敏感字段（token, password, secret, key, auth），防止信息泄露
  - 注意：系统设计为内网使用，草稿管理端点暂未启用 API Key 认证

### Removed
- **移除 API Key 认证中间件** (2026-03-27):
  - 删除 `internal/middleware/auth.go` 和 `auth_test.go`
  - 移除 `config.yaml` 中的 `auth` 配置段
  - 移除 `internal/calibre/types.go` 中的 `AuthConfig` 结构
  - 移除草稿管理路由中的 API Key 认证中间件
  - 原因：系统设计为内网自用，暂不需要 API Key 认证。后续如需认证将使用 Gin 统一解决方案

- **移除 Qdrant 向量数据库支持** (2026-03-03):
  - 删除 `internal/semantic/qdrant/` 整个包（client.go, searcher.go）
  - 删除 `internal/tasks/qdrant_sync.go` 同步任务实现
  - 删除 `docs/QDRANT_COLLECTION_SETUP.md` 配置文档
  - 删除 `create_indexes.sh` 索引创建脚本
  - 移除 `internal/container/container.go` 中的 Qdrant 初始化逻辑
  - 统一使用 MeiliSearch 作为搜索引擎
  - 任务类型重命名：`qdrant_sync` → `semantic_sync`
  - 前端任务显示更新为 "Semantic Sync"
  - 清理 Docker Compose 中的 Qdrant 服务配置
  - 更新所有相关文档（README, AGENTS, ARCHITECTURE 等）

- **移除旧版 BookService** (2026-03-02):
  - 删除 `internal/service/book_service.go` 中的旧版 `bookService` 实现
  - 删除 `internal/calibre/book_handler.go` 中的旧版 Handler 方法
  - 简化适配器逻辑，移除降级机制
  - 统一使用 Repository 模式的新架构
  - 减少代码约 300 行，提升代码清晰度
  - 保留转换函数供其他模块使用
  - 所有功能通过 `BookHandlerV2` 和 `bookServiceV2` 实现

### Fixed
- **Qdrant 增量同步修复** (2025-01-04):
  - 修复增量同步无法正确同步缺失书籍的问题
  - **问题原因**: 旧实现使用 `GetMaxID()` 获取 Qdrant 中最大 ID，然后查询 `id:>maxID` 的书籍。这种方法无法发现 ID < maxID 但不在 Qdrant 中的书籍（如之前同步失败的）
  - **修复方案**: 新实现通过比较 Calibre 和 Qdrant 的所有 ID 差异，找出真正缺失的书籍进行同步
  - 新增 `findMissingBooks()` 方法，实现正确的增量同步逻辑
  - 新增单元测试 `TestFindMissingBooksLogic` 验证修复

### Added
- **后端架构重构 P2** (Spec 026, 2025-12-12):
  - **Repository 层**:
    - 新增 `internal/repository` 包，实现数据抽象层
    - BookRepository 接口（8 个方法：FindByID, FindRecent, FindRandom, FindAllWithCursor, SearchByKeyword, Update, Delete）
    - QdrantBookRepository 实现
    - BookServiceV2 使用 Repository 抽象
    - 完整的四层架构（Handler → Service → Repository → Client）
  - **结构化日志**:
    - 新增 `pkg/logger` 包，基于 zerolog 实现
    - 支持日志级别配置（Debug, Info, Warn, Error, Fatal）
    - 实现 Gin 日志中间件（请求日志、恢复中间件、Context 日志器）
    - 结构化日志字段（request_id, method, path, status, latency 等）
    - 支持美化输出（开发模式）
  - **路由优化**:
    - 设计模块化路由结构（books, search, metadata, tasks, chat）
    - API 版本管理设计（/api/v1）
    - 路由文档注释完善
    - 保持向后兼容性
  - **单元测试**:
    - 新增 `pkg/errors/errors_test.go`（5 个测试，37.2% 覆盖率）
    - 新增 `pkg/response/response_test.go`（4 个测试，59.6% 覆盖率）
    - 所有测试通过（9/9）
  - **代码质量**:
    - 新增约 1,200 行高质量代码（P2）
    - 新增约 180 行测试代码
    - 编译测试通过，无 linter 错误
  - **测试验证**:
    - 编译测试: 100% 通过
    - 单元测试: 9/9 通过（100%）
    - errors 包覆盖率: 37.2%
    - response 包覆盖率: 59.6%
    - 创建测试验证报告（TEST_REPORT.md）

- **后端架构重构 P1** (Spec 026, 2025-12-12):
  - **统一错误处理**:
    - 新增 `pkg/errors` 包，定义 AppError 类型
    - 错误码分类（1xxx 通用、2xxx 业务、3xxx 搜索、4xxx 任务、5xxx 聊天、6xxx 验证）
    - 支持错误链、上下文信息、HTTP 状态码映射
    - 预定义常见业务错误（BookNotFound, SearchServiceNotAvailable, TaskNotFound 等）
  - **统一响应格式**:
    - 新增 `pkg/response` 包，定义 Response 和 PaginatedResponse 结构
    - 实现 Builder 模式构建响应
    - 便捷函数（Success, Error, Paginated, BadRequest, NotFound, ServiceUnavailable 等）
    - 自动集成 AppError，提取错误详情和上下文
  - **Service 层重构**:
    - 新增 `internal/service` 包，实现三层架构（Handler → Service → Repository）
    - BookService 接口和实现，包含 7 个业务方法
    - ContentAPI 接口抽象，解耦 content.Api 依赖
    - 业务逻辑从 Handler 分离到 Service 层
  - **Handler 层重构**:
    - 新增 BookHandlerV2，使用 Service 层和统一响应格式
    - Handler 只负责 HTTP 请求/响应转换，不包含业务逻辑
    - 保持向后兼容，支持优雅降级
  - **依赖注入增强**:
    - 容器支持创建 Service 和 Handler
    - contentAPIAdapter 适配器，桥接 content.Api 和 service.ContentAPI
    - 统一管理所有组件依赖关系
  - **代码质量**:
    - 新增约 1,400 行高质量代码
    - 重构约 500 行代码
    - 编译测试通过，无 linter 错误

### Added
- **元数据编辑功能** (Spec 010, 2025-12-10):
  - 单本书籍元数据编辑对话框
  - 批量元数据管理页面（含侧边栏导航入口）
  - **元数据网络搜索功能**：
    - 从豆瓣搜索书籍元数据（支持 ISBN/书名/作者）
    - 搜索结果展示（封面、标题、作者、出版社等）
    - 新旧元数据对比与选择
    - 每个字段独立选择"新"或"旧"数据
    - 作者和标签多选支持
  - 支持编辑：标题、作者、出版社、ISBN、标签、评分、描述
  - 保存后自动刷新数据

### Fixed
- **Spec 010 Bug 修复** (2025-12-10):
  - 修复批量元数据管理页面缺少导航入口
  - 修复搜索 API 参数错误（`query` → `q`）
  - 修复元数据搜索 API 响应格式处理（豆瓣 API 格式）
  
- **Spec 018 Chat 功能修复** (2025-12-11):
  - 修复 React Markdown 组件 className 属性错误
  - 添加 `/api/chat/stream` 路由以支持 AI SDK
  - 实现 ChatStream 处理器，兼容 Vercel AI SDK v5 SSE 格式
  - 修复 CGO 编译问题（SQLite 依赖）
  - 修复 chat.Message ID 类型错误
  - 修复 SSE 响应格式（`data: ` 前缀）
  - 添加工具调用支持（书籍搜索和推荐）
  - 修复聊天窗口布局（固定高度+内容滚动）
  - 移除发送按钮的 `disabled` 属性（依赖内部逻辑判断）
  - **测试验证**：
    - ✅ 后端 API 测试通过（curl）
    - ✅ 书籍推荐功能正常（返回24本科幻书籍）
    - ✅ SSE 流式响应正常
    - ✅ AI 生成回复质量良好

- **前端 UI/UX 优化** (2025-12-10):
  - **数据展示修复**: 修复 Discover 区域不显示数据的问题
  - **布局优化**: 从 6 列改为 5 列布局，减少拥挤感，防止图片变形
  - **动画优化**: 优化刷新 Skeleton 高度（从 320px 到 192px），添加淡入淡出动画
  - **分页优化**: 
    - 首页 Recently Added: 10 本 → 15 本（3行满）
    - All Books 页面: 12 本 → 20 本（4行满）
    - 优化原则：在主要屏幕尺寸下尽可能满行
  - **功能修复**:
    - Settings 页面路由（/setting → /settings）
    - 书籍详情页编辑/刷新按钮
    - 下载文件自动添加文件扩展名
  - **代码质量**: 修复所有 linter 错误，优化 React Hooks 使用
  - 详见: [specs/009-frontend-migration/OPTIMIZATION_SUMMARY.md](specs/009-frontend-migration/OPTIMIZATION_SUMMARY.md)

- **前端重构**: 从 Vue 3 + Element Plus 迁移到 Next.js 15 + Shadcn/UI
  - **新目录**: `web-next/` 完整的 Next.js 应用
  - **技术栈升级**:
    - Next.js 16 (App Router, Static Export)
    - React 19 + TypeScript 5
    - Shadcn/UI (20+ 组件)
    - Tailwind CSS 4 (Glassmorphism 样式系统)
  - **核心页面** (6/10 完成):
    - ✅ Home - 随机推荐 + 最近添加
    - ✅ Books - 基于游标的分页列表
    - ✅ Detail - 完整元数据展示和操作
    - ✅ Read - react-reader EPUB 阅读器
    - ✅ Search - 关键词/语义/混合三种模式
    - ✅ Chat - Vercel AI SDK 流式对话
  - **新增页面** (4/10 新完成):
    - ✅ Settings - 主题切换、API 配置、阅读器设置、搜索偏好
    - ✅ Tasks - 任务管理、实时状态更新、进度显示
    - ✅ Publisher - 出版社列表、统计信息、书籍筛选
    - ✅ Metadata Manager - 批量元数据编辑、预览变更
  - **功能亮点**:
    - Vercel AI SDK 无缝处理 SSE 流式响应
    - react-reader 集成 (epub.js)，支持阅读进度保存
    - Glassmorphism 样式系统（暗黑模式适配）
    - 响应式布局（桌面侧边栏 + 移动端 Sheet 抽屉）
    - 完整的 Zustand 状态管理
    - next-themes 主题切换支持
  - **Shadcn/UI 组件**: 新增 25+ 组件
    - Button, Card, Input, Label, Textarea
    - Select, Checkbox, Switch, Progress
    - Badge, Tabs, Dialog, Sheet
    - Skeleton, ScrollArea, NavigationMenu
    - Sonner (Toast 通知)
  - **静态导出配置**: 支持 Go 后端托管
  - 详见规格文档: [specs/009-frontend-migration/README.md](specs/009-frontend-migration/README.md)

### Changed
- **MCP 搜索优化**: `search_books` 工具改用纯语义搜索
  - 从关键词搜索 (`SearchByKeyword`) 切换到语义搜索 (`Search`)
  - 移除 `filter` 和 `offset` 参数，保留 `limit` 参数
  - 更智能的搜索体验，支持自然语言查询
  - 更新工具描述：明确说明使用向量相似度匹配
- **MCP 书籍信息增强**: `get_book` 工具新增目录（TOC）信息
  - 返回完整的书籍元数据 + 目录结构
  - 目录数据优先从 Qdrant 获取，缺失时从 EPUB 提取
  - 目录获取失败不影响基本元数据返回
  - 帮助 AI 更好地理解书籍结构并进行总结推荐

### Security
- **依赖升级**: 升级 Go 依赖包解决安全漏洞
  - 更新所有依赖至最新稳定版本
  - 修复 27 个安全漏洞（4 个高危，18 个中危，5 个低危）
  - 主要依赖版本：
    - `golang.org/x/crypto` v0.45.0
    - `golang.org/x/net` v0.47.0
    - `golang.org/x/sys` v0.38.0
    - `golang.org/x/term` v0.37.0
    - `golang.org/x/text` v0.31.0
    - `google.golang.org/protobuf` v1.36.10

### Fixed
- **日志格式**: 修复 Go 1.25 严格格式检查导致的编译错误
  - 修改 `log.Infof()` 调用使用常量格式字符串
  - 影响文件：`pkg/content/api.go`, `internal/calibre/metadata_handler.go`

### Improved
- **代码质量全面优化**: 对 8610 行 Go 代码进行了全面的质量优化
  - **错误处理**: 移除 panic 改为优雅退出，新增 9 个错误类型定义
  - **代码规范**: 统一格式化，新增 12 个常量定义，添加 50+ 行文档注释
  - **性能优化**: 确认并发安全，HTTP 客户端复用，添加性能基准测试
  - **结构优化**: 修复接口重复定义，添加配置验证函数
  - **测试覆盖**: 新增 16 个单元测试用例和 4 个性能基准测试
  - 详见 [CODE_QUALITY_IMPROVEMENTS.md](./CODE_QUALITY_IMPROVEMENTS.md)

## [1.2.0] - 2024-11-28

### 🎉 重大更新：MCP 框架迁移

#### Added
- **MCP 新架构**: 完全迁移到 `github.com/mark3labs/mcp-go@v0.43.1` 框架
- **多传输模式支持**: 
  - SSE (Server-Sent Events) 传输模式（默认）
  - StreamableHTTP 传输模式
  - 通过配置文件灵活切换
- **新 MCP 端点**:
  - `GET /mcp/sse` - SSE 连接端点
  - `POST /mcp/message` - 消息处理端点
- **6 个 MCP 工具**（阶段二）:
  - `search_books` - 搜索书籍（关键词 + 语义搜索）
  - `get_book` - 获取书籍详情
  - `random_books` - 随机推荐
  - `recent_books` - 最近更新
  - `get_isbn_metadata` - ISBN 元数据查询（豆瓣）
  - `search_metadata` - 在线元数据搜索（豆瓣）
  - **所有工具均为只读操作，保证安全**
- **配置增强**:
  - `mcp.transport`: 传输模式选择 (`sse` 或 `http`)
  - `mcp.sse_endpoint`: SSE 端点路径配置
  - `mcp.message_endpoint`: 消息端点路径配置
  - `mcp.version`: 更新为 `1.2.0`
- **CORS 支持**: 添加跨域支持，兼容 MCP Inspector 等浏览器客户端

#### Changed
- **依赖替换**: 移除 `gin-mcp`，使用官方 `mcp-go` 库
- **MCP 端点变更**: 从 `/mcp` 迁移到 `/mcp/sse` (SSE模式)
- **配置结构**: `MCPConfig` 新增 `Transport`, `SSEEndpoint`, `MessageEndpoint` 字段
- **版本号**: 从 `1.1.0` 升级到 `1.2.0`

#### Removed
- 移除 `github.com/ckanthony/gin-mcp` 依赖
- 删除 `internal/calibre/mcp_handler.go`
- 删除 `internal/calibre/mcp_enhanced_tools.go`
- 移除旧的 `/mcp` 单一端点
- **出于安全考虑移除危险工具**:
  - `update_book_metadata` - 更新元数据（风险高，应通过 Web UI 操作）
  - `delete_book` - 删除书籍（不可逆，应通过 Web UI 确认后操作）

#### Fixed
- MCP 协议标准化，遵循官方规范
- 更好的传输层抽象和扩展性

#### Technical Details
- 创建 `internal/calibre/mcp_server.go` 核心模块（120 行）
- 创建 `internal/calibre/mcp_tools.go` 工具实现（533 行）
- 实现 `MCPServer` 结构体封装 mcp-go 服务器
- 支持 SSE 和 StreamableHTTP 两种传输实现
- 工具注册框架和 6 个安全工具完整实现
- 保持所有现有 HTTP API 端点向后兼容
- 添加 `github.com/gin-contrib/cors` 中间件

#### Breaking Changes
⚠️ **重要**: MCP 客户端需要更新连接配置
- **旧配置**: `http://localhost:8080/mcp`
- **新配置 (SSE)**: `http://localhost:8080/mcp/sse`
- 建议在 MCP Inspector 中更新服务器地址

#### Migration Guide
1. 更新 `config.yaml`，确保包含新的 MCP 配置字段
2. MCP 客户端更新连接地址到 `/mcp/sse`
3. 如需使用 HTTP 模式，设置 `mcp.transport: "http"`
4. 所有 REST API 端点保持不变，前端应用无需修改

#### Notes
- 工具注册功能框架已就绪，具体工具将在后续版本实现
- 资源和提示管理将在后续版本适配
- 建议使用 MCP Inspector 测试新端点

---

## [1.1.0] - 2024-11-27

### Added
- 智能问答功能 (Chat Agent)
- LLM 集成 (OpenAI, Ollama)
- 语义搜索增强
- Qdrant 向量数据库支持

### Changed
- 改进搜索性能
- 优化缓存管理

---

## [1.0.0] - 2024-11-01

### Added
- 初始版本发布
- 基础书籍管理功能
- Calibre Content Server 集成
- 书籍搜索功能
- 元数据管理
- Web 前端界面

---

## 版本说明

- **Major**: 重大架构变更或不兼容的 API 变更
- **Minor**: 向后兼容的新功能
- **Patch**: 向后兼容的错误修复

## 链接

- [GitHub Repository](https://github.com/jianyun8023/calibre-api)
- [Issue Tracker](https://github.com/jianyun8023/calibre-api/issues)
- [MCP 文档](docs/MCP_README.md)

