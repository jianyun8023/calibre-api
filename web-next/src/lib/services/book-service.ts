/**
 * Book Service
 * 
 * This service handles all book-related API operations including:
 * - Fetching book lists (recent, random, all)
 * - Getting book details
 * - Updating and deleting books
 * - Searching books (hybrid and semantic)
 * - File operations (TOC, chapter content, download)
 * 
 * @example
 * ```typescript
 * import { bookService } from '@/lib/services/book-service'
 * 
 * // Get book details
 * const book = await bookService.getBook('1')
 * 
 * // Search books
 * const results = await bookService.searchBooks({
 *   q: 'typescript',
 *   mode: 'hybrid',
 *   pagination: { page: 1, page_size: 20 }
 * })
 * ```
 */

import { BaseApiService } from './base-service'
import { UnifiedApiClient, apiClient } from '../api-client-v2'
import { ErrorHandler, errorHandler } from '../error-handler'
import {
  PaginatedResponse,
  CursorPaginatedResponse,
  PaginationParams,
  CursorPaginationParams,
  SearchQuery,
} from '@/types/api-v2'
import { Book } from '@/types/book'

// ============================================================================
// Type Definitions
// ============================================================================

/**
 * Book list response (for backward compatibility with legacy formats)
 */
export interface BookListResponse {
  total: number
  records: Book[]
  next_cursor?: string
}

/**
 * Book TOC (Table of Contents) response
 */
export interface TocResponse {
  id: string
  title: string
  chapters: TocChapter[]
}

/**
 * TOC Chapter
 */
export interface TocChapter {
  title: string
  path: string
  level: number
  children?: TocChapter[]
}

/**
 * Book update data
 */
export interface BookUpdateData {
  title?: string
  authors?: string[]
  isbn?: string
  publisher?: string
  pubdate?: string
  rating?: number
  tags?: string[]
  comments?: string
  cover?: string
}

// ============================================================================
// Book Service Class
// ============================================================================

export class BookService extends BaseApiService {
  constructor(client: UnifiedApiClient, errorHandler: ErrorHandler) {
    super(client, errorHandler)
  }

  // ==========================================================================
  // Book List Operations
  // ==========================================================================

  /**
   * Get recent books
   * 
   * @param pagination - Pagination parameters (new format)
   * @returns Paginated response with books
   * 
   * @example
   * ```typescript
   * const books = await bookService.getRecentBooks({ page: 1, page_size: 20 })
   * ```
   */
  async getRecentBooks(pagination: PaginationParams): Promise<PaginatedResponse<Book>> {
    return this.handleRequest(async () => {
      const url = this.buildUrl('/api/recently', this.buildPaginationParams(pagination))
      const response = await this.client.request<Book[]>({ url, method: 'GET' })

      // Adapt response to PaginatedResponse format
      return this.adaptToPaginatedResponse(response.data, response.code, response.message)
    })
  }

  /**
   * Get random books
   * 
   * @param limit - Number of books to fetch (default: 5)
   * @returns Array of random books
   * 
   * @example
   * ```typescript
   * const books = await bookService.getRandomBooks(10)
   * ```
   */
  async getRandomBooks(limit: number = 5): Promise<Book[]> {
    return this.handleRequest(async () => {
      const url = this.buildUrl('/api/random', { limit })
      return this.client.get<Book[]>(url, { cache: 'no-store' })
    })
  }

  /**
   * Get all books with cursor pagination
   * 
   * @param pagination - Cursor pagination parameters
   * @returns Cursor paginated response with books
   * 
   * @example
   * ```typescript
   * const response = await bookService.getAllBooks({ limit: 20 })
   * // Next page
   * const nextResponse = await bookService.getAllBooks({ 
   *   limit: 20, 
   *   cursor: response.pagination.next_cursor 
   * })
   * ```
   */
  async getAllBooks(pagination: CursorPaginationParams): Promise<CursorPaginatedResponse<Book>> {
    return this.handleRequest(async () => {
      const url = this.buildUrl('/api/books/all', this.buildCursorPaginationParams(pagination))
      const response = await this.client.request<Book[]>({ url, method: 'GET' })

      // Adapt response to CursorPaginatedResponse format
      return this.adaptToCursorPaginatedResponse(response.data, response.code, response.message)
    })
  }

  // ==========================================================================
  // Book CRUD Operations
  // ==========================================================================

  /**
   * Get book by ID
   * 
   * @param id - Book ID
   * @returns Book details
   * 
   * @example
   * ```typescript
   * const book = await bookService.getBook('123')
   * ```
   */
  async getBook(id: string | number): Promise<Book> {
    return this.handleRequest(async () => {
      return this.client.get<Book>(`/api/book/${id}`)
    })
  }

  /**
   * Update book
   * 
   * @param id - Book ID
   * @param data - Book data to update
   * @returns Updated book
   * 
   * @example
   * ```typescript
   * const updated = await bookService.updateBook('123', {
   *   title: 'New Title',
   *   tags: ['tag1', 'tag2']
   * })
   * ```
   */
  async updateBook(id: string | number, data: BookUpdateData): Promise<Book> {
    return this.handleRequest(async () => {
      return this.client.post<Book>(`/api/book/${id}/update`, data)
    })
  }

  /**
   * Delete book
   * 
   * @param id - Book ID
   * 
   * @example
   * ```typescript
   * await bookService.deleteBook('123')
   * ```
   */
  async deleteBook(id: string | number): Promise<void> {
    return this.handleRequest(async () => {
      return this.client.post<void>(`/api/book/${id}/delete`, {})
    })
  }

  // ==========================================================================
  // Search Operations
  // ==========================================================================

  /**
   * Search books
   * 
   * @param query - Search query with filters and pagination
   * @returns Paginated search results
   * 
   * @example
   * ```typescript
   * const results = await bookService.searchBooks({
   *   q: 'javascript',
   *   mode: 'hybrid',
   *   filters: ['tag:programming'],
   *   sort: ['rating:desc'],
   *   pagination: { page: 1, page_size: 20 }
   * })
   * ```
   */
  async searchBooks(query: SearchQuery): Promise<PaginatedResponse<Book>> {
    return this.handleRequest(async () => {
      const { q, mode = 'hybrid', filters = [], sort = [], pagination } = query

      const url = this.buildUrl('/api/search', { q, mode })

      const requestBody = {
        Filter: filters,
        Sort: sort,
        ...this.convertToLegacyPagination(pagination),
      }

      const response = await this.client.request<Book[]>({
        url,
        method: 'POST',
        body: JSON.stringify(requestBody),
      })

      // Adapt response to PaginatedResponse format
      return this.adaptToPaginatedResponse(response.data, response.code, response.message)
    })
  }

  /**
   * Semantic search
   * 
   * @param query - Search query string
   * @param limit - Number of results (default: 12)
   * @returns Paginated search results
   * 
   * @example
   * ```typescript
   * const results = await bookService.searchSemantic('machine learning books', 20)
   * ```
   */
  async searchSemantic(query: string, limit: number = 12): Promise<PaginatedResponse<Book>> {
    return this.handleRequest(async () => {
      const url = this.buildUrl('/api/search/semantic', {
        q: query,
        limit
      })

      const response = await this.client.request<Book[]>({ url, method: 'GET' })

      // Adapt response to PaginatedResponse format
      return this.adaptToPaginatedResponse(response.data, response.code, response.message)
    })
  }

  // ==========================================================================
  // File Operations
  // ==========================================================================

  /**
   * Get book table of contents
   * 
   * @param id - Book ID
   * @returns TOC response
   * 
   * @example
   * ```typescript
   * const toc = await bookService.getBookToc('123')
   * ```
   */
  async getBookToc(id: string | number): Promise<TocResponse> {
    return this.handleRequest(async () => {
      // TOC API may not follow standard response format, handle specially
      const response = await fetch(`/api/read/${id}/toc`)

      if (!response.ok) {
        throw new Error(`Failed to fetch TOC: ${response.status} ${response.statusText}`)
      }

      return response.json()
    })
  }

  /**
   * Get chapter content
   * 
   * @param bookId - Book ID
   * @param filePath - Chapter file path
   * @returns Chapter content as HTML string
   * 
   * @example
   * ```typescript
   * const content = await bookService.getChapterContent('123', 'chapter1.xhtml')
   * ```
   */
  async getChapterContent(bookId: string | number, filePath: string): Promise<string> {
    return this.handleRequest(async () => {
      // Remove the /read/{id}/file/ prefix if present
      const cleanPath = filePath.replace(`/read/${bookId}/file/`, '')
      const response = await fetch(`/api/read/${bookId}/file/${cleanPath}`)

      if (!response.ok) {
        throw new Error(`Failed to fetch chapter content: ${response.status} ${response.statusText}`)
      }

      return response.text()
    })
  }

  /**
   * Download book file
   * 
   * @param id - Book ID
   * @returns Blob containing the book file
   * 
   * @example
   * ```typescript
   * const blob = await bookService.downloadBook('123')
   * const url = URL.createObjectURL(blob)
   * window.open(url)
   * ```
   */
  async downloadBook(id: string | number): Promise<Blob> {
    return this.handleRequest(async () => {
      const response = await fetch(`/api/book/${id}/download`)

      if (!response.ok) {
        throw new Error(`Failed to download book: ${response.status} ${response.statusText}`)
      }

      return response.blob()
    })
  }

  /**
   * Get publishers list
   * 
   * @returns Array of publisher names
   * 
   * @example
   * ```typescript
   * const publishers = await bookService.getPublishers()
   * ```
   */
  async getPublishers(): Promise<string[]> {
    return this.handleRequest(async () => {
      return this.client.get<string[]>('/api/publisher')
    })
  }

  // ==========================================================================
  // Helper Methods
  // ==========================================================================

  /**
   * Adapt legacy response to PaginatedResponse format
   */
  private adaptToPaginatedResponse<T>(
    data: T[] | any,
    code: number,
    message: string
  ): PaginatedResponse<T> {
    // If data is already in the correct format (has records, total, etc.)
    if (data && typeof data === 'object' && 'records' in data) {
      const legacyData = data as unknown as BookListResponse
      return {
        code,
        message,
        data: legacyData.records as unknown as T[],
        pagination: {
          total: legacyData.total,
          page: 1, // Default, as legacy format doesn't have page
          page_size: legacyData.records.length,
          total_pages: Math.ceil(legacyData.total / (legacyData.records.length || 1)),
        },
      }
    }

    // If data is just an array
    if (Array.isArray(data)) {
      return {
        code,
        message,
        data,
        pagination: {
          total: data.length,
          page: 1,
          page_size: data.length,
          total_pages: 1,
        },
      }
    }

    // Fallback
    return {
      code,
      message,
      data: [],
      pagination: {
        total: 0,
        page: 1,
        page_size: 0,
        total_pages: 0,
      },
    }
  }

  /**
   * Adapt legacy response to CursorPaginatedResponse format
   */
  private adaptToCursorPaginatedResponse<T>(
    data: T[] | any,
    code: number,
    message: string
  ): CursorPaginatedResponse<T> {
    // If data is already in the correct format
    if (data && typeof data === 'object' && 'records' in data) {
      const legacyData = data as unknown as BookListResponse
      return {
        code,
        message,
        data: legacyData.records as unknown as T[],
        pagination: {
          total: legacyData.total,
          next_cursor: legacyData.next_cursor,
          has_more: !!legacyData.next_cursor,
        },
      }
    }

    // If data is just an array
    if (Array.isArray(data)) {
      return {
        code,
        message,
        data,
        pagination: {
          total: data.length,
          has_more: false,
        },
      }
    }

    // Fallback
    return {
      code,
      message,
      data: [],
      pagination: {
        total: 0,
        has_more: false,
      },
    }
  }
}

// ============================================================================
// Default Service Instance
// ============================================================================

/**
 * Default book service instance
 * 
 * @example
 * ```typescript
 * import { bookService } from '@/lib/services/book-service'
 * 
 * const book = await bookService.getBook('1')
 * ```
 */
export const bookService = new BookService(apiClient, errorHandler)

