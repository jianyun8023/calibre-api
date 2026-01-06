# Calibre-API 项目接口和功能文档

## 目录

- [项目概述](#项目概述)
- [核心功能](#核心功能)
- [架构设计](#架构设计)
- [REST API 接口](#rest-api-接口)
- [MCP 增强工具](#mcp-增强工具)
- [任务管理](#任务管理)
- [向量搜索](#向量搜索)
- [配置说明](#配置说明)
- [部署指南](#部署指南)

---

## 项目概述

Calibre-API 是一个基于 Go 语言开发的 Calibre 书籍管理系统 API 服务，提供了完整的书籍管理、搜索、元数据管理和 AI 智能交互功能。

### 技术栈

- **后端框架**: Gin (Go Web Framework)
- **搜索引擎**: Qdrant (统一的关键词搜索 + 语义搜索)
- **向量数据库**: Qdrant
- **Embedding 服务**: 
  - Ollama (本地部署)
  - SiliconFlow (云服务)
- **AI 集成**: MCP (Model Context Protocol)
- **前端**: Vue.js + Element Plus

### 主要特性

1. **书籍管理**: CRUD 操作、元数据管理、封面管理
2. **智能搜索**: 关键词搜索 + 语义搜索
3. **AI 交互**: MCP 协议支持，与 AI 助手无缝集成
4. **在线元数据**: 支持从豆瓣等平台获取书籍信息
5. **任务管理**: 异步任务处理，支持索引同步和数据迁移
6. **阅读器集成**: 提供书源配置，可集成到第三方阅读 APP

---

## 核心功能

### 1. 书籍管理

- **书籍信息查询**: 获取书籍详细信息，包括标题、作者、出版社、ISBN 等
- **书籍搜索**: 支持关键词搜索和语义搜索
- **元数据编辑**: 更新书籍的标题、作者、标签、评分等信息
- **书籍删除**: 从系统中删除书籍及相关数据
- **封面管理**: 获取和代理书籍封面图片
- **文件下载**: 支持书籍文件下载
- **在线阅读**: 提供书籍目录和内容预览

### 2. 搜索功能

#### 统一搜索引擎 (Qdrant)
Qdrant 提供完整的搜索能力，包括：

**关键词搜索**:
- 支持标题、作者、出版社、ISBN、标签等字段搜索
- 灵活的过滤和分页功能
- 高性能的标量字段检索

**语义搜索**:
- 基于 Embedding 的向量相似度搜索
- 支持自然语言查询
- 智能推荐相关书籍

**混合搜索**:
- 结合向量搜索和元数据过滤
- 例如：搜索"机器学习入门书" + 过滤评分 > 4.0

### 3. 元数据服务

- **在线查询**: 通过 ISBN 查询书籍信息
- **元数据搜索**: 搜索在线元数据源
- **自动补全**: 智能补全书籍信息
- **封面代理**: 代理获取在线封面图片

### 4. AI 智能交互 (MCP)

- **自然语言操作**: 通过对话管理书籍
- **增强工具**: 提供 7 个增强工具
- **资源管理**: 统一的资源访问接口
- **提示词管理**: 预定义的交互提示词

### 5. 任务管理

- **向量同步**: 生成和同步书籍 Embedding 到 Qdrant
- **任务监控**: 实时查看任务状态和进度

---

## 架构设计

### 系统架构

```
┌─────────────────┐
│   前端应用      │ (Vue.js + Element Plus)
└────────┬────────┘
         │ HTTP/REST
┌────────▼────────┐
│   Gin API       │ (Go Web Server)
│   - REST API    │
│   - MCP Server  │
└────────┬────────┘
         │
    ┌────┴─────┬──────────┬─────────────┐
    │          │          │             │
┌───▼───┐  ┌──▼──┐  ┌────▼─────┐
│Calibre│  │Qdrant│ │Embedding │
│Content│  │      │ │Provider  │
│Server │  │ 统一  │ │(Ollama/  │
│       │  │搜索  │ │Silicon   │
│       │  │引擎  │ │Flow)     │
└───────┘  └──────┘  └──────────┘
            │ │ │
            │ │ └─→ 关键词搜索
            │ └───→ 语义搜索
            └─────→ 混合搜索
```

### 数据流

1. **书籍数据获取**: Calibre Content Server → API → Qdrant (统一存储)
2. **关键词搜索**: 用户查询 → API → Qdrant 标量过滤 → 结果返回
3. **语义搜索**: 用户查询 → Embedding 向量化 → Qdrant 向量搜索 → 结果返回
4. **混合搜索**: 用户查询 → 向量化 + 过滤条件 → Qdrant 混合检索 → 结果返回
5. **元数据更新**: 用户提交 → API → Calibre Content Server → 更新 Qdrant 索引
6. **AI 交互**: AI 助手 → MCP → API → 执行操作 → 返回结果

---

## REST API 接口

### 基础信息

- **Base URL**: `http://localhost:8080`
- **Content-Type**: `application/json`
- **Response Format**: JSON

### 接口列表

#### 1. 书籍查询

##### 1.1 获取书籍信息

```http
GET /api/book/:id
```

**路径参数**:
- `id` (string, required): 书籍ID

**响应示例**:
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "id": 123,
    "title": "深度学习",
    "authors": ["Ian Goodfellow", "Yoshua Bengio"],
    "publisher": "人民邮电出版社",
    "pubdate": "2017-07-01T00:00:00Z",
    "isbn": "9787115461476",
    "tags": ["机器学习", "深度学习", "人工智能"],
    "rating": 4.5,
    "comments": "深度学习领域的经典教材...",
    "languages": ["中文"],
    "last_modified": "2024-01-15T10:30:00Z",
    "cover": "/api/get/cover/123.jpg",
    "file_path": "/api/download/book/123.epub"
  }
}
```

##### 1.2 搜索书籍

```http
GET /api/search?q=深度学习&limit=20&offset=0&filter=title
POST /api/search
```

**查询参数** (GET):
- `q` (string, required): 搜索关键词
- `limit` (integer, default=20): 每页结果数量 (1-100)
- `offset` (integer, default=0): 结果偏移量
- `filter` (string): 过滤字段 (title/author/isbn/tags)
- `sort` (string): 排序字段

**请求体** (POST):
```json
{
  "q": "深度学习",
  "limit": 20,
  "offset": 0,
  "filter": "title"
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "records": [
      {
        "id": 123,
        "title": "深度学习",
        "authors": ["Ian Goodfellow"],
        "publisher": "人民邮电出版社",
        "isbn": "9787115461476",
        "rating": 4.5,
        "tags": ["机器学习", "深度学习"],
        ...
      }
    ],
    "total": 150,
    "limit": 20,
    "offset": 0
  }
}
```

**说明**: 
- 此接口使用 Qdrant 实现，支持标题、作者、出版社、ISBN 等字段的精确和模糊匹配
- `filter` 参数指定搜索字段：`title`(标题), `author`(作者), `publisher`(出版社), `isbn`(ISBN), `tags`(标签)
- Qdrant 通过 Scroll API 和内存过滤实现关键词搜索功能

##### 1.3 最近更新的书籍

```http
GET /api/recently?limit=10&offset=0
```

**查询参数**:
- `limit` (integer, default=10): 结果数量 (1-50)
- `offset` (integer, default=0): 结果偏移量

##### 1.4 随机书籍推荐

```http
GET /api/random?limit=10
```

**查询参数**:
- `limit` (integer, default=10): 推荐数量 (1-50)

#### 2. 书籍管理

##### 2.1 更新书籍元数据

```http
POST /api/book/:id/update
```

**请求体**:
```json
{
  "title": "深度学习（第二版）",
  "authors": ["Ian Goodfellow", "Yoshua Bengio", "Aaron Courville"],
  "publisher": "人民邮电出版社",
  "pubdate": "2020-01-01T00:00:00Z",
  "isbn": "9787115461476",
  "tags": ["机器学习", "深度学习", "人工智能", "神经网络"],
  "rating": 4.8,
  "comments": "更新的描述信息..."
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "元数据更新成功",
  "data": true
}
```

##### 2.2 删除书籍

```http
POST /api/book/:id/delete
```

**响应示例**:
```json
{
  "code": 200,
  "message": "ok",
  "data": true
}
```

#### 3. 文件和资源

##### 3.1 获取封面

```http
GET /api/get/cover/:id
```

**路径参数**:
- `id` (string): 书籍ID (可带 .jpg 后缀)

**响应**: 返回 JPEG 图片二进制数据

##### 3.2 代理封面

```http
GET /api/proxy/cover/*path
```

**路径参数**:
- `path` (string): 封面URL路径

**用途**: 代理获取在线封面图片，添加必要的请求头

##### 3.3 下载书籍文件

```http
GET /api/download/book/:id
```

**路径参数**:
- `id` (string): 书籍ID (可带 .epub 后缀)

**响应**: 返回 EPUB 文件二进制数据

##### 3.4 获取书籍目录

```http
GET /api/read/:id/toc
```

**路径参数**:
- `id` (string): 书籍ID

**响应示例**:
```json
{
  "points": [
    {
      "text": "第一章 引言",
      "content": {
        "src": "/read/123/file/OEBPS/chapter1.xhtml"
      }
    }
  ],
  "metadata": {...},
  "manifest": {...},
  "baseDir": "OEBPS"
}
```

##### 3.5 读取书籍内容

```http
GET /api/read/:id/file/*path
GET /api/book/content?id=123&path=OEBPS/content.opf
```

**路径参数** (方式1):
- `id` (string): 书籍ID
- `path` (string): 文件路径

**查询参数** (方式2):
- `id` (string, required): 书籍ID
- `path` (string, default="OEBPS/content.opf"): 文件路径

#### 4. 元数据服务

##### 4.1 根据 ISBN 获取元数据

```http
GET /api/metadata/isbn/:isbn
```

**路径参数**:
- `isbn` (string, required): ISBN号

**响应示例**:
```json
{
  "id": "1234567",
  "title": "深度学习",
  "author": ["Ian Goodfellow"],
  "publisher": "人民邮电出版社",
  "pubdate": "2017-07-01",
  "isbn13": "9787115461476",
  "summary": "深度学习领域的经典教材...",
  "images": {
    "large": "https://example.com/cover.jpg"
  }
}
```

##### 4.2 搜索元数据

```http
GET /api/metadata/search?query=深度学习&limit=10
```

**查询参数**:
- `query` (string, required): 搜索查询
- `limit` (integer, default=10): 结果数量 (1-50)

#### 5. 出版社管理

##### 5.1 获取出版社列表

```http
GET /api/publisher?limit=50&offset=0
```

**查询参数**:
- `limit` (integer, default=50): 结果数量 (1-100)
- `offset` (integer, default=0): 结果偏移量

**响应示例**:
```json
{
  "code": 200,
  "data": [
    {
      "name": "人民邮电出版社",
      "count": 150
    },
    {
      "name": "机械工业出版社",
      "count": 98
    }
  ]
}
```

#### 6. 数据同步

##### 6.1 同步书籍到 Qdrant

**说明**: 使用任务管理 API 来执行数据同步，详见 [任务管理](#8-任务管理) 章节。

**推荐方式**:
```bash
# 启动向量数据同步任务
curl -X POST http://localhost:8080/api/tasks/start \
  -H 'Content-Type: application/json' \
  -d '{"type": "qdrant_sync", "mode": "full"}'

# 查看同步进度
curl http://localhost:8080/api/tasks
```

**说明**: 
- Qdrant 统一管理所有数据，包括向量和元数据
- 不再需要维护多个索引
- 数据同步通过任务管理系统执行

#### 7. 语义搜索

##### 7.1 语义搜索

```http
GET /api/search/semantic?q=关于人工智能的入门书籍&limit=10
```

**查询参数**:
- `q` (string, required): 搜索查询（自然语言）
- `limit` (integer, default=10): 结果数量

**响应示例**:
```json
{
  "code": 200,
  "data": [
    {
      "id": 123,
      "title": "人工智能：一种现代方法",
      "authors": ["Stuart Russell", "Peter Norvig"],
      "publisher": "人民邮电出版社",
      "isbn": "9787111234760",
      "rating": 4.8,
      "score": 0.95,
      "rank": 1,
      ...
    }
  ]
}
```

**说明**:
- 使用 Qdrant 向量搜索实现
- 自动将查询文本转换为 Embedding 向量
- 返回最相似的书籍，按相似度排序
- `score` 为相似度分数（0-1之间，越接近1越相似）

##### 7.2 混合搜索（语义 + 过滤）

虽然没有单独的端点，但语义搜索支持元数据过滤。可以在客户端实现或通过扩展 API 支持。

**未来扩展示例**:
```http
POST /api/search/hybrid
{
  "query": "机器学习入门",
  "filter": {
    "rating": {"$gte": 4.0},
    "publisher": "O'Reilly"
  },
  "limit": 10
}
```

#### 8. 任务管理

##### 8.1 获取任务列表

```http
GET /api/tasks
```

**响应示例**:
```json
{
  "code": 200,
  "data": [
    {
      "id": "task-001",
      "type": "qdrant_sync",
      "mode": "incremental",
      "state": "running",
      "progress": 0.45,
      "message": "正在生成向量 2250/5000",
      "start_time": "2024-01-15T10:00:00Z"
    }
  ]
}
```

##### 8.2 启动任务

```http
POST /api/tasks/start
```

**请求体**:
```json
{
  "type": "qdrant_sync",
  "mode": "incremental"
}
```

**任务类型**:
- `qdrant_sync`: 向量数据生成和同步（为书籍生成 Embedding 并同步到 Qdrant）

**任务模式**:
- `full`: 全量同步
- `incremental`: 增量同步

**响应示例**:
```json
{
  "code": 200,
  "message": "Task started",
  "data": {
    "id": "task-001"
  }
}
```

**使用示例**:
```bash
# 全量生成向量（为所有书籍生成 Embedding）
curl -X POST http://localhost:8080/api/tasks/start \
  -H 'Content-Type: application/json' \
  -d '{"type": "qdrant_sync", "mode": "full"}'

# 增量生成向量（只为新书生成 Embedding）
curl -X POST http://localhost:8080/api/tasks/start \
  -H 'Content-Type: application/json' \
  -d '{"type": "qdrant_sync", "mode": "incremental"}'
```

##### 8.3 停止任务

```http
POST /api/tasks/:id/stop
```

**路径参数**:
- `id` (string, required): 任务ID

**响应示例**:
```json
{
  "code": 200,
  "message": "Task stopped"
}
```

---

## MCP 增强工具

### MCP 接口概述

MCP (Model Context Protocol) 是 AI 助手与外部服务交互的标准协议。Calibre-API 提供了完整的 MCP 支持，允许 AI 助手通过自然语言管理书籍。

### 启用 MCP 模式

#### 方式 1: 命令行参数
```bash
./calibre-api --mcp
```

#### 方式 2: 环境变量
```bash
MCP_MODE=true ./calibre-api
```

#### 方式 3: 配置文件
```yaml
mcp:
  enabled: true
```

### MCP 端点

```http
GET /mcp
```

AI 客户端连接到此端点以发现可用的工具、资源和提示词。

### 增强工具列表

#### 1. search_books_enhanced

**描述**: 增强的书籍搜索工具，支持多种搜索方式和结果分析

**输入参数**:
```json
{
  "query": "机器学习",
  "limit": 10,
  "offset": 0,
  "sort": "relevance",
  "include_resources": true
}
```

**输出**:
```json
{
  "books": [...],
  "total": 150,
  "query": "机器学习",
  "limit": 10,
  "offset": 0,
  "has_more": true
}
```

**关联提示词**:
- `search_books_by_topic`: 按主题搜索书籍
- `search_books_by_author`: 按作者搜索书籍

#### 2. get_book_details_enhanced

**描述**: 获取书籍详细信息，包括元数据、封面、目录等资源

**输入参数**:
```json
{
  "book_id": "123",
  "include_cover": true,
  "include_toc": true,
  "include_metadata": true
}
```

**关联资源**:
- `cover`: 书籍封面图片
- `toc`: 书籍目录结构
- `metadata`: 完整元数据

**关联提示词**:
- `get_book_details`: 获取书籍详情

#### 3. manage_book_enhanced

**描述**: 增强的书籍管理工具，支持更新元数据、删除等操作

**输入参数**:
```json
{
  "action": "update",
  "book_id": "123",
  "metadata": {
    "title": "新标题",
    "tags": ["标签1", "标签2"]
  }
}
```

**操作类型**:
- `update`: 更新元数据
- `delete`: 删除书籍
- `analyze`: 分析书籍

**关联提示词**:
- `update_book_metadata`: 更新书籍元数据
- `delete_book`: 删除书籍

#### 4. get_recommendations_enhanced

**描述**: 智能书籍推荐工具，支持多种推荐策略

**输入参数**:
```json
{
  "type": "similar",
  "limit": 10,
  "book_id": "123",
  "tags": ["机器学习", "深度学习"]
}
```

**推荐类型**:
- `recent`: 最近添加的书籍
- `random`: 随机推荐
- `similar`: 相似书籍（基于 book_id）
- `popular`: 热门书籍

**关联提示词**:
- `get_recent_books`: 获取最近书籍
- `get_random_books`: 获取随机书籍
- `find_similar_books`: 查找相似书籍

#### 5. metadata_services_enhanced

**描述**: 增强的元数据服务，支持在线搜索和 ISBN 查询

**输入参数**:
```json
{
  "service": "isbn",
  "isbn": "9787115461476",
  "limit": 5
}
```

**服务类型**:
- `search`: 搜索在线元数据
- `isbn`: 通过 ISBN 查询
- `enrich`: 丰富书籍元数据

**关联提示词**:
- `search_metadata_online`: 在线搜索元数据
- `get_metadata_by_isbn`: 通过 ISBN 获取元数据

#### 6. analyze_collection_enhanced

**描述**: 书籍收藏分析工具，提供统计信息和洞察

**输入参数**:
```json
{
  "analysis_type": "overview",
  "group_by": "author",
  "limit": 20
}
```

**分析类型**:
- `overview`: 收藏概览
- `authors`: 作者分析
- `publishers`: 出版社分析
- `topics`: 主题分析
- `timeline`: 时间线分析

**分组方式**:
- `author`: 按作者分组
- `publisher`: 按出版社分组
- `tag`: 按标签分组
- `year`: 按年份分组

**关联提示词**:
- `analyze_book_collection`: 分析书籍收藏

#### 7. export_data_enhanced

**描述**: 数据导出工具，支持多种格式和字段选择

**输入参数**:
```json
{
  "format": "json",
  "fields": ["title", "authors", "publisher", "isbn"],
  "filters": {...},
  "include_resources": false
}
```

**导出格式**:
- `json`: JSON 格式
- `csv`: CSV 格式
- `xml`: XML 格式
- `bibtex`: BibTeX 格式

**关联提示词**:
- `export_book_list`: 导出书籍列表

### 使用增强工具

#### HTTP 接口

##### 获取工具列表

```http
GET /api/mcp/tools/enhanced
```

##### 执行工具

```http
POST /api/mcp/tools/enhanced/:tool
```

**请求体**:
```json
{
  "args": {
    "query": "深度学习",
    "limit": 10
  }
}
```

### MCP 资源

资源提供了对书籍相关数据的统一访问接口。

#### 资源 URI 格式

```
calibre://books/{book_id}/{resource_type}
```

#### 资源类型

1. **metadata**: 书籍元数据
   ```
   calibre://books/123/metadata
   ```

2. **cover**: 书籍封面
   ```
   calibre://books/123/cover
   ```

3. **toc**: 书籍目录
   ```
   calibre://books/123/toc
   ```

4. **content**: 书籍内容
   ```
   calibre://books/123/content
   ```

### MCP 提示词

预定义的交互提示词，帮助 AI 助手更好地理解用户意图。

#### 提示词分类

1. **搜索相关**
   - `search_books_by_topic`: 按主题搜索书籍
   - `search_books_by_author`: 按作者搜索书籍

2. **书籍管理**
   - `get_book_details`: 获取书籍详情
   - `update_book_metadata`: 更新书籍元数据
   - `delete_book`: 删除书籍

3. **推荐相关**
   - `get_recent_books`: 获取最近书籍
   - `get_random_books`: 获取随机书籍
   - `find_similar_books`: 查找相似书籍

4. **元数据服务**
   - `search_metadata_online`: 在线搜索元数据
   - `get_metadata_by_isbn`: 通过 ISBN 获取元数据

5. **分析相关**
   - `analyze_book_collection`: 分析书籍收藏

6. **导出相关**
   - `export_book_list`: 导出书籍列表

---

## 任务管理

### 任务类型详解

#### 1. 向量生成和同步 (vector_sync)

**功能**: 为缺少 Embedding 向量的书籍生成向量

**模式**:
- **全量同步** (`full`):
  - 为所有书籍重新生成 Embedding
  - 适用于更换 Embedding 模型
  
- **增量同步** (`incremental`):
  - 只为新书或缺少向量的书籍生成 Embedding
  - 节省计算资源

**流程**:
1. 查询 Qdrant 中缺少向量的书籍
2. 提取书籍文本（标题、作者、描述、标签等）
3. 调用 Embedding 服务生成 4096 维向量
4. 更新 Qdrant 中对应书籍的向量
5. 实时报告进度

**Embedding 提取策略**:
```
组合文本 = 标题 + 作者 + 出版社 + 标签 + 简介摘要
```

**使用场景**:
- 新添加的书籍需要生成向量
- Embedding 模型升级
- 启用语义搜索功能

### 任务状态

- `running`: 任务正在执行
- `completed`: 任务已完成
- `error`: 任务执行出错
- `stopped`: 任务被手动停止

### 任务监控

**实时进度**: 通过 WebSocket 或轮询获取任务进度

**状态查询**:
```http
GET /api/tasks
```

**性能指标**:
- 处理速度：每秒处理的记录数
- 完成度：已处理/总数的百分比
- 预计剩余时间：基于当前速度估算

---

## 向量搜索

### Qdrant 统一搜索架构

Calibre-API v2.0 采用 Qdrant 作为统一的搜索引擎，实现了关键词搜索和语义搜索的完美结合。

#### 为什么选择 Qdrant？

**Qdrant 的优势**：
- **统一存储**：向量 + 元数据存储在同一个 Qdrant Collection
- **简化架构**：单一搜索引擎，降低运维复杂度
- **灵活搜索**：支持纯向量、纯标量、混合搜索三种模式
- **高性能**：HNSW 索引 + NVMe 优化，搜索延迟 < 50ms
- **易扩展**：原生支持分布式和复制
- **语义理解**：基于 Embedding 的深度语义搜索

#### Qdrant Collection 结构

```json
{
  "collection_name": "books",
  "vectors": {
    "size": 4096,
    "distance": "Cosine"
  },
  "payload": {
    "book_id": "int64",
    "title": "string",
    "authors": ["string"],
    "publisher": "string",
    "isbn": "string",
    "rating": "float",
    "tags": ["string"],
    "languages": ["string"],
    "comments": "string",
    "pubdate": "datetime",
    "last_modified": "datetime",
    "series_index": "float",
    "size": "int64",
    "identifiers": {"isbn": "string", ...},
    "cover": "string",
    "file_path": "string"
  }
}
```

**数据组织**：
- **Point ID**: 书籍 ID（`book_id`）
- **Vector**: 4096 维 Embedding（来自 Qwen3-Embedding-8B）
- **Payload**: 完整的书籍元数据（JSON 格式）

#### Qdrant 搜索能力

| 特性 | 能力 | 说明 |
|------|------|------|
| **搜索类型** | 向量 + 标量 | 支持语义搜索和关键词过滤 |
| **语义理解** | ✅ 优秀 | 基于 Embedding 的语义相似度 |
| **关键词搜索** | ✅ 支持 | Scroll + Payload 过滤 |
| **混合搜索** | ✅ 原生支持 | 向量相似度 + 标量过滤 |
| **存储效率** | ✅ 统一存储 | 向量和元数据存储在一起 |
| **索引速度** | 170本/秒 | 包含 Embedding 生成 |
| **查询延迟** | < 50ms (P95) | HNSW 索引优化 |
| **扩展性** | ✅ 分布式 | 支持集群部署 |
| **运维复杂度** | ⭐⭐ 低 | 单一服务，易于维护 |

**优势**：Qdrant 提供统一的搜索架构，同时支持语义搜索和关键词过滤，简化了系统复杂度，降低了运维成本。

### Qdrant 集成

### 搜索流程

1. **用户查询**: 用户输入自然语言查询
2. **向量化**: 将查询文本转换为 Embedding 向量
3. **向量搜索**: 在 Qdrant 中查找相似向量
4. **结果排序**: 根据相似度分数排序
5. **返回结果**: 返回最相关的书籍

### 搜索模式

Qdrant 作为统一的搜索引擎，支持三种搜索模式：

#### 1. 关键词搜索（标量过滤）

基于 Qdrant Payload 字段的精确或模糊匹配

**实现方式**:
- 使用 Qdrant Scroll API 遍历数据
- 在内存中进行字段匹配和过滤
- 支持标题、作者、出版社、ISBN、标签等字段

**特点**:
- 精确/模糊匹配
- 支持多字段搜索
- 支持分页和排序

**使用场景**:
- 搜索特定书名：`GET /api/search?q=深度学习&filter=title`
- 查找特定作者：`GET /api/search?q=Ian Goodfellow&filter=author`
- ISBN 查询：`GET /api/search?q=9787115461476&filter=isbn`

**性能说明**:
- Qdrant 主要设计用于向量搜索，关键词搜索通过 Scroll + 内存过滤实现
- 对于大规模数据集，关键词搜索性能可能不如专门的全文搜索引擎
- 适用于中小规模数据（< 100万条）

#### 2. 语义搜索（向量相似度）

基于 Embedding 向量的相似度搜索

**实现方式**:
1. 将用户查询转换为 4096 维向量
2. 在 Qdrant 中执行向量相似度搜索（Cosine 距离）
3. 返回最相似的书籍

**特点**:
- 理解语义和上下文
- 支持自然语言查询
- 模糊匹配和智能推荐

**使用场景**:
- 自然语言查询：`GET /api/search/semantic?q=推荐一些关于机器学习的入门书籍`
- 相似书籍推荐：基于某本书的内容找相似书籍
- 主题探索：`GET /api/search/semantic?q=人工智能伦理与社会影响`

**性能说明**:
- Qdrant 针对向量搜索优化，使用 HNSW 索引
- 搜索延迟：< 100ms（P95）
- 支持高并发查询

#### 3. 混合搜索（向量 + 过滤）

结合语义搜索和元数据过滤

**实现方式**:
1. 将查询文本转换为向量
2. 在 Qdrant 中执行向量搜索，同时应用元数据过滤条件
3. 返回符合条件的相似书籍

**特点**:
- 语义理解 + 精确过滤
- 最灵活的搜索方式
- 性能和准确性的平衡

**使用场景**:
- 有条件的语义搜索："找评分高于4.5的机器学习书籍"
- 特定出版社的主题书籍："O'Reilly 出版的数据科学书籍"
- 时间范围内的相关书籍："2020年后出版的深度学习书籍"

**API 示例**（未来扩展）:
```bash
curl -X POST http://localhost:8080/api/search/hybrid \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "机器学习入门",
    "filter": {
      "rating": {"$gte": 4.5},
      "pubdate": {"$gte": "2020-01-01"},
      "publisher": "O'\''Reilly"
    },
    "limit": 10
  }'
```

### Embedding 提供商

#### 1. Ollama（推荐用于本地部署）

**配置**:
```yaml
embedding:
  provider: "ollama"
  ollama:
    api_url: "http://localhost:11434/api/embeddings"
    model: "qwen3-embedding:8b"
```

**优点**:
- 完全本地化
- 无需网络
- 数据隐私

**缺点**:
- 需要本地资源
- 速度取决于硬件

#### 2. SiliconFlow（推荐用于云端部署）

**配置**:
```yaml
embedding:
  provider: "siliconflow"
  siliconflow:
    api_url: "https://api.siliconflow.cn/v1/embeddings"
    model: "Qwen/Qwen3-Embedding-8B"
    api_token: "your_api_token"
```

**优点**:
- 无需本地资源
- 高性能
- 易于扩展

**缺点**:
- 需要网络连接
- 有 API 调用成本

### 向量配置

**向量维度**: 4096 维
- 匹配 Qwen3-Embedding-8B 模型输出
- 高维度保证更好的语义表达能力

**距离度量**: Cosine 相似度
- 适合文本语义比较
- 归一化向量，便于解释相似度分数

**HNSW 索引参数**:
```yaml
m: 32                    # 图连接数（NVMe 优化）
ef_construct: 200        # 构建质量
full_scan_threshold: 10000
on_disk: false           # 向量索引在内存中
```

### 性能优化

#### Qdrant 优化配置

**存储优化**:
- `on_disk_payload: true` - Payload 存储在磁盘（NVMe 高速）
- `on_disk: false` - 向量索引在内存（搜索快）
- `indexing_threshold: 50000` - 批量索引阈值
- `max_segment_size: 200000` - 段大小优化

**性能指标**（实测）:
- 数据规模：130,637 本书
- 索引时间：约 13 分钟
- 搜索延迟：< 50ms (P95)
- 并发能力：100+ QPS
- 存储占用：~2GB（向量） + ~1GB（Payload）

#### 应用层优化

1. **批量处理**: 
   - 批量生成 Embedding（500本/批）
   - 批量写入 Qdrant（UpsertPoints）
   - 提高吞吐量 10倍+

2. **缓存机制**: 
   - 缓存热门查询的 Embedding
   - 减少重复的向量化计算
   - TTL：1小时

3. **连接池管理**:
   - HTTP 客户端连接复用
   - 超时控制：30秒
   - 自动重试机制

4. **并发控制**: 
   - 限制 Embedding 并发数（避免过载）
   - 使用 worker pool 处理批量任务
   - 动态调整批次大小

---

## 配置说明

### 完整配置示例

```yaml
# 服务器配置
address: :8080
debug: false
staticDir: "/app/static"
tmpDir: ".files"

# Calibre Content Server 配置
content:
  server: https://lib.example.com

# 元数据服务配置
metadata:
  doubanurl: https://api.douban.com

# MCP 服务器配置
mcp:
  enabled: false
  server_name: "calibre-mcp-server"
  version: "1.1.0"
  base_url: "http://localhost:8080"
  timeout: 30

# Qdrant 配置（统一的搜索和向量数据库）
qdrant:
  url: "http://127.0.0.1:6333"
  collection: "books"
  timeout: 30

# Embedding 服务配置
embedding:
  provider: "ollama"  # 或 "siliconflow"
  ollama:
    api_url: "http://localhost:11434/api/embeddings"
    model: "qwen3-embedding:8b"
  siliconflow:
    api_url: "https://api.siliconflow.cn/v1/embeddings"
    model: "Qwen/Qwen3-Embedding-8B"
    api_token: ""  # 建议使用环境变量
```

### 环境变量

所有配置项都可以通过环境变量覆盖：

```bash
# 基础配置
CALIBRE_ADDRESS=:8080
CALIBRE_DEBUG=false
CALIBRE_STATICDIR=/app/static
CALIBRE_TMP_DIR=.files

# Calibre Content Server
CALIBRE_CONTENT_SERVER=https://lib.example.com

# 元数据服务
CALIBRE_METADATA_DOUBANURL=https://api.douban.com

# MCP 配置
CALIBRE_MCP_ENABLED=false
CALIBRE_MCP_BASE_URL=http://localhost:8080
MCP_MODE=true  # 快速启用 MCP 模式

# Qdrant 配置（统一搜索引擎）
CALIBRE_QDRANT_URL=http://localhost:6333
CALIBRE_QDRANT_COLLECTION=books
CALIBRE_QDRANT_TIMEOUT=30

# Embedding 配置
CALIBRE_EMBEDDING_PROVIDER=ollama
CALIBRE_EMBEDDING_OLLAMA_API_URL=http://localhost:11434/api/embeddings
CALIBRE_EMBEDDING_OLLAMA_MODEL=qwen3-embedding:8b
SILICONFLOW_API_TOKEN=your_token  # SiliconFlow API Token（如使用 SiliconFlow）
```

### 配置文件优先级

配置文件按以下优先级查找：

1. `/etc/calibre-api/config.yaml`
2. `$HOME/.calibre-api/config.yaml`
3. `./config.yaml`

环境变量的优先级高于配置文件。

---

## 部署指南

### 本地开发

#### 1. 安装依赖

```bash
# 安装 Go 1.21+
go version

# 克隆项目
git clone https://github.com/jianyun8023/calibre-api.git
cd calibre-api

# 下载依赖
go mod download
```

#### 2. 配置服务

**启动 Qdrant**（必需）:
```bash
docker run -d \
  --name calibre-qdrant \
  -p 6333:6333 \
  -p 6334:6334 \
  -v $(pwd)/qdrant_storage:/qdrant/storage \
  qdrant/qdrant:latest
```

**验证 Qdrant**:
```bash
# 检查健康状态
curl http://localhost:6333/health

# 访问 Web UI
open http://localhost:6333/dashboard
```

**创建 Collection**:
```bash
curl -X PUT http://localhost:6333/collections/books \
  -H 'Content-Type: application/json' \
  -d '{
    "vectors": {
      "size": 4096,
      "distance": "Cosine",
      "hnsw_config": {
        "m": 32,
        "ef_construct": 200,
        "on_disk": false
      }
    },
    "optimizers_config": {
      "indexing_threshold": 50000,
      "max_segment_size": 200000
    },
    "on_disk_payload": true
  }'
```

**启动 Ollama**（用于 Embedding，可选）:
```bash
# 安装 Ollama
curl -fsSL https://ollama.com/install.sh | sh

# 启动服务
ollama serve

# 下载 Embedding 模型
ollama pull qwen3-embedding:8b
```

#### 3. 配置文件

创建 `config.yaml`:
```yaml
address: :8080
debug: true
staticDir: "./app/calibre-pages/dist"
tmpDir: ".files"

content:
  server: https://your-calibre-server.com

metadata:
  doubanurl: https://api.douban.com

qdrant:
  url: http://localhost:6333
  collection: books
  timeout: 30

embedding:
  provider: ollama
  ollama:
    api_url: http://localhost:11434/api/embeddings
    model: qwen3-embedding:8b

mcp:
  enabled: false
  server_name: "calibre-mcp-server"
  version: "1.1.0"
  base_url: "http://localhost:8080"
  timeout: 30
```

#### 4. 初始化数据

**同步书籍到 Qdrant**:
```bash
# 启动服务
./calibre-api

# 确保 Ollama 已启动并加载了模型
# 在另一个终端执行向量生成任务
curl -X POST http://localhost:8080/api/tasks/start \
  -H 'Content-Type: application/json' \
  -d '{"type": "qdrant_sync", "mode": "full"}'

# 查看同步进度
curl http://localhost:8080/api/tasks | jq
```

#### 5. 运行服务

```bash
# 开发模式
go run main.go

# 或构建后运行
make build
./calibre-api

# 查看日志
tail -f app.log
```

**访问服务**:
- API: http://localhost:8080
- Web UI: http://localhost:8080
- Qdrant Dashboard: http://localhost:6333/dashboard

### Docker 部署

#### 1. 构建镜像

```bash
docker build -t calibre-api:latest .
```

#### 2. 运行容器

```bash
docker run -d \
  --name calibre-api \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/.files:/app/.files \
  calibre-api:latest
```

### Docker Compose 部署

```yaml
version: '3.8'

services:
  calibre-api:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/app/config.yaml
      - ./.files:/app/.files
      - ./app/calibre-pages/dist:/app/static
    environment:
      - CALIBRE_CONTENT_SERVER=https://lib.example.com
      - CALIBRE_QDRANT_URL=http://qdrant:6333
      - CALIBRE_QDRANT_COLLECTION=books
      - CALIBRE_EMBEDDING_PROVIDER=ollama
      - CALIBRE_EMBEDDING_OLLAMA_API_URL=http://ollama:11434/api/embeddings
    depends_on:
      - qdrant
      - ollama
    restart: unless-stopped

  qdrant:
    image: qdrant/qdrant:latest
    ports:
      - "6333:6333"
      - "6334:6334"
    volumes:
      - qdrant_storage:/qdrant/storage
    environment:
      - QDRANT__SERVICE__GRPC_PORT=6334
    restart: unless-stopped

  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama
    restart: unless-stopped
    # 启动后需要手动拉取模型：
    # docker exec -it ollama ollama pull qwen3-embedding:8b

volumes:
  qdrant_storage:
  ollama_data:
```

运行：
```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f calibre-api

# 在 Ollama 容器中下载模型
docker exec -it ollama ollama pull qwen3-embedding:8b

# 初始化 Qdrant Collection
curl -X PUT http://localhost:6333/collections/books \
  -H 'Content-Type: application/json' \
  -d @- << 'EOF'
{
  "vectors": {
    "size": 4096,
    "distance": "Cosine",
    "hnsw_config": {
      "m": 32,
      "ef_construct": 200,
      "on_disk": false
    }
  },
  "on_disk_payload": true
}
EOF

# 生成书籍向量数据
curl -X POST http://localhost:8080/api/tasks/start \
  -H 'Content-Type: application/json' \
  -d '{"type": "qdrant_sync", "mode": "full"}'
```

### 生产环境部署

#### 1. 系统要求

- **CPU**: 4核及以上
- **内存**: 8GB 及以上
- **磁盘**: 根据书籍数量，建议 100GB+
- **网络**: 稳定的网络连接

#### 2. 性能优化

**Go 服务器**:
```bash
# 设置 GOMAXPROCS
export GOMAXPROCS=4

# 调整垃圾回收
export GOGC=100

# 增加文件句柄限制
ulimit -n 65536
```

**Qdrant 优化**:
```bash
# Docker 运行时优化
docker run -d \
  --name calibre-qdrant \
  -p 6333:6333 \
  -p 6334:6334 \
  -v /path/to/qdrant_storage:/qdrant/storage \
  --memory=8g \
  --cpus=4 \
  -e QDRANT__SERVICE__MAX_REQUEST_SIZE_MB=32 \
  -e QDRANT__SERVICE__GRPC_PORT=6334 \
  qdrant/qdrant:latest
```

**Qdrant Collection 优化**（生产环境）:
```json
{
  "vectors": {
    "size": 4096,
    "distance": "Cosine",
    "hnsw_config": {
      "m": 48,
      "ef_construct": 300,
      "full_scan_threshold": 20000,
      "max_indexing_threads": 8,
      "on_disk": false
    }
  },
  "optimizers_config": {
    "indexing_threshold": 100000,
    "max_segment_size": 500000,
    "memmap_threshold": 100000,
    "flush_interval_sec": 30
  },
  "on_disk_payload": true,
  "replication_factor": 2
}
```

#### 3. 监控和日志

**日志配置**:
```yaml
debug: false  # 生产环境关闭 debug
```

**日志输出**:
```bash
# 重定向日志到文件
./calibre-api > app.log 2>&1 &
```

**监控指标**:
- CPU 使用率
- 内存使用率
- 磁盘 I/O
- 网络流量
- API 响应时间
- 错误率

#### 4. 备份策略

**数据备份**:
```bash
# 备份 Qdrant 数据
tar -czf qdrant_backup_$(date +%Y%m%d).tar.gz qdrant_storage/

# 或使用 Qdrant 快照功能
curl -X POST http://localhost:6333/collections/books/snapshots

# 下载快照
curl -o books_snapshot.snapshot \
  http://localhost:6333/collections/books/snapshots/{snapshot_name}

# 备份配置文件
cp config.yaml config.yaml.bak
```

**定期备份脚本**:
```bash
#!/bin/bash
DATE=$(date +%Y%m%d)

# 创建 Qdrant 快照
SNAPSHOT=$(curl -X POST http://localhost:6333/collections/books/snapshots | jq -r '.result.name')

# 下载快照
curl -o backup/books_snapshot_$DATE.snapshot \
  http://localhost:6333/collections/books/snapshots/$SNAPSHOT

# 备份配置
cp config.yaml backup/config_$DATE.yaml

# 删除旧快照（保留最近7天）
find backup/ -name "*.snapshot" -mtime +7 -delete

echo "Backup completed: $DATE"
```

**恢复数据**:
```bash
# 恢复 Qdrant 快照
curl -X PUT http://localhost:6333/collections/books/snapshots/upload \
  -H 'Content-Type: multipart/form-data' \
  -F 'snapshot=@books_snapshot.snapshot'
```

#### 5. 安全配置

**API 认证**:
- 使用 API Key 或 JWT Token
- HTTPS 加密传输
- 限制 IP 访问

**防火墙**:
```bash
# 只开放必要端口
ufw allow 8080/tcp
ufw allow 6333/tcp
ufw allow 7700/tcp
ufw enable
```

**反向代理** (Nginx):
```nginx
server {
    listen 80;
    server_name api.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### AI 助手集成

#### Claude Desktop 配置

编辑 `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "calibre-api": {
      "command": "/path/to/calibre-api",
      "args": ["--mcp"],
      "env": {
        "CALIBRE_CONTENT_SERVER": "https://lib.example.com"
      }
    }
  }
}
```

#### Cursor 配置

在 Cursor 设置中添加 MCP 服务器：

```json
{
  "mcp.servers": [
    {
      "name": "calibre-api",
      "url": "http://localhost:8080/mcp"
    }
  ]
}
```

### 阅读 APP 集成

#### 书源配置

导入以下书源配置到支持书源的阅读 APP（如阅读3.0）：

```json
{
  "bookSourceUrl": "http://localhost:8080",
  "bookSourceType": 0,
  "bookSourceName": "Calibre书库",
  "bookSourceGroup": "calibre",
  "bookSourceComment": "个人书库",
  "searchUrl": "search?q={{key}}&sort=id:desc",
  "enabled": true,
  "ruleSearch": {
    "bookList": "$.hits",
    "name": "$.title",
    "author": "$.authors",
    "intro": "$.comments",
    "coverUrl": "/get/cover/{{$.id}}.jpg",
    "bookUrl": "/book/{{$.id}}"
  },
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

---

## 附录

### A. 错误码说明

| 错误码 | 说明 | 处理建议 |
|--------|------|----------|
| 200 | 成功 | - |
| 400 | 请求参数错误 | 检查请求参数格式和必填项 |
| 404 | 资源未找到 | 确认资源 ID 是否正确 |
| 500 | 服务器内部错误 | 查看日志，联系管理员 |
| 503 | 服务不可用 | 检查依赖服务是否正常运行 |

### B. 性能指标

**API 响应时间**:
- 书籍查询: < 100ms
- 搜索请求: < 500ms
- 封面获取: < 200ms
- 文件下载: 取决于文件大小

**并发能力**:
- 单实例支持 1000+ 并发连接
- 搜索 QPS: 500+

**数据容量**:
- 支持管理 100,000+ 书籍
- 向量搜索延迟: < 100ms

### C. 常见问题

**Q: 关键词搜索速度慢？**
A: 
1. Qdrant 主要优化向量搜索，关键词搜索通过 Scroll + 内存过滤实现
2. 对于超大数据集（>100万），可能需要优化搜索策略
3. 考虑添加专门的全文搜索引擎（如 Elasticsearch）作为补充
4. 使用更精确的过滤条件减少扫描范围

**Q: 语义搜索结果不相关？**
A:
1. 检查书籍是否已生成 Embedding 向量
2. 确认 Embedding 服务（Ollama/SiliconFlow）正常运行
3. 尝试调整查询文本，使用更具体的描述
4. 检查 Qdrant 向量索引是否完成构建

**Q: 封面图片无法显示？**
A:
1. 检查 Calibre Content Server 连接：`curl https://your-calibre-server.com`
2. 使用封面代理接口：`/api/proxy/cover/*path`
3. 确认封面文件存在：访问 `/api/get/cover/{book_id}`
4. 检查网络和防火墙设置

**Q: Qdrant 内存占用过高？**
A:
1. 检查配置：`on_disk_payload: true`（将 Payload 存到磁盘）
2. 调整 `hnsw_config.on_disk: true`（将向量索引存到磁盘，会降低性能）
3. 减少 `m` 和 `ef_construct` 参数值
4. 增加服务器内存或使用 NVMe 固态硬盘

**Q: 数据同步失败？**
A:
1. 检查 Calibre Content Server 是否可访问
2. 查看任务日志：`curl http://localhost:8080/api/tasks`
3. 检查 Qdrant 存储空间是否充足
4. 验证网络连接和超时设置
5. 尝试使用增量模式：`{"mode": "incremental"}`

### D. 更新日志

#### v2.0.0 (2024-11-21) - 重大更新
- 🎉 **完全迁移到 Qdrant**
- ✨ 统一的搜索引擎（关键词 + 语义搜索）
- ✨ 移除 MeiliSearch 依赖
- ✨ 优化的向量同步任务
- 🔧 HNSW 索引优化（m=32, ef_construct=200）
- 🔧 NVMe 存储优化配置
- ⚡ 搜索性能提升：P95 < 50ms
- 📊 新增 Prometheus 监控支持
- 🐛 修复多个性能和稳定性问题

#### v1.1.0 (2024-01-15)
- ✨ 新增 Qdrant 向量搜索支持（实验性）
- ✨ 新增 MCP 增强工具
- ✨ 新增任务管理系统
- 🔧 优化搜索性能
- 🐛 修复若干 bug

#### v1.0.0 (2023-12-01)
- 🎉 首次发布
- ✨ 基本的书籍管理功能
- ✨ 搜索功能集成
- ✨ MCP 基础支持

### E. 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### F. 许可证

本项目采用 MIT 许可证。详见 [LICENSE](../LICENSE) 文件。

### G. 联系方式

- **项目地址**: https://github.com/jianyun8023/calibre-api
- **问题反馈**: https://github.com/jianyun8023/calibre-api/issues
- **文档**: https://github.com/jianyun8023/calibre-api/docs

---

**最后更新**: 2024-11-21
**版本**: 2.0.0

---

**最后更新**: 2024-11-21
**版本**: 2.0.0

