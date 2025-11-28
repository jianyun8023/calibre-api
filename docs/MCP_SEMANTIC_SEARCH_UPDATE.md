# MCP 语义搜索和书籍信息增强

> 更新日期: 2024-11-28

## 📋 概述

本次更新优化了 MCP 工具的搜索策略和书籍信息返回，使 AI 助手能够更智能地搜索书籍并获取完整的书籍结构信息。

## 🔄 主要变更

### 1. search_books 工具 - 改用语义搜索

#### 变更前
- 使用关键词搜索 (`searcher.SearchByKeyword`)
- 需要指定 `filter` 参数（title, author, publisher 等）
- 支持 `offset` 分页参数
- 适合精确匹配场景

#### 变更后
- 使用语义搜索 (`searcher.Search`)
- 支持自然语言查询
- 使用向量相似度匹配
- 更智能的理解用户意图

#### 参数变化

**之前：**
```json
{
  "query": "machine learning",
  "filter": "title",      // 可选：title, author, publisher, isbn, tags
  "limit": 20,           // 可选
  "offset": 0            // 可选
}
```

**现在：**
```json
{
  "query": "关于机器学习和深度学习的书籍",  // 支持自然语言
  "limit": 20                              // 可选，默认 20
}
```

#### 使用示例

```bash
# 自然语言查询
curl -X POST http://localhost:8080/mcp/message \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "search_books",
      "arguments": {
        "query": "Python 编程入门教程",
        "limit": 10
      }
    }
  }'

# 技术概念查询
curl -X POST http://localhost:8080/mcp/message \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "search_books",
      "arguments": {
        "query": "distributed systems and microservices",
        "limit": 5
      }
    }
  }'
```

### 2. get_book 工具 - 新增目录信息

#### 变更前
- 仅返回书籍基本元数据（标题、作者、出版社等）
- AI 无法了解书籍的章节结构

#### 变更后
- 返回完整元数据 + 目录结构（TOC）
- 目录信息优先从 Qdrant 获取，缺失时从 EPUB 文件提取
- 目录获取失败不影响基本信息返回

#### 返回数据结构

```json
{
  "id": 123,
  "title": "Python 编程：从入门到实践",
  "authors": ["Eric Matthes"],
  "publisher": "人民邮电出版社",
  "pubdate": "2023-01-01T00:00:00Z",
  "isbn": "9787115551234",
  "tags": ["编程", "Python", "入门"],
  "rating": 4.5,
  "comments": "适合初学者的 Python 教程...",
  "languages": ["zh"],
  "cover": "/api/get/cover/123.jpg",
  "file_path": "/api/download/book/123.epub",
  "toc": {                           // 新增目录信息
    "chapter": [
      {
        "title": "第一章 起步",
        "src": "chapter1.xhtml",
        "children": [
          {
            "title": "1.1 搭建编程环境",
            "src": "chapter1.xhtml#section1-1"
          },
          {
            "title": "1.2 在不同操作系统中运行 Python",
            "src": "chapter1.xhtml#section1-2"
          }
        ]
      },
      {
        "title": "第二章 变量和简单数据类型",
        "src": "chapter2.xhtml"
      }
    ]
  }
}
```

#### 使用示例

```bash
curl -X POST http://localhost:8080/mcp/message \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "get_book",
      "arguments": {
        "id": "123"
      }
    }
  }'
```

## 🎯 使用场景

### 语义搜索适用场景

1. **自然语言查询**
   - "关于机器学习的书"
   - "Python 编程入门教程"
   - "如何学习数据结构和算法"

2. **概念查询**
   - "distributed systems" → 找到分布式系统相关书籍
   - "design patterns" → 找到设计模式相关书籍
   - "web development" → 找到 Web 开发相关书籍

3. **模糊查询**
   - "学习 Java"
   - "前端开发"
   - "人工智能入门"

### 目录信息的价值

1. **书籍总结**
   - AI 可以根据目录结构生成书籍内容概要
   - 了解书籍的章节组织和知识体系

2. **内容定位**
   - 快速找到特定章节
   - 了解某个主题在书中的位置

3. **推荐系统**
   - 基于目录结构相似度推荐相关书籍
   - 判断书籍深度和覆盖范围

4. **学习路径**
   - 根据目录为用户规划学习路线
   - 识别前置知识和进阶内容

## 🔧 技术实现

### 语义搜索实现

```go
// performSemanticSearch 执行语义搜索
func (m *MCPServer) performSemanticSearch(query string, limit int) ([]Book, error) {
    searcher, ok := m.api.semanticSearcher.(*qdrant.Searcher)
    if !ok || searcher == nil {
        return nil, fmt.Errorf("search service not available")
    }

    // 使用语义搜索
    results, err := searcher.Search(query, limit)
    if err != nil {
        return nil, err
    }

    // 转换结果
    books := make([]Book, 0, len(results))
    for _, result := range results {
        books = append(books, convertSemanticToBook(result.Book))
    }

    return books, nil
}
```

### 目录获取实现

```go
// handleGetBook 获取书籍详情（含目录）
func (m *MCPServer) handleGetBook(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // ... 获取书籍基本信息 ...

    // 获取目录信息（如果可用）
    toc, tocErr := m.api.GetBookTocData(id)
    if tocErr != nil {
        log.Debugf("Failed to get TOC for book %s: %v", id, tocErr)
        // TOC 获取失败不影响基本信息返回
    }

    // 构建完整响应
    result := map[string]interface{}{
        "id":            book.ID,
        "title":         book.Title,
        // ... 其他元数据 ...
    }

    // 如果成功获取目录，添加到结果中
    if tocErr == nil && toc != nil {
        result["toc"] = toc
    }

    return mcp.NewToolResultText(formatToolResult(result)), nil
}
```

### 目录数据来源

1. **优先从 Qdrant 获取**
   - 已同步的书籍目录存储在 Qdrant 的 payload 中
   - 访问速度快，无需文件解析

2. **回退到 EPUB 提取**
   - 如果 Qdrant 中没有目录数据
   - 从 EPUB 文件实时提取
   - 异步更新到 Qdrant

3. **容错处理**
   - 目录获取失败不影响元数据返回
   - 仅记录日志，不抛出错误

## 📊 性能影响

### 语义搜索性能

- **优势**：
  - 更智能的搜索结果
  - 支持自然语言，用户体验更好
  - 基于向量相似度，结果更相关

- **考虑**：
  - 需要向量化查询（调用 Embedding 服务）
  - 向量搜索比关键词搜索稍慢（通常 < 100ms）

### 目录获取性能

- **Qdrant 获取**：< 10ms（内存读取）
- **EPUB 提取**：50-200ms（首次访问）
- **缓存机制**：后续访问直接从 Qdrant 获取

## 🧪 测试

### 运行测试脚本

```bash
# 确保服务已启动
./calibre-api

# 在另一个终端运行测试
./test_mcp_semantic_search.sh
```

### 测试内容

1. ✅ 语义搜索 - 自然语言查询
2. ✅ 语义搜索 - 技术概念查询
3. ✅ 获取书籍详情（包含目录）
4. ✅ 验证响应包含 toc 字段

### 预期结果

```bash
====================================
测试 1: 语义搜索 - 自然语言查询
====================================
工具: search_books
参数: {"query": "关于机器学习和深度学习的书籍", "limit": 5}
✓ 测试通过

====================================
测试 2: 语义搜索 - 技术概念
====================================
工具: search_books
参数: {"query": "Python 编程入门教程", "limit": 10}
✓ 测试通过

====================================
测试 3: 获取书籍详情（含目录）
====================================
工具: get_book
参数: {"id": "123"}
✓ 测试通过
✓ 响应包含目录（TOC）信息
```

## 🔄 迁移指南

### 对现有 MCP 客户端的影响

#### search_books 工具

**不兼容变更**：
- 移除了 `filter` 参数
- 移除了 `offset` 参数

**兼容性处理**：
如果客户端仍传递这些参数，会被忽略，不会报错。

**迁移建议**：
```javascript
// 旧代码
await callTool('search_books', {
  query: 'Python',
  filter: 'title',
  limit: 20,
  offset: 0
});

// 新代码
await callTool('search_books', {
  query: 'Python 编程书籍',  // 使用更自然的描述
  limit: 20
});
```

#### get_book 工具

**向后兼容**：
- 所有原有字段保持不变
- 新增 `toc` 字段（可选）
- 旧客户端可以忽略 `toc` 字段

**使用新功能**：
```javascript
const result = await callTool('get_book', { id: '123' });

// 检查是否有目录信息
if (result.toc) {
  console.log('书籍目录:', result.toc);
  // 展示目录结构
  displayTableOfContents(result.toc);
}
```

## 📝 相关文档

- [MCP 集成指南](./MCP_README.md)
- [语义搜索原理](./HYBRID_SEARCH_STRATEGY.md)
- [API 文档](./API_DOCUMENTATION.md)
- [变更日志](../CHANGELOG.md)

## 🤔 常见问题

### Q1: 为什么要从关键词搜索改为语义搜索？

**A:** 语义搜索能理解用户意图，而不只是匹配关键词。例如：
- 查询 "学习 Python" 可以找到 "Python 编程入门"、"Python 实战"等书籍
- 查询 "机器学习" 可以找到 "深度学习"、"神经网络"等相关书籍
- 支持跨语言搜索（英文查询找到中文书籍）

### Q2: 如果还需要关键词搜索怎么办？

**A:** 关键词搜索仍然可用：
- REST API: `GET /api/search?q=Python&mode=keyword`
- 混合搜索: `GET /api/search?q=Python&mode=hybrid`
- MCP 工具专注于提供最佳搜索体验（语义搜索）

### Q3: 目录信息从哪里来？

**A:** 目录信息来源优先级：
1. Qdrant 数据库（已同步的书籍）
2. EPUB 文件实时提取（首次访问）
3. 自动异步更新到 Qdrant（后续快速访问）

### Q4: 如果书籍没有目录怎么办？

**A:** 
- PDF 格式的书籍可能没有目录结构
- `toc` 字段不会出现在响应中
- 不影响其他元数据的正常返回

### Q5: 语义搜索的准确性如何？

**A:** 准确性取决于：
- Embedding 模型质量（使用 Ollama 或 SiliconFlow）
- 向量维度（默认 4096 维）
- 书籍元数据的完整性
- 通常情况下，语义搜索比关键词搜索更准确

## 📈 后续优化

### 计划中的改进

1. **搜索结果排序**
   - 支持按评分、出版日期等排序
   - 个性化推荐

2. **目录搜索**
   - 在书籍目录中搜索特定章节
   - 跨书籍目录搜索

3. **混合模式**
   - 提供可选的混合搜索模式
   - 结合关键词和语义搜索优势

4. **缓存优化**
   - 目录信息缓存策略
   - 搜索结果缓存

---

**版本**: v1.2.1-next  
**作者**: jianyun8023  
**最后更新**: 2024-11-28

