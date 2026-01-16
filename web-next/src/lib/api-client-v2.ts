/**
 * Unified API Client (V2)
 * 
 * This module provides a unified HTTP client for making API requests.
 * It handles request/response interceptors, error handling, retries, and caching.
 * 
 * @example
 * ```typescript
 * const client = new UnifiedApiClient({ baseURL: '/api' })
 * 
 * // Simple GET request
 * const book = await client.get<Book>('/book/1')
 * 
 * // POST request with data
 * const newBook = await client.post<Book>('/book', bookData)
 * 
 * // With custom config
 * const books = await client.get<Book[]>('/books', {
 *   timeout: 5000,
 *   useCache: true,
 *   cacheTTL: 60000
 * })
 * ```
 */

import {
  ApiResponse,
  ApiError,
  NetworkError,
  RequestConfig,
  RequestInterceptor,
  ResponseInterceptor,
  InterceptorManager,
} from '@/types/api-v2'

// ============================================================================
// API Client Configuration
// ============================================================================

export interface ApiClientConfig {
  /** Base URL for all requests */
  baseURL?: string
  /** Default timeout in milliseconds */
  timeout?: number
  /** Default number of retry attempts */
  retryAttempts?: number
  /** Whether to enable caching by default */
  cacheEnabled?: boolean
  /** Default headers */
  headers?: Record<string, string>
}

// ============================================================================
// Interceptor Manager Implementation
// ============================================================================

class InterceptorManagerImpl<T> implements InterceptorManager<T> {
  private interceptors: Array<((value: T) => T | Promise<T>) | null> = []

  use(interceptor: (value: T) => T | Promise<T>): number {
    this.interceptors.push(interceptor)
    return this.interceptors.length - 1
  }

  eject(id: number): void {
    if (this.interceptors[id]) {
      this.interceptors[id] = null
    }
  }

  async execute(value: T): Promise<T> {
    let result = value
    for (const interceptor of this.interceptors) {
      if (interceptor) {
        result = await interceptor(result)
      }
    }
    return result
  }
}

// ============================================================================
// Unified API Client
// ============================================================================

export class UnifiedApiClient {
  private config: Required<ApiClientConfig>

  /** Request interceptors */
  public interceptors = {
    request: new InterceptorManagerImpl<RequestConfig>(),
    response: new InterceptorManagerImpl<ApiResponse<any>>(),
  }

  constructor(config: ApiClientConfig = {}) {
    this.config = {
      baseURL: config.baseURL || '',
      timeout: config.timeout || 30000,
      retryAttempts: config.retryAttempts || 0,
      cacheEnabled: config.cacheEnabled || false,
      headers: config.headers || {},
    }
  }

  // ==========================================================================
  // Core Request Method
  // ==========================================================================

  /**
   * Make an HTTP request
   * 
   * @template T - The expected response data type
   * @param config - Request configuration
   * @returns Promise resolving to the API response
   * 
   * @throws {ApiError} When the API returns an error response
   * @throws {NetworkError} When a network-level error occurs
   */
  async request<T>(config: RequestConfig): Promise<ApiResponse<T>> {
    // Apply request interceptors
    const processedConfig = await this.interceptors.request.execute(config)

    // Build full URL
    const url = this.buildUrl(processedConfig.url || '', processedConfig)

    // Prepare request options
    const options: RequestInit = {
      ...processedConfig,
      headers: {
        'Content-Type': 'application/json',
        ...this.config.headers,
        ...processedConfig.headers,
      },
    }

    // Execute request with timeout and retries
    const response = await this.executeWithRetry(() =>
      this.fetchWithTimeout(url, options, processedConfig.timeout || this.config.timeout),
      processedConfig.retryAttempts ?? this.config.retryAttempts
    )

    // Parse and validate response
    const apiResponse = await this.parseResponse<T>(response, processedConfig)

    // Apply response interceptors
    const processedResponse = await this.interceptors.response.execute(apiResponse)

    // Check for API errors
    if (processedResponse.code >= 400) {
      throw new ApiError(processedResponse, processedConfig)
    }

    return processedResponse
  }

  // ==========================================================================
  // Convenience Methods
  // ==========================================================================

  /**
   * Make a GET request
   * 
   * @template T - The expected response data type
   * @param url - Request URL (relative to baseURL)
   * @param config - Optional request configuration
   * @returns Promise resolving to the response data
   */
  async get<T>(url: string, config?: RequestConfig): Promise<T> {
    const response = await this.request<T>({
      ...config,
      url,
      method: 'GET',
    })
    return response.data as T
  }

  /**
   * Make a POST request
   * 
   * @template T - The expected response data type
   * @param url - Request URL (relative to baseURL)
   * @param data - Request body data
   * @param config - Optional request configuration
   * @returns Promise resolving to the response data
   */
  async post<T>(url: string, data?: any, config?: RequestConfig): Promise<T> {
    const response = await this.request<T>({
      ...config,
      url,
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    })
    return response.data as T
  }

  /**
   * Make a PUT request
   * 
   * @template T - The expected response data type
   * @param url - Request URL (relative to baseURL)
   * @param data - Request body data
   * @param config - Optional request configuration
   * @returns Promise resolving to the response data
   */
  async put<T>(url: string, data?: any, config?: RequestConfig): Promise<T> {
    const response = await this.request<T>({
      ...config,
      url,
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    })
    return response.data as T
  }

  /**
   * Make a DELETE request
   * 
   * @template T - The expected response data type
   * @param url - Request URL (relative to baseURL)
   * @param config - Optional request configuration
   * @returns Promise resolving to the response data
   */
  async delete<T>(url: string, config?: RequestConfig): Promise<T> {
    const response = await this.request<T>({
      ...config,
      url,
      method: 'DELETE',
    })
    return response.data as T
  }

  // ==========================================================================
  // Private Helper Methods
  // ==========================================================================

  /**
   * Build full URL from base URL and path
   */
  private buildUrl(path: string, config: RequestConfig): string {
    // If path is absolute URL, return as-is
    if (path.startsWith('http://') || path.startsWith('https://')) {
      return path
    }

    // Combine base URL and path
    const baseURL = config.baseURL || this.config.baseURL
    const url = `${baseURL}${path}`

    return url
  }

  /**
   * Execute fetch with timeout
   */
  private async fetchWithTimeout(
    url: string,
    options: RequestInit,
    timeout: number
  ): Promise<Response> {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), timeout)

    try {
      const response = await fetch(url, {
        ...options,
        signal: controller.signal,
      })
      clearTimeout(timeoutId)
      return response
    } catch (error) {
      clearTimeout(timeoutId)

      if (error instanceof Error) {
        if (error.name === 'AbortError') {
          throw new NetworkError('timeout', `Request timeout after ${timeout}ms`, options)
        }
        throw new NetworkError('connection', error.message, options)
      }

      throw error
    }
  }

  /**
   * Execute operation with retry logic
   */
  private async executeWithRetry<T>(
    operation: () => Promise<T>,
    maxAttempts: number
  ): Promise<T> {
    let lastError: Error | undefined

    for (let attempt = 0; attempt <= maxAttempts; attempt++) {
      try {
        return await operation()
      } catch (error) {
        lastError = error as Error

        // Don't retry on the last attempt
        if (attempt >= maxAttempts) {
          break
        }

        // Don't retry client errors (4xx)
        if (error instanceof ApiError && error.code >= 400 && error.code < 500) {
          break
        }

        // Wait before retrying (exponential backoff)
        const delay = Math.min(1000 * Math.pow(2, attempt), 10000)
        await this.sleep(delay)
      }
    }

    throw lastError
  }

  /**
   * Parse response body and validate format
   */
  private async parseResponse<T>(
    response: Response,
    config: RequestConfig
  ): Promise<ApiResponse<T>> {
    try {
      const json = await response.json()

      // Validate V2 format
      if (!this.isValidV2Response(json)) {
        console.warn('Invalid API response format:', json)

        // Attempt to adapt legacy format
        return this.adaptLegacyResponse<T>(json, response.status)
      }

      return json as ApiResponse<T>
    } catch (error) {
      // If JSON parsing fails, treat as error response
      throw new NetworkError(
        'connection',
        'Failed to parse response body',
        config
      )
    }
  }

  /**
   * Validate if response conforms to V2 format
   */
  private isValidV2Response(response: any): boolean {
    return (
      typeof response === 'object' &&
      response !== null &&
      typeof response.code === 'number' &&
      typeof response.message === 'string' &&
      (response.data !== undefined || response.error !== undefined)
    )
  }

  /**
   * Adapt legacy response format to V2
   */
  private adaptLegacyResponse<T>(legacyResponse: any, statusCode: number): ApiResponse<T> {
    // Try to extract data from various legacy formats
    let data: T | undefined
    let error: any

    if (legacyResponse.data !== undefined) {
      data = legacyResponse.data
    } else if (Array.isArray(legacyResponse)) {
      data = legacyResponse as any
    } else if (legacyResponse.records !== undefined) {
      // Legacy paginated format
      data = {
        records: legacyResponse.records,
        total: legacyResponse.total,
        next_cursor: legacyResponse.next_cursor,
      } as any
    }

    if (legacyResponse.error) {
      error = {
        code: 'LEGACY_ERROR',
        message: typeof legacyResponse.error === 'string'
          ? legacyResponse.error
          : legacyResponse.error.message || 'Unknown error',
      }
    }

    return {
      code: legacyResponse.code || statusCode,
      message: legacyResponse.message || (statusCode >= 400 ? 'error' : 'success'),
      data,
      error,
    }
  }

  /**
   * Sleep helper for retry delays
   */
  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms))
  }
}

// ============================================================================
// Default Client Instance
// ============================================================================

/**
 * Default API client instance
 * 
 * This instance can be used throughout the application for making API requests.
 * 
 * @example
 * ```typescript
 * import { apiClient } from '@/lib/api-client-v2'
 * 
 * const book = await apiClient.get<Book>('/api/book/1')
 * ```
 */
export const apiClient = new UnifiedApiClient({
  baseURL: '',
  timeout: 30000,
  retryAttempts: 2,
})

// Add default request interceptor to log requests in development
if (process.env.NODE_ENV === 'development') {
  apiClient.interceptors.request.use((config) => {
    console.log('[API Request]', config.method, config.url)
    return config
  })

  apiClient.interceptors.response.use((response) => {
    console.log('[API Response]', response.code, response.message)
    return response
  })
}

