# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
- **配置增强**:
  - `mcp.transport`: 传输模式选择 (`sse` 或 `http`)
  - `mcp.sse_endpoint`: SSE 端点路径配置
  - `mcp.message_endpoint`: 消息端点路径配置
  - `mcp.version`: 更新为 `1.2.0`

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

#### Fixed
- MCP 协议标准化，遵循官方规范
- 更好的传输层抽象和扩展性

#### Technical Details
- 创建 `internal/calibre/mcp_server.go` 核心模块
- 实现 `MCPServer` 结构体封装 mcp-go 服务器
- 支持 SSE 和 StreamableHTTP 两种传输实现
- 预留工具、资源、提示注册接口框架
- 保持所有现有 HTTP API 端点向后兼容

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

