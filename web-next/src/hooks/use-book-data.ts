/**
 * Use Book Data Hook
 * 
 * This hook automatically fetches complete book data when incomplete data is detected.
 * It checks if essential fields (like cover) are missing and fetches from API if needed.
 */

import { useState, useEffect } from 'react'
import { Book } from '@/types/book'
import { fetchBook } from '@/lib/api/books'

interface UseBookDataOptions {
  /** Enable auto-fetch when data is incomplete */
  autoFetch?: boolean
  /** Fields to check for completeness */
  requiredFields?: (keyof Book)[]
}

/**
 * Check if book data is complete
 */
function isBookDataComplete(book: Book, requiredFields: (keyof Book)[]): boolean {
  for (const field of requiredFields) {
    const value = book[field]
    
    // Check if field is missing or empty
    if (value === undefined || value === null) {
      return false
    }
    
    // For arrays, check if empty
    if (Array.isArray(value) && value.length === 0) {
      return false
    }
    
    // For strings, check if empty
    if (typeof value === 'string' && value.trim() === '') {
      return false
    }
  }
  
  return true
}

/**
 * Hook to manage book data with auto-fetch capability
 * 
 * @param initialBook - Initial book data (may be incomplete)
 * @param options - Configuration options
 * @returns Enhanced book data and loading state
 * 
 * @example
 * ```typescript
 * const { book, loading, isComplete } = useBookData(initialBook, {
 *   autoFetch: true,
 *   requiredFields: ['cover', 'authors', 'publisher']
 * })
 * ```
 */
export function useBookData(
  initialBook: Book,
  options: UseBookDataOptions = {}
): {
  book: Book
  loading: boolean
  isComplete: boolean
  error: Error | null
  refetch: () => Promise<void>
} {
  const {
    autoFetch = true,
    requiredFields = ['cover', 'authors', 'publisher', 'pubdate'],
  } = options

  const [book, setBook] = useState<Book>(initialBook)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)
  const [hasFetched, setHasFetched] = useState(false)

  // Check if current data is complete
  const isComplete = isBookDataComplete(book, requiredFields)

  // Fetch complete book data
  const fetchCompleteData = async () => {
    if (!book.id) {
      console.warn('[useBookData] Cannot fetch: book ID is missing')
      return
    }

    setLoading(true)
    setError(null)

    try {
      const completeBook = await fetchBook(String(book.id))
      setBook(completeBook)
      setHasFetched(true)
    } catch (err) {
      console.error('[useBookData] Fetch error:', err)
      setError(err as Error)
      // Keep using incomplete data on error
    } finally {
      setLoading(false)
    }
  }

  // Auto-fetch when data is incomplete
  useEffect(() => {
    if (autoFetch && !isComplete && !hasFetched && !loading && book.id) {
      fetchCompleteData()
    }
  }, [autoFetch, isComplete, hasFetched, loading, book.id])

  // Update book when initialBook changes
  useEffect(() => {
    setBook(initialBook)
    setHasFetched(false)
  }, [initialBook.id]) // Only reset when ID changes

  return {
    book,
    loading,
    isComplete,
    error,
    refetch: fetchCompleteData,
  }
}

