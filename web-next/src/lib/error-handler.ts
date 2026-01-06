/**
 * Error Handler
 * 
 * This module provides unified error handling for API errors and network errors.
 * It converts technical errors into user-friendly messages and provides retry logic.
 * 
 * @example
 * ```typescript
 * import { errorHandler } from '@/lib/error-handler'
 * 
 * try {
 *   await apiClient.get('/book/1')
 * } catch (error) {
 *   const friendlyError = errorHandler.toUserFriendlyError(error)
 *   toast.error(friendlyError.title, { description: friendlyError.message })
 * }
 * ```
 */

import { ApiError, NetworkError, UserFriendlyError, RequestConfig } from '@/types/api-v2'

// ============================================================================
// Error Context
// ============================================================================

export interface ErrorContext {
  /** Request URL */
  url: string
  /** HTTP method */
  method: string
  /** Request ID (optional) */
  requestId?: string
  /** Timestamp when error occurred */
  timestamp: number
}

// ============================================================================
// Error Handler Configuration
// ============================================================================

export interface ErrorHandlerConfig {
  /** Whether to log errors to console */
  logErrors?: boolean
  /** Whether to report errors to external service */
  reportErrors?: boolean
  /** Custom error reporter function */
  errorReporter?: (error: Error, context: ErrorContext) => void
}

// ============================================================================
// Error Handler Class
// ============================================================================

export class ErrorHandler {
  private config: Required<ErrorHandlerConfig>

  constructor(config: ErrorHandlerConfig = {}) {
    this.config = {
      logErrors: config.logErrors ?? true,
      reportErrors: config.reportErrors ?? false,
      errorReporter: config.errorReporter || this.defaultErrorReporter,
    }
  }

  // ==========================================================================
  // Main Error Handling Methods
  // ==========================================================================

  /**
   * Convert any error to a user-friendly error
   * 
   * @param error - The error to convert
   * @param context - Optional error context
   * @returns User-friendly error object
   */
  toUserFriendlyError(error: unknown, context?: Partial<ErrorContext>): UserFriendlyError {
    const fullContext = this.buildErrorContext(error, context)

    // Log error if configured
    if (this.config.logErrors) {
      console.error('[Error Handler]', error, fullContext)
    }

    // Report error if configured
    if (this.config.reportErrors && error instanceof Error) {
      this.config.errorReporter(error, fullContext)
    }

    // Convert to user-friendly error
    if (error instanceof ApiError) {
      return this.handleApiError(error, fullContext)
    }

    if (error instanceof NetworkError) {
      return this.handleNetworkError(error, fullContext)
    }

    if (error instanceof Error) {
      return this.handleGenericError(error, fullContext)
    }

    // Unknown error type
    return {
      title: '未知错误',
      message: '发生了未知错误，请稍后重试',
      canRetry: true,
    }
  }

  /**
   * Handle API errors (4xx, 5xx responses)
   */
  handleApiError(error: ApiError, context: ErrorContext): UserFriendlyError {
    const { code, errorCode, details } = error

    // Map error codes to user-friendly messages
    const errorMessage = this.getErrorMessageByCode(errorCode, code)
    
    // Determine if the error is retryable
    const canRetry = this.canRetry(error)

    return {
      title: this.getErrorTitle(code),
      message: errorMessage || error.message || '操作失败',
      canRetry,
      action: canRetry ? {
        label: '重试',
        handler: () => {
          // Handler will be provided by the caller
        },
      } : undefined,
    }
  }

  /**
   * Handle network errors (timeout, connection, etc.)
   */
  handleNetworkError(error: NetworkError, context: ErrorContext): UserFriendlyError {
    const errorMessages = {
      timeout: {
        title: '请求超时',
        message: '服务器响应时间过长，请检查网络连接后重试',
      },
      connection: {
        title: '网络错误',
        message: '无法连接到服务器，请检查网络连接',
      },
      abort: {
        title: '请求已取消',
        message: '请求被取消',
      },
    }

    const { title, message } = errorMessages[error.type] || errorMessages.connection

    return {
      title,
      message,
      canRetry: error.type !== 'abort',
    }
  }

  /**
   * Handle generic errors
   */
  handleGenericError(error: Error, context: ErrorContext): UserFriendlyError {
    return {
      title: '操作失败',
      message: error.message || '发生了未知错误',
      canRetry: true,
    }
  }

  // ==========================================================================
  // Error Analysis Methods
  // ==========================================================================

  /**
   * Determine if an error is retryable
   */
  canRetry(error: unknown): boolean {
    if (error instanceof ApiError) {
      // Client errors (4xx) are generally not retryable
      if (error.code >= 400 && error.code < 500) {
        // Except for these specific cases
        return [408, 429].includes(error.code)
      }

      // Server errors (5xx) are retryable
      return error.code >= 500
    }

    if (error instanceof NetworkError) {
      // Network errors are retryable (except abort)
      return error.type !== 'abort'
    }

    // Unknown errors are retryable
    return true
  }

  /**
   * Get retry delay in milliseconds
   * 
   * Uses exponential backoff strategy.
   */
  getRetryDelay(attempt: number): number {
    return Math.min(1000 * Math.pow(2, attempt), 10000)
  }

  // ==========================================================================
  // Error Message Mapping
  // ==========================================================================

  /**
   * Get user-friendly error message by error code
   */
  private getErrorMessageByCode(errorCode: string | undefined, httpCode: number): string {
    // Map specific error codes to messages
    const codeMessages: Record<string, string> = {
      // Book errors
      BOOK_NOT_FOUND: '找不到指定的书籍',
      BOOK_ALREADY_EXISTS: '书籍已存在',
      BOOK_OPERATION_FAILED: '书籍操作失败',
      
      // Search errors
      SEARCH_FAILED: '搜索失败，请稍后重试',
      INVALID_SEARCH_QUERY: '搜索条件无效',
      
      // File errors
      FILE_NOT_FOUND: '文件未找到',
      FILE_READ_ERROR: '读取文件失败',
      FILE_WRITE_ERROR: '写入文件失败',
      
      // Database errors
      DATABASE_ERROR: '数据库操作失败',
      DATABASE_CONNECTION_ERROR: '数据库连接失败',
      
      // Validation errors
      INVALID_INPUT: '输入数据无效',
      VALIDATION_ERROR: '数据验证失败',
      MISSING_REQUIRED_FIELD: '缺少必填字段',
      
      // Authentication errors
      UNAUTHORIZED: '未授权，请先登录',
      FORBIDDEN: '没有权限执行此操作',
      
      // Rate limiting
      RATE_LIMIT_EXCEEDED: '请求过于频繁，请稍后重试',
      
      // Server errors
      INTERNAL_SERVER_ERROR: '服务器内部错误',
      SERVICE_UNAVAILABLE: '服务暂时不可用',
      
      // External service errors
      EXTERNAL_SERVICE_ERROR: '外部服务错误',
      DOUBAN_API_ERROR: '豆瓣 API 错误',
      
      // Task errors
      TASK_NOT_FOUND: '任务未找到',
      TASK_ALREADY_RUNNING: '任务已在运行中',
      TASK_FAILED: '任务执行失败',
      
      // Chat errors
      CONVERSATION_NOT_FOUND: '对话未找到',
      MESSAGE_SEND_FAILED: '消息发送失败',
      
      // Legacy error
      LEGACY_ERROR: '操作失败',
    }

    if (errorCode && codeMessages[errorCode]) {
      return codeMessages[errorCode]
    }

    // Fall back to HTTP status code messages
    return this.getErrorMessageByHttpCode(httpCode)
  }

  /**
   * Get error message by HTTP status code
   */
  private getErrorMessageByHttpCode(code: number): string {
    if (code >= 400 && code < 500) {
      const clientErrors: Record<number, string> = {
        400: '请求参数错误',
        401: '未授权，请先登录',
        403: '没有权限执行此操作',
        404: '请求的资源不存在',
        408: '请求超时',
        409: '资源冲突',
        422: '数据验证失败',
        429: '请求过于频繁',
      }
      return clientErrors[code] || '客户端请求错误'
    }

    if (code >= 500) {
      const serverErrors: Record<number, string> = {
        500: '服务器内部错误',
        502: '网关错误',
        503: '服务暂时不可用',
        504: '网关超时',
      }
      return serverErrors[code] || '服务器错误'
    }

    return '操作失败'
  }

  /**
   * Get error title by HTTP status code
   */
  private getErrorTitle(code: number): string {
    if (code >= 400 && code < 500) {
      return '请求错误'
    }

    if (code >= 500) {
      return '服务器错误'
    }

    return '错误'
  }

  // ==========================================================================
  // Error Context Building
  // ==========================================================================

  /**
   * Build error context from error and partial context
   */
  private buildErrorContext(error: unknown, partial?: Partial<ErrorContext>): ErrorContext {
    let url = partial?.url || ''
    let method = partial?.method || 'UNKNOWN'

    // Extract context from error if available
    if (error instanceof ApiError && error.request) {
      url = url || error.request.url || ''
      method = method || error.request.method || 'UNKNOWN'
    }

    if (error instanceof NetworkError && error.request) {
      url = url || error.request.url || ''
      method = method || error.request.method || 'UNKNOWN'
    }

    return {
      url,
      method,
      requestId: partial?.requestId,
      timestamp: partial?.timestamp || Date.now(),
    }
  }

  // ==========================================================================
  // Error Reporting
  // ==========================================================================

  /**
   * Default error reporter (logs to console)
   */
  private defaultErrorReporter(error: Error, context: ErrorContext): void {
    console.error('[Error Report]', {
      error: {
        name: error.name,
        message: error.message,
        stack: error.stack,
      },
      context,
    })
  }

  /**
   * Report error to external service
   * 
   * This can be overridden to send errors to services like Sentry, LogRocket, etc.
   */
  reportError(error: Error, context: ErrorContext): void {
    this.config.errorReporter(error, context)
  }
}

// ============================================================================
// Default Error Handler Instance
// ============================================================================

/**
 * Default error handler instance
 * 
 * This instance can be used throughout the application for error handling.
 * 
 * @example
 * ```typescript
 * import { errorHandler } from '@/lib/error-handler'
 * 
 * try {
 *   await someApiCall()
 * } catch (error) {
 *   const friendlyError = errorHandler.toUserFriendlyError(error)
 *   console.log(friendlyError)
 * }
 * ```
 */
export const errorHandler = new ErrorHandler({
  logErrors: process.env.NODE_ENV === 'development',
  reportErrors: process.env.NODE_ENV === 'production',
})

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Create a retry handler function
 * 
 * @param operation - The operation to retry
 * @param maxAttempts - Maximum number of attempts
 * @returns Promise that resolves to the operation result
 * 
 * @example
 * ```typescript
 * const result = await withRetry(
 *   () => apiClient.get('/book/1'),
 *   3
 * )
 * ```
 */
export async function withRetry<T>(
  operation: () => Promise<T>,
  maxAttempts: number = 3
): Promise<T> {
  let lastError: Error | undefined

  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    try {
      return await operation()
    } catch (error) {
      lastError = error as Error

      // Check if error is retryable
      if (!errorHandler.canRetry(error)) {
        throw error
      }

      // Don't wait after the last attempt
      if (attempt < maxAttempts - 1) {
        const delay = errorHandler.getRetryDelay(attempt)
        await new Promise(resolve => setTimeout(resolve, delay))
      }
    }
  }

  throw lastError
}

