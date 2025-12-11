"use client"

import { useState, useEffect, useCallback, Suspense } from "react"
import { useSearchParams, useRouter } from "next/navigation"
import { fetchBooks, searchSemantic } from "@/lib/api/books"
import type { Book } from "@/types/book"
import { BookCard } from "@/components/book-card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Skeleton } from "@/components/ui/skeleton"
import { Search as SearchIcon, Filter } from "lucide-react"
import { SearchFilters } from "./components/search-filters"
import { SearchHistory } from "./components/search-history"
import { useSearchFilters } from "./hooks/use-search-filters"
import { useSearchHistory } from "./hooks/use-search-history"

function SearchPageContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  
  const query = searchParams.get('q') || ''
  const mode = searchParams.get('mode') || 'hybrid'
  const author = searchParams.get('author')
  const publisher = searchParams.get('publisher')
  const tag = searchParams.get('tag')

  const [input, setInput] = useState(query)
  const [allBooks, setAllBooks] = useState<Book[]>([])
  const [loading, setLoading] = useState(false)
  const [filtersOpen, setFiltersOpen] = useState(true)
  const [historyOpen, setHistoryOpen] = useState(false)

  // Initialize search filters hook
  const {
    filters,
    setFilters,
    applyFilters,
    availableAuthors,
    availableTags,
    updateURL
  } = useSearchFilters(allBooks)

  // Initialize search history hook
  const { addToHistory } = useSearchHistory()

  // Apply filters to get displayed books
  const books = applyFilters(allBooks)

  const performSearch = useCallback(async () => {
    setLoading(true)
    try {
      let data;
      if (mode === 'semantic') {
        data = await searchSemantic(query)
      } else {
        // 构建过滤器
        const filters = []
        if (author) filters.push(`authors:${author}`)
        if (publisher) filters.push(`publisher:${publisher}`)
        if (tag) filters.push(`tags:${tag}`)
        
        data = await fetchBooks(query, filters, 20, 0, [], mode)
      }
      const books = data.records || data || []
      setAllBooks(books)
      
      // Add to search history if there's a query
      if (query) {
        addToHistory(query, mode, books.length)
      }
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }, [query, mode, author, publisher, tag, addToHistory])

  useEffect(() => {
    if (query || author || publisher || tag) {
      performSearch()
    }
  }, [query, author, publisher, tag, performSearch])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    const params = new URLSearchParams()
    if (input) params.set('q', input)
    params.set('mode', mode)
    router.push(`/search?${params.toString()}`)
  }

  const handleHistorySearch = (query: string, searchMode: string) => {
    setInput(query)
    const params = new URLSearchParams()
    params.set('q', query)
    params.set('mode', searchMode)
    router.push(`/search?${params.toString()}`)
    setHistoryOpen(false)
  }

  const handleModeChange = (newMode: string) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set('mode', newMode)
    router.push(`/search?${params.toString()}`)
  }

  // Update URL when filters change
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      updateURL()
    }, 300) // Debounce URL updates
    
    return () => clearTimeout(timeoutId)
  }, [filters, updateURL])

  return (
    <div className="space-y-6">
      {/* Search Header */}
      <div className="flex flex-col items-center space-y-4 pt-8">
        <h1 className="text-3xl font-bold">Search Library</h1>
        <div className="w-full max-w-2xl">
          <form onSubmit={handleSearch} className="flex gap-2">
            <Input 
              value={input} 
              onChange={(e) => setInput(e.target.value)} 
              placeholder="Search books, authors, publishers..." 
              className="h-12 text-lg"
            />
            <Button type="submit" size="lg" className="h-12 px-8">
              <SearchIcon className="mr-2 h-5 w-5" />
              Search
            </Button>
            <Button
              type="button"
              variant="outline"
              size="lg"
              className="h-12 px-4 md:hidden"
              onClick={() => setFiltersOpen(!filtersOpen)}
            >
              <Filter className="h-5 w-5" />
            </Button>
          </form>
        </div>
        
        <Tabs value={mode} onValueChange={handleModeChange} className="w-full max-w-md">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="keyword">Keyword</TabsTrigger>
            <TabsTrigger value="semantic">Semantic</TabsTrigger>
            <TabsTrigger value="hybrid">Hybrid</TabsTrigger>
          </TabsList>
        </Tabs>
      </div>

      {/* Main Content */}
      <div className="flex gap-6">
        {/* Left Sidebar - Desktop */}
        <div className={`hidden md:block transition-all duration-300 ${filtersOpen ? 'w-80' : 'w-0 overflow-hidden'}`}>
          <div className="space-y-4">
            <SearchFilters
              filters={filters}
              onFiltersChange={setFilters}
              availableAuthors={availableAuthors}
              availableTags={availableTags}
              isOpen={filtersOpen}
              onToggle={() => setFiltersOpen(!filtersOpen)}
            />
            <SearchHistory
              onSearchSelect={handleHistorySearch}
              isOpen={historyOpen}
              onToggle={() => setHistoryOpen(!historyOpen)}
            />
          </div>
        </div>

        {/* Search Results */}
        <div className="flex-1 min-w-0">
          {/* Results Header */}
          {(query || author || publisher || tag) && (
            <div className="flex items-center justify-between mb-4">
              <div className="text-sm text-muted-foreground">
                {loading ? (
                  "搜索中..."
                ) : (
                  <>
                    找到 {books.length} 本书籍
                    {allBooks.length !== books.length && (
                      <span className="ml-2">
                        (从 {allBooks.length} 本中筛选)
                      </span>
                    )}
                  </>
                )}
              </div>
              <Button
                variant="ghost"
                size="sm"
                className="hidden md:flex"
                onClick={() => setFiltersOpen(!filtersOpen)}
              >
                <Filter className="w-4 h-4 mr-2" />
                {filtersOpen ? '隐藏过滤器' : '显示过滤器'}
              </Button>
            </div>
          )}

          {/* Books Grid */}
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6">
            {loading
              ? Array.from({ length: 10 }).map((_, i) => (
                  <div key={i} className="h-[280px]">
                    <Skeleton className="h-full w-full rounded-xl" />
                  </div>
                ))
              : books.map((book) => (
                  <div key={book.id} className="h-[280px]">
                    <BookCard book={book} moreInfo={true} />
                  </div>
                ))}
          </div>
          
          {/* Empty State */}
          {!loading && books.length === 0 && (query || author || publisher) && (
            <div className="text-center py-20 text-muted-foreground">
              <SearchIcon className="w-16 h-16 mx-auto mb-4 opacity-50" />
              <p className="text-lg mb-2">未找到匹配的书籍</p>
              <p className="text-sm">
                尝试调整搜索关键词或过滤条件
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Mobile Filters Modal */}
      {filtersOpen && (
        <div className="fixed inset-0 z-50 md:hidden">
          <div className="fixed inset-0 bg-black/50" onClick={() => setFiltersOpen(false)} />
          <div className="fixed right-0 top-0 h-full w-80 bg-background border-l overflow-y-auto">
            <div className="space-y-4 p-4">
              <SearchFilters
                filters={filters}
                onFiltersChange={setFilters}
                availableAuthors={availableAuthors}
                availableTags={availableTags}
                isOpen={true}
                onToggle={() => setFiltersOpen(false)}
              />
              <SearchHistory
                onSearchSelect={handleHistorySearch}
                isOpen={historyOpen}
                onToggle={() => setHistoryOpen(!historyOpen)}
              />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default function SearchPage() {
  return (
    <Suspense fallback={
      <div className="container py-8">
        <Skeleton className="h-12 w-full mb-6" />
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6">
          {[...Array(10)].map((_, idx) => (
            <Skeleton key={`skeleton-${idx}`} className="h-[320px] w-full" />
          ))}
        </div>
      </div>
    }>
      <SearchPageContent />
    </Suspense>
  )
}

