# MCP 语义搜索和目录功能测试指南

## 概述

本文档说明如何测试 v1.2.1 中的 MCP 工具优化：
- `search_books` 工具改用纯语义搜索
- `get_book` 工具增强，返回完整元数据和目录（TOC）

## 前置条件

1. **服务器运行**：确保 calibre-api 服务正常运行
   ```bash
   ./calibre-api
   ```

2. **Qdrant 配置**：确保已配置 Qdrant 向量数据库
   - 检查 `config.yaml` 中的 `qdrant.url` 配置
   - 确保已同步书籍到 Qdrant（运行 `qdrant_sync` 任务）

3. **数据准备**：
   - 至少有一些书籍数据
   - 理想情况下，有一些 EPUB 格式的书籍（用于测试 TOC）

## 快速测试

### 方法 1: 使用测试脚本（推荐）

运行专门的测试脚本：

```bash
cd /path/to/calibre-api
chmod +x examples/test_semantic_search_and_toc.sh
./examples/test_semantic_search_and_toc.sh
```

**测试内容**：
- ✅ 语义搜索查询（自然语言）
- ✅ 搜索结果验证
- ✅ get_book 返回 TOC
- ✅ 工具 Schema 验证

### 方法 2: 使用 MCP Inspector

1. **打开 MCP Inspector**：
   - 在浏览器中打开 `examples/mcp_sse_client.html`
   - 或访问在线版本：https://mcp-inspector.netlify.app/

2. **连接到服务器**：
   ```
   Server URL: http://localhost:8080/mcp/sse
   Transport: SSE
   ```

3. **测试语义搜索**：
   - 选择工具：`search_books`
   - 输入查询：
     ```json
     {
       "query": "关于 Python 编程的书",
       "limit": 5
     }
     ```
   - 点击 "Call Tool"
   - 观察返回结果

4. **测试 get_book 获取 TOC**：
   - 从搜索结果中复制一个书籍 ID
   - 选择工具：`get_book`
   - 输入参数：
     ```json
     {
       "id": "123"  // 替换为实际 ID
     }
     ```
   - 点击 "Call Tool"
   - 检查返回结果中是否包含 `toc` 字段

### 方法 3: 使用 cURL 手动测试

#### 1. 初始化连接

```bash
SESSION_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')

curl -X POST "http://localhost:8080/mcp/message?sessionId=${SESSION_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
      }
    }
  }' | jq .
```

#### 2. 测试语义搜索

```bash
curl -X POST "http://localhost:8080/mcp/message?sessionId=${SESSION_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "search_books",
      "arguments": {
        "query": "机器学习算法",
        "limit": 5
      }
    }
  }' | jq .
```

**预期结果**：
- 返回基于语义相似度的搜索结果
- 结果中包含 `books` 数组和 `count` 字段
- 能理解自然语言查询（不需要精确的关键词匹配）

#### 3. 测试 get_book 获取 TOC

```bash
# 替换 BOOK_ID 为实际的书籍 ID
BOOK_ID="123"

curl -X POST "http://localhost:8080/mcp/message?sessionId=${SESSION_ID}" \
  -H "Content-Type: application/json" \
  -d "{
    \"jsonrpc\": \"2.0\",
    \"id\": 3,
    \"method\": \"tools/call\",
    \"params\": {
      \"name\": \"get_book\",
      \"arguments\": {
        \"id\": \"$BOOK_ID\"
      }
    }
  }" | jq .
```

**预期结果**：
- 返回完整的书籍元数据（title, authors, publisher, isbn, comments 等）
- 如果是 EPUB 格式且已提取目录，包含 `toc` 字段
- `toc` 字段结构示例：
  ```json
  {
    "toc": {
      "chapter1": "第一章 介绍",
      "chapter2": "第二章 基础概念",
      ...
    }
  }
  ```

## 验证要点

### ✅ 语义搜索验证

1. **自然语言理解**：
   - 查询 "关于 Python 的书" 应该能找到 Python 相关书籍
   - 查询 "机器学习入门" 应该能找到机器学习书籍
   - 不需要精确的关键词匹配

2. **参数简化**：
   - 工具只需要 `query` 和 `limit` 参数
   - 不再有 `filter` 和 `offset` 参数

3. **工具描述**：
   - 工具描述明确说明使用"语义搜索"和"向量相似度匹配"

### ✅ TOC 功能验证

1. **完整元数据**：
   - get_book 返回所有书籍字段
   - 包括 id, title, authors, publisher, isbn, tags, rating, comments 等

2. **目录信息**：
   - 如果书籍有目录，返回结果包含 `toc` 字段
   - TOC 格式可能是对象或数组，取决于提取结果

3. **容错处理**：
   - TOC 获取失败不影响基本元数据返回
   - 日志中会记录 TOC 获取失败的原因

## 常见问题

### Q: 搜索返回 "search service not available"

**A**: Qdrant 服务未启动或未配置
- 检查 `config.yaml` 中的 `qdrant.url`
- 确保 Qdrant 服务正在运行（默认端口 6333）
- 运行 `docker run -d -p 6333:6333 qdrant/qdrant`

### Q: 语义搜索无结果

**A**: 可能需要同步数据到 Qdrant
- 访问 Web UI 的任务管理页面
- 启动 "同步书籍到 Qdrant" 任务
- 或通过 API：`POST /api/tasks/start?type=qdrant_sync`

### Q: get_book 没有返回 TOC

**A**: 可能的原因：
1. 书籍格式不是 EPUB（只有 EPUB 支持 TOC 提取）
2. 未运行过 TOC 提取任务
3. Qdrant 中没有该书籍的 TOC 数据

**解决方案**：
- 运行 TOC 提取任务：`POST /api/tasks/start?type=toc_extract`
- 或首次访问书籍详情页，会自动触发提取并缓存

### Q: 语义搜索结果不准确

**A**: 可能需要优化 embedding 模型
- 检查 `config.yaml` 中的 embedding 配置
- 尝试更换 embedding 模型（如从 Ollama 切换到 SiliconFlow）
- 确保使用的模型维度与 Qdrant 集合配置一致（默认 4096）

## 性能基准

基于测试环境的参考性能：

| 操作 | 平均响应时间 | 备注 |
|------|------------|------|
| 语义搜索 (10 结果) | 200-500ms | 包含向量化时间 |
| 语义搜索 (50 结果) | 300-800ms | 与向量数据库大小相关 |
| get_book (有 TOC) | 100-300ms | 如果 TOC 在 Qdrant 中 |
| get_book (提取 TOC) | 1-3s | 首次提取 EPUB 目录 |

## 进一步测试

### 集成测试

运行完整的 MCP 工具测试：

```bash
./examples/test_mcp_tools.sh
```

### 压力测试

使用 Apache Bench 测试搜索性能：

```bash
# 测试语义搜索并发性能
ab -n 100 -c 10 -p search_payload.json -T application/json \
  http://localhost:8080/api/semantic-search?q=Python&limit=10
```

### 前端集成测试

在 Web UI 中测试：
1. 打开搜索页面
2. 输入自然语言查询
3. 点击书籍查看详情
4. 确认能看到目录信息

## 相关文档

- [MCP README](./MCP_README.md) - MCP 集成完整文档
- [API 文档](./API_DOCUMENTATION.md) - REST API 参考
- [混合搜索策略](./HYBRID_SEARCH_STRATEGY.md) - 搜索算法说明
- [CHANGELOG](../CHANGELOG.md) - 版本变更历史

## 反馈

如果在测试过程中发现问题，请：
1. 检查服务器日志：`tail -f app.log`
2. 提交 Issue：https://github.com/jianyun8023/calibre-api/issues
3. 包含以下信息：
   - calibre-api 版本
   - Qdrant 版本
   - 测试步骤
   - 错误信息
   - 预期结果 vs 实际结果

