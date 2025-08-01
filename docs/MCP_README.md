# Calibre MCP Server

本项目为 Calibre API 实现了 MCP (Model Context Protocol) 功能，让 AI 助手能够直接与您的 Calibre 书库交互。

## 功能特性

MCP 服务器提供以下工具：

### 书籍搜索和浏览
- `search_books` - 搜索书籍（按标题、作者、ISBN等）
- `get_book` - 获取书籍详细信息
- `get_recent_books` - 获取最近更新的书籍
- `get_random_books` - 获取随机书籍推荐

### 书籍管理
- `update_book_metadata` - 更新书籍元数据
- `delete_book` - 删除书籍
- `get_publishers` - 获取所有出版社列表

### 元数据服务
- `search_metadata` - 在线搜索书籍元数据（豆瓣等）
- `get_metadata_by_isbn` - 根据ISBN获取元数据

### 系统管理
- `update_search_index` - 更新搜索索引

## 安装和配置

MCP 功能已集成到主 Calibre API 服务器中，支持两种运行方式：

### 方式一: 集成模式（推荐）

```bash
# 构建主服务器（包含 MCP 功能）
make build

# 或者直接使用 go build
go build -o calibre-api .
```

### 方式二: 独立模式

```bash
# 构建独立 MCP 服务器
make build-mcp

# 或者直接使用 go build
go build -o calibre-mcp-server ./cmd/mcp-server
```

### 2. 配置文件

确保您的 `config.yaml` 文件包含必要的配置：

```yaml
address: :8080
debug: false
staticDir: "/app/static"
tmpDir: ".files"

# Calibre Content Server 配置
content:
  server: https://your-calibre-server.com

# MeiliSearch 配置
search:
  host: http://localhost:7700
  apikey: ""
  index: books

# 元数据服务配置
metadata:
  doubanurl: https://api.douban.com

# MCP 配置（集成模式专用）
mcp:
  enabled: false                        # 是否默认启用 MCP 模式
  server_name: "calibre-mcp-server"     # MCP 服务器名称
  version: "1.1.0"                      # MCP 服务器版本
  base_url: "http://localhost:8080"     # API 基础 URL
  timeout: 30                           # API 请求超时时间（秒）
```

### 3. 环境变量

您也可以使用环境变量进行配置：

```bash
export CALIBRE_CONTENT_SERVER="https://your-calibre-server.com"
export CALIBRE_SEARCH_HOST="http://localhost:7700"
export CALIBRE_SEARCH_INDEX="books"
export CALIBRE_METADATA_DOUBANURL="https://api.douban.com"
```

## 使用方法

### 启动 MCP 服务器

#### 集成模式（推荐）

```bash
# 方法1: 使用命令行参数
./calibre-api --mcp

# 方法2: 使用环境变量  
MCP_MODE=true ./calibre-api

# 方法3: 在配置文件中设置 mcp.enabled: true
./calibre-api
```

#### 独立模式

```bash
# 直接运行独立服务器
./calibre-mcp-server

# 或者使用配置文件
./calibre-mcp-server --config /path/to/config.yaml
```

### 在 AI 助手中使用

#### Claude Desktop 配置

**集成模式配置（推荐）：**

```json
{
  "mcpServers": {
    "calibre-api": {
      "command": "/path/to/calibre-api",
      "args": ["--mcp"],
      "env": {
        "CALIBRE_MCP_BASE_URL": "http://localhost:8080"
      }
    }
  }
}
```

**独立模式配置：**

```json
{
  "mcpServers": {
    "calibre-api": {
      "command": "/path/to/calibre-mcp-server",
      "args": [],
      "env": {
        "CALIBRE_MCP_BASE_URL": "http://localhost:8080"
      }
    }
  }
}
```

#### 其他 MCP 客户端

MCP 服务器遵循标准的 MCP 协议，可以与任何支持 MCP 的 AI 助手或客户端集成。

## 工具详解

### search_books

搜索书籍库中的书籍。

**参数：**
- `query` (必需): 搜索关键词
- `limit` (可选): 返回结果数量，默认10
- `offset` (可选): 分页偏移量，默认0
- `sort` (可选): 排序方式，如 "id:desc", "title:asc"

**示例：**
```json
{
  "query": "机器学习",
  "limit": 5,
  "sort": "pubdate:desc"
}
```

### get_book

获取指定书籍的详细信息。

**参数：**
- `id` (必需): 书籍ID

### update_book_metadata

更新书籍的元数据信息。

**参数：**
- `id` (必需): 书籍ID
- `title` (可选): 书籍标题
- `authors` (可选): 作者列表
- `publisher` (可选): 出版社
- `isbn` (可选): ISBN号码
- `comments` (可选): 书籍简介
- `tags` (可选): 标签列表
- `rating` (可选): 评分（0-10分）
- `pubdate` (可选): 出版日期（ISO 8601格式）

### search_metadata

在线搜索书籍元数据信息。

**参数：**
- `query` (必需): 搜索关键词

### get_metadata_by_isbn

根据ISBN获取书籍元数据。

**参数：**
- `isbn` (必需): ISBN号码

## 常见用例

### 1. 书籍推荐
"推荐一些关于Python编程的书籍"

AI助手会使用 `search_books` 工具搜索相关书籍并提供推荐。

### 2. 书库管理
"帮我整理一下重复的书籍"

AI助手可以搜索书库，识别重复的书籍，并帮助您进行管理。

### 3. 元数据补全
"这本书的ISBN是9787111544937，帮我补全它的元数据"

AI助手会使用 `get_metadata_by_isbn` 获取详细信息，然后使用 `update_book_metadata` 更新书籍信息。

### 4. 阅读计划
"根据我的阅读历史，推荐下一本要读的书"

AI助手可以分析您的书库和阅读模式，提供个性化推荐。

## 故障排除

### 常见问题

1. **连接错误**
   - 检查 Calibre Content Server 是否正常运行
   - 确认配置文件中的服务器地址正确

2. **搜索不工作**
   - 确保 MeiliSearch 服务正常运行
   - 检查搜索索引是否已创建和更新

3. **元数据服务异常**
   - 检查网络连接
   - 确认元数据服务URL是否可访问

### 调试模式

启用调试模式以获取更多日志信息：

```yaml
debug: true
```

或者使用环境变量：

```bash
export CALIBRE_DEBUG=true
```

## 开发和扩展

### 添加新工具

1. 在 `internal/mcp/tools.go` 中添加新的工具定义
2. 实现对应的处理函数
3. 在 `GetTools()` 方法中注册新工具
4. 在 `CallTool()` 方法中添加调用逻辑

### 自定义配置

您可以扩展配置结构以支持更多自定义选项。

## 许可证

本项目采用与主项目相同的许可证。

## 支持

如果您遇到问题或有功能建议，请在项目的 GitHub 仓库中创建 issue。