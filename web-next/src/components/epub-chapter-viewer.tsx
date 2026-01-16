"use client"

import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { ChevronLeft, ChevronRight, List, Download } from "lucide-react"
import { fetchBookToc } from "@/lib/api/books"

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
      const data = await fetchBookToc(bookId)
      const anyData = data as any
      if (anyData && anyData.points && Array.isArray(anyData.points)) {
        setToc(anyData.points)
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
      const filePath = srcPath.replace(`/read/${bookId}/file/`, '')

      const response = await fetch(`/api/read/${bookId}/file/${filePath}`)
      if (response.ok) {
        const content = await response.text()
        setChapterContent(content)
      } else {
        setChapterContent(`<p>Failed to load chapter content (HTTP ${response.status})</p>`)
      }
    } catch (error) {
      console.error("Failed to load chapter:", error)
      setChapterContent(`<p>Error loading chapter: ${error instanceof Error ? error.message : 'Unknown error'}</p>`)
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
      <div className="h-[calc(100vh-4rem)] flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p>Loading book chapters...</p>
        </div>
      </div>
    )
  }

  if (toc.length === 0) {
    return (
      <div className="h-[calc(100vh-4rem)] flex items-center justify-center">
        <div className="text-center">
          <List className="w-12 h-12 mx-auto mb-4 opacity-50" />
          <h2 className="text-xl font-semibold mb-2">No Chapters Found</h2>
          <p className="text-muted-foreground mb-4">
            This book doesn't have a readable table of contents.
          </p>
          <Button onClick={() => window.open(`/api/download/book/${bookId}.epub`, '_blank')}>
            <Download className="w-4 h-4 mr-2" />
            Download EPUB
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="h-[calc(100vh-4rem)] flex">
      {/* Chapter Navigation Sidebar */}
      <div className="w-80 border-r bg-muted/30">
        <div className="p-4 border-b">
          <h2 className="font-semibold truncate" title={bookTitle}>
            {bookTitle}
          </h2>
          <p className="text-sm text-muted-foreground">
            {toc.length} chapters
          </p>
        </div>
        <ScrollArea className="h-[calc(100%-80px)]">
          <div className="p-2">
            {toc.map((chapter, index) => (
              <button
                key={index}
                onClick={() => setCurrentChapter(index)}
                className={`w-full text-left p-3 rounded-md mb-1 transition-colors ${currentChapter === index
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
      <div className="flex-1 flex flex-col">
        {/* Chapter Header */}
        <div className="p-4 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-lg font-semibold">
                {toc[currentChapter]?.text || `Chapter ${currentChapter + 1}`}
              </h1>
              <p className="text-sm text-muted-foreground">
                Chapter {currentChapter + 1} of {toc.length}
              </p>
            </div>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={goToPreviousChapter}
                disabled={currentChapter === 0}
              >
                <ChevronLeft className="w-4 h-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={goToNextChapter}
                disabled={currentChapter === toc.length - 1}
              >
                <ChevronRight className="w-4 h-4" />
              </Button>
            </div>
          </div>
        </div>

        {/* Chapter Content */}
        <ScrollArea className="flex-1">
          <div className="p-6 max-w-4xl mx-auto">
            {contentLoading ? (
              <div className="text-center py-8">
                <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary mx-auto mb-4"></div>
                <p>Loading chapter...</p>
              </div>
            ) : (
              <div
                className="prose prose-gray dark:prose-invert max-w-none"
                dangerouslySetInnerHTML={{ __html: chapterContent }}
              />
            )}
          </div>
        </ScrollArea>
      </div>
    </div>
  )
}