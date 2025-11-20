# EnhancedTool 注册到 MCP 完整指南

## 概述

EnhancedTool 是 Calibre API 的增强工具系统，提供了比基础 API 更丰富的功能。本文档详细说明如何将 EnhancedTool 注册到 MCP (Model Context Protocol) 系统中。

## 架构设计

### 1. 核心组件

```go
// EnhancedTool 增强工具定义
type EnhancedTool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
    Resources   []string               `json:"resources,omitempty"`
    Prompts     []string               `json:"prompts,omitempty"`
}

// EnhancedToolManager 增强工具管理器
type EnhancedToolManager struct {
    api         *Api
    resourceMgr *ResourceManager
    promptMgr   *PromptManager
}
```

### 2. 注册流程

```
1. 定义 EnhancedTool
   ↓
2. 在 EnhancedToolManager 中注册
   ↓
3. 创建 HTTP 端点
   ↓
4. 注册到 gin-mcp
   ↓
5. MCP 客户端自动发现
```

## 实现步骤

### 步骤 1: 定义增强工具

在 `internal/calibre/mcp_enhanced_tools.go` 中定义工具：

```go
func (etm *EnhancedToolManager) GetEnhancedTools() []EnhancedTool {
    return []EnhancedTool{
        {
            Name:        "search_books_enhanced",
            Description: "增强的书籍搜索工具，支持多种搜索方式和结果分析",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "query": map[string]interface{}{
                        "type":        "string",
                        "description": "搜索关键词",
                    },
                    "limit": map[string]interface{}{
                        "type":        "integer",
                        "description": "返回结果数量",
                        "default":     10,
                    },
                },
                "required": []string{"query"},
            },
            Prompts: []string{"search_books_by_topic", "search_books_by_author"},
        },
        // ... 更多工具
    }
}
```

### 步骤 2: 实现工具执行逻辑

```go
func (etm *EnhancedToolManager) ExecuteEnhancedTool(toolName string, args map[string]interface{}) (interface{}, error) {
    switch toolName {
    case "search_books_enhanced":
        return etm.executeSearchBooksEnhanced(args)
    case "get_book_details_enhanced":
        return etm.executeGetBookDetailsEnhanced(args)
    // ... 更多工具
    default:
        return nil, fmt.Errorf("未知的工具: %s", toolName)
    }
}
```

### 步骤 3: 创建 HTTP 端点

在 `internal/calibre/api.go` 中添加路由：

```go
func (c *Api) SetupRouter(r *gin.Engine) {
    base := r.Group("/api")
    // ... 其他路由
    
    // Enhanced Tools MCP 端点
    base.GET("/mcp/tools/enhanced", c.getEnhancedTools)
    base.POST("/mcp/tools/enhanced/:tool", c.executeEnhancedTool)
}
```

### 步骤 4: 实现端点处理函数

```go
// getEnhancedTools 获取增强工具列表
func (c *Api) getEnhancedTools(context *gin.Context) {
    etm := NewEnhancedToolManager(c)
    tools := etm.GetEnhancedTools()
    
    context.JSON(http.StatusOK, gin.H{
        "code": 200,
        "data": tools,
    })
}

// executeEnhancedTool 执行增强工具
func (c *Api) executeEnhancedTool(context *gin.Context) {
    toolName := context.Param("tool")
    
    var args map[string]interface{}
    if err := context.ShouldBindJSON(&args); err != nil {
        context.JSON(http.StatusBadRequest, gin.H{
            "code":    400,
            "message": "参数解析失败: " + err.Error(),
        })
        return
    }
    
    etm := NewEnhancedToolManager(c)
    result, err := etm.ExecuteEnhancedTool(toolName, args)
    if err != nil {
        context.JSON(http.StatusInternalServerError, gin.H{
            "code":    500,
            "message": "工具执行失败: " + err.Error(),
        })
        return
    }
    
    context.JSON(http.StatusOK, gin.H{
        "code": 200,
        "data": result,
    })
}
```

### 步骤 5: 注册到 gin-mcp

在 `main.go` 中注册参数模式：

```go
func registerMCPSchemas(mcp *ginmcp.GinMCP) {
    // ... 其他注册
    
    // Enhanced Tools 接口
    mcp.RegisterSchema("GET", "/api/mcp/tools/enhanced", nil, nil)
    mcp.RegisterSchema("POST", "/api/mcp/tools/enhanced/:tool", nil, calibre.EnhancedToolRequest{})
}
```

### 步骤 6: 定义参数结构体

在 `internal/calibre/schemas.go` 中添加：

```go
// EnhancedToolRequest 增强工具请求参数
type EnhancedToolRequest struct {
    Args map[string]interface{} `json:"args" jsonschema:"description=工具参数"`
}
```

## 可用的增强工具

### 1. search_books_enhanced
- **功能**: 增强的书籍搜索
- **参数**: query, limit, offset, sort, include_resources
- **返回**: 搜索结果 + 资源信息

### 2. get_book_details_enhanced
- **功能**: 获取书籍详细信息
- **参数**: book_id, include_cover, include_toc, include_metadata
- **返回**: 完整书籍信息 + 资源

### 3. manage_book_enhanced
- **功能**: 增强的书籍管理
- **参数**: action, book_id, metadata
- **返回**: 操作结果

### 4. get_recommendations_enhanced
- **功能**: 智能书籍推荐
- **参数**: type, limit, book_id, tags
- **返回**: 推荐书籍列表

### 5. metadata_services_enhanced
- **功能**: 增强的元数据服务
- **参数**: service, query, isbn, limit
- **返回**: 元数据信息

### 6. analyze_collection_enhanced
- **功能**: 书籍收藏分析
- **参数**: analysis_type, group_by, limit
- **返回**: 分析结果

### 7. export_data_enhanced
- **功能**: 数据导出工具
- **参数**: format, fields, filters, include_resources
- **返回**: 导出数据

## 使用示例

### 1. 获取工具列表

```bash
curl -X GET "http://localhost:8080/api/mcp/tools/enhanced"
```

### 2. 执行增强搜索

```bash
curl -X POST "http://localhost:8080/api/mcp/tools/enhanced/search_books_enhanced" \
  -H "Content-Type: application/json" \
  -d '{
    "args": {
      "query": "机器学习",
      "limit": 10,
      "include_resources": true
    }
  }'
```

### 3. 获取书籍详情

```bash
curl -X POST "http://localhost:8080/api/mcp/tools/enhanced/get_book_details_enhanced" \
  -H "Content-Type: application/json" \
  -d '{
    "args": {
      "book_id": "123",
      "include_cover": true,
      "include_toc": true
    }
  }'
```

## MCP 客户端集成

### Claude Desktop 配置

```json
{
  "mcpServers": {
    "calibre-enhanced": {
      "command": "/path/to/calibre-api",
      "args": ["--mcp"],
      "env": {
        "CALIBRE_MCP_BASE_URL": "http://localhost:8080"
      }
    }
  }
}
```

### 工具发现

MCP 客户端会自动发现以下端点：
- `GET /api/mcp/tools/enhanced` - 获取工具列表
- `POST /api/mcp/tools/enhanced/:tool` - 执行工具

## 优势特性

### 1. 资源集成
- 工具可以直接访问书籍资源（封面、目录、文件等）
- 支持资源缓存和优化

### 2. 提示模板
- 每个工具关联相关的提示模板
- 帮助 AI 助手更好地理解工具用途

### 3. 智能分析
- 提供数据分析和洞察功能
- 支持多种分析维度

### 4. 多格式支持
- 支持多种数据导出格式
- 灵活的参数配置

## 扩展开发

### 添加新工具

1. 在 `GetEnhancedTools()` 中定义工具
2. 在 `ExecuteEnhancedTool()` 中添加执行逻辑
3. 实现具体的工具功能
4. 更新文档和测试

### 自定义资源

1. 在 `ResourceManager` 中添加资源类型
2. 实现资源的读取和管理逻辑
3. 在工具中集成资源访问

### 提示模板

1. 在 `PromptManager` 中添加提示模板
2. 定义模板参数和渲染逻辑
3. 关联到相应的工具

## 总结

EnhancedTool 通过以下方式成功注册到 MCP：

1. **HTTP 端点**: 创建标准的 REST API 端点
2. **参数模式**: 使用 jsonschema 定义详细的参数说明
3. **自动发现**: gin-mcp 自动发现和注册工具
4. **资源集成**: 提供丰富的资源访问能力
5. **提示模板**: 关联相关的提示模板

这种设计使得 AI 助手能够：
- 自动发现和使用增强工具
- 获得详细的参数说明和约束
- 访问书籍的实际资源
- 使用预设的提示模板
- 执行复杂的分析和推荐功能

通过这种方式，Calibre API 的 MCP 实现变得更加完整和实用，为 AI 助手提供了更丰富的交互能力。 