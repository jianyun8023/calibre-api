# 前端 API 接口清理与迁移设计文档

## 概述

本设计文档详细说明了前端 API 接口清理与迁移的技术方案。在 Spec 026 后端代码优化完成后，后端已实现统一的 API 响应格式（V2 格式）和错误处理机制。前端需要相应地更新 API 接口调用逻辑，清理冗余代码，统一错误处理，并适配新的响应格式。

本设计采用渐进式迁移策略，确保在迁移过程中系统的稳定性和可用性，同时提供向后兼容性支持。

## 架构

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Next.js Frontend                         │
├─────────────────────────────────────────────────────────────┤
│  React Components                                           │
│  ├── Book Components (BookGrid, BookDetail, etc.)          │
│  ├── Chat Components (ChatPage, MessageList, etc.)         │
│  ├── Task Components (TaskStream, TaskManager, etc.)       │
│  └── Metadata Components (MetadataDialog, etc.)            │
├─────────────────────────────────────────────────────────────┤
│  API Layer (New Unified Architecture)                      │
│  ├── Unified API Client                                    │
│  │   ├── Request Interceptor                               │
│  │   ├── Response Interceptor                              │
│  │   ├── Error Handler                                     │
│  │   └── Cache Manager                                     │
│  ├── API Services                                          │
│  │   ├── BookService                                       │
│  │   ├── MetadataService                                   │
│  │   ├── ChatService                                       │
│  │   ├── TaskService                                       │
│  │   └── FileService                                       │
│  └── Type Definitions                                      │
│      ├── API Response Types                                │
│      ├── Error Types                                       │
│      ├── Pagination Types                                  │
│      └── Domain Types                                      │
├─────────────────────────────────────────────────────────────┤
│  Compatibility Layer (Temporary)                           │
│  ├── Format Adapter                                        │
│  ├── Migration Helper                                      │
│  └── Legacy Support                                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  Go Backend (V2 API)                       │
│  ├── Unified Response Format                               │
│  ├── Structured Error Handling                             │
│  ├── New Pagination Format                                 │
│  └── Consistent HTTP Status Codes                          │
└─────────────────────────────────────────────────────────────┘
```

### 分层设计

1. **组件层 (Component Layer)**
   - React 组件负责 UI 渲染和用户交互
   - 通过 API 服务层获取数据
   - 不直接调用 HTTP 接口

2. **API 服务层 (API Service Layer)**
   - 提供业务领域相关的 API 方法
   - 封装具体的 HTTP 调用逻辑
   - 处理数据转换和业务逻辑

3. **统一 API 客户端 (Unified API Client)**
   - 统一的 HTTP 请求处理
   - 自动的请求/响应拦截
   - 统一的错误处理和缓存

4. **兼容性层 (Compatibility Layer)**
   - 临时支持新旧格式转换
   - 渐进式迁移支持
   - 迁移完成后移除

## 组件和接口

### 核心组件

#### 1. 统一 API 客户端 (UnifiedApiClient)

```typescript
interface ApiClientConfig {
  baseURL?: string
  timeout?: number
  retryAttempts?: number
  cacheEnabled?: boolean
}

class UnifiedApiClient {
  constructor(config: ApiClientConfig)
  
  // 核心请求方法
  async request<T>(config: RequestConfig): Promise<ApiResponse<T>>
  async get<T>(url: string, config?: RequestConfig): Promise<T>
  async post<T>(url: string, data?: any, config?: RequestConfig): Promise<T>
  async put<T>(url: string, data?: any, config?: RequestConfig): Promise<T>
  async delete<T>(url: string, config?: RequestConfig): Promise<T>
  
  // 拦截器
  interceptors: {
    request: InterceptorManager<RequestConfig>
    response: InterceptorManager<ApiResponse<any>>
  }
}
```

#### 2. 错误处理器 (ErrorHandler)

```typescript
interface ErrorContext {
  url: string
  method: string
  requestId?: string
  timestamp: number
}

class ErrorHandler {
  // 处理 API 错误
  handleApiError(error: ApiError, context: ErrorContext): UserFriendlyError
  
  // 处理网络错误
  handleNetworkError(error: NetworkError, context: ErrorContext): UserFriendlyError
  
  // 错误恢复
  canRetry(error: ApiError): boolean
  getRetryDelay(attempt: number): number
  
  // 错误报告
  reportError(error: Error, context: ErrorContext): void
}
```

#### 3. 缓存管理器 (CacheManager)

```typescript
interface CacheConfig {
  ttl: number // Time to live in milliseconds
  maxSize: number
  strategy: 'lru' | 'fifo' | 'custom'
}

class CacheManager {
  // 缓存操作
  get<T>(key: string): Promise<T | null>
  set<T>(key: string, value: T, ttl?: number): Promise<void>
  delete(key: string): Promise<void>
  clear(): Promise<void>
  
  // 缓存策略
  shouldCache(request: RequestConfig): boolean
  generateKey(request: RequestConfig): string
  isExpired(entry: CacheEntry): boolean
}
```

#### 4. API 服务基类 (BaseApiService)

```typescript
abstract class BaseApiService {
  protected client: UnifiedApiClient
  protected errorHandler: ErrorHandler
  
  constructor(client: UnifiedApiClient, errorHandler: ErrorHandler)
  
  // 通用方法
  protected async handleRequest<T>(request: () => Promise<T>): Promise<T>
  protected buildUrl(path: string, params?: Record<string, any>): string
  protected validateResponse<T>(response: ApiResponse<T>): T
}
```

### API 服务接口

#### 1. 书籍服务 (BookService)

```typescript
interface BookService {
  // 书籍列表
  getRecentBooks(pagination: PaginationParams): Promise<PaginatedResponse<Book>>
  getRandomBooks(limit?: number): Promise<Book[]>
  getAllBooks(pagination: CursorPaginationParams): Promise<CursorPaginatedResponse<Book>>
  
  // 书籍操作
  getBook(id: string): Promise<Book>
  updateBook(id: string, data: Partial<Book>): Promise<Book>
  deleteBook(id: string): Promise<void>
  
  // 搜索
  searchBooks(query: SearchQuery): Promise<PaginatedResponse<Book>>
  searchSemantic(query: string, limit?: number): Promise<PaginatedResponse<Book>>
  
  // 文件操作
  getBookToc(id: string): Promise<TocResponse>
  getChapterContent(bookId: string, filePath: string): Promise<string>
  downloadBook(id: string): Promise<Blob>
}
```

#### 2. 元数据服务 (MetadataService)

```typescript
interface MetadataService {
  searchByISBN(isbn: string): Promise<DoubanBook>
  searchByKeyword(query: string): Promise<MetadataSearchResponse>
  
  // 数据映射
  mapToInternalFormat(externalData: DoubanBook): Book
}
```

#### 3. 聊天服务 (ChatService)

```typescript
interface ChatService {
  // 对话管理
  getConversations(): Promise<Conversation[]>
  createConversation(data: CreateConversationRequest): Promise<Conversation>
  deleteConversation(id: string): Promise<void>
  
  // 消息管理
  getMessages(conversationId: string): Promise<Message[]>
  sendMessage(conversationId: string, content: string): Promise<Message>
  deleteMessage(id: string): Promise<void>
  
  // 流式聊天
  streamChat(request: ChatStreamRequest): Promise<ReadableStream>
}
```

#### 4. 任务服务 (TaskService)

```typescript
interface TaskService {
  getTasks(): Promise<Task[]>
  startTask(type: string, mode: string): Promise<Task>
  stopTask(id: string): Promise<void>
  
  // 任务流
  subscribeToTaskUpdates(): Promise<EventSource>
}
```

## 数据模型

### API 响应类型

#### 1. 基础响应类型

```typescript
// 新的统一响应格式 (V2)
interface ApiResponse<T> {
  code: number
  message: string
  data?: T
  error?: ErrorInfo
  trace_id?: string
}

// 错误信息
interface ErrorInfo {
  code: string
  message: string
  details?: string
  context?: Record<string, any>
}

// 分页响应
interface PaginatedResponse<T> {
  code: number
  message: string
  data: T[]
  pagination: Pagination
}

interface Pagination {
  total: number
  page: number
  page_size: number
  total_pages: number
}

// 游标分页响应
interface CursorPaginatedResponse<T> {
  code: number
  message: string
  data: T[]
  pagination: CursorPagination
}

interface CursorPagination {
  total: number
  next_cursor?: string
  has_more: boolean
}
```

#### 2. 兼容性类型

```typescript
// 旧格式 (V1) - 用于兼容性
interface LegacyApiResponse<T> {
  code: number
  message: string
  data?: T | LegacyPaginatedData<T>
  error?: string
}

interface LegacyPaginatedData<T> {
  records: T[]
  total: number
  limit: number
  offset: number
}

// 格式适配器
type AdaptedResponse<T> = ApiResponse<T> | LegacyApiResponse<T>
```

#### 3. 请求参数类型

```typescript
// 分页参数
interface PaginationParams {
  page: number
  page_size: number
}

interface LegacyPaginationParams {
  limit: number
  offset: number
}

// 游标分页参数
interface CursorPaginationParams {
  limit: number
  cursor?: string
}

// 搜索查询
interface SearchQuery {
  q: string
  mode?: 'hybrid' | 'semantic' | 'text'
  filters?: string[]
  sort?: string[]
  pagination: PaginationParams
}
```

### 错误类型

```typescript
// API 错误
class ApiError extends Error {
  code: number
  errorCode?: string
  details?: string
  context?: Record<string, any>
  traceId?: string
  
  constructor(response: ApiResponse<any>, request?: RequestConfig)
}

// 网络错误
class NetworkError extends Error {
  type: 'timeout' | 'connection' | 'abort'
  request?: RequestConfig
  
  constructor(type: string, message: string, request?: RequestConfig)
}

// 用户友好错误
interface UserFriendlyError {
  title: string
  message: string
  action?: {
    label: string
    handler: () => void
  }
  canRetry: boolean
}
```

## 正确性属性

*属性是一个特征或行为，应该在系统的所有有效执行中保持为真——本质上，是关于系统应该做什么的正式声明。属性作为人类可读规范和机器可验证正确性保证之间的桥梁。*

基于预工作分析，以下是需要验证的正确性属性：

### 属性反思

在生成具体属性之前，让我识别和消除冗余：

**冗余分析：**
1. **响应格式验证属性**：多个属性都在验证响应格式的一致性，可以合并为更全面的属性
2. **错误处理属性**：错误处理相关的属性可以合并，避免重复验证相同的错误处理逻辑
3. **分页处理属性**：分页相关的多个属性可以合并为一个综合的分页处理属性
4. **API 客户端属性**：API 客户端的多个行为可以合并为更全面的客户端行为属性

**合并后的属性：**

**属性 1: 统一响应格式一致性**
*对于任何* API 响应，无论成功或失败，都应该符合新的 V2 响应格式规范，包含正确的字段结构和类型
**验证需求: 1.1, 1.2, 1.3, 2.2**

**属性 2: 错误处理统一性**
*对于任何* API 错误场景，错误处理器应该优先使用新格式的错误信息，提供一致的错误提示和恢复机制
**验证需求: 2.3, 5.1, 5.2, 5.3, 5.4, 5.5**

**属性 3: 分页处理现代化**
*对于任何* 分页请求和响应，应该使用新的 page/page_size 参数格式和 pagination 对象结构，而不是旧的 limit/offset 和 records 结构
**验证需求: 1.4, 6.1, 6.2, 6.3, 6.4, 6.5**

**属性 4: API 客户端统一性**
*对于任何* HTTP 请求，都应该通过统一的 API 客户端发起，自动处理请求头、响应解析和错误处理
**验证需求: 2.1, 2.4, 2.5**

**属性 5: 书籍 API 现代化**
*对于任何* 书籍相关的 API 操作（获取、更新、删除、搜索），都应该使用新的响应格式和错误处理机制
**验证需求: 3.1, 3.2, 3.3, 3.4, 3.5**

**属性 6: 元数据搜索一致性**
*对于任何* 元数据搜索操作（ISBN 或关键词），都应该使用统一的错误处理和响应格式处理
**验证需求: 4.1, 4.2, 4.3, 4.4, 4.5**

**属性 7: 任务和聊天 API 统一性**
*对于任何* 任务管理或聊天相关的 API 操作，都应该使用统一的 API 客户端和标准化的响应处理
**验证需求: 8.1, 8.2, 8.3, 8.4, 8.5**

**属性 8: 文件操作一致性**
*对于任何* 文件相关操作（目录、内容、下载、封面），都应该使用统一的错误处理和响应格式
**验证需求: 9.1, 9.2, 9.3, 9.4, 9.5**

**属性 9: 向后兼容性支持**
*对于任何* 混合格式的 API 响应，适配器应该能够自动识别格式并正确转换为统一格式
**验证需求: 11.1, 11.2, 11.3, 11.4**

**属性 10: 性能优化有效性**
*对于任何* API 请求，缓存、批量处理和预加载机制应该有效减少网络开销和提升响应速度
**验证需求: 12.1, 12.2, 12.3, 12.4**

## 错误处理

### 错误分类

1. **API 错误 (4xx, 5xx)**
   - 客户端错误：400, 401, 403, 404, 422
   - 服务器错误：500, 502, 503, 504

2. **网络错误**
   - 连接超时
   - 网络中断
   - DNS 解析失败

3. **业务逻辑错误**
   - 数据验证失败
   - 业务规则违反
   - 状态不一致

### 错误处理策略

#### 1. 分层错误处理

```typescript
// 1. API 客户端层 - 处理 HTTP 错误
class UnifiedApiClient {
  private async handleResponse<T>(response: Response): Promise<ApiResponse<T>> {
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new ApiError(response.status, errorData, response.url)
    }
    
    const data = await response.json()
    return this.validateResponseFormat(data)
  }
}

// 2. 服务层 - 处理业务逻辑错误
class BookService extends BaseApiService {
  async getBook(id: string): Promise<Book> {
    try {
      return await this.client.get<Book>(`/api/book/${id}`)
    } catch (error) {
      if (error instanceof ApiError && error.code === 404) {
        throw new BookNotFoundError(id)
      }
      throw error
    }
  }
}

// 3. 组件层 - 处理用户界面错误
function BookDetail({ id }: { id: string }) {
  const [error, setError] = useState<UserFriendlyError | null>(null)
  
  const handleError = useCallback((error: Error) => {
    const friendlyError = errorHandler.toUserFriendlyError(error)
    setError(friendlyError)
  }, [])
}
```

#### 2. 错误恢复机制

```typescript
// 自动重试
class RetryHandler {
  async executeWithRetry<T>(
    operation: () => Promise<T>,
    config: RetryConfig
  ): Promise<T> {
    let lastError: Error
    
    for (let attempt = 1; attempt <= config.maxAttempts; attempt++) {
      try {
        return await operation()
      } catch (error) {
        lastError = error
        
        if (!this.shouldRetry(error, attempt, config)) {
          break
        }
        
        await this.delay(this.calculateDelay(attempt, config))
      }
    }
    
    throw lastError
  }
}

// 错误边界
class ApiErrorBoundary extends React.Component {
  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    if (error instanceof ApiError) {
      // 处理 API 错误
      this.handleApiError(error)
    } else {
      // 处理其他错误
      this.handleUnexpectedError(error, errorInfo)
    }
  }
}
```

### 用户体验优化

#### 1. 错误提示设计

```typescript
// 错误提示组件
interface ErrorDisplayProps {
  error: UserFriendlyError
  onRetry?: () => void
  onDismiss?: () => void
}

function ErrorDisplay({ error, onRetry, onDismiss }: ErrorDisplayProps) {
  return (
    <Alert variant="destructive">
      <AlertCircle className="h-4 w-4" />
      <AlertTitle>{error.title}</AlertTitle>
      <AlertDescription>
        {error.message}
        {error.canRetry && onRetry && (
          <Button variant="outline" size="sm" onClick={onRetry} className="mt-2">
            重试
          </Button>
        )}
      </AlertDescription>
    </Alert>
  )
}
```

#### 2. 加载状态管理

```typescript
// 统一的加载状态 Hook
function useApiCall<T>(
  apiCall: () => Promise<T>,
  dependencies: any[] = []
) {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<UserFriendlyError | null>(null)
  
  const execute = useCallback(async () => {
    setLoading(true)
    setError(null)
    
    try {
      const result = await apiCall()
      setData(result)
    } catch (err) {
      const friendlyError = errorHandler.toUserFriendlyError(err)
      setError(friendlyError)
    } finally {
      setLoading(false)
    }
  }, dependencies)
  
  return { data, loading, error, execute, retry: execute }
}
```

## 测试策略

### 双重测试方法

本项目采用单元测试和属性测试相结合的综合测试策略：

- **单元测试**：验证具体示例、边缘情况和错误条件
- **属性测试**：验证应该在所有输入中保持的通用属性
- **集成测试**：验证组件间的交互和端到端流程

### 单元测试策略

#### 1. API 客户端测试

```typescript
// 测试成功响应处理
describe('UnifiedApiClient', () => {
  test('should handle successful response', async () => {
    const mockResponse = {
      code: 200,
      message: 'success',
      data: { id: 1, title: 'Test Book' }
    }
    
    fetchMock.mockResponseOnce(JSON.stringify(mockResponse))
    
    const client = new UnifiedApiClient()
    const result = await client.get<Book>('/api/book/1')
    
    expect(result).toEqual(mockResponse.data)
  })
  
  // 测试错误响应处理
  test('should handle error response', async () => {
    const mockErrorResponse = {
      code: 404,
      message: 'error',
      error: {
        code: 'BOOK_NOT_FOUND',
        message: 'Book not found',
        details: 'Book with ID 1 does not exist'
      }
    }
    
    fetchMock.mockResponseOnce(
      JSON.stringify(mockErrorResponse),
      { status: 404 }
    )
    
    const client = new UnifiedApiClient()
    
    await expect(client.get('/api/book/1')).rejects.toThrow(ApiError)
  })
})
```

#### 2. 错误处理测试

```typescript
describe('ErrorHandler', () => {
  test('should prioritize new error format', () => {
    const newFormatError = {
      code: 404,
      message: 'error',
      error: {
        code: 'BOOK_NOT_FOUND',
        message: 'Book not found'
      }
    }
    
    const handler = new ErrorHandler()
    const friendlyError = handler.toUserFriendlyError(new ApiError(newFormatError))
    
    expect(friendlyError.message).toBe('Book not found')
  })
})
```

### 属性测试策略

我们将使用 **fast-check** 作为 TypeScript 的属性测试库，配置每个属性测试运行最少 100 次迭代。

#### 1. 响应格式属性测试

```typescript
import fc from 'fast-check'

describe('API Response Format Properties', () => {
  // **Feature: frontend-api-interface-cleanup, Property 1: 统一响应格式一致性**
  test('all API responses should conform to V2 format', () => {
    fc.assert(fc.property(
      fc.record({
        code: fc.integer({ min: 200, max: 599 }),
        message: fc.string(),
        data: fc.anything(),
        error: fc.option(fc.record({
          code: fc.string(),
          message: fc.string(),
          details: fc.option(fc.string()),
          context: fc.option(fc.dictionary(fc.string(), fc.anything()))
        }))
      }),
      (response) => {
        const validator = new ResponseValidator()
        expect(() => validator.validateV2Format(response)).not.toThrow()
      }
    ), { numRuns: 100 })
  })
})
```

#### 2. 错误处理属性测试

```typescript
describe('Error Handling Properties', () => {
  // **Feature: frontend-api-interface-cleanup, Property 2: 错误处理统一性**
  test('error handler should consistently process all error types', () => {
    fc.assert(fc.property(
      fc.oneof(
        // API 错误
        fc.record({
          code: fc.integer({ min: 400, max: 599 }),
          error: fc.record({
            code: fc.string(),
            message: fc.string()
          })
        }),
        // 网络错误
        fc.record({
          type: fc.constantFrom('timeout', 'connection', 'abort'),
          message: fc.string()
        })
      ),
      (errorData) => {
        const handler = new ErrorHandler()
        const result = handler.toUserFriendlyError(errorData)
        
        // 所有错误都应该有用户友好的消息
        expect(result.message).toBeTruthy()
        expect(typeof result.message).toBe('string')
        expect(result.title).toBeTruthy()
        expect(typeof result.canRetry).toBe('boolean')
      }
    ), { numRuns: 100 })
  })
})
```

#### 3. 分页处理属性测试

```typescript
describe('Pagination Properties', () => {
  // **Feature: frontend-api-interface-cleanup, Property 3: 分页处理现代化**
  test('pagination should use new format consistently', () => {
    fc.assert(fc.property(
      fc.record({
        page: fc.integer({ min: 1, max: 1000 }),
        page_size: fc.integer({ min: 1, max: 100 })
      }),
      (paginationParams) => {
        const service = new BookService(mockClient, mockErrorHandler)
        const request = service.buildPaginationRequest(paginationParams)
        
        // 应该使用新的分页参数格式
        expect(request.params).toHaveProperty('page')
        expect(request.params).toHaveProperty('page_size')
        expect(request.params).not.toHaveProperty('limit')
        expect(request.params).not.toHaveProperty('offset')
      }
    ), { numRuns: 100 })
  })
})
```

### 测试配置

#### Jest 配置

```javascript
// jest.config.js
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/src/test/setup.ts'],
  testMatch: [
    '<rootDir>/src/**/__tests__/**/*.{ts,tsx}',
    '<rootDir>/src/**/*.{test,spec}.{ts,tsx}'
  ],
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.d.ts',
    '!src/test/**/*'
  ],
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80
    }
  }
}
```

#### Fast-check 配置

```typescript
// src/test/property-test-config.ts
import fc from 'fast-check'

export const propertyTestConfig = {
  numRuns: 100, // 每个属性测试运行 100 次
  timeout: 5000, // 5 秒超时
  seed: 42, // 固定种子以确保可重现性
  verbose: true // 详细输出
}

// 自定义生成器
export const generators = {
  apiResponse: <T>(dataGen: fc.Arbitrary<T>) => fc.record({
    code: fc.integer({ min: 200, max: 599 }),
    message: fc.string(),
    data: fc.option(dataGen),
    error: fc.option(fc.record({
      code: fc.string(),
      message: fc.string(),
      details: fc.option(fc.string())
    }))
  }),
  
  book: () => fc.record({
    id: fc.integer({ min: 1 }),
    title: fc.string({ minLength: 1 }),
    authors: fc.array(fc.string({ minLength: 1 }), { minLength: 1 }),
    isbn: fc.string(),
    publisher: fc.string(),
    pubdate: fc.date().map(d => d.toISOString()),
    rating: fc.float({ min: 0, max: 10 }),
    tags: fc.array(fc.string()),
    comments: fc.string()
  })
}
```

### 测试执行要求

1. **属性测试标注**：每个属性测试必须包含注释，明确标识对应的设计文档属性
2. **迭代次数**：所有属性测试配置为运行至少 100 次迭代
3. **测试隔离**：每个测试应该独立运行，不依赖其他测试的状态
4. **模拟数据**：使用智能生成器创建符合业务规则的测试数据
5. **覆盖率要求**：单元测试覆盖率不低于 80%，属性测试覆盖核心业务逻辑