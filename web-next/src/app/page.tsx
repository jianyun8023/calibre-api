"use client"

import { useEffect, useState, useCallback } from "react"
import { BookCard } from "@/components/book-card"
import { fetchRandomBooks, fetchRecentBooks } from "@/lib/api/books"
import type { Book } from "@/types/book"
import { Skeleton } from "@/components/ui/skeleton"
import { Button } from "@/components/ui/button"
import { ArrowRight, RefreshCcw } from "lucide-react"
import Link from "next/link"

const RANDOM_BOOKS_COUNT = 5  // 与 API limit 保持一致
const RECENT_BOOKS_COUNT = 15  // 优化分页：15 在 5列(3行), 3列(5行), 2列(7.5行) 下接近满行

export default function Home() {
  const [randomBooks, setRandomBooks] = useState<Book[]>([])
  const [recentBooks, setRecentBooks] = useState<Book[]>([])
  const [loadingRandom, setLoadingRandom] = useState(true)
  const [loadingRecent, setLoadingRecent] = useState(true)
  const [fadeIn, setFadeIn] = useState(false)

  const loadRandomBooks = useCallback(async () => {
    // 先淡出
    setFadeIn(false)
    // 等待淡出动画完成
    await new Promise(resolve => setTimeout(resolve, 150))
    
    setLoadingRandom(true)
    try {
      const data = await fetchRandomBooks()
      setRandomBooks(data.records || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoadingRandom(false)
      // 加载完成后淡入
      setTimeout(() => setFadeIn(true), 50)
    }
  }, [])

  const loadRecentBooks = useCallback(async () => {
    setLoadingRecent(true)
    try {
      const data = await fetchRecentBooks(RECENT_BOOKS_COUNT, 0)
      setRecentBooks(data.records || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoadingRecent(false)
    }
  }, [])

  useEffect(() => {
    loadRandomBooks()
    loadRecentBooks()
  }, [loadRandomBooks, loadRecentBooks])
  
  // 初始加载完成后淡入
  useEffect(() => {
    if (!loadingRandom) {
      setFadeIn(true)
    }
  }, [loadingRandom])

  return (
    <div className="space-y-10">
      {/* Hero / Random Section */}
      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold tracking-tight">Discover</h2>
          <Button variant="ghost" size="sm" onClick={loadRandomBooks} disabled={loadingRandom}>
            <RefreshCcw className={`h-4 w-4 mr-2 ${loadingRandom ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
        
        <div 
          className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6 transition-opacity duration-300"
          style={{ opacity: loadingRandom || fadeIn ? 1 : 0 }}
        >
          {loadingRandom
            ? Array.from({ length: RANDOM_BOOKS_COUNT }, (_, i) => `skeleton-random-${i}`).map((id) => (
                <Skeleton key={id} className="h-48 w-full rounded-xl" />
              ))
            : randomBooks.slice(0, RANDOM_BOOKS_COUNT).map((book) => (
                <BookCard key={book.id} book={book} />
              ))}
        </div>
      </section>

      {/* Recent Books Section */}
      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold tracking-tight">Recently Added</h2>
          <Button variant="link" asChild>
            <Link href="/books">
              View All <ArrowRight className="ml-2 h-4 w-4" />
            </Link>
          </Button>
        </div>

        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6">
          {loadingRecent
            ? Array.from({ length: RECENT_BOOKS_COUNT }, (_, i) => `skeleton-recent-${i}`).map((id) => (
                <Skeleton key={id} className="h-48 w-full rounded-xl" />
              ))
            : recentBooks.slice(0, RECENT_BOOKS_COUNT).map((book) => (
                <BookCard key={book.id} book={book} moreInfo />
              ))}
        </div>
      </section>
    </div>
  )
}
