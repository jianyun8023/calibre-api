---
status: complete
created: '2025-12-08'
tags: []
priority: medium
created_at: '2025-12-08T13:03:56.634Z'
updated_at: '2025-12-08T13:08:00.000Z'
completed_at: '2025-12-08T13:08:00.000Z'
completed: '2025-12-08'
transitions:
  - status: complete
    at: '2025-12-08T13:08:00.000Z'
depends_on:
  - 001-book-management
  - 002-search-functionality
---

# MCP 协议集成

> **Status**: ✅ Complete · **Priority**: Medium · **Created**: 2025-12-08

## Overview

**问题**: AI 助手（Claude、ChatGPT）无法直接访问 Calibre 书库，需要用户手动复制粘贴信息。

**解决方案**: 实现 MCP (Model Context Protocol) 服务器，暴露 Tools、Resources、Prompts 供 AI 助手调用。

**为什么现在**: MCP 是 Anthropic 推出的 AI-工具连接标准，Cursor/Claude Desktop 已原生支持。

## Design

### MCP 核心概念

**三大组件**:
1. **Tools**: AI 可调用的函数（如搜索书籍、获取详情）
2. **Resources**: AI 可读取的资源（如单本书籍、搜索结果）
3. **Prompts**: 预定义的提示模板（如书籍推荐、搜索指导）

### 传输层选择

**SSE vs stdio**:
- stdio: MCP 标准传输，适合本地进程
- SSE (Server-Sent Events): HTTP 长连接，适合 Web 集成

**权衡**: 选择 SSE
- 复用现有 HTTP 服务器（Gin）
- 支持 Web 前端直接连接（调试方便）
- 可通过 MCP Inspector 测试
- 代价：相比 stdio 多一层 HTTP 开销（可忽略）

### 注册的 Tools

**书籍操作** (7 个):
- `search_books`: 搜索书籍（支持混合策略）
- `get_book`: 获取单本详情
- `random_books`: 随机推荐
- `update_book_metadata`: 更新元数据
- `delete_book`: 删除书籍
- `get_isbn_metadata`: ISBN 查询
- `search_metadata`: 在线元数据搜索

**设计原则**:
- 每个 Tool 对应一个明确的业务动作
- 参数使用 JSON Schema 严格校验
- 返回结果包含完整上下文（AI 可直接使用）

### 注册的 Resources

**动态资源**:
- `book://{id}`: 单本书籍 JSON（包含元数据、目录、文件列表）
- `search_results://{query}`: 搜索结果缓存

**为何需要 Resources**?
- Tools 用于"执行动作"，Resources 用于"读取数据"
- AI 可以先列举资源再按需读取（减少 Token 消耗）

### 注册的 Prompts

**模板示例**:
```
book_search:
  "请帮我在书库中搜索关于 {topic} 的书籍，重点关注 {focus_area}。"

book_recommendation:
  "基于用户喜好 {preferences}，推荐 5 本书籍，并说明推荐理由。"
```

**作用**: 降低用户提示门槛，提供最佳实践模板。

## Plan

- [x] 集成 mcp-go SDK
- [x] 实现 SSE 传输层（/mcp/sse, /mcp/message）
- [x] 注册 7 个核心 Tools
- [x] 实现 Tool 执行器（映射到 Handler）
- [x] 注册 Resources（book, search_results）
- [x] 注册 Prompts 模板
- [x] 添加 MCP Inspector 测试页面
- [x] 更新 Cursor 配置文档

## Test

- [x] MCP Inspector 成功连接
- [x] Tools 列表正确显示 7 个工具
- [x] search_books 工具返回正确结果
- [x] get_book 工具返回完整书籍信息
- [x] Resources 可通过 URI 访问
- [x] Prompts 可正确渲染参数
- [x] Claude Desktop 集成测试通过

## Notes

**端点配置**:
```
SSE Endpoint: http://localhost:8080/mcp/sse
Message Endpoint: http://localhost:8080/mcp/message
```

**Cursor 配置**:
```json
{
  "mcpServers": {
    "calibre": {
      "transport": {
        "type": "sse",
        "url": "http://localhost:8080/mcp/sse"
      }
    }
  }
}
```

**安全考虑**:
- 删除/更新操作需要额外确认（Tool Schema 标注）
- 暂不支持身份验证（本地部署场景）
- 未来可添加 API Key 认证

**性能指标**:
- Tool 调用延迟: ~100ms（取决于底层 Handler）
- SSE 连接保持: 长期稳定（WebSocket 级别）
- 并发支持: 单个连接串行，多连接并发

**已知限制**:
- SSE 不支持双向通信（暂无需求）
- Tool 执行超时: 30s（硬编码）
- 不支持流式返回（MCP 1.0 协议限制）

**相关文档**:
- `docs/MCP_README.md`: MCP 集成完整指南
- `docs/features/MCP_V1.2.0_SUMMARY.md`: 版本总结
- `examples/mcp_sse_client.html`: 测试客户端

**依赖项**:
- `001-book-management`: 书籍操作 API
- `002-search-functionality`: 搜索能力
- `github.com/metoro-io/mcp-go`: MCP SDK
