# MCP Inspector 使用指南

> **目标**: 使用 MCP Inspector 测试 Calibre API 的 MCP 功能

---

## 📋 前置条件

### 1. 确保服务器已启动

```bash
cd /path/to/calibre-api
./calibre-api
```

**验证日志输出**:
```
[INFO] Creating MCP Server: name=calibre-mcp-server, version=1.2.0, transport=sse
[INFO] Mounting MCP SSE endpoints: /mcp/sse, /mcp/message
[INFO] MCP Server enabled: transport=sse
[INFO] route: GET /mcp/sse
[INFO] route: POST /mcp/message
```

### 2. 验证 CORS 配置

检查 `main.go` 中是否包含 CORS 中间件：

```go
import "github.com/gin-contrib/cors"

r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"*"},
    AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
    ExposeHeaders:    []string{"Content-Length", "Content-Type"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}))
```

### 3. 测试端点可访问性

```bash
# 测试 SSE 端点
curl -N http://localhost:8080/mcp/sse

# 应该看到 SSE 连接建立
```

---

## 🚀 使用 MCP Inspector

### 方式一：在线版本（推荐）

1. **访问 MCP Inspector**
   
   打开浏览器访问: https://inspector.modelcontextprotocol.io/

2. **配置连接**
   
   在 "Server URL" 输入框中填入:
   ```
   http://localhost:8080/mcp/sse
   ```

3. **点击 Connect**
   
   如果配置正确，应该看到连接成功提示

4. **查看服务器信息**
   
   连接成功后，Inspector 会显示:
   - Server Name: `calibre-mcp-server`
   - Version: `1.2.0`
   - Protocol Version: `2024-11-05`

### 方式二：本地部署

```bash
# 克隆 MCP Inspector
git clone https://github.com/modelcontextprotocol/inspector.git
cd inspector

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 访问 http://localhost:5173
```

---

## 🔍 测试功能

### 1. 查看服务器能力

连接成功后，在 Inspector 界面查看:

- **Tools**: 当前注册的工具列表（目前为占位符）
- **Resources**: 可用资源列表（待实现）
- **Prompts**: 提示模板列表（待实现）

### 2. 测试初始化

在 Inspector 的 "Initialize" 标签页：

```json
{
  "protocolVersion": "2024-11-05",
  "capabilities": {
    "tools": {},
    "resources": {},
    "prompts": {}
  },
  "clientInfo": {
    "name": "mcp-inspector",
    "version": "1.0.0"
  }
}
```

点击 "Send" 发送初始化请求。

### 3. 调用工具（待工具实现后）

在 "Tools" 标签页：

1. 选择一个工具（如 `search_books`）
2. 填写参数（如 `{"query": "golang", "limit": 10}`）
3. 点击 "Call Tool"
4. 查看返回结果

---

## 🐛 常见问题

### 问题 1: CORS 错误

**错误信息**:
```
Access to fetch at 'http://localhost:8080/mcp/sse' from origin 'https://inspector.modelcontextprotocol.io' 
has been blocked by CORS policy: No 'Access-Control-Allow-Origin' header is present on the requested resource.
```

**解决方案**:
1. 确保 `main.go` 中已添加 CORS 中间件
2. 重启服务器
3. 清除浏览器缓存后重试

### 问题 2: 连接超时

**可能原因**:
- 服务器未启动
- 端口被占用
- 防火墙阻止连接

**解决方案**:
```bash
# 检查服务器是否运行
ps aux | grep calibre-api

# 检查端口是否监听
lsof -i :8080

# 测试端点
curl http://localhost:8080/ping
```

### 问题 3: SSE 连接断开

**错误信息**:
```
SSE connection closed unexpectedly
```

**解决方案**:
1. 检查服务器日志是否有错误
2. 确认 MCP 配置正确:
   ```yaml
   mcp:
     enabled: true
     transport: "sse"
     sse_endpoint: "/mcp/sse"
     message_endpoint: "/mcp/message"
   ```
3. 重启服务器

### 问题 4: 工具列表为空

**说明**: 这是正常的！当前阶段（v1.2.0-rc1）只完成了基础框架，工具注册将在后续版本实现。

**验证方法**:
- 连接成功即表示框架工作正常
- 可以看到服务器信息
- 工具列表为空是预期行为

---

## 📊 测试检查清单

使用以下清单验证 MCP 功能：

### 基础连接
- [ ] Inspector 可以连接到 `/mcp/sse`
- [ ] 无 CORS 错误
- [ ] 显示服务器名称和版本
- [ ] SSE 连接保持稳定

### 协议功能
- [ ] Initialize 请求成功
- [ ] 获取服务器能力列表
- [ ] Ping/Pong 保持连接

### 工具功能（待实现）
- [ ] 工具列表可见
- [ ] 工具调用成功
- [ ] 返回结果正确

### 资源功能（待实现）
- [ ] 资源列表可见
- [ ] 资源读取成功

### 提示功能（待实现）
- [ ] 提示列表可见
- [ ] 提示渲染成功

---

## 🎯 预期行为

### v1.2.0-rc1（当前版本）

**✅ 应该工作**:
- SSE 连接建立
- 服务器信息显示
- Initialize 请求响应
- 无 CORS 错误

**⏳ 待实现**:
- 工具注册和调用
- 资源管理
- 提示模板

### v1.2.0（计划中）

将实现完整的工具、资源和提示功能。

---

## 📝 调试技巧

### 1. 启用 Debug 模式

```yaml
# config.yaml
debug: true
```

重启服务器后，会输出详细日志。

### 2. 使用 curl 测试

```bash
# 测试 SSE 连接
curl -N -H "Accept: text/event-stream" http://localhost:8080/mcp/sse

# 测试 POST 消息
curl -X POST http://localhost:8080/mcp/message \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
```

### 3. 浏览器开发者工具

打开浏览器开发者工具（F12）：
- **Network** 标签: 查看 SSE 连接状态
- **Console** 标签: 查看 JavaScript 错误
- **Application** 标签: 检查 CORS 策略

### 4. 查看服务器日志

```bash
# 实时查看日志
tail -f app.log

# 过滤 MCP 相关日志
grep "MCP" app.log
```

---

## 🔗 相关资源

- [MCP Inspector 官方仓库](https://github.com/modelcontextprotocol/inspector)
- [MCP 协议规范](https://spec.modelcontextprotocol.io/)
- [mcp-go 文档](https://github.com/mark3labs/mcp-go)
- [项目迁移记录](./MCP_MIGRATION_V1.2.0.md)
- [CHANGELOG](../../CHANGELOG.md)

---

## 💡 提示

1. **首次使用**: 建议先使用在线版 Inspector，无需安装
2. **本地开发**: 如需修改 Inspector 代码，使用本地部署版本
3. **生产环境**: CORS 配置应限制具体域名，不要使用 `*`
4. **性能监控**: 可以使用 Inspector 观察 SSE 消息流量

---

**文档版本**: 1.0.0  
**最后更新**: 2024-11-28  
**维护者**: jianyun8023

