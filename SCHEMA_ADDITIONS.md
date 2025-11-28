# Calibre API Schema 添加总结

## 概述
为 Calibre API 的读取书籍目录和读取书籍内容接口以及其他相关接口添加了 MCP Schema 支持，以便为 MCP 工具提供详细的参数说明。

## 新增的 Schema 结构体

### 1. 读取书籍目录接口
- **接口路径**: `GET /api/read/:id/toc`
- **Schema 结构体**: `BookTocRequest`
- **参数**: 
  - `ID` (string): 书籍ID (必需)

### 2. 读取书籍内容接口
- **接口路径**: `GET /api/read/:id/file/*path`
- **Schema 结构体**: `BookContentRequest`
- **参数**:
  - `ID` (string): 书籍ID (必需)
  - `Path` (string): 内容文件路径 (必需)

### 3. 获取封面接口
- **接口路径**: `GET /api/get/cover/:id`
- **Schema 结构体**: `GetCoverRequest`
- **参数**:
  - `ID` (string): 书籍ID (必需)

### 4. 代理封面接口
- **接口路径**: `GET /api/proxy/cover/*path`
- **Schema 结构体**: `ProxyCoverRequest`
- **参数**:
  - `Path` (string): 封面文件路径 (必需)

### 5. 获取书籍文件接口
- **接口路径**: `GET /api/download/book/:id`
- **Schema 结构体**: `GetBookFileRequest`
- **参数**:
  - `ID` (string): 书籍ID (必需)

### 6. 获取书籍信息接口
- **接口路径**: `GET /api/book/:id`
- **Schema 结构体**: `GetBookRequest`
- **参数**:
  - `ID` (string): 书籍ID (必需)

### 7. 删除书籍接口
- **接口路径**: `POST /api/book/:id/delete`
- **Schema 结构体**: `DeleteBookRequest`
- **参数**:
  - `ID` (string): 书籍ID (必需)

### 8. 获取ISBN信息接口
- **接口路径**: `GET /api/metadata/isbn/:isbn`
- **Schema 结构体**: `GetISBNRequest`
- **参数**:
  - `ISBN` (string): ISBN号 (必需)

## 修改的文件

### 1. `internal/calibre/schemas.go`
添加了以下新的结构体定义：
- `BookTocRequest`
- `BookContentRequest`
- `GetCoverRequest`
- `ProxyCoverRequest`
- `GetBookFileRequest`
- `GetBookRequest`
- `DeleteBookRequest`
- `GetISBNRequest`

### 2. `main.go`
在 `registerMCPSchemas` 函数中添加了对应的 schema 注册：
- 为每个新接口注册了相应的请求参数 schema
- 所有 schema 都包含了中文描述和参数验证规则

## 验证
- ✅ 代码编译通过
- ✅ 所有新添加的结构体都有正确的 JSON 标签和 jsonschema 描述
- ✅ 所有接口的 schema 都已正确注册到 MCP 服务器

## 使用示例

### 读取书籍目录
```bash
GET /api/read/12345/toc
```

### 读取书籍内容
```bash
GET /api/read/12345/file/OEBPS/chapter1.xhtml
```

### 获取封面
```bash
GET /api/get/cover/12345
```

这些 schema 将为 MCP 客户端提供详细的接口参数说明，提高 API 的可用性和文档化程度。 