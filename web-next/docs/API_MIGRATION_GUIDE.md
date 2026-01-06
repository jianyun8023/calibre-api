# API 迁移指南

本指南帮助开发者从旧的 API 格式迁移到新的 V2 格式。

## 目录

- [概述](#概述)
- [主要变化](#主要变化)
- [迁移步骤](#迁移步骤)
- [代码示例](#代码示例)
- [常见问题](#常见问题)

## 概述

在 Spec 026 和 Spec 027 中，我们对前后端 API 接口进行了全面优化和统一：

- **后端 (Spec 026)**: 统一的响应格式和错误处理
- **前端 (Spec 027)**: 统一的 API 客户端和服务层架构

## 主要变化

### 1. API 响应格式

**旧格式 (V1)**:
```typescript
{
  code: 200,
  message: "success",
  data: {...} | { records: [], total: 100, limit: 20, offset: 0 },
  error: "错误消息字符串"  // 简单字符串
}
```

**新格式 (V2)**:
```typescript
{
  code: 200,
  message: "success",
  data: {...},
  error: {                    // 结构化错误对象
    code: "ERROR_CODE",
    message: "错误消息",
    details: "详细信息",
    context: {...}
  },
  trace_id: "abc123"         // 追踪ID
}
```

### 2. 分页格式

**旧格式**:
```typescript
// 请求参数
{ limit: 20, offset: 0 }

// 响应格式
{
  records: [...],
  total: 100,
  limit: 20,
  offset: 0
}
```

**新格式**:
```typescript
// 请求参数
{ page: 1, page_size: 20 }

// 响应格式
{
  code: 200,
  message: "success",
  data: [...],
  pagination: {
    total: 100,
    page: 1,
    page_size: 20,
    total_pages: 5
  }
}
```

### 3. API 调用方式

**旧方式**:
```typescript
import { apiRequest } from '@/lib/api-client'

const books = await apiRequest<Book[]>('/api/books')
```

**新方式（推荐）**:
```typescript
import { bookService } from '@/lib/services'

const books = await bookService.getRecentBooks({ page: 1, page_size: 20 })
```

**新方式（兼容）**:
```typescript
import { fetchRecentBooks } from '@/lib/api/books'

const result = await fetchRecentBooks(20, 0)  // 仍然使用旧参数格式
```

## 迁移步骤

### 步骤 1: 更新导入语句

**替换旧的导入**:
```typescript
import { apiRequest } from '@/lib/api-client'
import { fetchBooks } from '@/lib/api/books'
```

**使用新的导入**:
```typescript
import { bookService } from '@/lib/services'
// 或者使用兼容的API函数
import { fetchBooks } from '@/lib/api/books'  // 内部已使用新服务
```

### 步骤 2: 更新 API 调用

**旧代码**:
```typescript
// 获取书籍列表
const response = await apiRequest<BookListResponse>(
  '/api/recently?limit=20&offset=0'
)
const books = response.records

// 搜索书籍
const searchResult = await apiRequest<BookListResponse>(
  '/api/search?q=keyword',
  {
    method: 'POST',
    body: JSON.stringify({
      Filter: ['tag:programming'],
      Limit: 20,
      Offset: 0,
    })
  }
)
```

**新代码**:
```typescript
// 获取书籍列表
const response = await bookService.getRecentBooks({
  page: 1,
  page_size: 20
})
const books = response.data

// 搜索书籍
const searchResult = await bookService.searchBooks({
  q: 'keyword',
  mode: 'hybrid',
  filters: ['tag:programming'],
  pagination: { page: 1, page_size: 20 }
})
```

### 步骤 3: 更新错误处理

**旧代码**:
```typescript
try {
  const book = await apiRequest('/api/book/1')
} catch (error) {
  toast.error(error.message)  // 简单的错误消息
}
```

**新代码**:
```typescript
import { errorHandler } from '@/lib/error-handler'

try {
  const book = await bookService.getBook('1')
} catch (error) {
  const friendlyError = errorHandler.toUserFriendlyError(error)
  toast.error(friendlyError.title, {
    description: friendlyError.message,
    action: friendlyError.action
  })
}
```

### 步骤 4: 更新分页逻辑

**旧代码**:
```typescript
const [limit] = useState(20)
const [offset, setOffset] = useState(0)

const loadMore = () => {
  setOffset(offset + limit)
}

useEffect(() => {
  fetchBooks(limit, offset)
}, [limit, offset])
```

**新代码**:
```typescript
const [page, setPage] = useState(1)
const [pageSize] = useState(20)

const loadMore = () => {
  setPage(page + 1)
}

useEffect(() => {
  bookService.getRecentBooks({ page, page_size: pageSize })
}, [page, pageSize])
```

## 代码示例

### 示例 1: 书籍列表组件

```typescript
import { useState, useEffect } from 'react'
import { bookService } from '@/lib/services'
import { errorHandler } from '@/lib/error-handler'
import { toast } from 'sonner'

function BookList() {
  const [books, setBooks] = useState([])
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(0)

  const loadBooks = async () => {
    setLoading(true)
    try {
      const response = await bookService.getRecentBooks({
        page,
        page_size: 20
      })

      setBooks(response.data)
      setTotalPages(response.pagination.total_pages)
    } catch (error) {
      const friendlyError = errorHandler.toUserFriendlyError(error)
      toast.error(friendlyError.title, {
        description: friendlyError.message
      })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadBooks()
  }, [page])

  return (
    <div>
      {loading ? <Spinner /> : <BookGrid books={books} />}
      <Pagination 
        currentPage={page}
        totalPages={totalPages}
        onPageChange={setPage}
      />
    </div>
  )
}
```

### 示例 2: 搜索组件

```typescript
import { useState } from 'react'
import { bookService } from '@/lib/services'
import { useDebounce } from '@/hooks/use-debounce'

function BookSearch() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState([])
  const [searching, setSearching] = useState(false)
  
  const debouncedQuery = useDebounce(query, 500)

  useEffect(() => {
    if (!debouncedQuery) {
      setResults([])
      return
    }

    const search = async () => {
      setSearching(true)
      try {
        const response = await bookService.searchBooks({
          q: debouncedQuery,
          mode: 'hybrid',
          pagination: { page: 1, page_size: 10 }
        })
        setResults(response.data)
      } catch (error) {
        console.error('Search failed:', error)
      } finally {
        setSearching(false)
      }
    }

    search()
  }, [debouncedQuery])

  return (
    <div>
      <input 
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="搜索书籍..."
      />
      {searching ? <Spinner /> : <SearchResults results={results} />}
    </div>
  )
}
```

### 示例 3: 元数据搜索

```typescript
import { metadataService } from '@/lib/services'

async function searchBookMetadata(isbn: string) {
  try {
    // 通过 ISBN 搜索
    const book = await metadataService.searchByISBN(isbn)
    
    // 映射到内部格式
    const internalBook = metadataService.mapToInternalFormat(book)
    
    return internalBook
  } catch (error) {
    const friendlyError = errorHandler.toUserFriendlyError(error)
    toast.error(friendlyError.message)
    return null
  }
}

async function searchByKeyword(keyword: string) {
  try {
    const response = await metadataService.searchByKeyword(keyword)
    
    // 映射所有结果
    const books = metadataService.mapManyToInternalFormat(response.books)
    
    return books
  } catch (error) {
    console.error('Metadata search failed:', error)
    return []
  }
}
```

### 示例 4: 任务管理

```typescript
import { taskService } from '@/lib/services'
import { useState, useEffect } from 'react'

function TaskManager() {
  const [tasks, setTasks] = useState([])
  const [eventSource, setEventSource] = useState<EventSource | null>(null)

  // 启动任务
  const startSync = async () => {
    try {
      const task = await taskService.startTask('qdrant_sync', 'full')
      toast.success(`任务已启动: ${task.id}`)
    } catch (error) {
      toast.error('启动任务失败')
    }
  }

  // 订阅任务更新
  useEffect(() => {
    const es = taskService.subscribeToTaskUpdates()
    
    es.addEventListener('message', (event) => {
      const update = JSON.parse(event.data)
      setTasks(prev => 
        prev.map(task => 
          task.id === update.task_id ? { ...task, ...update } : task
        )
      )
    })

    es.addEventListener('error', () => {
      console.error('SSE connection error')
      es.close()
    })

    setEventSource(es)

    return () => {
      es.close()
    }
  }, [])

  return (
    <div>
      <button onClick={startSync}>启动同步</button>
      <TaskList tasks={tasks} />
    </div>
  )
}
```

## 常见问题

### Q1: 我必须立即迁移所有代码吗？

**A**: 不需要。我们提供了向后兼容的 API 函数（如 `fetchBooks`、`fetchRecentBooks` 等），它们内部已经使用新的服务层，但保持了旧的函数签名。你可以逐步迁移。

### Q2: 如何检查迁移进度？

**A**: 在开发模式下，打开浏览器控制台，每 5 分钟会自动打印迁移报告：

```javascript
// 手动查看迁移统计
import { MigrationHelper } from '@/lib/adapters/legacy-adapter'

MigrationHelper.printReport()
```

### Q3: 新的 API 有什么优势？

**A**:
- ✅ 统一的错误处理和用户提示
- ✅ 自动重试机制
- ✅ 请求/响应拦截器
- ✅ 缓存支持
- ✅ 更好的类型安全
- ✅ 更易于测试

### Q4: 旧的 `api-client.ts` 还能用吗？

**A**: 可以，但不推荐。旧的 `apiRequest` 函数仍然可用，但缺少新客户端的功能（重试、缓存、拦截器等）。

### Q5: 如何处理现有组件中的错误？

**A**: 使用新的 `errorHandler`:

```typescript
import { errorHandler } from '@/lib/error-handler'

try {
  // ... API 调用
} catch (error) {
  const friendlyError = errorHandler.toUserFriendlyError(error)
  // 显示友好的错误消息
  toast.error(friendlyError.title, {
    description: friendlyError.message
  })
}
```

### Q6: 分页逻辑需要全部改写吗？

**A**: 如果使用兼容的 API 函数（如 `fetchBooks`），不需要。如果直接使用服务（推荐），需要从 `limit/offset` 改为 `page/page_size`。

### Q7: 测试怎么办？

**A**: 新的服务类更易于测试，可以轻松 mock：

```typescript
const mockBookService = {
  getBook: jest.fn().mockResolvedValue(mockBook),
  getRecentBooks: jest.fn().mockResolvedValue(mockResponse),
}
```

### Q8: 有性能提升吗？

**A**: 有：
- 内置缓存减少重复请求
- 请求合并减少网络开销
- 连接池复用减少延迟

## 相关文档

- [后端 API 文档](../../docs/API_DOCUMENTATION.md)
- [Spec 026 - 后端代码优化](../../specs/026-backend-code-optimization/)
- [Spec 027 - 前端 API 清理](../../specs/027-frontend-api-interface-cleanup/)

## 获取帮助

如有问题，请：
1. 查看本文档的常见问题部分
2. 查看服务类的 JSDoc 注释
3. 查看测试文件中的示例
4. 在项目 Issues 中提问

