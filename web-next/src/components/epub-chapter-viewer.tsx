"use client"

import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { ChevronLeft, ChevronRight, List, Download } from "lucide-react"
import { fetchBookToc, fetchChapterContent } from "@/lib/api/books"

interface TocItem {
  text: string
  content: {
    src: string
  }
}

interface EpubChapterViewerProps {
  bookId: string
  bookTitle: string
}

// Backend TOC response format
interface TocResponseData {
  points?: TocItem[]
  [key: string]: unknown
}

export function EpubChapterViewer({ bookId, bookTitle }: EpubChapterViewerProps) {
  const [toc, setToc] = useState<TocItem[]>([])
  const [currentChapter, setCurrentChapter] = useState<number>(0)
  const [chapterContent, setChapterContent] = useState<string>("")
  const [loading, setLoading] = useState(true)
  const [contentLoading, setContentLoading] = useState(false)

  useEffect(() => {
    loadToc()
  }, [bookId])

  useEffect(() => {
    if (toc.length > 0) {
      loadChapter(currentChapter)
    }
  }, [currentChapter, toc])

  const loadToc = async () => {
    try {
      const data = await fetchBookToc(bookId) as unknown as TocResponseData
      if (data && data.points && Array.isArray(data.points)) {
        setToc(data.points)
      }
    } catch (error) {
      console.error("Failed to load TOC:", error)
    } finally {
      setLoading(false)
    }
  }

  const loadChapter = async (chapterIndex: number) => {
    if (!toc[chapterIndex]) return
    
    setContentLoading(true)
    try {
      const chapter = toc[chapterIndex]
      // Extract the file path from the src URL
      const srcPath = chapter.content.src
      // Remove the URL prefix to get the actual file path
      // Format could be: /read/69431/file/OEBPS/Text/chapter1.xhtml
      const filePath = srcPath.replace(/^\/read\/[^/]+\/file\//, '')
      
      // Use the API helper function which handles the request properly
      const content = await fetchChapterContent(bookId, filePath)
      setChapterContent(content)
    } catch (error) {
      console.error("Failed to load chapter:", error)
      setChapterContent(`<p class="text-destructive">Error loading chapter: ${error instanceof Error ? error.message : 'Unknown error'}</p>`)
    } finally {
      setContentLoading(false)
    }
  }

  const goToPreviousChapter = () => {
    if (currentChapter > 0) {
      setCurrentChapter(currentChapter - 1)
    }
  }

  const goToNextChapter = () => {
    if (currentChapter < toc.length - 1) {
      setCurrentChapter(currentChapter + 1)
    }
  }

  if (loading) {
    return (
      <div className="h-full flex items-center justify-center p-2">
        <div className="text-center">
          <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-primary mx-auto mb-2"></div>
          <p className="text-xs">Loading book chapters...</p>
        </div>
      </div>
    )
  }

  if (toc.length === 0) {
    return (
      <div className="h-full flex items-center justify-center p-2">
        <div className="text-center max-w-md">
          <List className="w-6 h-6 mx-auto mb-2 opacity-50" />
          <h2 className="text-sm font-semibold mb-1 leading-tight">No Chapters Found</h2>
          <p className="text-xs text-muted-foreground mb-2 leading-tight">
            This book doesn't have a readable table of contents.
          </p>
          <Button size="sm" onClick={() => window.open(`/api/download/book/${bookId}.epub`, '_blank')} className="h-7 text-xs">
            <Download className="w-3 h-3 mr-1.5" />
            Download EPUB
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex">
      {/* Chapter Navigation Sidebar */}
      <div className="w-80 border-r bg-muted/30 flex flex-col">
        <div className="px-2 py-1 border-b shrink-0">
          <h2 className="text-xs font-semibold truncate leading-none mb-0.5" title={bookTitle}>
            {bookTitle}
          </h2>
          <p className="text-xs text-muted-foreground leading-none">
            {toc.length} chapters
          </p>
        </div>
        <ScrollArea className="flex-1">
          <div className="p-2">
            {toc.map((chapter, index) => (
              <button
                key={index}
                onClick={() => setCurrentChapter(index)}
                className={`w-full text-left p-3 rounded-md mb-1 transition-colors ${
                  currentChapter === index
                    ? 'bg-primary text-primary-foreground'
                    : 'hover:bg-accent'
                }`}
              >
                <div className="font-medium text-sm truncate">
                  {chapter.text || `Chapter ${index + 1}`}
                </div>
              </button>
            ))}
          </div>
        </ScrollArea>
      </div>

      {/* Chapter Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Chapter Header */}
        <div className="px-2 py-1 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 shrink-0">
          <div className="flex items-center justify-between">
            <div className="flex-1 min-w-0">
              <h1 className="text-xs font-semibold truncate leading-none mb-0.5">
                {toc[currentChapter]?.text || `Chapter ${currentChapter + 1}`}
              </h1>
              <p className="text-xs text-muted-foreground leading-none">
                Chapter {currentChapter + 1} of {toc.length}
              </p>
            </div>
            <div className="flex gap-0.5 ml-2 shrink-0">
              <Button
                variant="outline"
                size="sm"
                onClick={goToPreviousChapter}
                disabled={currentChapter === 0}
                className="h-6 w-6 p-0"
              >
                <ChevronLeft className="w-3 h-3" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={goToNextChapter}
                disabled={currentChapter === toc.length - 1}
                className="h-6 w-6 p-0"
              >
                <ChevronRight className="w-3 h-3" />
              </Button>
            </div>
          </div>
        </div>

        {/* Chapter Content */}
        <ScrollArea className="flex-1">
          <div className="p-4 max-w-4xl mx-auto">
            {contentLoading ? (
              <div className="text-center py-4">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary mx-auto mb-2"></div>
                <p className="text-xs">Loading chapter...</p>
              </div>
            ) : (
              <div 
                className="prose prose-sm prose-gray dark:prose-invert max-w-none"
                dangerouslySetInnerHTML={{ __html: chapterContent }}
              />
            )}
          </div>
        </ScrollArea>
      </div>
    </div>
  )
}