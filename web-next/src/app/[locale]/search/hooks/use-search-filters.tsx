"use client"

import { useState, useEffect, useCallback, useMemo } from "react"
import { useSearchParams, useRouter } from "next/navigation"
import type { Book } from "@/types/book"
import type { SearchFiltersState } from "../components/search-filters"

export interface UseSearchFiltersReturn {
  filters: SearchFiltersState
  setFilters: (filters: SearchFiltersState) => void
  applyFilters: (books: Book[]) => Book[]
  availableAuthors: string[]
  availableTags: string[]
  updateURL: () => void
}

const defaultFilters: SearchFiltersState = {
  publishers: [],
  authors: [],
  tags: [],
  ratingRange: [0, 5],
  yearRange: [1900, new Date().getFullYear()],
  sortBy: 'relevance',
  sortOrder: 'desc'
}

export function useSearchFilters(books: Book[] = []): UseSearchFiltersReturn {
  const searchParams = useSearchParams()
  const router = useRouter()
  
  const [filters, setFiltersState] = useState<SearchFiltersState>(defaultFilters)

  // Initialize filters from URL parameters
  useEffect(() => {
    const urlFilters: SearchFiltersState = { ...defaultFilters }
    
    // Parse publishers
    const publishersParam = searchParams.get('publishers')
    if (publishersParam) {
      urlFilters.publishers = publishersParam.split(',').filter(Boolean)
    }
    
    // Parse authors
    const authorsParam = searchParams.get('authors')
    if (authorsParam) {
      urlFilters.authors = authorsParam.split(',').filter(Boolean)
    }
    
    // Parse tags
    const tagsParam = searchParams.get('tags')
    if (tagsParam) {
      urlFilters.tags = tagsParam.split(',').filter(Boolean)
    }
    
    // Parse rating range
    const ratingParam = searchParams.get('rating')
    if (ratingParam) {
      const [min, max] = ratingParam.split('-').map(Number)
      if (!isNaN(min) && !isNaN(max)) {
        urlFilters.ratingRange = [min, max]
      }
    }
    
    // Parse year range
    const yearParam = searchParams.get('year')
    if (yearParam) {
      const [min, max] = yearParam.split('-').map(Number)
      if (!isNaN(min) && !isNaN(max)) {
        urlFilters.yearRange = [min, max]
      }
    }
    
    // Parse sort
    const sortParam = searchParams.get('sort')
    if (sortParam) {
      const [sortBy, sortOrder] = sortParam.split('-')
      if (sortBy && sortOrder) {
        urlFilters.sortBy = sortBy
        urlFilters.sortOrder = sortOrder as 'asc' | 'desc'
      }
    }
    
    setFiltersState(urlFilters)
  }, [searchParams])

  // Extract available authors and tags from books
  const availableAuthors = useMemo(() => {
    const authors = new Set<string>()
    books.forEach(book => {
      book.authors?.forEach(author => authors.add(author))
    })
    return Array.from(authors).sort()
  }, [books])

  const availableTags = useMemo(() => {
    const tags = new Set<string>()
    books.forEach(book => {
      book.tags?.forEach(tag => tags.add(tag))
    })
    return Array.from(tags).sort()
  }, [books])

  // Update URL with current filters
  const updateURL = useCallback(() => {
    const params = new URLSearchParams(searchParams.toString())
    
    // Update publishers
    if (filters.publishers.length > 0) {
      params.set('publishers', filters.publishers.join(','))
    } else {
      params.delete('publishers')
    }
    
    // Update authors
    if (filters.authors.length > 0) {
      params.set('authors', filters.authors.join(','))
    } else {
      params.delete('authors')
    }
    
    // Update tags
    if (filters.tags.length > 0) {
      params.set('tags', filters.tags.join(','))
    } else {
      params.delete('tags')
    }
    
    // Update rating range
    if (filters.ratingRange[0] > 0 || filters.ratingRange[1] < 5) {
      params.set('rating', `${filters.ratingRange[0]}-${filters.ratingRange[1]}`)
    } else {
      params.delete('rating')
    }
    
    // Update year range
    if (filters.yearRange[0] > 1900 || filters.yearRange[1] < new Date().getFullYear()) {
      params.set('year', `${filters.yearRange[0]}-${filters.yearRange[1]}`)
    } else {
      params.delete('year')
    }
    
    // Update sort
    if (filters.sortBy !== 'relevance' || filters.sortOrder !== 'desc') {
      params.set('sort', `${filters.sortBy}-${filters.sortOrder}`)
    } else {
      params.delete('sort')
    }
    
    router.push(`/search?${params.toString()}`)
  }, [filters, searchParams, router])

  // Set filters and update URL
  const setFilters = useCallback((newFilters: SearchFiltersState) => {
    setFiltersState(newFilters)
  }, [])

  // Apply filters to books array
  const applyFilters = useCallback((books: Book[]): Book[] => {
    let filteredBooks = [...books]

    // Apply publisher filter
    if (filters.publishers.length > 0) {
      filteredBooks = filteredBooks.filter(book => 
        book.publisher && filters.publishers.includes(book.publisher)
      )
    }

    // Apply author filter
    if (filters.authors.length > 0) {
      filteredBooks = filteredBooks.filter(book => 
        book.authors?.some(author => filters.authors.includes(author))
      )
    }

    // Apply tag filter
    if (filters.tags.length > 0) {
      filteredBooks = filteredBooks.filter(book => 
        book.tags?.some(tag => filters.tags.includes(tag))
      )
    }

    // Apply rating filter
    if (filters.ratingRange[0] > 0 || filters.ratingRange[1] < 5) {
      filteredBooks = filteredBooks.filter(book => {
        const rating = (book.rating || 0) / 2 // Convert from 0-10 to 0-5 scale
        return rating >= filters.ratingRange[0] && rating <= filters.ratingRange[1]
      })
    }

    // Apply year filter
    if (filters.yearRange[0] > 1900 || filters.yearRange[1] < new Date().getFullYear()) {
      filteredBooks = filteredBooks.filter(book => {
        if (!book.pubdate) return false
        const year = new Date(book.pubdate).getFullYear()
        return year >= filters.yearRange[0] && year <= filters.yearRange[1]
      })
    }

    // Apply sorting
    if (filters.sortBy !== 'relevance') {
      filteredBooks.sort((a, b) => {
        let aValue: any, bValue: any

        switch (filters.sortBy) {
          case 'pubdate':
            aValue = a.pubdate ? new Date(a.pubdate).getTime() : 0
            bValue = b.pubdate ? new Date(b.pubdate).getTime() : 0
            break
          case 'rating':
            aValue = a.rating || 0
            bValue = b.rating || 0
            break
          case 'title':
            aValue = a.title?.toLowerCase() || ''
            bValue = b.title?.toLowerCase() || ''
            break
          default:
            return 0
        }

        if (filters.sortOrder === 'asc') {
          return aValue > bValue ? 1 : aValue < bValue ? -1 : 0
        } else {
          return aValue < bValue ? 1 : aValue > bValue ? -1 : 0
        }
      })
    }

    return filteredBooks
  }, [filters])

  return {
    filters,
    setFilters,
    applyFilters,
    availableAuthors,
    availableTags,
    updateURL
  }
}