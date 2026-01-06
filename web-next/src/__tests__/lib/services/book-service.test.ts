/**
 * BookService Tests
 * 
 * Test suite for the book service.
 */

import { BookService } from '@/lib/services/book-service'
import { UnifiedApiClient } from '@/lib/api-client-v2'
import { ErrorHandler } from '@/lib/error-handler'
import { Book } from '@/types/book'

describe('BookService', () => {
  let bookService: BookService
  let mockClient: jest.Mocked<UnifiedApiClient>
  let mockErrorHandler: jest.Mocked<ErrorHandler>

  beforeEach(() => {
    // Create mock client
    mockClient = {
      get: jest.fn(),
      post: jest.fn(),
      put: jest.fn(),
      delete: jest.fn(),
      request: jest.fn(),
      interceptors: {
        request: { use: jest.fn(), eject: jest.fn() },
        response: { use: jest.fn(), eject: jest.fn() },
      },
    } as any

    // Create mock error handler
    mockErrorHandler = {
      toUserFriendlyError: jest.fn(),
      handleApiError: jest.fn(),
      handleNetworkError: jest.fn(),
      handleGenericError: jest.fn(),
      canRetry: jest.fn(),
      getRetryDelay: jest.fn(),
      reportError: jest.fn(),
    } as any

    bookService = new BookService(mockClient, mockErrorHandler)
  })

  describe('getBook', () => {
    it('should fetch book by ID', async () => {
      const mockBook: Book = {
        id: 1,
        title: 'Test Book',
        authors: ['Author 1'],
        isbn: '1234567890',
        publisher: 'Test Publisher',
        pubdate: '2024-01-01',
        rating: 8.5,
        tags: ['tag1', 'tag2'],
        comments: 'Test comments',
        cover: 'cover.jpg',
      }

      mockClient.get.mockResolvedValue(mockBook)

      const result = await bookService.getBook('1')

      expect(mockClient.get).toHaveBeenCalledWith('/api/book/1')
      expect(result).toEqual(mockBook)
    })

    it('should handle errors when fetching book', async () => {
      const error = new Error('Book not found')
      mockClient.get.mockRejectedValue(error)

      await expect(bookService.getBook('999')).rejects.toThrow('Book not found')
    })
  })

  describe('getRecentBooks', () => {
    it('should fetch recent books with pagination', async () => {
      const mockResponse = {
        code: 200,
        message: 'success',
        data: [
          { id: 1, title: 'Book 1' } as Book,
          { id: 2, title: 'Book 2' } as Book,
        ],
      }

      mockClient.request.mockResolvedValue(mockResponse as any)

      const result = await bookService.getRecentBooks({ page: 1, page_size: 20 })

      expect(mockClient.request).toHaveBeenCalledWith(
        expect.objectContaining({
          method: 'GET',
        })
      )
      expect(result.data).toHaveLength(2)
      expect(result.pagination).toBeDefined()
    })
  })

  describe('getRandomBooks', () => {
    it('should fetch random books', async () => {
      const mockBooks = [
        { id: 1, title: 'Book 1' } as Book,
        { id: 2, title: 'Book 2' } as Book,
      ]

      mockClient.get.mockResolvedValue(mockBooks)

      const result = await bookService.getRandomBooks(5)

      expect(mockClient.get).toHaveBeenCalledWith(
        expect.stringContaining('limit=5')
      )
      expect(result).toEqual(mockBooks)
    })
  })

  describe('searchBooks', () => {
    it('should search books with query and filters', async () => {
      const mockResponse = {
        code: 200,
        message: 'success',
        data: [
          { id: 1, title: 'TypeScript Book' } as Book,
        ],
      }

      mockClient.request.mockResolvedValue(mockResponse as any)

      const result = await bookService.searchBooks({
        q: 'typescript',
        mode: 'hybrid',
        filters: ['tag:programming'],
        sort: ['rating:desc'],
        pagination: { page: 1, page_size: 20 },
      })

      expect(mockClient.request).toHaveBeenCalledWith(
        expect.objectContaining({
          method: 'POST',
          body: expect.stringContaining('Filter'),
        })
      )
      expect(result.data).toHaveLength(1)
    })
  })

  describe('updateBook', () => {
    it('should update book with new data', async () => {
      const updatedBook: Book = {
        id: 1,
        title: 'Updated Title',
        authors: ['Author 1'],
        isbn: '1234567890',
        publisher: 'Test Publisher',
        pubdate: '2024-01-01',
        rating: 9.0,
        tags: ['new-tag'],
        comments: 'Updated comments',
        cover: 'cover.jpg',
      }

      mockClient.post.mockResolvedValue(updatedBook)

      const result = await bookService.updateBook('1', {
        title: 'Updated Title',
        rating: 9.0,
      })

      expect(mockClient.post).toHaveBeenCalledWith(
        '/api/book/1/update',
        expect.objectContaining({
          title: 'Updated Title',
          rating: 9.0,
        })
      )
      expect(result).toEqual(updatedBook)
    })
  })

  describe('deleteBook', () => {
    it('should delete book by ID', async () => {
      mockClient.post.mockResolvedValue(undefined)

      await bookService.deleteBook('1')

      expect(mockClient.post).toHaveBeenCalledWith(
        '/api/book/1/delete',
        {}
      )
    })
  })

  describe('getPublishers', () => {
    it('should fetch publishers list', async () => {
      const mockPublishers = ['Publisher A', 'Publisher B', 'Publisher C']

      mockClient.get.mockResolvedValue(mockPublishers)

      const result = await bookService.getPublishers()

      expect(mockClient.get).toHaveBeenCalledWith('/api/publisher')
      expect(result).toEqual(mockPublishers)
    })
  })
})

