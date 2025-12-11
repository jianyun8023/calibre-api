"use client"

import { useState, useEffect } from "react"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { ScrollArea } from "@/components/ui/scroll-area"
import { fetchBookToc, fetchChapterContent } from "@/lib/api/books"
import { toast } from "sonner"
import { List, ChevronRight, ChevronDown, FileText, ExternalLink, Eye } from "lucide-react"
import { useRouter } from "next/navigation"

interface BookTocDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  bookId: string
  bookTitle: string
}

interface ChapterContentViewerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  bookId: string
  chapterTitle: string
  chapterPath: string
}

interface TocItem {
  title: string
  href?: string
  level?: number
  children?: TocItem[]
}

interface EpubNavPoint {
  text: string
  content: {
    src: string
  }
}

interface TocResponse {
  points?: EpubNavPoint[]
  metadata?: any
  manifest?: any
  baseDir?: string
}

export function BookTocDialog({ open, onOpenChange, bookId, bookTitle }: BookTocDialogProps) {
  const router = useRouter()
  const [toc, setToc] = useState<TocItem[] | null>(null)
  const [loading, setLoading] = useState(false)
  const [expandedItems, setExpandedItems] = useState<Set<string>>(new Set())
  const [chapterViewerOpen, setChapterViewerOpen] = useState(false)
  const [selectedChapter, setSelectedChapter] = useState<{title: string, path: string} | null>(null)

  useEffect(() => {
    if (open && bookId) {
      loadToc()
    }
  }, [open, bookId])

  const loadToc = async () => {
    setLoading(true)
    try {
      const data: TocResponse = await fetchBookToc(bookId)
      
      // Handle different TOC data formats
      let tocData: TocItem[] = []
      
      if (data && data.points && Array.isArray(data.points)) {
        // Handle EPUB NavPoint format from backend
        tocData = data.points.map((point: EpubNavPoint) => ({
          title: point.text || 'Untitled',
          href: point.content?.src || undefined,
          level: 1
        }))
      } else if (Array.isArray(data)) {
        // Handle direct array format
        tocData = data.map((item: any) => ({
          title: item.title || item.text || item.Text || item.name || 'Untitled',
          href: item.href || item.src || item.content?.src || item.Content?.Src || undefined,
          level: item.level || 1,
          children: item.children || undefined
        }))
      } else if (data && typeof data === 'object') {
        // Handle other object formats
        if (data.toc && Array.isArray(data.toc)) {
          tocData = data.toc
        } else if (data.chapters && Array.isArray(data.chapters)) {
          tocData = data.chapters
        } else if (data.contents && Array.isArray(data.contents)) {
          tocData = data.contents
        } else {
          // Convert object to array format
          tocData = Object.entries(data).map(([key, value]) => ({
            title: key,
            href: typeof value === 'string' ? value : undefined,
            level: 1
          }))
        }
      }
      
      setToc(tocData)
      
      if (tocData.length === 0) {
        toast.info("No table of contents available for this book")
      }
    } catch (error) {
      console.error("Failed to load TOC:", error)
      toast.error("Failed to load table of contents")
      setToc([])
    } finally {
      setLoading(false)
    }
  }

  const toggleExpanded = (itemKey: string) => {
    const newExpanded = new Set(expandedItems)
    if (newExpanded.has(itemKey)) {
      newExpanded.delete(itemKey)
    } else {
      newExpanded.add(itemKey)
    }
    setExpandedItems(newExpanded)
  }

  const handleItemClick = (item: TocItem) => {
    if (item.href) {
      // Open chapter content viewer using backend API
      setSelectedChapter({
        title: item.title,
        path: item.href
      })
      setChapterViewerOpen(true)
    } else {
      // If no href, just navigate to the reader
      router.push(`/read/${bookId}`)
      onOpenChange(false)
    }
  }

  const renderTocItem = (item: TocItem, index: number, level: number = 0) => {
    const itemKey = `${level}-${index}-${item.title}`
    const hasChildren = item.children && item.children.length > 0
    const isExpanded = expandedItems.has(itemKey)
    const paddingLeft = level * 20 + 12

    return (
      <div key={itemKey} className="select-none">
        <div
          className={`
            flex items-center py-2 px-3 hover:bg-accent/50 cursor-pointer rounded-md transition-colors
            ${item.href ? 'hover:bg-accent' : ''}
          `}
          style={{ paddingLeft: `${paddingLeft}px` }}
          onClick={() => {
            if (hasChildren) {
              toggleExpanded(itemKey)
            } else if (item.href) {
              handleItemClick(item)
            }
          }}
        >
          {hasChildren ? (
            isExpanded ? (
              <ChevronDown className="w-4 h-4 mr-2 text-muted-foreground" />
            ) : (
              <ChevronRight className="w-4 h-4 mr-2 text-muted-foreground" />
            )
          ) : (
            <FileText className="w-4 h-4 mr-2 text-muted-foreground" />
          )}
          
          <span className="flex-1 text-sm font-medium truncate" title={item.title}>
            {item.title}
          </span>
          
          {item.href && (
            <Eye className="w-3 h-3 ml-2 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
          )}
        </div>
        
        {hasChildren && isExpanded && (
          <div className="ml-2">
            {item.children!.map((child, childIndex) => 
              renderTocItem(child, childIndex, level + 1)
            )}
          </div>
        )}
      </div>
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <List className="w-5 h-5" />
            Table of Contents
          </DialogTitle>
          <DialogDescription>
            {bookTitle}
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-hidden">
          {loading ? (
            <div className="space-y-2 p-4">
              {Array.from({ length: 8 }).map((_, i) => (
                <div key={i} className="flex items-center space-x-2">
                  <Skeleton className="h-4 w-4" />
                  <Skeleton className="h-4 flex-1" />
                </div>
              ))}
            </div>
          ) : toc && toc.length > 0 ? (
            <ScrollArea className="h-full">
              <div className="p-2 space-y-1">
                {toc.map((item, index) => renderTocItem(item, index))}
              </div>
            </ScrollArea>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
              <List className="w-12 h-12 mb-4 opacity-50" />
              <p className="text-lg mb-2">No Table of Contents</p>
              <p className="text-sm text-center">
                This book doesn't have a table of contents available,<br />
                or it hasn't been extracted yet.
              </p>
              <Button 
                variant="outline" 
                className="mt-4"
                onClick={() => {
                  toast.info("TOC extraction can be triggered from the Tasks page")
                  onOpenChange(false)
                }}
              >
                Learn More
              </Button>
            </div>
          )}
        </div>

        <div className="flex justify-end gap-2 pt-4 border-t">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
          {toc && toc.length > 0 && (
            <Button onClick={() => {
              router.push(`/read/${bookId}`)
              onOpenChange(false)
            }}>
              Start Reading
            </Button>
          )}
        </div>
      </DialogContent>
      
      {/* Chapter Content Viewer */}
      {selectedChapter && (
        <ChapterContentViewer
          open={chapterViewerOpen}
          onOpenChange={setChapterViewerOpen}
          bookId={bookId}
          chapterTitle={selectedChapter.title}
          chapterPath={selectedChapter.path}
        />
      )}
    </Dialog>
  )
}

// Chapter Content Viewer Component
function ChapterContentViewer({ open, onOpenChange, bookId, chapterTitle, chapterPath }: ChapterContentViewerProps) {
  const [content, setContent] = useState<string>("")
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (open && chapterPath) {
      loadChapterContent()
    }
  }, [open, chapterPath])

  const loadChapterContent = async () => {
    setLoading(true)
    try {
      const htmlContent = await fetchChapterContent(bookId, chapterPath)
      setContent(htmlContent)
    } catch (error) {
      console.error("Failed to load chapter content:", error)
      setContent(`<p class="error">Failed to load chapter content: ${error instanceof Error ? error.message : 'Unknown error'}</p>`)
      toast.error("Failed to load chapter content")
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FileText className="w-5 h-5" />
            {chapterTitle}
          </DialogTitle>
          <DialogDescription>
            Chapter content loaded from backend API
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-hidden">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
              <span className="ml-3">Loading chapter content...</span>
            </div>
          ) : (
            <ScrollArea className="h-full">
              <div 
                className="prose prose-gray dark:prose-invert max-w-none p-6"
                dangerouslySetInnerHTML={{ __html: content }}
                style={{
                  fontSize: '16px',
                  lineHeight: '1.6',
                  fontFamily: 'Georgia, serif'
                }}
              />
            </ScrollArea>
          )}
        </div>

        <div className="flex justify-end gap-2 pt-4 border-t">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
          <Button onClick={() => {
            window.open(`/api/read/${bookId}/file/${chapterPath.replace(`/read/${bookId}/file/`, '')}`, '_blank')
          }}>
            Open in New Tab
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}