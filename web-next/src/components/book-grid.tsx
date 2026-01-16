"use client"

import { memo, useMemo } from 'react'
import { Book } from '@/types/book'
import { BookCard } from '@/components/book-card'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'

interface BookGridProps {
  books: Book[]
  loading?: boolean
  emptyMessage?: React.ReactNode
  moreInfo?: boolean
  proxyImage?: boolean
  showSummaryButton?: boolean
  /** Auto-fetch complete data for incomplete books (useful for chat responses) */
  autoFetchCompleteData?: boolean
  skeletonCount?: number
  columns?: {
    base?: number
    sm?: number
    md?: number
    lg?: number
    xl?: number
  }
  className?: string
}

export const BookGrid = memo(function BookGrid({
  books,
  loading = false,
  emptyMessage = "No books found.",
  moreInfo = false,
  proxyImage = false,
  showSummaryButton = false,
  autoFetchCompleteData = false,
  skeletonCount = 10,
  columns = {
    base: 2,
    sm: 2,
    md: 3,
    lg: 4,
    xl: 5
  },
  className
}: BookGridProps) {
  // 使用 useMemo 缓存 grid class 计算
  // 注意：Tailwind 需要完整的类名，不能动态拼接
  const gridClass = useMemo(() => {
    const { base = 2, sm = 2, md = 3, lg = 4, xl = 5 } = columns

    // 映射列数到完整的 Tailwind 类名
    const baseClass = base === 2 ? 'grid-cols-2' : base === 3 ? 'grid-cols-3' : base === 4 ? 'grid-cols-4' : 'grid-cols-5'
    const smClass = sm === 2 ? 'sm:grid-cols-2' : sm === 3 ? 'sm:grid-cols-3' : sm === 4 ? 'sm:grid-cols-4' : 'sm:grid-cols-5'
    const mdClass = md === 2 ? 'md:grid-cols-2' : md === 3 ? 'md:grid-cols-3' : md === 4 ? 'md:grid-cols-4' : 'md:grid-cols-5'
    const lgClass = lg === 2 ? 'lg:grid-cols-2' : lg === 3 ? 'lg:grid-cols-3' : lg === 4 ? 'lg:grid-cols-4' : 'lg:grid-cols-5'
    const xlClass = xl === 2 ? 'xl:grid-cols-2' : xl === 3 ? 'xl:grid-cols-3' : xl === 4 ? 'xl:grid-cols-4' : 'xl:grid-cols-5'

    return cn(
      "grid gap-6",
      baseClass,
      smClass,
      mdClass,
      lgClass,
      xlClass,
      className
    )
  }, [columns, className])

  // 加载状态 - 显示骨架屏
  if (loading) {
    return (
      <div className={gridClass}>
        {Array.from({ length: skeletonCount }).map((_, i) => (
          <div
            key={`skeleton-${i}`}
            className="flex p-3 gap-3 min-h-[140px] rounded-xl border bg-card/50"
          >
            {/* Cover Skeleton */}
            <Skeleton className="shrink-0 w-20 h-[120px] rounded-md" />

            {/* Content Skeleton */}
            <div className="flex-1 flex flex-col justify-between py-1">
              <div className="space-y-2">
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-3 w-1/2" />
              </div>
              <div className="flex gap-2 items-center mt-2">
                <Skeleton className="h-5 w-12 rounded-full" />
                <Skeleton className="h-3 w-10" />
              </div>
            </div>
          </div>
        ))}
      </div>
    )
  }

  // 空状态 - 显示空消息
  if (books.length === 0) {
    return (
      <div className="text-center py-20 text-muted-foreground">
        {emptyMessage}
      </div>
    )
  }

  // 正常状态 - 显示书籍网格
  return (
    <div className={gridClass}>
      {books.map((book) => (
        <BookCard
          key={book.id}
          book={book}
          moreInfo={moreInfo}
          proxyImage={proxyImage}
          showSummaryButton={showSummaryButton}
          autoFetchCompleteData={autoFetchCompleteData}
        />
      ))}
    </div>
  )
})
