"use client"

import { useEffect, useState, useCallback } from "react"
import { useTranslations } from 'next-intl'
import { BookGrid } from "@/components/book-grid"
import { fetchRandomBooks, fetchRecentBooks } from "@/lib/api/books"
import type { Book } from "@/types/book"
import { Button } from "@/components/ui/button"
import { ArrowRight, RefreshCcw } from "lucide-react"
import Link from "next/link"
import { cn } from "@/lib/utils"

const RANDOM_BOOKS_COUNT = 5  // 与 API limit 保持一致
const RECENT_BOOKS_COUNT = 15  // 优化分页：15 在 5列(3行), 3列(5行), 2列(7.5行) 下接近满行

export default function HomeClient({ locale }: { locale: string }) {
  const t = useTranslations('common')
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
          <h2 className="text-2xl font-bold tracking-tight text-gradient dark:text-foreground">Discover</h2>
          <Button
            variant="ghost"
            size="sm"
            onClick={loadRandomBooks}
            disabled={loadingRandom}
            className="group hover:bg-primary/10 transition-all duration-300"
          >
            <RefreshCcw className={cn(
              "h-4 w-4 mr-2 transition-transform duration-500",
              loadingRandom ? 'animate-spin' : 'group-hover:rotate-180'
            )} />
            Refresh
          </Button>
        </div>

        <div
          className="transition-opacity duration-300"
          style={{ opacity: loadingRandom || fadeIn ? 1 : 0 }}
        >
          <BookGrid
            books={randomBooks}
            loading={loadingRandom}
            skeletonCount={RANDOM_BOOKS_COUNT}
            emptyMessage="No books available"
          />
        </div>
      </section>

      {/* Recent Books Section */}
      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold tracking-tight text-gradient dark:text-foreground">Recently Added</h2>
          <Button variant="link" asChild className="group">
            <Link href={`/${locale}/books`} className="flex items-center">
              View All <ArrowRight className="ml-2 h-4 w-4 transition-transform group-hover:translate-x-1" />
            </Link>
          </Button>
        </div>

        <BookGrid
          books={recentBooks}
          loading={loadingRecent}
          skeletonCount={RECENT_BOOKS_COUNT}
          moreInfo
          emptyMessage="No recent books"
        />
      </section>
    </div>
  )
}
