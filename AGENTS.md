# AI Agent Instructions

## Project: calibre-api

Calibre 电子书管理系统的 AI 原生升级版，提供 RESTful + MCP 双协议 API。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.24.4, Gin 1.11 |
| 前端 | Next.js 16.0, React 19.2, Shadcn/UI ([web-next/](web-next/README.md)) |
| 数据源 | Calibre Content Server (HTTP API) |
| 搜索 | MeiliSearch 0.36 (全文+语义搜索) |
| Embedding | Ollama/SiliconFlow (向量化服务) |
| 协议 | MCP v1.2.0 (SSE, mark3labs/mcp-go 0.43) |

## 核心文档

| 文档 | 用途 |
|------|------|
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | 系统架构 |
| [DEVELOPMENT_GUIDE.md](docs/DEVELOPMENT_GUIDE.md) | 代码规范 |
| [CODE_STRUCTURE.md](docs/CODE_STRUCTURE.md) | 代码结构 |
| [API_DOCUMENTATION.md](docs/API_DOCUMENTATION.md) | REST API |
| [MCP_README.md](docs/MCP_README.md) | MCP 协议 |
| [QUICK_START.md](docs/QUICK_START.md) | 部署指南 |
| [CHANGELOG.md](CHANGELOG.md) | 变更历史 |

## 架构分层

```
Next.js → Gin Router → Handlers → Services → Data Sources
                         ↓           ↓          ↓
                    Book/Search   MeiliSearch  Calibre Server
                    MCP           Embedding    FileCache
                    Task          Cache
```

## 快速开始

```bash
# 后端
go mod download && make build && ./calibre-api

# 前端
cd web-next && pnpm install && pnpm dev

# MeiliSearch
docker run -d -p 7700:7700 getmeili/meilisearch:latest
```

## 开发任务

### 添加 API 端点
1. `internal/calibre/` 创建 handler
2. `SetupRouter()` 注册路由
3. 更新 `CHANGELOG.md`

### 添加 MCP Tool
1. `mcp_tools.go` 添加定义
2. 实现逻辑（只读工具）
3. 更新 `docs/MCP_README.md`

## 代码规范

- **命名**: 包名小写，导出类型 PascalCase，接口 `-er` 后缀
- **错误**: `fmt.Errorf` 包装，早返回
- **并发**: `sync.RWMutex` 保护共享状态，`context.Context` 控制超时
- **API**: RESTful 约定，统一响应格式

## 关键约定

- 敏感信息用环境变量
- Embedding 维度：默认 4096（由提供商决定）
- MCP 只暴露只读工具
- 所有变更记录到 `CHANGELOG.md`
- 搜索引擎：统一使用 MeiliSearch（已移除 Qdrant）

---

**版本**: 1.2.1 | **维护者**: jianyun8023
