# Calibre MCP Server 快速开始

本指南将帮助您快速设置和运行 Calibre MCP Server。现在支持两种部署方式：

1. **集成模式** - MCP 功能集成在主 Calibre API 服务器中（推荐）
2. **独立模式** - 运行独立的 MCP 服务器

## 前置条件

1. **Go 语言环境** (Go 1.19+)
2. **Calibre Content Server** 正在运行
3. **MeiliSearch** 服务器正在运行
4. **支持 MCP 的 AI 客户端** (如 Claude Desktop)

## 方式一: 集成模式（推荐）

### 步骤 1: 构建主服务器

```bash
# 克隆项目（如果还没有）
git clone <your-repo-url>
cd calibre-api

# 构建主服务器（包含 MCP 功能）
make build

# 或者使用 go build
go build -o calibre-api .
```

## 步骤 2: 准备配置文件

创建或修改 `config.yaml` 文件：

```yaml
# config.yaml
address: :8080
debug: true
staticDir: "./static"
tmpDir: "./.files"

content:
  server: "http://localhost:8083"  # 您的 Calibre Content Server 地址

search:
  host: "http://localhost:7700"    # MeiliSearch 地址
  apikey: ""                       # MeiliSearch API Key (如需要)
  index: "books"

metadata:
  doubanurl: "https://api.douban.com"
```

## 步骤 3: 启动 Calibre Content Server

确保您的 Calibre Content Server 正在运行：

```bash
# 启动 Calibre Content Server（示例）
calibre-server --port=8083 /path/to/your/calibre/library
```

## 步骤 4: 启动 MeiliSearch

```bash
# 使用 Docker 启动 MeiliSearch
docker run -it --rm \
    -p 7700:7700 \
    -v $(pwd)/meili_data:/meili_data \
    getmeili/meilisearch:v1.5
```

## 步骤 5: 初始化搜索索引

在 MCP 服务器运行之前，需要创建和配置搜索索引：

```bash
# 启动主 Calibre API 服务器
./calibre-api

# 更新搜索索引
curl -X POST "http://localhost:8080/api/index/update"
```

## 步骤 6A: 运行集成模式的 MCP 服务器

```bash
# 方法1: 使用命令行参数
./calibre-api --mcp

# 方法2: 使用环境变量  
MCP_MODE=true ./calibre-api

# 方法3: 在配置文件中设置 mcp.enabled: true
./calibre-api
```

## 方式二: 独立模式

### 构建独立 MCP 服务器

```bash
# 构建独立 MCP 服务器
make build-mcp

# 或者使用 go build
go build -o dist/calibre-mcp-server ./cmd/mcp-server
```

### 运行独立 MCP 服务器

```bash
# 设置环境变量（可选）
export CALIBRE_MCP_BASE_URL="http://localhost:8080"

# 运行独立 MCP 服务器
./dist/calibre-mcp-server
```

## 步骤 7: 配置 AI 客户端

### Claude Desktop 配置

编辑 Claude Desktop 配置文件：

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%/Claude/claude_desktop_config.json`

#### 集成模式配置（推荐）

```json
{
  "mcpServers": {
    "calibre-api": {
      "command": "/path/to/your/calibre-api",
      "args": ["--mcp"],
      "env": {
        "CALIBRE_MCP_BASE_URL": "http://localhost:8080"
      }
    }
  }
}
```

#### 独立模式配置

```json
{
  "mcpServers": {
    "calibre-api": {
      "command": "/path/to/your/calibre-mcp-server",
      "args": [],
      "env": {
        "CALIBRE_MCP_BASE_URL": "http://localhost:8080"
      }
    }
  }
}
```

### 其他 MCP 客户端

对于其他支持 MCP 的客户端，请参考其文档配置 MCP 服务器。

## 测试

重启 Claude Desktop 后，您应该能看到 Calibre MCP 工具可用。尝试以下命令：

- "搜索关于 Python 的书籍"
- "显示书籍 ID 123 的详细信息"
- "获取最近添加的 5 本书"
- "随机推荐几本书"

## 故障排除

### 1. MCP 服务器无法启动

- 检查配置文件路径和格式
- 确保 Calibre Content Server 和 MeiliSearch 正在运行
- 查看服务器日志获取详细错误信息

### 2. AI 客户端无法连接

- 检查 MCP 客户端配置文件路径是否正确
- 确认 MCP 服务器可执行文件路径正确
- 重启 AI 客户端应用

### 3. 搜索功能不工作

```bash
# 检查 MeiliSearch 状态
curl "http://localhost:7700/health"

# 检查搜索索引
curl "http://localhost:7700/indexes"

# 重建搜索索引
curl -X POST "http://localhost:8080/api/index/update"
```

### 4. API 调用失败

- 确认 `CALIBRE_MCP_BASE_URL` 环境变量设置正确
- 检查 Calibre API 服务器是否正在运行
- 测试 API 端点是否可访问：

```bash
curl "http://localhost:8080/api/search?q=test"
```

## 环境变量参考

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `CALIBRE_MCP_BASE_URL` | Calibre API 服务器地址 | `http://localhost:8080` |
| `CALIBRE_DEBUG` | 启用调试模式 | `false` |
| `CALIBRE_CONTENT_SERVER` | Calibre Content Server 地址 | 从配置文件读取 |
| `CALIBRE_SEARCH_HOST` | MeiliSearch 服务器地址 | 从配置文件读取 |

## 下一步

- 阅读完整的 [MCP 使用文档](./MCP_README.md)
- 了解如何自定义和扩展 MCP 工具
- 集成到您的工作流程中

## 获取帮助

如果您遇到问题：

1. 检查所有服务是否正在运行
2. 查看服务器日志
3. 参考故障排除部分
4. 在项目仓库中创建 issue