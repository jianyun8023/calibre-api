# Calibre MCP SSE 服务

Calibre API 现在支持通过 **Server-Sent Events (SSE)** 提供 MCP (Model Context Protocol) 服务，这使得 Web 应用可以轻松地与 MCP 服务器进行实时交互。

## 🌟 特性

- **实时通信**: 基于 SSE 的实时双向通信
- **HTTP API**: 支持标准 HTTP 请求进行 MCP 操作
- **Web 友好**: 可以直接在浏览器中使用，无需特殊客户端
- **多客户端**: 支持多个客户端同时连接
- **自动重连**: 客户端断线后自动重连机制

## 🚀 快速开始

### 1. 启用 SSE MCP 服务

在 `config.yaml` 中启用 MCP 服务：

```yaml
mcp:
  enabled: true                        # 启用 MCP 服务
  server_name: "calibre-mcp-server"    # 服务器名称
  version: "1.1.0"                     # 版本
  base_url: "http://localhost:8080"    # API 基础 URL
  timeout: 30                          # 超时时间（秒）
```

### 2. 启动服务器

```bash
# 启动 HTTP 模式（包含 SSE MCP 服务）
./calibre-api

# 或者明确指定 HTTP 模式
./calibre-api --mcp=false
```

### 3. 访问 SSE MCP 端点

服务器启动后，SSE MCP 服务将在以下端点可用：

- **SSE 连接**: `GET /api/mcp/connect`
- **初始化**: `POST /api/mcp/initialize`
- **工具列表**: `GET /api/mcp/tools/list`
- **工具调用**: `POST /api/mcp/tools/call`

## 📡 API 端点详解

### SSE 连接

```javascript
// 建立 SSE 连接
const eventSource = new EventSource('http://localhost:8080/api/mcp/connect');

eventSource.addEventListener('connected', function(event) {
    const data = JSON.parse(event.data);
    console.log('连接已建立:', data.client_id);
});

eventSource.addEventListener('tool_result', function(event) {
    const result = JSON.parse(event.data);
    console.log('工具执行结果:', result);
});
```

### HTTP API

#### 初始化 MCP 服务

```bash
curl -X POST http://localhost:8080/api/mcp/initialize \
  -H "Content-Type: application/json" \
  -d '{
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {
      "name": "web-client",
      "version": "1.0.0"
    }
  }'
```

#### 获取工具列表

```bash
curl http://localhost:8080/api/mcp/tools/list
```

响应示例：
```json
{
  "tools": [
    {
      "name": "search_books",
      "description": "搜索书籍。可以按标题、作者、ISBN等搜索，支持分页和排序。",
      "inputSchema": {
        "type": "object",
        "properties": {
          "query": {
            "type": "string",
            "description": "搜索查询词，可以是书名、作者名、ISBN等"
          },
          "limit": {
            "type": "integer",
            "description": "返回结果数量限制",
            "default": 10
          }
        },
        "required": ["query"]
      }
    }
  ],
  "count": 6
}
```

#### 调用工具

```bash
curl -X POST http://localhost:8080/api/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "search_books",
    "arguments": {
      "query": "Python",
      "limit": 5
    }
  }'
```

## 🌐 Web 客户端示例

我们提供了一个完整的 Web 客户端示例，位于 `examples/mcp_sse_client.html`。

### 使用方法

1. 启动 Calibre API 服务器
2. 在浏览器中打开 `examples/mcp_sse_client.html`
3. 点击"连接 SSE"按钮
4. 初始化 MCP 服务
5. 获取工具列表
6. 调用工具进行测试

### 功能特性

- **实时连接状态显示**
- **自动重连机制**
- **工具参数预设**
- **结果实时显示**
- **完整的事件日志**

## 🔧 JavaScript 客户端库

### 基础用法

```javascript
class CalibreMCPClient {
    constructor(baseUrl = 'http://localhost:8080') {
        this.baseUrl = baseUrl;
        this.eventSource = null;
        this.clientId = null;
    }

    // 连接 SSE
    connect() {
        return new Promise((resolve, reject) => {
            this.eventSource = new EventSource(`${this.baseUrl}/api/mcp/connect`);
            
            this.eventSource.addEventListener('connected', (event) => {
                const data = JSON.parse(event.data);
                this.clientId = data.client_id;
                resolve(data);
            });

            this.eventSource.onerror = (error) => {
                reject(error);
            };
        });
    }

    // 初始化
    async initialize() {
        const response = await fetch(`${this.baseUrl}/api/mcp/initialize`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                protocolVersion: "2024-11-05",
                capabilities: {},
                clientInfo: { name: "js-client", version: "1.0.0" }
            })
        });
        return response.json();
    }

    // 获取工具列表
    async getTools() {
        const response = await fetch(`${this.baseUrl}/api/mcp/tools/list`);
        return response.json();
    }

    // 调用工具
    async callTool(name, args) {
        const response = await fetch(`${this.baseUrl}/api/mcp/tools/call`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, arguments: args })
        });
        return response.json();
    }
}

// 使用示例
const client = new CalibreMCPClient();

async function example() {
    // 连接
    await client.connect();
    
    // 初始化
    await client.initialize();
    
    // 获取工具
    const tools = await client.getTools();
    console.log('可用工具:', tools);
    
    // 搜索书籍
    const result = await client.callTool('search_books', {
        query: 'Python',
        limit: 5
    });
    console.log('搜索结果:', result);
}
```

## 🛠️ 可用工具

SSE MCP 服务提供以下工具：

### 1. search_books - 搜索书籍
```json
{
  "name": "search_books",
  "arguments": {
    "query": "Python",           // 搜索关键词
    "limit": 10,                 // 返回数量限制
    "offset": 0,                 // 分页偏移
    "sort": "id:desc"           // 排序方式
  }
}
```

### 2. get_book - 获取书籍详情
```json
{
  "name": "get_book",
  "arguments": {
    "id": "123"                  // 书籍ID
  }
}
```

### 3. get_recent_books - 获取最近书籍
```json
{
  "name": "get_recent_books",
  "arguments": {
    "limit": 20                  // 返回数量限制
  }
}
```

### 4. update_book_metadata - 更新书籍元数据
```json
{
  "name": "update_book_metadata",
  "arguments": {
    "id": "123",                 // 书籍ID
    "metadata": {
      "title": "新标题",
      "authors": ["作者1", "作者2"],
      "tags": ["标签1", "标签2"]
    }
  }
}
```

### 5. delete_book - 删除书籍
```json
{
  "name": "delete_book",
  "arguments": {
    "id": "123"                  // 书籍ID
  }
}
```

### 6. search_metadata - 搜索元数据
```json
{
  "name": "search_metadata",
  "arguments": {
    "query": "Python编程",       // 搜索查询
    "source": "douban"          // 元数据源
  }
}
```

## 🔄 事件类型

SSE 连接会发送以下类型的事件：

- **connected**: 连接建立时发送，包含客户端ID
- **ping**: 定期心跳包，保持连接活跃
- **tool_result**: 工具执行结果广播
- **response**: MCP 协议响应消息

## ⚙️ 配置选项

在 `config.yaml` 中可以配置以下 MCP 相关选项：

```yaml
mcp:
  enabled: true                        # 是否启用 MCP 服务
  server_name: "calibre-mcp-server"    # 服务器名称
  version: "1.1.0"                     # 服务器版本
  base_url: "http://localhost:8080"    # API 基础 URL
  timeout: 30                          # 请求超时时间（秒）
```

## 🚀 部署建议

### 开发环境
```bash
# 直接运行
./calibre-api

# 或使用 Docker
docker run -p 8080:8080 -v $(pwd)/config.yaml:/app/config.yaml calibre-api
```

### 生产环境
```bash
# 使用 Docker Compose
version: '3.8'
services:
  calibre-api:
    image: ghcr.io/jianyun8023/calibre-api:latest
    ports:
      - "8080:8080"
    environment:
      - MCP_ENABLED=true
      - MCP_BASE_URL=https://your-domain.com
    volumes:
      - ./config.yaml:/app/config.yaml
      - ./data:/app/data
```

## 🔒 安全考虑

1. **CORS 配置**: 确保正确配置 CORS 策略
2. **身份验证**: 在生产环境中添加适当的身份验证
3. **速率限制**: 考虑添加 API 速率限制
4. **HTTPS**: 在生产环境中使用 HTTPS

## 🐛 故障排除

### 常见问题

1. **连接失败**
   - 检查服务器是否启动
   - 确认 MCP 服务已启用
   - 检查防火墙设置

2. **工具调用失败**
   - 确保已初始化 MCP 服务
   - 检查工具参数格式
   - 查看服务器日志

3. **SSE 连接断开**
   - 检查网络连接
   - 查看浏览器控制台错误
   - 确认服务器稳定性

### 调试模式

启用调试模式获取更多日志信息：

```yaml
debug: true
```

## 📚 更多资源

- [MCP 协议规范](https://modelcontextprotocol.io/)
- [Server-Sent Events 文档](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events)
- [Calibre API 文档](../README.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进 SSE MCP 服务！