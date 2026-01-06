/**
 * Legacy API Adapter
 * 
 * This module provides adapters to convert between legacy (V1) and new (V2) API formats.
 * It helps with gradual migration by allowing components to work with both formats.
 * 
 * @example
 * ```typescript
 * import { adaptLegacyResponse, adaptToLegacyFormat } from '@/lib/adapters/legacy-adapter'
 * 
 * // Convert legacy response to V2
 * const v2Response = adaptLegacyResponse(legacyResponse)
 * 
 * // Convert V2 response to legacy format
 * const legacyFormat = adaptToLegacyFormat(v2Response)
 * ```
 */

import {
  ApiResponse,
  PaginatedResponse,
  CursorPaginatedResponse,
  LegacyApiResponse,
  LegacyPaginatedData,
  LegacyPaginationParams,
  PaginationParams,
  CursorPaginationParams,
} from '@/types/api-v2'

// ============================================================================
// Response Format Adapters
// ============================================================================

/**
 * Adapt legacy API response to V2 format
 * 
 * @template T - The data type
 * @param legacy - Legacy response
 * @returns V2 API response
 */
export function adaptLegacyResponse<T>(legacy: LegacyApiResponse<T>): ApiResponse<T> {
  return {
    code: legacy.code,
    message: legacy.message,
    data: legacy.data as T,
    error: legacy.error ? {
      code: 'LEGACY_ERROR',
      message: legacy.error,
    } : undefined,
  }
}

/**
 * Adapt V2 API response to legacy format
 * 
 * @template T - The data type
 * @param v2 - V2 response
 * @returns Legacy API response
 */
export function adaptToLegacyFormat<T>(v2: ApiResponse<T>): LegacyApiResponse<T> {
  return {
    code: v2.code,
    message: v2.message,
    data: v2.data,
    error: v2.error?.message,
  }
}

// ============================================================================
// Paginated Response Adapters
// ============================================================================

/**
 * Adapt legacy paginated data to V2 paginated response
 * 
 * @template T - The item type
 * @param legacy - Legacy paginated data
 * @returns V2 paginated response
 */
export function adaptLegacyPaginatedData<T>(
  legacy: LegacyPaginatedData<T>
): PaginatedResponse<T> {
  const page = Math.floor(legacy.offset / legacy.limit) + 1
  const totalPages = Math.ceil(legacy.total / legacy.limit)

  return {
    code: 200,
    message: 'success',
    data: legacy.records,
    pagination: {
      total: legacy.total,
      page,
      page_size: legacy.limit,
      total_pages: totalPages,
    },
  }
}

/**
 * Adapt V2 paginated response to legacy format
 * 
 * @template T - The item type
 * @param v2 - V2 paginated response
 * @returns Legacy paginated data
 */
export function adaptToLegacyPaginatedData<T>(
  v2: PaginatedResponse<T>
): LegacyPaginatedData<T> {
  const offset = (v2.pagination.page - 1) * v2.pagination.page_size

  return {
    records: v2.data,
    total: v2.pagination.total,
    limit: v2.pagination.page_size,
    offset,
  }
}

// ============================================================================
// Pagination Parameters Adapters
// ============================================================================

/**
 * Convert legacy pagination params to V2 format
 * 
 * @param legacy - Legacy pagination params (limit, offset)
 * @returns V2 pagination params (page, page_size)
 */
export function adaptLegacyPaginationParams(
  legacy: LegacyPaginationParams
): PaginationParams {
  const page = Math.floor(legacy.offset / legacy.limit) + 1

  return {
    page,
    page_size: legacy.limit,
  }
}

/**
 * Convert V2 pagination params to legacy format
 * 
 * @param v2 - V2 pagination params (page, page_size)
 * @returns Legacy pagination params (limit, offset)
 */
export function adaptToLegacyPaginationParams(
  v2: PaginationParams
): LegacyPaginationParams {
  const offset = (v2.page - 1) * v2.page_size

  return {
    limit: v2.page_size,
    offset,
  }
}

// ============================================================================
// Cursor Pagination Adapters
// ============================================================================

/**
 * Extract cursor pagination info from legacy response
 * 
 * @template T - The item type
 * @param legacy - Legacy response with cursor info
 * @returns V2 cursor paginated response
 */
export function adaptLegacyCursorPaginatedData<T>(legacy: {
  records: T[]
  total: number
  next_cursor?: string
}): CursorPaginatedResponse<T> {
  return {
    code: 200,
    message: 'success',
    data: legacy.records,
    pagination: {
      total: legacy.total,
      next_cursor: legacy.next_cursor,
      has_more: !!legacy.next_cursor,
    },
  }
}

// ============================================================================
// Response Format Detection
// ============================================================================

/**
 * Detect if response is in V2 format
 * 
 * @param response - Response to check
 * @returns True if V2 format, false otherwise
 */
export function isV2Format(response: any): response is ApiResponse<any> {
  return (
    typeof response === 'object' &&
    response !== null &&
    typeof response.code === 'number' &&
    typeof response.message === 'string' &&
    (response.data !== undefined || response.error !== undefined) &&
    (response.error === undefined || typeof response.error === 'object')
  )
}

/**
 * Detect if response is in legacy format
 * 
 * @param response - Response to check
 * @returns True if legacy format, false otherwise
 */
export function isLegacyFormat(response: any): response is LegacyApiResponse<any> {
  return (
    typeof response === 'object' &&
    response !== null &&
    typeof response.code === 'number' &&
    typeof response.message === 'string' &&
    (response.error === undefined || typeof response.error === 'string')
  )
}

/**
 * Detect if data is in legacy paginated format
 * 
 * @param data - Data to check
 * @returns True if legacy paginated format, false otherwise
 */
export function isLegacyPaginatedData(data: any): data is LegacyPaginatedData<any> {
  return (
    typeof data === 'object' &&
    data !== null &&
    Array.isArray(data.records) &&
    typeof data.total === 'number' &&
    typeof data.limit === 'number' &&
    typeof data.offset === 'number'
  )
}

// ============================================================================
// Auto-Adapt Functions
// ============================================================================

/**
 * Automatically adapt response to V2 format if needed
 * 
 * @template T - The data type
 * @param response - Response in any format
 * @returns V2 API response
 */
export function autoAdaptToV2<T>(response: any): ApiResponse<T> {
  if (isV2Format(response)) {
    return response
  }

  if (isLegacyFormat(response)) {
    return adaptLegacyResponse(response)
  }

  // If neither format, try to construct a V2 response
  return {
    code: 200,
    message: 'success',
    data: response as T,
  }
}

/**
 * Automatically adapt pagination params to V2 format if needed
 * 
 * @param params - Pagination params in any format
 * @returns V2 pagination params
 */
export function autoAdaptPaginationParams(
  params: PaginationParams | LegacyPaginationParams | any
): PaginationParams {
  // Check if already V2 format
  if ('page' in params && 'page_size' in params) {
    return params as PaginationParams
  }

  // Check if legacy format
  if ('limit' in params && 'offset' in params) {
    return adaptLegacyPaginationParams(params as LegacyPaginationParams)
  }

  // Default pagination
  return {
    page: 1,
    page_size: 20,
  }
}

// ============================================================================
// Migration Helper Class
// ============================================================================

/**
 * Migration Helper
 * 
 * Provides utilities to help with API format migration.
 */
export class MigrationHelper {
  private static usageLog: Map<string, { v1: number; v2: number }> = new Map()

  /**
   * Log format usage for monitoring migration progress
   * 
   * @param endpoint - API endpoint
   * @param format - Format used ('v1' or 'v2')
   */
  static logFormatUsage(endpoint: string, format: 'v1' | 'v2'): void {
    if (!this.usageLog.has(endpoint)) {
      this.usageLog.set(endpoint, { v1: 0, v2: 0 })
    }

    const stats = this.usageLog.get(endpoint)!
    stats[format]++
  }

  /**
   * Get migration statistics
   * 
   * @returns Migration statistics by endpoint
   */
  static getStats(): Record<string, { v1: number; v2: number; percentage: number }> {
    const stats: Record<string, { v1: number; v2: number; percentage: number }> = {}

    this.usageLog.forEach((counts, endpoint) => {
      const total = counts.v1 + counts.v2
      const percentage = total > 0 ? (counts.v2 / total) * 100 : 0

      stats[endpoint] = {
        v1: counts.v1,
        v2: counts.v2,
        percentage: Math.round(percentage * 100) / 100,
      }
    })

    return stats
  }

  /**
   * Get overall migration progress
   * 
   * @returns Overall percentage of V2 usage
   */
  static getOverallProgress(): number {
    let totalV1 = 0
    let totalV2 = 0

    this.usageLog.forEach((counts) => {
      totalV1 += counts.v1
      totalV2 += counts.v2
    })

    const total = totalV1 + totalV2
    return total > 0 ? Math.round((totalV2 / total) * 10000) / 100 : 0
  }

  /**
   * Print migration report to console
   */
  static printReport(): void {
    const stats = this.getStats()
    const overall = this.getOverallProgress()

    console.group('[API Migration Report]')
    console.log(`Overall Progress: ${overall}%`)
    console.table(stats)
    console.groupEnd()
  }

  /**
   * Clear usage statistics
   */
  static clearStats(): void {
    this.usageLog.clear()
  }
}

// Print migration report in development mode (every 5 minutes)
if (typeof window !== 'undefined' && process.env.NODE_ENV === 'development') {
  setInterval(() => {
    if (MigrationHelper.getOverallProgress() > 0) {
      MigrationHelper.printReport()
    }
  }, 300000) // 5 minutes
}

