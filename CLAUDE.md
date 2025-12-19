# Calibre API - AI 助手开发指南

> 本文档为 AI 助手提供项目概览和快速参考

## 🎯 快速导航

**我需要理解项目** → 阅读 [项目概述](#📋-项目概述) 和 [架构设计](docs/ARCHITECTURE.md)  
**我需要编写代码** → 参考 [开发指南](docs/DEVELOPMENT_GUIDE.md)  
**我需要添加功能** → 查看 [开发工作流](#🚀-开发工作流) 和 [Specs 文档](specs/)  
**我需要了解变更** → 查看 [CHANGELOG.md](CHANGELOG.md)

## 📋 项目概述

**Calibre API** 是一个将 Calibre 从传统电子书管理工具升级为 AI 原生智能书库系统的项目。

### 核心价值
- 📚 **现代化 API**: RESTful + MCP 双协议支持
- 🔍 **智能搜索**: 关键词 + 语义混合搜索（先召回后精排）
- 🤖 **AI 问答**: 基于书库内容的智能推荐和对话
- 🚀 **开箱即用**: Docker 一键部署，跨平台支持

### 技术栈
- **后端**: Go 1.24.4, Gin Web Framework
- **前端**: Vue.js 3 (详见 [前端指南](app/AGENTS.md))
- **数据库**: SQLite (Calibre DB, Chat DB)
- **向量数据库**: Qdrant (语义搜索)
- **AI/LLM**: OpenAI API, Ollama (本地部署)
- **协议**: MCP v1.2.0 (SSE 传输)

### 项目规模
- **代码量**: ~13K 行 (Go 8K + Vue 5K)
- **性能**: 100+ 并发，P95 < 500ms
- **功能模块**: 7 个核心 Specs
- **API 端点**: 20+ RESTful + 6 MCP Tools
- **测试覆盖**: 60%+ 单元测试

## 📚 文档导航

### 核心指南
- [AGENTS.md](AGENTS.md) - AI 助手快速参考（LeanSpec 工作流）
- [CLAUDE.md](CLAUDE.md) - 本文档（项目概览）
- [CHANGELOG.md](CHANGELOG.md) - 版本变更历史

### 详细文档
- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - 系统架构和模块设计
- [DEVELOPMENT_GUIDE.md](docs/DEVELOPMENT_GUIDE.md) - 代码规范和开发流程
- [API_DOCUMENTATION.md](docs/API_DOCUMENTATION.md) - RESTful API 参考
- [CODE_STRUCTURE.md](docs/CODE_STRUCTURE.md) - 项目目录说明
- [PHASE_WORKFLOW.md](docs/PHASE_WORKFLOW.md) - 阶段工作流指南（PDPI-spec）

### 功能规格 (`specs/` 目录)
- [项目概览](specs/007-000-project-overview/) - 系统架构和技术栈
- [书籍管理](specs/001-book-management/) - CRUD 操作和文件处理
- [搜索功能](specs/002-search-functionality/) - 混合搜索策略
- [MCP 集成](specs/003-mcp-integration/) - AI 助手协议支持
- [智能问答](specs/004-chat-agent/) - LLM 对话和工具调用
- [向量搜索](specs/005-qdrant-vector-search/) - Qdrant 语义搜索
- [任务管理](specs/006-task-management/) - 异步任务和性能优化
- [示例规格](specs/example-detailed-spec/) - 完整的 PDPI-spec 工作流示例

### 用户文档 (`docs/` 目录)
- [QUICK_START.md](docs/QUICK_START.md) - 部署和配置指南
- [MCP_README.md](docs/MCP_README.md) - MCP 协议集成和使用
- [MCP_INSPECTOR_GUIDE.md](docs/features/MCP_INSPECTOR_GUIDE.md) - MCP 工具测试
- [QDRANT_COLLECTION_SETUP.md](docs/QDRANT_COLLECTION_SETUP.md) - 向量数据库设置

## 🏗️ 架构概览

### 系统分层
```
前端层 (Vue.js) 
    ↓ HTTP/REST API
API 网关层 (Gin Router)
    ↓
业务逻辑层 (Handlers)
    ├─ BookHandler: CRUD 操作
    ├─ SearchHandler: 混合搜索（先召回后精排）
    ├─ ChatHandler: 智能问答
    ├─ MCPServer: MCP 协议
    └─ TaskHandler: 异步任务
    ↓
服务层 (Services)
    ├─ ContentAPI: Calibre 交互
    ├─ QdrantClient: 向量搜索
    ├─ ChatAgent: LLM 对话
    ├─ CacheManager: 文件缓存
    └─ TaskManager: 任务调度
    ↓
数据层 (Storage)
    ├─ Calibre DB (SQLite)
    ├─ Qdrant DB (向量)
    ├─ Chat DB (SQLite)
    └─ File Cache (Disk)
```

详细架构请参考：[ARCHITECTURE.md](docs/ARCHITECTURE.md)

## 🚀 开发工作流

### 1. 环境设置
```bash
# 克隆项目
git clone https://github.com/jianyun8023/calibre-api.git
cd calibre-api

# 安装依赖
go mod download

# 配置环境
cp config.yaml.example config.yaml

# 启动 Qdrant
docker run -d -p 6333:6333 qdrant/qdrant

# 构建运行
make build
./calibre-api
```

### 2. 开发流程
```bash
# 1. 创建功能分支
git checkout -b feature/new-feature

# 2. 开发功能（遵循代码规范）

# 3. 运行测试
go test ./...

# 4. 更新 CHANGELOG.md
# 在 [Unreleased] 部分记录变更

# 5. 提交代码
git commit -m "feat: add new feature"

# 6. 推送并创建 PR
git push origin feature/new-feature
```

### 3. 常见开发任务

#### 添加新的 API 端点
1. 在 `internal/calibre/` 创建或更新 handler
2. 在 `SetupRouter()` 中注册路由
3. 更新 API 文档
4. 在 `CHANGELOG.md` 记录变更

#### 添加新的 MCP Tool
1. 在 `mcp_tools.go` 中添加工具定义
2. 实现工具逻辑（调用现有 Handler）
3. 使用 MCP Inspector 测试
4. 更新 `docs/MCP_README.md`
5. 在 `CHANGELOG.md` 记录变更

**注意**: 只添加只读工具，写操作通过 Web UI 进行

#### 添加新的搜索策略
1. 在 `search_handler.go` 实现搜索逻辑
2. 更新 `search()` 函数
3. 测试搜索质量
4. 在 `CHANGELOG.md` 记录变更

#### 添加新的任务类型
1. 在 `internal/tasks/` 实现任务逻辑
2. 在 `TaskManager` 注册任务类型
3. 在 `task_handler.go` 添加启动入口
4. 在 `CHANGELOG.md` 记录变更

## 📐 代码规范要点

### 命名规范
- 包名: 全小写
- 导出类型: 首字母大写驼峰
- 接口: 以 `-er` 结尾
- 方法: 驼峰命名，动词开头

### 错误处理
- 定义明确的错误类型
- 使用 `fmt.Errorf` 包装错误
- 早返回模式
- 记录或处理所有错误

### 并发安全
- 使用 `sync.RWMutex` 保护共享资源
- 使用 channel 进行 goroutine 通信
- 使用 `context.Context` 控制超时

### API 设计
- 遵循 RESTful 约定
- 统一的响应格式
- 正确使用 HTTP 状态码

详细规范请参考：[DEVELOPMENT_GUIDE.md](docs/DEVELOPMENT_GUIDE.md)

## 🧪 测试

### 单元测试
```bash
go test ./...
```

### 集成测试
```bash
# MCP 测试
./examples/test_mcp_full.sh

# Chat API 测试
./test_chat_api.sh
```

## 🐛 调试

### 启用调试模式
```yaml
# config.yaml
debug: true
```

### 查看日志
```bash
tail -f app.log
grep "ERROR" app.log
```

### MCP Inspector
```bash
./examples/test_mcp_full.sh
open examples/mcp_sse_client.html
```

## 📝 重要约定

### 开发规范
1. **配置管理**: 敏感信息使用环境变量
2. **数据库迁移**: Chat DB 使用 SQL 迁移文件
3. **向量维度**: Embedding 维度与 Qdrant 一致（4096）
4. **缓存清理**: CacheManager 自动清理
5. **MCP 端点**: 使用 SSE 传输
6. **并发安全**: 共享状态必须加锁
7. **变更日志**: 更新 `CHANGELOG.md`
8. **MCP 安全**: 只暴露只读工具

### 性能考虑
1. 使用 `context.Context` 控制超时
2. 大量数据使用流式处理或分页
3. 频繁访问的数据使用缓存
4. Qdrant 批量操作提高性能
5. 避免循环中进行网络请求

### 安全注意
1. 验证所有用户输入
2. 使用参数化查询防止 SQL 注入
3. 限制文件上传大小和类型
4. 日志不记录敏感信息
5. API 密钥使用环境变量

## 📖 版本历史

### v1.2.0 (2024-11-28) - MCP 框架重构
- 迁移到官方 `mcp-go` 框架
- 支持 SSE 和 StreamableHTTP 传输
- 移除危险工具，保留 6 个只读工具
- 优化 `search_books` 使用纯语义搜索
- `get_book` 新增 TOC 目录信息

### v1.1.0 (2024-11-27) - AI 智能增强
- 添加 Chat Agent 智能问答
- 集成 LLM (OpenAI, Ollama)
- Qdrant 向量数据库支持
- TOC 提取性能优化（5-7x 提升）

### v1.0.0 (2024-11-01) - 初始版本
- 基础书籍管理功能
- Calibre Content Server 集成
- 搜索和元数据管理
- Vue.js Web 界面

## 🔄 最近更新 (Unreleased)

### 文档维护
- 重构文档结构，拆分 CLAUDE.md
- 新增 ARCHITECTURE.md（架构设计）
- 新增 DEVELOPMENT_GUIDE.md（开发指南）
- 清理过时的临时文档
- 更新 README.md 文档导航

### 安全性
- 升级所有 Go 依赖，修复 27 个安全漏洞
- 更新 `golang.org/x/*` 和 `google.golang.org/protobuf`

### 代码质量
- 全面代码质量优化（8610 行 Go 代码）
- 新增 9 个错误类型，50+ 行文档注释
- 新增 16 个单元测试和 4 个性能基准测试

---

**当前版本**: 1.2.1 (开发中)  
**最后更新**: 2024-12-08  
**维护者**: jianyun8023  
**项目地址**: https://github.com/jianyun8023/calibre-api
