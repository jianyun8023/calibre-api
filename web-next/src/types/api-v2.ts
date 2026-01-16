/**
 * API V2 Type Definitions
 * 
 * This file contains all type definitions for the new V2 API format.
 * These types align with the backend's unified response format implemented in Spec 026.
 * 
 * @see docs/API_DOCUMENTATION.md for API documentation
 * @see specs/026-backend-code-optimization/ for backend implementation details
 */

// ============================================================================
// Base Response Types
// ============================================================================

/**
 * Standard API Response Format (V2)
 * 
 * All API endpoints should return this format for consistency.
 * 
 * @template T - The type of data returned in successful responses
 * 
 * @example
 * ```typescript
 * // Success response
 * {
 *   code: 200,
 *   message: "success",
 *   data: { id: 1, title: "Book Title" }
 * }
 * 
 * // Error response
 * {
 *   code: 404,
 *   message: "error",
 *   error: {
 *     code: "BOOK_NOT_FOUND",
 *     message: "Book not found",
 *     details: "Book with ID 1 does not exist"
 *   },
 *   trace_id: "abc123"
 * }
 * ```
 */
export interface ApiResponse<T> {
  /** HTTP status code */
  code: number
  /** Human-readable message */
  message: string
  /** Response data (present in successful responses) */
  data?: T
  /** Error information (present in error responses) */
  error?: ErrorInfo
  /** Trace ID for debugging (optional) */
  trace_id?: string
}

/**
 * Error Information
 * 
 * Provides structured error details for API error responses.
 */
export interface ErrorInfo {
  /** Error code (e.g., "BOOK_NOT_FOUND", "INVALID_INPUT") */
  code: string
  /** Human-readable error message */
  message: string
  /** Additional error details (optional) */
  details?: string
  /** Contextual information about the error (optional) */
  context?: Record<string, any>
}

// ============================================================================
// Pagination Types
// ============================================================================

/**
 * Paginated Response Format
 * 
 * Used for APIs that return paginated data with page numbers.
 * 
 * @template T - The type of items in the data array
 * 
 * @example
 * ```typescript
 * {
 *   code: 200,
 *   message: "success",
 *   data: [{ id: 1, title: "Book 1" }, { id: 2, title: "Book 2" }],
 *   pagination: {
 *     total: 100,
 *     page: 1,
 *     page_size: 20,
 *     total_pages: 5
 *   }
 * }
 * ```
 */
export interface PaginatedResponse<T> {
  code: number
  message: string
  data: T[]
  pagination: Pagination
}

/**
 * Pagination Metadata
 * 
 * Contains information about the current page and total pages.
 */
export interface Pagination {
  /** Total number of items */
  total: number
  /** Current page number (1-based) */
  page: number
  /** Number of items per page */
  page_size: number
  /** Total number of pages */
  total_pages: number
}

/**
 * Pagination Parameters
 * 
 * Parameters for requesting paginated data.
 */
export interface PaginationParams {
  /** Page number (1-based) */
  page: number
  /** Number of items per page */
  page_size: number
}

// ============================================================================
// Cursor Pagination Types
// ============================================================================

/**
 * Cursor Paginated Response Format
 * 
 * Used for APIs that return paginated data with cursor-based pagination.
 * This is useful for infinite scrolling scenarios.
 * 
 * @template T - The type of items in the data array
 * 
 * @example
 * ```typescript
 * {
 *   code: 200,
 *   message: "success",
 *   data: [{ id: 1, title: "Book 1" }],
 *   pagination: {
 *     total: 100,
 *     next_cursor: "eyJpZCI6MTB9",
 *     has_more: true
 *   }
 * }
 * ```
 */
export interface CursorPaginatedResponse<T> {
  code: number
  message: string
  data: T[]
  pagination: CursorPagination
}

/**
 * Cursor Pagination Metadata
 * 
 * Contains information about cursor-based pagination.
 */
export interface CursorPagination {
  /** Total number of items */
  total: number
  /** Cursor for the next page (if available) */
  next_cursor?: string
  /** Whether there are more items to load */
  has_more: boolean
}

/**
 * Cursor Pagination Parameters
 * 
 * Parameters for requesting cursor-paginated data.
 */
export interface CursorPaginationParams {
  /** Number of items to fetch */
  limit: number
  /** Cursor for pagination (optional, omit for first page) */
  cursor?: string
}

// ============================================================================
// Search Types
// ============================================================================

/**
 * Search Query Parameters
 * 
 * Parameters for performing a search operation.
 */
export interface SearchQuery {
  /** Search query string */
  q: string
  /** Search mode */
  mode?: 'hybrid' | 'semantic' | 'text'
  /** Filter conditions (optional) */
  filters?: string[]
  /** Sort conditions (optional) */
  sort?: string[]
  /** Pagination parameters */
  pagination: PaginationParams
}

// ============================================================================
// Request Configuration
// ============================================================================

/**
 * HTTP Request Configuration
 * 
 * Extended fetch options with additional configuration.
 */
export interface RequestConfig extends RequestInit {
  /** Request URL (relative to baseURL) */
  url?: string
  /** Base URL for the request */
  baseURL?: string
  /** Request timeout in milliseconds */
  timeout?: number
  /** Number of retry attempts */
  retryAttempts?: number
  /** Whether to use cache for this request */
  useCache?: boolean
  /** Cache TTL in milliseconds */
  cacheTTL?: number
  /** Custom error handler */
  onError?: (error: ApiError) => void
}

// ============================================================================
// Error Types
// ============================================================================

/**
 * API Error
 * 
 * Represents an error returned by the API.
 */
export class ApiError extends Error {
  /** HTTP status code */
  code: number
  /** Error code from API */
  errorCode?: string
  /** Additional error details */
  details?: string
  /** Error context */
  context?: Record<string, any>
  /** Trace ID for debugging */
  traceId?: string
  /** Original request configuration */
  request?: RequestConfig

  constructor(response: ApiResponse<any>, request?: RequestConfig) {
    super(response.error?.message || response.message)
    this.name = 'ApiError'
    this.code = response.code
    this.errorCode = response.error?.code
    this.details = response.error?.details
    this.context = response.error?.context
    this.traceId = response.trace_id
    this.request = request
  }
}

/**
 * Network Error
 * 
 * Represents a network-level error (not an API error).
 */
export class NetworkError extends Error {
  /** Error type */
  type: 'timeout' | 'connection' | 'abort'
  /** Original request configuration */
  request?: RequestConfig

  constructor(type: 'timeout' | 'connection' | 'abort', message: string, request?: RequestConfig) {
    super(message)
    this.name = 'NetworkError'
    this.type = type
    this.request = request
  }
}

/**
 * User-Friendly Error
 * 
 * An error formatted for display to end users.
 */
export interface UserFriendlyError {
  /** Error title */
  title: string
  /** User-friendly error message */
  message: string
  /** Optional action the user can take */
  action?: {
    label: string
    handler: () => void
  }
  /** Whether the operation can be retried */
  canRetry: boolean
}

// ============================================================================
// Legacy Format Support (for backward compatibility)
// ============================================================================

/**
 * Legacy API Response Format (V1)
 * 
 * @deprecated Use ApiResponse<T> instead
 * This is kept for backward compatibility during the migration period.
 */
export interface LegacyApiResponse<T> {
  code: number
  message: string
  data?: T | LegacyPaginatedData<T>
  error?: string
}

/**
 * Legacy Paginated Data Format
 * 
 * @deprecated Use PaginatedResponse<T> instead
 */
export interface LegacyPaginatedData<T> {
  records: T[]
  total: number
  limit: number
  offset: number
}

/**
 * Legacy Pagination Parameters
 * 
 * @deprecated Use PaginationParams instead
 */
export interface LegacyPaginationParams {
  limit: number
  offset: number
}

/**
 * Adapted Response Type
 * 
 * A union type that accepts both new and legacy formats.
 * Used during the migration period.
 */
export type AdaptedResponse<T> = ApiResponse<T> | LegacyApiResponse<T>

// ============================================================================
// Cache Types
// ============================================================================

/**
 * Cache Configuration
 */
export interface CacheConfig {
  /** Time to live in milliseconds */
  ttl: number
  /** Maximum cache size (number of entries) */
  maxSize: number
  /** Cache eviction strategy */
  strategy: 'lru' | 'fifo' | 'custom'
}

/**
 * Cache Entry
 */
export interface CacheEntry<T> {
  /** Cached data */
  data: T
  /** Timestamp when the entry was created */
  timestamp: number
  /** Time to live in milliseconds */
  ttl: number
  /** Cache key */
  key: string
}

// ============================================================================
// Interceptor Types
// ============================================================================

/**
 * Request Interceptor Function
 */
export type RequestInterceptor = (config: RequestConfig) => RequestConfig | Promise<RequestConfig>

/**
 * Response Interceptor Function
 */
export type ResponseInterceptor = <T>(response: ApiResponse<T>) => ApiResponse<T> | Promise<ApiResponse<T>>

/**
 * Interceptor Manager
 */
export interface InterceptorManager<T> {
  /** Add an interceptor */
  use(interceptor: (value: T) => T | Promise<T>): number
  /** Remove an interceptor by ID */
  eject(id: number): void
}

// ============================================================================
// Utility Types
// ============================================================================

/**
 * Extract data type from ApiResponse
 */
export type ExtractData<T> = T extends ApiResponse<infer D> ? D : never

/**
 * Extract data type from PaginatedResponse
 */
export type ExtractPaginatedData<T> = T extends PaginatedResponse<infer D> ? D : never

/**
 * Make all properties of T optional recursively
 */
export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P]
}

