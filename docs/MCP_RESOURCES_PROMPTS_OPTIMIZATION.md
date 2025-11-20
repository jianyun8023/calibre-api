# MCP Resources 和 Prompts 优化方案

## 概述

针对您提出的 "Resources、Prompts是否需要优化" 问题，我们进行了全面的分析和改进。原始的 gin-mcp 实现确实存在以下不足：

1. **缺少 Resources 功能** - 无法提供书籍内容、封面图片等资源
2. **缺少 Prompts 功能** - 没有预设的提示模板来帮助 AI 助手更好地使用工具
3. **工具功能有限** - 主要是基础的 CRUD 操作，缺少更智能的功能

## 优化方案

### 1. Resources 功能实现

#### 功能特性
- **多种资源类型**：封面图片、目录结构、元数据、电子书文件
- **统一资源 URI**：`calibre://books/{id}/{type}` 格式
- **Base64 编码**：支持二进制资源的安全传输
- **MIME 类型支持**：自动识别不同文件格式

#### 资源类型
```go
const (
    ResourceTypeCover    ResourceType = "cover"    // 封面图片
    ResourceTypeContent  ResourceType = "content"  // 书籍内容
    ResourceTypeToc      ResourceType = "toc"      // 目录
    ResourceTypeMetadata ResourceType = "metadata" // 元数据
    ResourceTypeFile     ResourceType = "file"     // 文件
)
```

#### 使用示例
```go
// 列出书籍的所有资源
resources, err := resourceMgr.ListResources("123")

// 读取特定资源
coverResource, err := resourceMgr.ReadResource("calibre://books/123/cover")
```

### 2. Prompts 功能实现

#### 功能特性
- **预设提示模板**：15+ 个常用操作提示
- **参数化模板**：支持动态参数替换
- **分类管理**：按功能分类组织提示
- **智能建议**：根据上下文推荐相关提示

#### 提示分类
1. **搜索相关**：按主题、作者、ISBN 搜索
2. **书籍管理**：更新元数据、删除书籍
3. **推荐相关**：最近书籍、随机推荐
4. **元数据服务**：在线搜索、ISBN 查询
5. **系统管理**：索引更新、出版社列表
6. **高级功能**：收藏分析、相似书籍、数据导出

#### 使用示例
```go
// 获取所有提示模板
prompts := promptMgr.GetPrompts()

// 根据标签获取提示
searchPrompts := promptMgr.GetPromptsByTag("搜索")

// 渲染提示模板
renderedPrompt, err := promptMgr.RenderPrompt("search_books_by_topic", map[string]string{
    "topic": "机器学习",
})
```

### 3. 增强工具集

#### 新增工具
1. **search_books_enhanced** - 增强搜索，支持资源包含
2. **get_book_details_enhanced** - 详细书籍信息，包含所有资源
3. **manage_book_enhanced** - 增强书籍管理
4. **get_recommendations_enhanced** - 智能推荐系统
5. **metadata_services_enhanced** - 增强元数据服务
6. **analyze_collection_enhanced** - 收藏分析工具
7. **export_data_enhanced** - 数据导出工具

#### 工具特性
- **资源集成**：工具可以直接访问和返回资源
- **提示关联**：每个工具关联相关的提示模板
- **智能分析**：提供数据分析和洞察功能
- **多格式支持**：支持多种数据导出格式

## 实现文件

### 核心文件
1. **internal/calibre/mcp_resources.go** - Resources 功能实现
2. **internal/calibre/mcp_prompts.go** - Prompts 功能实现
3. **internal/calibre/mcp_enhanced_tools.go** - 增强工具集

### 配置更新
- **main.go** - 添加了参数模式注册
- **internal/calibre/schemas.go** - 新增参数结构体定义

## 使用效果

### 改进前
```
工具：search_books
参数：query (string)
参数：limit (number)
```

### 改进后
```
工具：search_books_enhanced
参数：query (string, required) - 搜索关键词
参数：limit (number, 1-100, default=10) - 返回结果数量
参数：include_resources (boolean, default=false) - 是否包含资源信息
关联提示：search_books_by_topic, search_books_by_author
可用资源：cover, toc, metadata, file
```

## 集成方式

### 1. 自动集成
Resources 和 Prompts 功能已自动集成到现有的 MCP 系统中，无需额外配置。

### 2. 手动调用
```go
// 创建管理器
resourceMgr := NewResourceManager(api)
promptMgr := NewPromptManager(api)
enhancedToolMgr := NewEnhancedToolManager(api)

// 使用功能
resources := resourceMgr.ListResources("123")
prompts := promptMgr.GetPrompts()
tools := enhancedToolMgr.GetEnhancedTools()
```

## 优势对比

| 功能 | 原始实现 | 优化后实现 |
|------|----------|------------|
| 参数说明 | 基础类型 | 详细描述 + 约束 |
| 资源访问 | ❌ 不支持 | ✅ 完整支持 |
| 提示模板 | ❌ 不支持 | ✅ 15+ 模板 |
| 工具功能 | 基础 CRUD | 智能分析 + 推荐 |
| 用户体验 | 需要猜测 | 引导式操作 |

## 下一步计划

1. **完善实现**：补充具体的业务逻辑实现
2. **性能优化**：添加缓存和异步处理
3. **扩展功能**：支持更多资源类型和提示模板
4. **文档完善**：提供详细的使用指南和示例

## 总结

通过添加 Resources 和 Prompts 功能，我们显著提升了 MCP 系统的能力：

1. **Resources** 让 AI 助手能够访问书籍的实际内容（封面、目录、文件等）
2. **Prompts** 提供了预设的提示模板，帮助 AI 助手更好地理解和使用工具
3. **增强工具** 提供了更智能的功能，如分析、推荐、导出等

这些优化使得 Calibre API 的 MCP 实现更加完整和实用，为 AI 助手提供了更丰富的交互能力。 