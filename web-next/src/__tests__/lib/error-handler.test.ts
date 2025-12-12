/**
 * ErrorHandler Tests
 * 
 * Test suite for the error handler.
 */

import { ErrorHandler } from '@/lib/error-handler'
import { ApiError, NetworkError } from '@/types/api-v2'

describe('ErrorHandler', () => {
  let errorHandler: ErrorHandler

  beforeEach(() => {
    errorHandler = new ErrorHandler({
      logErrors: false, // Disable logging in tests
      reportErrors: false,
    })
  })

  describe('toUserFriendlyError', () => {
    it('should convert ApiError to user-friendly error', () => {
      const apiError = new ApiError({
        code: 404,
        message: 'error',
        error: {
          code: 'BOOK_NOT_FOUND',
          message: 'Book not found',
        },
      })

      const friendlyError = errorHandler.toUserFriendlyError(apiError)

      expect(friendlyError.title).toBe('请求错误')
      expect(friendlyError.message).toContain('找不到指定的书籍')
      expect(friendlyError.canRetry).toBe(false)
    })

    it('should convert NetworkError to user-friendly error', () => {
      const networkError = new NetworkError('timeout', 'Request timeout')

      const friendlyError = errorHandler.toUserFriendlyError(networkError)

      expect(friendlyError.title).toBe('请求超时')
      expect(friendlyError.message).toContain('服务器响应时间过长')
      expect(friendlyError.canRetry).toBe(true)
    })

    it('should convert generic Error to user-friendly error', () => {
      const error = new Error('Something went wrong')

      const friendlyError = errorHandler.toUserFriendlyError(error)

      expect(friendlyError.title).toBe('操作失败')
      expect(friendlyError.message).toBe('Something went wrong')
      expect(friendlyError.canRetry).toBe(true)
    })

    it('should handle unknown error types', () => {
      const unknownError = 'string error'

      const friendlyError = errorHandler.toUserFriendlyError(unknownError)

      expect(friendlyError.title).toBe('未知错误')
      expect(friendlyError.message).toBeDefined()
      expect(friendlyError.canRetry).toBe(true)
    })
  })

  describe('canRetry', () => {
    it('should return false for 4xx errors (except 408, 429)', () => {
      const error400 = new ApiError({ code: 400, message: 'Bad Request' })
      const error404 = new ApiError({ code: 404, message: 'Not Found' })

      expect(errorHandler.canRetry(error400)).toBe(false)
      expect(errorHandler.canRetry(error404)).toBe(false)
    })

    it('should return true for 408 and 429 errors', () => {
      const error408 = new ApiError({ code: 408, message: 'Request Timeout' })
      const error429 = new ApiError({ code: 429, message: 'Too Many Requests' })

      expect(errorHandler.canRetry(error408)).toBe(true)
      expect(errorHandler.canRetry(error429)).toBe(true)
    })

    it('should return true for 5xx errors', () => {
      const error500 = new ApiError({ code: 500, message: 'Internal Server Error' })
      const error503 = new ApiError({ code: 503, message: 'Service Unavailable' })

      expect(errorHandler.canRetry(error500)).toBe(true)
      expect(errorHandler.canRetry(error503)).toBe(true)
    })

    it('should return false for abort network errors', () => {
      const abortError = new NetworkError('abort', 'Request aborted')

      expect(errorHandler.canRetry(abortError)).toBe(false)
    })

    it('should return true for other network errors', () => {
      const timeoutError = new NetworkError('timeout', 'Request timeout')
      const connectionError = new NetworkError('connection', 'Connection failed')

      expect(errorHandler.canRetry(timeoutError)).toBe(true)
      expect(errorHandler.canRetry(connectionError)).toBe(true)
    })
  })

  describe('getRetryDelay', () => {
    it('should use exponential backoff', () => {
      const delay1 = errorHandler.getRetryDelay(0)
      const delay2 = errorHandler.getRetryDelay(1)
      const delay3 = errorHandler.getRetryDelay(2)

      expect(delay1).toBe(1000) // 2^0 * 1000
      expect(delay2).toBe(2000) // 2^1 * 1000
      expect(delay3).toBe(4000) // 2^2 * 1000
    })

    it('should cap delay at 10 seconds', () => {
      const delay = errorHandler.getRetryDelay(10)

      expect(delay).toBeLessThanOrEqual(10000)
    })
  })

  describe('Error code mapping', () => {
    it('should map specific error codes to Chinese messages', () => {
      const testCases = [
        { code: 'BOOK_NOT_FOUND', expected: '找不到指定的书籍' },
        { code: 'SEARCH_FAILED', expected: '搜索失败' },
        { code: 'DATABASE_ERROR', expected: '数据库操作失败' },
        { code: 'RATE_LIMIT_EXCEEDED', expected: '请求过于频繁' },
      ]

      testCases.forEach(({ code, expected }) => {
        const apiError = new ApiError({
          code: 400,
          message: 'error',
          error: {
            code,
            message: 'Error message',
          },
        })

        const friendlyError = errorHandler.toUserFriendlyError(apiError)
        expect(friendlyError.message).toContain(expected)
      })
    })
  })
})

