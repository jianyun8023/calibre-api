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
import { Search as SearchIcon } from "lucide-react"

function SearchPageContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  
  const query = searchParams.get('q') || ''
  const mode = searchParams.get('mode') || 'hybrid'
  const author = searchParams.get('author')
  const publisher = searchParams.get('publisher')
  const tag = searchParams.get('tag')

  const [input, setInput] = useState(query)
  const [books, setBooks] = useState<Book[]>([])
  const [loading, setLoading] = useState(false)

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
      setBooks(data.records || data ||[])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }, [query, mode, author, publisher, tag])

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

  const handleModeChange = (newMode: string) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set('mode', newMode)
    router.push(`/search?${params.toString()}`)
  }

  return (
    <div className="space-y-8">
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

      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6">
        {loading
          ? Array.from({ length: 10 }).map((_, i) => (
              <div key={i} className="h-[280px]">
                <Skeleton className="h-full w-full rounded-xl" />
              </div>
            ))
          : books.map((book, bookIndex) => (
              <div key={`book-${book.id}-${bookIndex}`} className="h-[280px]">
                <BookCard book={book} moreInfo={true} />
              </div>
            ))}
      </div>
      
      {!loading && books.length === 0 && (query || author || publisher) && (
        <div className="text-center py-20 text-muted-foreground">
          No results found.
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

