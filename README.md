# Calibre-API

基于 MeiliSearch 搭建的 Calibre 书籍管理系统，支持搜索、下载、预览和智能交互。

## ✨ 核心特性

### 📚 书籍管理
- 使用 Calibre Content Server 作为数据来源
- MeiliSearch 增强查询响应速度
- 支持书籍元数据的 CRUD 操作
- 在线元数据获取和补全
- 封面图片和文件下载

### 🤖 AI 智能交互（MCP 支持）
- **MCP (Model Context Protocol) 集成** - 与 AI 助手无缝交互
- **自然语言操作** - 通过对话管理书籍
- **智能推荐** - AI 驱动的书籍推荐
- **双模式部署** - 支持集成模式和独立模式
- **详细参数说明** - 为所有 API 接口提供完整的参数文档

### 🔧 开发特性
- RESTful API 接口
- Docker 容器化部署
- 多平台二进制发布
- 完整的 CI/CD 流程


## 🚀 快速开始

### 运行模式

#### 1. HTTP API 模式（默认）
```bash
# 构建和运行
make build
./calibre-api
```

#### 2. MCP 智能交互模式
```bash
# 使用命令行参数
./calibre-api --mcp

# 使用环境变量
MCP_MODE=true ./calibre-api

# 使用配置文件
# 在 config.yaml 中设置 mcp.enabled: true
./calibre-api
```

### 🤖 AI 助手集成

配置 Claude Desktop 或其他 MCP 客户端：

```json
{
  "mcpServers": {
    "calibre-api": {
      "command": "/path/to/calibre-api",
      "args": ["--mcp"]
    }
  }
}
```

现在您可以通过自然语言与 AI 助手交互：
- *"搜索关于机器学习的书籍"*
- *"帮我更新书籍 ID 123 的元数据"*
- *"推荐几本随机书籍"*
- *"根据 ISBN 获取书籍信息"*

## 📖 API 接口

```text
GET    /api/get/cover/:id            --> 获取书籍封面
GET    /api/get/book/:id             --> 下载书籍文件
GET    /api/read/:id/toc             --> 获取书籍目录（包含元信息、目录和地址）
GET    /api/read/:id/file/*path      --> 读取书籍中的文件
GET    /api/book/:id                 --> 获取书籍信息
GET    /api/search                   --> 搜索书籍
POST   /api/search                   --> 搜索书籍
GET    /api/recently                 --> 最近更新的书籍
GET    /api/random                   --> 随机书籍推荐
GET    /api/publisher                --> 获取出版社列表
GET    /api/metadata/isbn/:isbn      --> 根据 ISBN 获取元数据
GET    /api/metadata/search          --> 搜索在线元数据
POST   /api/book/:id/update          --> 更新书籍元数据
POST   /api/book/:id/delete          --> 删除书籍
POST   /api/index/update             --> 更新搜索索引
POST   /api/index/switch             --> 切换搜索索引
```

## 数据导入

创建索引，更新索引设置，该命令仅第一次使用需要执行。
```shell
## Create index
curl -X "POST" "http://localhost:7700/indexes" \
     -H 'Content-Type: application/json' \
     -d $'{
  "uid": "books"
}'
## Update settings
curl -X "PATCH" "http://localhost:7700/indexes/books/settings" \
     -H 'Content-Type: application/json' \
     -d $'{
  "displayedAttributes": [
    "*"
  ],
  "filterableAttributes": [
    "authors",
    "file_path",
    "id",
    "last_modified",
    "pubdate",
    "publisher",
    "isbn",
    "tags"
  ],
  "searchableAttributes": [
    "title",
    "authors"
  ],
  "sortableAttributes": [
    "authors_sort",
    "id",
    "last_modified",
    "pubdate",
    "publisher"
  ]
}'
```

使用下面命令更新索引
```shell
curl -X "POST" "http://localhost:8080/index/update" -H 'Content-Type: application/json' 
```

## 接口

### MCP 参数说明改进

本项目对 gin-mcp 包进行了增强，为所有 API 接口提供了详细的参数说明。这使得 AI 助手能够更好地理解和使用这些接口。

#### 改进内容

1. **参数结构体定义** - 在 `internal/calibre/schemas.go` 中定义了所有接口的参数结构体
2. **jsonschema 标签** - 为每个参数添加了详细的描述和约束
3. **自动注册** - 在启动时自动为所有接口注册参数模式

#### 示例对比

**改进前：**
```
参数：q (string)
参数：limit (number)
参数：offset (number)
```

**改进后：**
```
参数：q (string, required) - 搜索关键词
参数：limit (number, 1-100, default=20) - 每页结果数量
参数：offset (number, >=0, default=0) - 结果偏移量
参数：filter (string, optional) - 过滤条件
参数：sort (string, optional) - 排序字段
```

#### 使用方法

1. 启动服务器：`./calibre-api`
2. 在 MCP 客户端（如 Cursor）中连接到 `http://localhost:8080/mcp`
3. 所有 API 工具都会包含详细的参数说明

详细文档请参考：[MCP 参数说明改进方案](docs/MCP_SCHEMA_IMPROVEMENT.md)

## 🔨 构建和部署

### 本地构建
```bash
# 构建主程序（包含 MCP 功能）
make build

# 构建独立 MCP 服务器
make build-mcp

# 构建所有版本  
make build-all

# 直接使用 Go 构建
go build -o calibre-api .
go build -o calibre-mcp-server ./cmd/mcp-server
```

### Docker 部署
```bash
# 构建 Docker 镜像
docker build -t calibre-api:latest .

# 运行容器（HTTP 模式）
docker run -d -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  calibre-api:latest

# 运行容器（MCP 模式）
docker run -d \
  -e MCP_MODE=true \
  -v $(pwd)/config.yaml:/app/config.yaml \
  calibre-api:latest
```

### 预构建二进制
从 [Releases](https://github.com/jianyun8023/calibre-api/releases) 页面下载预构建的二进制文件：
- `calibre-api-*` - 主程序（包含 MCP 功能）
- `calibre-mcp-server-*` - 独立 MCP 服务器

## 配置

### 配置文件

配置文件会按优先级从下面查找:

- `/etc/calibre-api/config.yaml`
- `$HOME/.calibre-api`
- `./config.yaml`

配置内容

```yaml
address: :8080
debug: false
staticDir: "/app/static"
tmpDir: ".files"

# Calibre Content Server 配置
content:
  server: https://lib.pve.icu

# MeiliSearch 搜索引擎配置  
search:
  host: http://127.0.0.1:7700
  apikey: ""
  index: books

# 元数据服务配置
metadata:
  doubanurl: https://api.douban.com

# MCP 服务器配置
mcp:
  enabled: false                        # 是否默认启用 MCP 模式
  server_name: "calibre-mcp-server"     # MCP 服务器名称
  version: "1.1.0"                      # MCP 服务器版本  
  base_url: "http://localhost:8080"     # API 基础 URL
  timeout: 30                           # API 请求超时时间（秒）
```

### 环境变量

环境变量优先于配置文件，可以使用环境变量覆盖配置文件中的参数

```text
# 基础配置
CALIBRE_ADDRESS=:8080
CALIBRE_DEBUG=false
CALIBRE_STATICDIR=/app/static
CALIBRE_TMP_DIR=.files

# Calibre Content Server
CALIBRE_CONTENT_SERVER=https://your-calibre-server.com

# MeiliSearch 配置
CALIBRE_SEARCH_HOST=http://localhost:7700
CALIBRE_SEARCH_APIKEY=your-api-key
CALIBRE_SEARCH_INDEX=books

# 元数据服务
CALIBRE_METADATA_DOUBANURL=https://api.douban.com

# MCP 配置
CALIBRE_MCP_ENABLED=false
CALIBRE_MCP_BASE_URL=http://localhost:8080
MCP_MODE=true                    # 快速启用 MCP 模式
CALIBRE_MCP_MODE=true           # 快速启用 MCP 模式
```

## 适配阅读书源

支持添加书源的APP可以使用下面配置，将本服务引入

```json
{
  "bookSourceUrl": "http://localhost:8080",
  "bookSourceType": 0,
  "bookSourceName": "calibre书库",
  "bookSourceGroup": "calibre",
  "bookSourceComment": "",
  "loginUrl": "",
  "loginUi": "",
  "loginCheckJs": "",
  "concurrentRate": "",
  "header": "",
  "bookUrlPattern": "",
  "searchUrl": "search?q={{key}}&sort=id:desc",
  "exploreUrl": "",
  "enabled": true,
  "enabledExplore": false,
  "weight": 0,
  "customOrder": 0,
  "lastUpdateTime": 1661322926750,
  "ruleSearch": {
    "bookList": "$.hits",
    "name": "$.title",
    "author": "$.authors",
    "intro": "$.comments",
    "coverUrl": "/get/cover/{{$.id}}.jpg",
    "bookUrl": "/book/{{$.id}}"
  },
  "ruleExplore": {},
  "ruleBookInfo": {
    "name": "$.title",
    "author": "$.authors",
    "intro": "$.comments",
    "coverUrl": "/get/cover/{{$.id}}.jpg",
    "tocUrl": "/read/{{$.id}}/toc"
  },
  "ruleToc": {
    "chapterList": "$.points",
    "chapterName": "$.text",
    "chapterUrl": "$.content.src"
  },
  "ruleContent": {
    "content": "//body"
  }
}
```

## 📚 文档

- **[快速开始指南](docs/QUICK_START.md)** - 详细的设置和部署指南
- **[MCP 使用文档](docs/MCP_README.md)** - AI 助手集成的完整说明
- **[API 参考](docs/API.md)** - RESTful API 接口文档

## 🎯 使用场景

### 1. 个人书库管理
- 通过 Web 界面管理和搜索书籍
- 下载和预览书籍内容
- 元数据编辑和补全

### 2. AI 智能助手
- 自然语言搜索书籍
- AI 驱动的阅读推荐
- 智能元数据处理

### 3. 书源集成
- 作为阅读 APP 的书源
- API 接口供第三方应用调用

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License