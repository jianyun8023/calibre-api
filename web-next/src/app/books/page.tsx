"use client"

import { useEffect, useState, Suspense } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { BookCard } from "@/components/book-card"
import { fetchAllBooks } from "@/lib/api/books"
import type { Book } from "@/types/book"
import { Skeleton } from "@/components/ui/skeleton"
import { Button } from "@/components/ui/button"
import { ArrowLeft, ArrowRight } from "lucide-react"

// 分离组件以使用 useSearchParams (需要 Suspense)
function BookList() {
  const router = useRouter()
  const searchParams = useSearchParams()
  
  const [books, setBooks] = useState<Book[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  
  // State from URL or defaults
  const cursor = searchParams.get('cursor') || ''
  const page = parseInt(searchParams.get('page') || '1', 10)
  // 优化分页数量：20 在 5列(4行), 4列(5行), 2列(10行) 下都是满行
  const limit = 20

  // Internal state for next cursor
  const [nextCursor, setNextCursor] = useState('')
  // History stack for previous cursors to enable "Previous" button
  const [cursorHistory, setCursorStack] = useState<string[]>([])

  useEffect(() => {
    const loadBooks = async () => {
      setLoading(true)
      try {
        const data = await fetchAllBooks(limit, cursor)
        setBooks(data.records || [])
        setTotal(data.total)
        setNextCursor(data.next_cursor || '')
      } catch (e) {
        console.error(e)
      } finally {
        setLoading(false)
      }
    }
    
    loadBooks()
  }, [cursor])

  const handleNext = () => {
    if (nextCursor) {
      setCursorStack([...cursorHistory, cursor])
      updateUrl(nextCursor, page + 1)
    }
  }

  const handlePrev = () => {
    if (page > 1) {
      const prevCursor = cursorHistory[cursorHistory.length - 1] || ''
      setCursorStack(cursorHistory.slice(0, -1))
      updateUrl(prevCursor, page - 1)
    }
  }

  const updateUrl = (newCursor: string, newPage: number) => {
    const params = new URLSearchParams(searchParams.toString())
    if (newCursor) params.set('cursor', newCursor)
    else params.delete('cursor')
    params.set('page', newPage.toString())
    
    router.push(`/books?${params.toString()}`)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">All Books</h1>
        <div className="text-sm text-muted-foreground bg-muted/50 px-3 py-1 rounded-md">
          Total: {total}
        </div>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6 min-h-[500px]">
        {loading
          ? Array.from({ length: limit }, (_, i) => `skeleton-books-${i}`).map((id) => (
              <Skeleton key={id} className="h-48 w-full rounded-xl" />
            ))
          : books.map((book) => (
              <BookCard key={book.id} book={book} moreInfo />
            ))}
      </div>

      {!loading && books.length === 0 && (
        <div className="text-center py-20 text-muted-foreground">
          No books found.
        </div>
      )}

      {/* Pagination */}
      <div className="flex items-center justify-center gap-4 py-8">
        <Button 
          variant="outline" 
          onClick={handlePrev} 
          disabled={page <= 1 || loading}
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          Previous
        </Button>
        <span className="text-sm font-medium">Page {page}</span>
        <Button 
          variant="outline" 
          onClick={handleNext} 
          disabled={!nextCursor || loading}
        >
          Next
          <ArrowRight className="h-4 w-4 ml-2" />
        </Button>
      </div>
    </div>
  )
}

export default function BooksPage() {
  return (
    <Suspense fallback={
      <div className="p-8">
        <Skeleton className="h-10 w-40 mb-8" />
        <div className="grid grid-cols-4 gap-4">
          {Array.from({length:8}, (_, i) => `fallback-skeleton-${i}`).map((id) => (
            <Skeleton key={id} className="h-60" />
          ))}
        </div>
      </div>
    }>
      <BookList />
    </Suspense>
  )
}

