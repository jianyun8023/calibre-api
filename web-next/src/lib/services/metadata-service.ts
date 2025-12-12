/**
 * Metadata Service
 * 
 * This service handles metadata search operations:
 * - Search by ISBN
 * - Search by keyword (title, author, etc.)
 * - Map external metadata to internal format
 * 
 * @example
 * ```typescript
 * import { metadataService } from '@/lib/services/metadata-service'
 * 
 * // Search by ISBN
 * const book = await metadataService.searchByISBN('9787111633082')
 * 
 * // Search by keyword
 * const results = await metadataService.searchByKeyword('TypeScript')
 * ```
 */

import { BaseApiService } from './base-service'
import { UnifiedApiClient, apiClient } from '../api-client-v2'
import { ErrorHandler, errorHandler } from '../error-handler'
import { Book } from '@/types/book'

// ============================================================================
// Type Definitions
// ============================================================================

/**
 * Douban Book Metadata
 */
export interface DoubanBook {
  title: string
  sub_title?: string
  author: string[]
  translator?: string[]
  publisher: string
  pubdate: string
  isbn13: string
  isbn10?: string
  summary: string
  image: string
  rating: {
    average: number
    numRaters: number
  }
  tags: Array<{ name: string; count: number }>
  pages?: number
  price?: string
  binding?: string
  series?: string | null
  url?: string
  catalog?: string | null
  origin_title?: string | null
  author_intro?: string | null
  ebook_url?: string | null
  ebook_price?: string | null
}

/**
 * Metadata Search Response
 */
export interface MetadataSearchResponse {
  success: boolean
  books: DoubanBook[]
  count: number
  start: number
  total: number
}

/**
 * ISBN Metadata Response
 */
export type MetadataISBNResponse = DoubanBook

// ============================================================================
// Metadata Service Class
// ============================================================================

export class MetadataService extends BaseApiService {
  constructor(client: UnifiedApiClient, errorHandler: ErrorHandler) {
    super(client, errorHandler)
  }

  // ==========================================================================
  // Metadata Search Operations
  // ==========================================================================

  /**
   * Search metadata by ISBN
   * 
   * @param isbn - ISBN (10 or 13 digits, hyphens will be removed)
   * @returns Douban book metadata
   * 
   * @example
   * ```typescript
   * const book = await metadataService.searchByISBN('978-7-111-63308-2')
   * ```
   */
  async searchByISBN(isbn: string): Promise<MetadataISBNResponse> {
    return this.handleRequest(async () => {
      // Clean ISBN (remove hyphens)
      const cleanISBN = isbn.replace(/-/g, '')

      // Note: Douban API may not follow standard response format
      // We use direct fetch for now
      const response = await fetch(`/api/metadata/isbn/${cleanISBN}`)

      if (!response.ok) {
        const errorText = await response.text().catch(() => 'Unknown error')
        throw new Error(`Failed to fetch metadata: HTTP ${response.status} - ${errorText}`)
      }

      return response.json()
    })
  }

  /**
   * Search metadata by keyword
   * 
   * Supports searching by title, author, ISBN, and other fields.
   * 
   * @param query - Search query string
   * @returns Search response with matching books
   * 
   * @example
   * ```typescript
   * const results = await metadataService.searchByKeyword('TypeScript Programming')
   * console.log(`Found ${results.total} books`)
   * results.books.forEach(book => console.log(book.title))
   * ```
   */
  async searchByKeyword(query: string): Promise<MetadataSearchResponse> {
    return this.handleRequest(async () => {
      // Note: Douban API may not follow standard response format
      // We use direct fetch for now
      const url = `/api/metadata/search?query=${encodeURIComponent(query)}`
      const response = await fetch(url)

      if (!response.ok) {
        const errorText = await response.text().catch(() => 'Unknown error')
        throw new Error(`Failed to search metadata: HTTP ${response.status} - ${errorText}`)
      }

      return response.json()
    })
  }

  // ==========================================================================
  // Data Mapping Operations
  // ==========================================================================

  /**
   * Map Douban book metadata to internal Book format
   * 
   * @param doubanBook - Douban book metadata
   * @returns Internal Book object
   * 
   * @example
   * ```typescript
   * const doubanBook = await metadataService.searchByISBN('9787111633082')
   * const book = metadataService.mapToInternalFormat(doubanBook)
   * ```
   */
  mapToInternalFormat(doubanBook: DoubanBook): Partial<Book> {
    return {
      id: 0, // Will be set when creating/updating
      title: this.joinTitle(doubanBook.title, doubanBook.sub_title),
      authors: doubanBook.author || [],
      isbn: doubanBook.isbn13 || '',
      publisher: doubanBook.publisher || '',
      pubdate: this.parseDateString(doubanBook.pubdate).toISOString(),
      rating: doubanBook.rating?.average || 0,
      tags: doubanBook.tags?.map(tag => tag.name) || [],
      comments: doubanBook.summary || '',
      cover: this.convertImageUrl(doubanBook.image),
    }
  }

  /**
   * Map multiple Douban books to internal format
   * 
   * @param doubanBooks - Array of Douban book metadata
   * @returns Array of internal Book objects
   */
  mapManyToInternalFormat(doubanBooks: DoubanBook[]): Partial<Book>[] {
    return doubanBooks.map(book => this.mapToInternalFormat(book))
  }

  // ==========================================================================
  // Helper Methods
  // ==========================================================================

  /**
   * Join title and subtitle
   */
  private joinTitle(title: string, subTitle?: string): string {
    if (!subTitle) {
      return title
    }
    
    // Don't append if subtitle is too long
    if (subTitle.length > 16) {
      return title
    }
    
    return `${title}：${subTitle}`
  }

  /**
   * Parse date string to Date object
   * 
   * Handles various date formats:
   * - YYYY-MM-DD
   * - YYYY-MM
   * - YYYY
   */
  private parseDateString(dateString: string): Date {
    if (!dateString) {
      return new Date()
    }

    const dateParts = dateString.split('-')
    
    if (dateParts.length === 0) {
      return new Date()
    }

    const year = parseInt(dateParts[0], 10)
    const month = dateParts.length >= 2 ? parseInt(dateParts[1], 10) - 1 : 0
    const day = dateParts.length >= 3 ? parseInt(dateParts[2], 10) : 1

    return new Date(year, month, day)
  }

  /**
   * Convert Douban image URL from large to small size
   * 
   * Douban uses different URL patterns for different image sizes.
   * We convert to smaller size for faster loading.
   */
  private convertImageUrl(imageUrl: string): string {
    if (!imageUrl) {
      return ''
    }

    // Convert large image to small image
    return imageUrl.replace('subject/l/public', 'subject/s/public')
  }

  // ==========================================================================
  // Validation Methods
  // ==========================================================================

  /**
   * Validate ISBN format
   * 
   * @param isbn - ISBN string
   * @returns True if valid, false otherwise
   */
  validateISBN(isbn: string): boolean {
    const cleanISBN = isbn.replace(/-/g, '')
    
    // ISBN-10 or ISBN-13
    return /^\d{10}$/.test(cleanISBN) || /^\d{13}$/.test(cleanISBN)
  }

  /**
   * Check if metadata is complete
   * 
   * @param book - Douban book metadata
   * @returns True if all required fields are present
   */
  isMetadataComplete(book: DoubanBook): boolean {
    return !!(
      book.title &&
      book.author &&
      book.author.length > 0 &&
      book.publisher &&
      book.pubdate
    )
  }

  /**
   * Get metadata completeness score (0-100)
   * 
   * @param book - Douban book metadata
   * @returns Completeness score
   */
  getMetadataScore(book: DoubanBook): number {
    let score = 0
    const maxScore = 100

    // Required fields (60 points)
    if (book.title) score += 15
    if (book.author && book.author.length > 0) score += 15
    if (book.publisher) score += 15
    if (book.pubdate) score += 15

    // Optional but important fields (40 points)
    if (book.isbn13 || book.isbn10) score += 10
    if (book.image) score += 10
    if (book.summary) score += 10
    if (book.rating && book.rating.average > 0) score += 5
    if (book.tags && book.tags.length > 0) score += 5

    return Math.min(score, maxScore)
  }
}

// ============================================================================
// Default Service Instance
// ============================================================================

/**
 * Default metadata service instance
 * 
 * @example
 * ```typescript
 * import { metadataService } from '@/lib/services/metadata-service'
 * 
 * const book = await metadataService.searchByISBN('9787111633082')
 * ```
 */
export const metadataService = new MetadataService(apiClient, errorHandler)

