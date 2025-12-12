/**
 * Metadata API Functions (V2)
 * 
 * This module provides backward-compatible API functions using the new V2 service layer.
 * 
 * @deprecated These functions are maintained for backward compatibility.
 * New code should use metadataService directly from '@/lib/services'
 */

import { 
  metadataService, 
  DoubanBook as ServiceDoubanBook, 
  MetadataSearchResponse as ServiceMetadataSearchResponse,
  MetadataISBNResponse as ServiceMetadataISBNResponse,
} from '@/lib/services'

// ============================================================================
// Re-export Types for Backward Compatibility
// ============================================================================

/**
 * 豆瓣元数据结构
 * @deprecated Use DoubanBook from '@/lib/services' instead
 */
export type DoubanBook = ServiceDoubanBook

/**
 * Metadata search response
 * @deprecated Use MetadataSearchResponse from '@/lib/services' instead
 */
export type MetadataSearchResponse = ServiceMetadataSearchResponse

/**
 * Metadata ISBN response
 * @deprecated Use MetadataISBNResponse from '@/lib/services' instead
 */
export type MetadataISBNResponse = ServiceMetadataISBNResponse

// ============================================================================
// API Functions
// ============================================================================

/**
 * 通过 ISBN 搜索元数据
 * 
 * @param isbn - ISBN (10 or 13 digits, hyphens will be removed)
 * @returns Douban book metadata
 * 
 * @example
 * ```typescript
 * const book = await searchMetadataByISBN('978-7-111-63308-2')
 * console.log(book.title)
 * ```
 */
export async function searchMetadataByISBN(isbn: string): Promise<MetadataISBNResponse> {
  return metadataService.searchByISBN(isbn)
}

/**
 * 通过关键词搜索元数据（支持标题、作者、ISBN）
 * 
 * @param query - Search query string
 * @returns Search response with matching books
 * 
 * @example
 * ```typescript
 * const results = await searchMetadata('TypeScript Programming')
 * console.log(`Found ${results.total} books`)
 * results.books.forEach(book => console.log(book.title))
 * ```
 */
export async function searchMetadata(query: string): Promise<MetadataSearchResponse> {
  return metadataService.searchByKeyword(query)
}

