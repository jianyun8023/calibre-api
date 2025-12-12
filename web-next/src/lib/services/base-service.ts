/**
 * Base API Service
 * 
 * This abstract class provides common functionality for all API services.
 * It handles request execution, error handling, URL building, and response validation.
 * 
 * @example
 * ```typescript
 * class BookService extends BaseApiService {
 *   async getBook(id: string): Promise<Book> {
 *     return this.handleRequest(() => 
 *       this.client.get<Book>(`/api/book/${id}`)
 *     )
 *   }
 * }
 * ```
 */

import { UnifiedApiClient } from '../api-client-v2'
import { ErrorHandler } from '../error-handler'
import { ApiResponse, RequestConfig, PaginationParams, CursorPaginationParams } from '@/types/api-v2'

// ============================================================================
// Base API Service Class
// ============================================================================

export abstract class BaseApiService {
  protected client: UnifiedApiClient
  protected errorHandler: ErrorHandler

  constructor(client: UnifiedApiClient, errorHandler: ErrorHandler) {
    this.client = client
    this.errorHandler = errorHandler
  }

  // ==========================================================================
  // Request Handling
  // ==========================================================================

  /**
   * Handle request with unified error handling
   * 
   * @template T - The expected response type
   * @param request - Function that returns a promise
   * @returns Promise resolving to the response data
   */
  protected async handleRequest<T>(
    request: () => Promise<T>
  ): Promise<T> {
    try {
      return await request()
    } catch (error) {
      // Convert error to user-friendly format
      const friendlyError = this.errorHandler.toUserFriendlyError(error)
      
      // Log error in development
      if (process.env.NODE_ENV === 'development') {
        console.error('[Service Error]', friendlyError)
      }

      // Re-throw the original error for the caller to handle
      throw error
    }
  }

  // ==========================================================================
  // URL Building
  // ==========================================================================

  /**
   * Build URL with query parameters
   * 
   * @param path - Base path
   * @param params - Query parameters
   * @returns Full URL with query string
   * 
   * @example
   * ```typescript
   * const url = this.buildUrl('/api/books', { page: 1, limit: 20 })
   * // Returns: '/api/books?page=1&limit=20'
   * ```
   */
  protected buildUrl(path: string, params?: Record<string, any>): string {
    if (!params || Object.keys(params).length === 0) {
      return path
    }

    const queryString = Object.entries(params)
      .filter(([_, value]) => value !== undefined && value !== null && value !== '')
      .map(([key, value]) => {
        if (Array.isArray(value)) {
          // Handle array parameters
          return value.map(v => `${encodeURIComponent(key)}=${encodeURIComponent(v)}`).join('&')
        }
        return `${encodeURIComponent(key)}=${encodeURIComponent(value)}`
      })
      .join('&')

    return queryString ? `${path}?${queryString}` : path
  }

  /**
   * Build pagination query parameters
   * 
   * @param pagination - Pagination parameters
   * @returns Query parameters object
   */
  protected buildPaginationParams(pagination: PaginationParams): Record<string, any> {
    return {
      page: pagination.page,
      page_size: pagination.page_size,
    }
  }

  /**
   * Build cursor pagination query parameters
   * 
   * @param pagination - Cursor pagination parameters
   * @returns Query parameters object
   */
  protected buildCursorPaginationParams(pagination: CursorPaginationParams): Record<string, any> {
    return {
      limit: pagination.limit,
      cursor: pagination.cursor,
    }
  }

  // ==========================================================================
  // Response Validation
  // ==========================================================================

  /**
   * Validate API response
   * 
   * @template T - The expected data type
   * @param response - API response to validate
   * @returns The data from the response
   * @throws {Error} If response is invalid or contains errors
   */
  protected validateResponse<T>(response: ApiResponse<T>): T {
    if (response.code >= 400) {
      throw new Error(response.error?.message || response.message || 'Request failed')
    }

    if (!response.data) {
      throw new Error('Response data is missing')
    }

    return response.data
  }

  // ==========================================================================
  // Data Transformation
  // ==========================================================================

  /**
   * Transform request data before sending
   * 
   * Override this method in subclasses to apply custom transformations.
   * 
   * @param data - Raw request data
   * @returns Transformed data
   */
  protected transformRequestData<T>(data: T): T {
    return data
  }

  /**
   * Transform response data after receiving
   * 
   * Override this method in subclasses to apply custom transformations.
   * 
   * @param data - Raw response data
   * @returns Transformed data
   */
  protected transformResponseData<T>(data: T): T {
    return data
  }

  // ==========================================================================
  // Utility Methods
  // ==========================================================================

  /**
   * Convert legacy pagination params to new format
   * 
   * @deprecated This is for backward compatibility only
   */
  protected convertLegacyPagination(limit: number, offset: number): PaginationParams {
    const page = Math.floor(offset / limit) + 1
    return {
      page,
      page_size: limit,
    }
  }

  /**
   * Convert new pagination params to legacy format
   * 
   * @deprecated This is for backward compatibility only
   */
  protected convertToLegacyPagination(pagination: PaginationParams): { limit: number; offset: number } {
    return {
      limit: pagination.page_size,
      offset: (pagination.page - 1) * pagination.page_size,
    }
  }

  /**
   * Retry a request with exponential backoff
   * 
   * @template T - The expected response type
   * @param request - Function that returns a promise
   * @param maxAttempts - Maximum number of retry attempts
   * @returns Promise resolving to the response data
   */
  protected async retryRequest<T>(
    request: () => Promise<T>,
    maxAttempts: number = 3
  ): Promise<T> {
    let lastError: Error | undefined

    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      try {
        return await request()
      } catch (error) {
        lastError = error as Error

        // Check if error is retryable
        if (!this.errorHandler.canRetry(error)) {
          throw error
        }

        // Don't wait after the last attempt
        if (attempt < maxAttempts - 1) {
          const delay = this.errorHandler.getRetryDelay(attempt)
          await this.sleep(delay)
        }
      }
    }

    throw lastError
  }

  /**
   * Sleep for specified milliseconds
   */
  protected sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms))
  }

  // ==========================================================================
  // Request Configuration Helpers
  // ==========================================================================

  /**
   * Create request config with caching
   */
  protected withCache(ttl?: number): RequestConfig {
    return {
      useCache: true,
      cacheTTL: ttl,
    }
  }

  /**
   * Create request config with custom timeout
   */
  protected withTimeout(timeout: number): RequestConfig {
    return {
      timeout,
    }
  }

  /**
   * Create request config with retry
   */
  protected withRetry(attempts: number): RequestConfig {
    return {
      retryAttempts: attempts,
    }
  }

  /**
   * Combine multiple request configs
   */
  protected combineConfigs(...configs: RequestConfig[]): RequestConfig {
    return configs.reduce((acc, config) => ({
      ...acc,
      ...config,
      headers: {
        ...acc.headers,
        ...config.headers,
      },
    }), {})
  }
}

