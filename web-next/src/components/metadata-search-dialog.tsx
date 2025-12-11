"use client"

import { useState, useEffect } from "react"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { toast } from "sonner"
import { searchMetadata, type DoubanBook } from "@/lib/api/metadata"
import type { Book } from "@/types/book"
import { Search, Loader2 } from "lucide-react"
import Image from "next/image"

interface MetadataSearchDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  book: Book
  onSelect: (metadata: DoubanBook) => void
}

export function MetadataSearchDialog({ open, onOpenChange, book, onSelect }: MetadataSearchDialogProps) {
  const [query, setQuery] = useState("")
  const [searchResults, setSearchResults] = useState<DoubanBook[]>([])
  const [searching, setSearching] = useState(false)
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null)

  // Initialize search query when dialog opens
  useEffect(() => {
    if (open && book) {
      // Priority: ISBN > Title
      if (book.isbn) {
        const cleanISBN = book.isbn.replace(/-/g, '')
        setQuery(cleanISBN)
      } else if (book.title) {
        const authorQuery = book.authors && book.authors.length > 0 ? ` ${book.authors[0]}` : ""
        setQuery(book.title + authorQuery)
      }
    }
  }, [open, book])

  const handleSearch = async () => {
    if (!query.trim()) {
      toast.error("请输入搜索关键词")
      return
    }

    setSearching(true)
    setSearchResults([])
    setSelectedIndex(null)

    try {
      const response = await searchMetadata(query)
      
      // 豆瓣 API 返回格式：{ success: boolean, books: [], message?: string }
      if (response.success && response.books && response.books.length > 0) {
        setSearchResults(response.books)
        toast.success(`找到 ${response.books.length} 本相关书籍`)
      } else {
        toast.warning("未找到相关书籍")
        setSearchResults([])
      }
    } catch (error) {
      console.error("搜索元数据失败:", error)
      const message = error instanceof Error ? error.message : "搜索失败"
      toast.error(`搜索失败: ${message}`)
    } finally {
      setSearching(false)
    }
  }

  const handleConfirm = () => {
    if (selectedIndex === null) {
      toast.error("请选择一本书")
      return
    }

    onSelect(searchResults[selectedIndex])
    onOpenChange(false)
  }

  const handleClose = () => {
    setSearchResults([])
    setSelectedIndex(null)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-6xl max-h-[90vh] overflow-y-auto w-[95vw]">
        <DialogHeader>
          <DialogTitle>搜索元数据</DialogTitle>
          <DialogDescription>
            从豆瓣搜索书籍元数据，可以使用 ISBN、书名或作者进行搜索
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Search Input */}
          <div className="flex gap-2">
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSearch()}
              placeholder="输入 ISBN、书名或作者..."
              disabled={searching}
            />
            <Button onClick={handleSearch} disabled={searching}>
              {searching ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  搜索中...
                </>
              ) : (
                <>
                  <Search className="w-4 h-4 mr-2" />
                  搜索
                </>
              )}
            </Button>
          </div>

          {/* Search Results */}
          {searchResults.length > 0 && (
            <div className="border rounded-lg overflow-hidden">
              <div className="max-h-[500px] overflow-y-auto">
                {searchResults.map((result, index) => (
                  <div
                    key={index}
                    onClick={() => setSelectedIndex(index)}
                    className={`
                      flex gap-4 p-4 cursor-pointer transition-colors border-b last:border-b-0
                      ${selectedIndex === index ? "bg-accent" : "hover:bg-accent/50"}
                    `}
                  >
                    {/* Cover */}
                    <div className="relative w-20 h-28 shrink-0">
                      <Image
                        src={`/api/proxy/cover/${result.image}`}
                        alt={result.title}
                        fill
                        className="object-cover rounded"
                        unoptimized
                      />
                    </div>

                    {/* Info */}
                    <div className="flex-1 min-w-0 space-y-1">
                      <h3 className="font-semibold line-clamp-2 leading-tight">
                        {result.title}
                        {result.sub_title && (
                          <span className="text-muted-foreground">：{result.sub_title}</span>
                        )}
                      </h3>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-1 text-sm text-muted-foreground">
                        <p className="truncate">
                          <span className="font-medium">作者：</span>
                          {result.author.join(", ")}
                        </p>
                        <p className="truncate">
                          <span className="font-medium">出版社：</span>
                          {result.publisher || "-"}
                        </p>
                        <p className="truncate">
                          <span className="font-medium">出版日期：</span>
                          {result.pubdate || "-"}
                        </p>
                        <p className="truncate">
                          <span className="font-medium">ISBN：</span>
                          {result.isbn13 || result.isbn10 || "-"}
                        </p>
                        {result.rating && result.rating.average > 0 && (
                          <p className="truncate md:col-span-2">
                            <span className="font-medium">评分：</span>
                            {result.rating.average} ({result.rating.numRaters} 人评价)
                          </p>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Empty State */}
          {!searching && searchResults.length === 0 && query && (
            <div className="text-center py-12 text-muted-foreground">
              <Search className="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p>点击搜索按钮开始查找元数据</p>
            </div>
          )}

          {/* Actions */}
          <div className="flex justify-end gap-2 pt-4">
            <Button variant="outline" onClick={handleClose}>
              取消
            </Button>
            <Button onClick={handleConfirm} disabled={selectedIndex === null}>
              确认选择
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

