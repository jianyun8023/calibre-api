"use client"

import { useState } from 'react'
import { Card } from '@/components/ui/card'
import { Book } from '@/types/book'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useRouter } from 'next/navigation'
import { Sparkles, RefreshCw } from 'lucide-react'
import Image from 'next/image'
import { useBookData } from '@/hooks/use-book-data'

interface BookCardProps {
  book: Book
  moreInfo?: boolean
  proxyImage?: boolean
  showSummaryButton?: boolean
  /** Auto-fetch complete data if incomplete (default: true for chat, false otherwise) */
  autoFetchCompleteData?: boolean
}

export function BookCard({ 
  book: initialBook, 
  moreInfo = false, 
  proxyImage = false, 
  showSummaryButton = false,
  autoFetchCompleteData = false
}: BookCardProps) {
  const router = useRouter()
  const [tilt, setTilt] = useState({ x: 0, y: 0 })
  
  // Use enhanced book data with auto-fetch capability
  const { book, loading: fetchingData, isComplete } = useBookData(initialBook, {
    autoFetch: autoFetchCompleteData,
    requiredFields: ['cover', 'authors', 'publisher'],
  })

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    const el = e.currentTarget
    const rect = el.getBoundingClientRect()
    const x = e.clientX - rect.left
    const y = e.clientY - rect.top
    
    // Calculate rotation (max 5 degrees)
    // tiltY depends on x position (left/right)
    // tiltX depends on y position (top/bottom)
    const tY = ((x / rect.width) - 0.5) * 10
    const tX = ((y / rect.height) - 0.5) * -10
    
    setTilt({ x: tX, y: tY })
  }

  const resetTilt = () => {
    setTilt({ x: 0, y: 0 })
  }

  const imageUrl = proxyImage && book.cover 
    ? `/api/proxy/cover/${book.cover}` 
    : book.cover

  // 简单的截断函数
  const truncate = (text: string, length: number) => {
    if (!text) return ''
    return text.length > length ? text.substring(0, length - 3) + '...' : text
  }

  // 总结此书 - 跳转到聊天页面并设置消息
  const handleSummarize = (e: React.MouseEvent) => {
    e.stopPropagation() // 防止触发卡片点击
    
    // 构建总结请求，包含书籍ID便于LLM获取详细信息
    const summaryPrompt = `请总结一下书籍 ID ${book.id} 《${book.title}》${book.authors?.length > 0 ? ' 作者：' + book.authors.join(', ') : ''}的内容`
    
    // 跳转到聊天页面，通过URL参数传递消息
    router.push(`/chat?message=${encodeURIComponent(summaryPrompt)}`)
  }

  return (
    <div 
      className="perspective-1000"
      onMouseMove={handleMouseMove}
      onMouseLeave={resetTilt}
    >
      <Card
        className={cn(
          "relative overflow-hidden cursor-pointer transition-all duration-300 ease-out border-border/50 bg-background/50 hover:bg-accent/5 hover:border-accent/50 hover:shadow-xl",
          "glass"
        )}
        style={{
          transform: `rotateX(${tilt.x}deg) rotateY(${tilt.y}deg)`,
          transition: tilt.x === 0 && tilt.y === 0 ? 'transform 0.5s ease' : 'none'
        }}
        onClick={() => router.push(`/detail/${book.id}`)}
      >
        <div className="flex p-3 gap-3 min-h-[140px]">
          {/* Cover Image */}
          <div className="shrink-0 w-20 h-[120px] relative rounded-md overflow-hidden bg-muted/30 border border-border/30">
            {/* Loading indicator */}
            {fetchingData && (
              <div className="absolute inset-0 bg-background/80 flex items-center justify-center z-10">
                <RefreshCw className="w-6 h-6 animate-spin text-muted-foreground" />
              </div>
            )}
            
            {imageUrl ? (
               <Image
                 src={imageUrl}
                 alt={book.title}
                 fill
                 className="object-cover transition-transform duration-500 hover:scale-110"
                 sizes="96px"
                 loading="lazy"
                 unoptimized
                 onError={(e) => {
                   // Hide broken image, show fallback
                   e.currentTarget.style.display = 'none'
                   const parent = e.currentTarget.parentElement
                   if (parent) {
                     const fallback = parent.querySelector('.cover-fallback')
                     if (fallback) {
                       (fallback as HTMLElement).style.display = 'flex'
                     }
                   }
                 }}
               />
            ) : null}
            {/* Fallback "No Cover" display */}
            <div 
              className={cn(
                "cover-fallback absolute inset-0 flex flex-col items-center justify-center text-muted-foreground/60 text-center p-2",
                imageUrl ? "hidden" : "flex"
              )}
            >
              <svg 
                className="w-8 h-8 mb-1" 
                fill="none" 
                stroke="currentColor" 
                viewBox="0 0 24 24"
              >
                <path 
                  strokeLinecap="round" 
                  strokeLinejoin="round" 
                  strokeWidth={1.5} 
                  d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" 
                />
              </svg>
              <span className="text-[10px] font-medium">No Cover</span>
            </div>
            <div className="absolute inset-0 bg-gradient-to-b from-transparent to-black/10 opacity-0 hover:opacity-100 transition-opacity" />
          </div>

          {/* Info */}
          <div className="flex-1 flex flex-col min-w-0 justify-between">
            {/* Top: Title and Author */}
            <div className="space-y-1">
              <h3 className="font-semibold text-sm leading-tight line-clamp-2 text-foreground">
                {book.title}
              </h3>
              {book.authors?.length > 0 && (
                <div className="text-xs text-muted-foreground/80 line-clamp-1">
                  {book.authors.join(', ')}
                </div>
              )}
            </div>

            {/* Bottom: Metadata and Button */}
            <div className="space-y-1.5 mt-2">
              {/* Publisher (if moreInfo) */}
              {moreInfo && book.publisher && (
                <div className="text-xs text-muted-foreground/70 line-clamp-1">
                  {book.publisher}
                </div>
              )}
              
              {/* Score Badge and Year */}
              <div className="flex items-center gap-1.5 flex-wrap">
                {book.score !== undefined && (
                  <Badge 
                    variant={book.score > 0.7 ? "default" : "secondary"} 
                    className="text-[10px] px-1.5 py-0.5 h-5 font-medium"
                  >
                    {(book.score * 100).toFixed(0)}%
                  </Badge>
                )}
                {moreInfo && book.pubdate && (
                  <span className="text-xs text-muted-foreground/60">
                    {new Date(book.pubdate).getFullYear()}
                  </span>
                )}
              </div>

              {/* Summary Button */}
              {showSummaryButton && (
                <Button
                  variant="outline"
                  size="sm"
                  className="w-full h-7 text-xs gap-1 hover:bg-primary hover:text-primary-foreground transition-colors"
                  onClick={handleSummarize}
                >
                  <Sparkles className="h-3 w-3" />
                  总结
                </Button>
              )}
            </div>
          </div>
        </div>
      </Card>
    </div>
  )
}

