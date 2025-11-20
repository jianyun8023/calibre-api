# Gin-MCP 参数说明改进方案

## 问题描述

原始的 gin-mcp 包虽然能够自动发现 Gin 路由并生成 MCP 工具，但是生成的工具缺少详细的参数说明，这导致 AI 助手在使用这些工具时无法了解：

1. 每个参数的具体含义
2. 参数的数据类型和约束
3. 哪些参数是必需的
4. 参数的取值范围和格式

## 解决方案

### 1. 创建参数结构体

在 `internal/calibre/schemas.go` 中定义了所有 API 接口的参数结构体，每个结构体都包含：

- `form` 标签：用于 Gin 的参数绑定
- `json` 标签：用于 JSON 序列化
- `jsonschema` 标签：用于生成 MCP 工具的参数说明

### 2. 注册参数模式

在 `main.go` 中通过 `registerMCPSchemas` 函数为每个 API 接口注册对应的参数模式：

```go
func registerMCPSchemas(mcp *ginmcp.GinMCP) {
    // 搜索相关接口
    mcp.RegisterSchema("GET", "/api/search", calibre.SearchRequest{}, nil)
    mcp.RegisterSchema("POST", "/api/search", nil, calibre.SearchRequest{})
    
    // 书籍管理相关接口
    mcp.RegisterSchema("POST", "/api/book/:id/update", nil, calibre.BookUpdateRequest{})
    
    // 元数据相关接口
    mcp.RegisterSchema("GET", "/api/metadata/search", calibre.MetadataSearchRequest{}, nil)
    
    // 索引管理相关接口
    mcp.RegisterSchema("POST", "/api/index/update", nil, calibre.IndexUpdateRequest{})
    mcp.RegisterSchema("POST", "/api/index/switch", nil, calibre.IndexSwitchRequest{})
    
    // 出版社列表接口
    mcp.RegisterSchema("GET", "/api/publisher", calibre.PublisherListRequest{}, nil)
    
    // 最近书籍接口
    mcp.RegisterSchema("GET", "/api/recently", calibre.RecentlyBooksRequest{}, nil)
    
    // 随机书籍接口
    mcp.RegisterSchema("GET", "/api/random", calibre.RandomBooksRequest{}, nil)
}
```

### 3. 参数结构体示例

#### 搜索请求参数

```go
type SearchRequest struct {
    Q           string `form:"q" json:"q" jsonschema:"description=搜索关键词,required"`
    Limit       int    `form:"limit,default=20" json:"limit,omitempty" jsonschema:"description=每页结果数量,minimum=1,maximum=100"`
    Offset      int    `form:"offset,default=0" json:"offset,omitempty" jsonschema:"description=结果偏移量,minimum=0"`
    Filter      string `form:"filter" json:"filter,omitempty" jsonschema:"description=过滤条件"`
    Sort        string `form:"sort" json:"sort,omitempty" jsonschema:"description=排序字段"`
    // ... 更多参数
}
```

#### 书籍更新请求参数

```go
type BookUpdateRequest struct {
    Title       string    `json:"title,omitempty" jsonschema:"description=书籍标题"`
    Authors     []string  `json:"authors,omitempty" jsonschema:"description=作者列表"`
    AuthorSort  string    `json:"author_sort,omitempty" jsonschema:"description=作者排序"`
    Comments    string    `json:"comments,omitempty" jsonschema:"description=书籍评论"`
    Publisher   string    `json:"publisher,omitempty" jsonschema:"description=出版社"`
    PubDate     time.Time `json:"pubdate,omitempty" jsonschema:"description=出版日期"`
    Isbn        string    `json:"isbn,omitempty" jsonschema:"description=ISBN号"`
    Languages   []string  `json:"languages,omitempty" jsonschema:"description=语言列表"`
    Tags        []string  `json:"tags,omitempty" jsonschema:"description=标签列表"`
    SeriesIndex float64   `json:"series_index,omitempty" jsonschema:"description=系列索引"`
    Rating      float64   `json:"rating,omitempty" jsonschema:"description=评分,minimum=0,maximum=5"`
    Identifiers map[string]string `json:"identifiers,omitempty" jsonschema:"description=标识符映射"`
}
```

## 改进效果

### 改进前

MCP 工具生成的参数说明可能是这样的：
```
参数：q (string)
参数：limit (number)
参数：offset (number)
```

### 改进后

MCP 工具生成的参数说明会包含详细信息：
```
参数：q (string, required) - 搜索关键词
参数：limit (number, 1-100, default=20) - 每页结果数量
参数：offset (number, >=0, default=0) - 结果偏移量
参数：filter (string, optional) - 过滤条件
参数：sort (string, optional) - 排序字段
```

## 使用方法

1. 确保您的 `config.yaml` 中启用了 MCP 功能
2. 启动服务器：`./calibre-api`
3. 在 MCP 客户端（如 Cursor）中连接到 `http://localhost:8080/mcp`
4. 现在所有的 API 工具都会包含详细的参数说明

## 扩展指南

如果需要为新的 API 接口添加参数说明：

1. 在 `internal/calibre/schemas.go` 中定义新的参数结构体
2. 在 `registerMCPSchemas` 函数中注册新的模式
3. 确保结构体包含适当的 `jsonschema` 标签

### jsonschema 标签说明

- `description`: 参数描述
- `required`: 标记必需参数
- `minimum`/`maximum`: 数值范围约束
- `pattern`: 字符串格式约束
- `enum`: 枚举值约束
- `default`: 默认值

## 注意事项

1. 确保参数结构体的字段名与 API 接口中使用的参数名一致
2. 对于路径参数，使用 `uri` 标签而不是 `form` 标签
3. 对于复杂的嵌套结构，可以定义嵌套的结构体
4. 定期更新参数说明以保持与 API 文档的一致性 