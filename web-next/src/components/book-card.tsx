"use client"

import { useState } from 'react'
import { Card } from '@/components/ui/card'
import { Book } from '@/types/book'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { useRouter } from 'next/navigation'
import Image from 'next/image'

interface BookCardProps {
  book: Book
  moreInfo?: boolean
  proxyImage?: boolean
}

export function BookCard({ book, moreInfo = false, proxyImage = false }: BookCardProps) {
  const router = useRouter()
  const [tilt, setTilt] = useState({ x: 0, y: 0 })

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

  return (
    <div 
      className="perspective-1000 h-full"
      onMouseMove={handleMouseMove}
      onMouseLeave={resetTilt}
    >
      <Card
        className={cn(
          "relative overflow-hidden cursor-pointer h-full transition-all duration-300 ease-out border-border/50 bg-background/50 hover:bg-accent/5 hover:border-accent/50 hover:shadow-xl",
          "glass"
        )}
        style={{
          transform: `rotateX(${tilt.x}deg) rotateY(${tilt.y}deg)`,
          transition: tilt.x === 0 && tilt.y === 0 ? 'transform 0.5s ease' : 'none'
        }}
        onClick={() => router.push(`/detail/${book.id}`)}
      >
        <div className="flex p-3 gap-4 h-full">
          {/* Cover Image */}
          <div className="shrink-0 w-20 relative rounded-sm overflow-hidden bg-muted aspect-[2/3]">
            {imageUrl ? (
               <Image
                 src={imageUrl}
                 alt={book.title}
                 fill
                 className="object-cover transition-transform duration-500 hover:scale-110"
                 sizes="80px"
                 loading="lazy"
                 unoptimized
               />
            ) : (
              <div className="flex items-center justify-center h-full text-muted-foreground">
                No Cover
              </div>
            )}
            <div className="absolute inset-0 bg-gradient-to-b from-transparent to-black/20 opacity-0 hover:opacity-100 transition-opacity" />
          </div>

          {/* Info */}
          <div className="flex-1 flex flex-col justify-between min-w-0 gap-2">
            <div>
              <h3 className="font-semibold text-base leading-tight line-clamp-2 mb-1 text-foreground">
                {book.title}
              </h3>
              {book.authors?.length > 0 && (
                <div className="text-sm text-muted-foreground flex items-center gap-1">
                  {/* Icon could go here */}
                  <span className="truncate">{book.authors.join(', ')}</span>
                </div>
              )}
            </div>

            <div className="flex flex-col gap-1 text-xs text-muted-foreground">
              {moreInfo && book.publisher && (
                <div className="truncate">{book.publisher}</div>
              )}
              {moreInfo && book.pubdate && (
                <div>{new Date(book.pubdate).getFullYear()}</div>
              )}
              
              {book.score !== undefined && (
                <div className="mt-1">
                  <Badge variant={book.score > 0.7 ? "default" : "secondary"} className="text-[10px] px-1 py-0 h-5">
                    Match: {(book.score * 100).toFixed(0)}%
                  </Badge>
                </div>
              )}
            </div>
          </div>
        </div>
      </Card>
    </div>
  )
}

