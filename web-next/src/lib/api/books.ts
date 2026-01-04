/**
 * Book API Functions (V2)
 * 
 * This module provides backward-compatible API functions using the new V2 service layer.
 * Components can continue using these functions while benefiting from unified error handling,
 * caching, and retry logic.
 * 
 * @deprecated These functions are maintained for backward compatibility.
 * New code should use bookService directly from '@/lib/services'
 */

import { bookService, BookListResponse, TocResponse, BookUpdateData } from '@/lib/services'
import { Book } from '@/types/book'
import { PaginationParams } from '@/types/api-v2'

// ============================================================================
// Publisher Operations
// ============================================================================

/**
 * Fetch publishers list
 * 
 * @returns Array of publisher names
 * 
 * @example
 * ```typescript
 * const publishers = await fetchPublishers()
 * ```
 */
export async function fetchPublishers(): Promise<string[]> {
  return bookService.getPublishers()
}

// ============================================================================
// Book List Operations
// ============================================================================

/**
 * Fetch random books
 * 
 * @param limit - Number of books to fetch (default: 5)
 * @returns Book list response
 * 
 * @example
 * ```typescript
 * const result = await fetchRandomBooks(10)
 * console.log(result.records)
 * ```
 */
export async function fetchRandomBooks(limit: number = 5): Promise<BookListResponse> {
  const books = await bookService.getRandomBooks(limit)
  return {
    total: books.length,
    records: books,
  }
}

/**
 * Fetch recent books
 * 
 * @param limit - Number of books per page
 * @param offset - Offset for pagination
 * @returns Book list response
 * 
 * @example
 * ```typescript
 * const result = await fetchRecentBooks(20, 0)
 * ```
 */
export async function fetchRecentBooks(limit: number, offset: number): Promise<BookListResponse> {
  // Convert legacy pagination to V2 format
  const page = Math.floor(offset / limit) + 1
  const pagination: PaginationParams = { page, page_size: limit }

  const response = await bookService.getRecentBooks(pagination)

  // Convert back to legacy format for backward compatibility
  return {
    total: response.pagination.total,
    records: response.data,
  }
}

/**
 * Fetch all books with cursor pagination
 * 
 * @param limit - Number of books to fetch
 * @param cursor - Cursor for pagination (empty for first page)
 * @returns Book list response
 * 
 * @example
 * ```typescript
 * const result = await fetchAllBooks(20, '')
 * // Next page
 * const nextResult = await fetchAllBooks(20, result.next_cursor || '')
 * ```
 */
export async function fetchAllBooks(limit: number, cursor: string = ''): Promise<BookListResponse> {
  const response = await bookService.getAllBooks({ limit, cursor })

  return {
    total: response.pagination.total,
    records: response.data,
    next_cursor: response.pagination.next_cursor,
  }
}

// ============================================================================
// Search Operations
// ============================================================================

/**
 * Search books
 * 
 * @param keyword - Search keyword
 * @param filter - Filter conditions
 * @param limit - Number of results per page
 * @param offset - Offset for pagination
 * @param sort - Sort conditions (optional)
 * @param mode - Search mode (optional, default: 'hybrid')
 * @returns Book list response
 * 
 * @example
 * ```typescript
 * const results = await fetchBooks('typescript', ['tag:programming'], 20, 0)
 * ```
 */
export async function fetchBooks(
  keyword: string,
  filter: string[],
  limit: number,
  offset: number,
  sort?: string[],
  mode?: string
): Promise<BookListResponse> {
  // Convert legacy pagination to V2 format
  const page = Math.floor(offset / limit) + 1
  const pagination: PaginationParams = { page, page_size: limit }

  const response = await bookService.searchBooks({
    q: keyword,
    mode: (mode as any) || 'hybrid',
    filters: filter,
    sort: sort || [],
    pagination,
  })

  // Convert back to legacy format
  return {
    total: response.pagination.total,
    records: response.data,
  }
}

/**
 * Semantic search
 * 
 * @param query - Search query
 * @param limit - Number of results (default: 12)
 * @returns Book list response
 * 
 * @example
 * ```typescript
 * const results = await searchSemantic('machine learning books', 20)
 * ```
 */
export async function searchSemantic(query: string, limit: number = 12): Promise<BookListResponse> {
  const response = await bookService.searchSemantic(query, limit)

  return {
    total: response.pagination.total,
    records: response.data,
  }
}

// ============================================================================
// Book CRUD Operations
// ============================================================================

/**
 * Fetch book by ID
 * 
 * @param id - Book ID
 * @returns Book details
 * 
 * @example
 * ```typescript
 * const book = await fetchBook('123')
 * ```
 */
export async function fetchBook(id: string): Promise<Book> {
  return bookService.getBook(id)
}

/**
 * Update book
 * 
 * @param id - Book ID
 * @param body - Book data to update
 * @returns Updated book
 * 
 * @example
 * ```typescript
 * const updated = await updateBook('123', { title: 'New Title' })
 * ```
 */
export async function updateBook(id: string | number, body: BookUpdateData): Promise<Book> {
  return bookService.updateBook(id, body)
}

/**
 * Delete book
 * 
 * @param bookId - Book ID
 * 
 * @example
 * ```typescript
 * await deleteBook(123)
 * ```
 */
export async function deleteBook(bookId: number): Promise<void> {
  return bookService.deleteBook(bookId)
}

// ============================================================================
// File Operations
// ============================================================================

/**
 * Fetch book table of contents
 * 
 * @param id - Book ID
 * @returns TOC response
 * 
 * @example
 * ```typescript
 * const toc = await fetchBookToc('123')
 * ```
 */
export async function fetchBookToc(id: string): Promise<TocResponse> {
  return bookService.getBookToc(id)
}

/**
 * Fetch chapter content
 * 
 * @param bookId - Book ID
 * @param filePath - Chapter file path
 * @returns Chapter content as HTML string
 * 
 * @example
 * ```typescript
 * const content = await fetchChapterContent('123', 'chapter1.xhtml')
 * ```
 */
export async function fetchChapterContent(bookId: string, filePath: string): Promise<string> {
  return bookService.getChapterContent(bookId, filePath)
}

// ============================================================================
// Metadata Extraction
// ============================================================================

/**
 * Extracted metadata from book file (copyright page)
 */
export interface ExtractedMetadata {
  book_id: number
  isbn: string
  book_title: string
  author: string
  translator: string
  publisher: string
  publish_date: string
}

/**
 * Extract metadata from book file (copyright page)
 * 
 * @param bookId - Book ID
 * @returns Extracted metadata result
 * 
 * @example
 * ```typescript
 * const result = await extractBookMetadata('123')
 * if (result.success && result.data) {
 *   console.log('Extracted ISBN:', result.data.isbn)
 * }
 * ```
 */
export async function extractBookMetadata(bookId: string): Promise<{
  success: boolean
  message: string
  data: ExtractedMetadata | null
}> {
  const response = await fetch(`/api/book/${bookId}/extract-metadata`, {
    method: 'POST',
  })

  if (!response.ok) {
    throw new Error(`Failed to extract metadata: ${response.statusText}`)
  }

  const result = await response.json()
  return result.data || result
}

// ============================================================================
// Deprecated Interface (for backward compatibility)
// ============================================================================

/**
 * @deprecated Use BookListResponse from '@/lib/services' instead
 */
export interface LegacyBookListResponse {
  total: number
  records: Book[]
  next_cursor?: string
}

