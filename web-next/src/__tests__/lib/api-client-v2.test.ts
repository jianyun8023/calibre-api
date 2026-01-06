/**
 * UnifiedApiClient Tests
 * 
 * Test suite for the unified API client.
 */

import { UnifiedApiClient } from '@/lib/api-client-v2'
import { ApiError, NetworkError } from '@/types/api-v2'

describe('UnifiedApiClient', () => {
  let client: UnifiedApiClient
  let fetchMock: jest.Mock

  beforeEach(() => {
    client = new UnifiedApiClient({
      baseURL: '/api',
      timeout: 5000,
      retryAttempts: 0, // Disable retry for faster tests
    })

    fetchMock = global.fetch as jest.Mock
    fetchMock.mockClear()
  })

  describe('GET requests', () => {
    it('should handle successful GET request', async () => {
      const mockResponse = {
        code: 200,
        message: 'success',
        data: { id: 1, title: 'Test Book' },
      }

      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      })

      const result = await client.get<any>('/book/1')

      expect(fetchMock).toHaveBeenCalledWith(
        '/api/book/1',
        expect.objectContaining({
          method: 'GET',
        })
      )
      expect(result).toEqual(mockResponse.data)
    })

    it('should handle GET request with query parameters', async () => {
      const mockResponse = {
        code: 200,
        message: 'success',
        data: [],
      }

      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      })

      await client.get('/books', { 
        url: '/books?page=1&limit=20' 
      })

      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining('page=1'),
        expect.any(Object)
      )
    })

    it('should throw ApiError on 404 response', async () => {
      const mockErrorResponse = {
        code: 404,
        message: 'error',
        error: {
          code: 'BOOK_NOT_FOUND',
          message: 'Book not found',
        },
      }

      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: async () => mockErrorResponse,
      })

      await expect(client.get('/book/999')).rejects.toThrow(ApiError)
    })
  })

  describe('POST requests', () => {
    it('should handle successful POST request', async () => {
      const mockResponse = {
        code: 200,
        message: 'success',
        data: { id: 1, title: 'New Book' },
      }

      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      })

      const data = { title: 'New Book' }
      const result = await client.post('/book', data)

      expect(fetchMock).toHaveBeenCalledWith(
        '/api/book',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(data),
        })
      )
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('Error handling', () => {
    it('should handle network errors', async () => {
      fetchMock.mockRejectedValueOnce(new Error('Network error'))

      await expect(client.get('/book/1')).rejects.toThrow(NetworkError)
    })

    it('should handle timeout errors', async () => {
      // Create client with very short timeout
      const shortTimeoutClient = new UnifiedApiClient({
        timeout: 1,
      })

      fetchMock.mockImplementationOnce(
        () => new Promise(resolve => setTimeout(resolve, 100))
      )

      await expect(shortTimeoutClient.get('/book/1')).rejects.toThrow(NetworkError)
    })
  })

  describe('Interceptors', () => {
    it('should apply request interceptors', async () => {
      const mockResponse = {
        code: 200,
        message: 'success',
        data: {},
      }

      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      })

      client.interceptors.request.use((config) => {
        return {
          ...config,
          headers: {
            ...config.headers,
            'X-Custom-Header': 'test',
          },
        }
      })

      await client.get('/book/1')

      expect(fetchMock).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            'X-Custom-Header': 'test',
          }),
        })
      )
    })

    it('should apply response interceptors', async () => {
      const mockResponse = {
        code: 200,
        message: 'success',
        data: { id: 1 },
      }

      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      })

      let interceptorCalled = false
      client.interceptors.response.use((response) => {
        interceptorCalled = true
        return response
      })

      await client.get('/book/1')

      expect(interceptorCalled).toBe(true)
    })
  })

  describe('Legacy format adaptation', () => {
    it('should adapt legacy response format to V2', async () => {
      const legacyResponse = {
        code: 200,
        message: 'success',
        data: { id: 1, title: 'Test' },
        error: 'some error', // Legacy error format (string)
      }

      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: async () => legacyResponse,
      })

      // Should not throw despite legacy format
      const result = await client.get('/book/1')
      expect(result).toBeDefined()
    })
  })
})

