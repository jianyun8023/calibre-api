"use client"

import { useState, useEffect } from "react"
import { useParams, useSearchParams } from "next/navigation"
import dynamic from "next/dynamic"
import { useTheme } from "next-themes"
import { Skeleton } from "@/components/ui/skeleton"
import { toast } from "sonner"
import { fetchBook } from "@/lib/api/books"

// 懒加载 ReactReader (大约 200KB)
const ReactReader = dynamic(
  () => import("react-reader").then((mod) => mod.ReactReader),
  {
    loading: () => (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <Skeleton className="h-8 w-48 mx-auto mb-4" />
          <p className="text-muted-foreground">Loading reader...</p>
        </div>
      </div>
    ),
    ssr: false,
  }
)

// 懒加载章节查看器
const EpubChapterViewer = dynamic(
  () => import("@/components/epub-chapter-viewer").then((mod) => mod.EpubChapterViewer),
  {
    loading: () => <Skeleton className="h-screen" />,
    ssr: false,
  }
)

export default function ReadBookPage() {
  const params = useParams()
  const searchParams = useSearchParams()
  const id = params.id as string
  const { theme } = useTheme()
  
  const [location, setLocation] = useState<string | number>(0)
  const [url, setUrl] = useState<string>("")
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [useAlternativeViewer, setUseAlternativeViewer] = useState(false)
  const [bookTitle, setBookTitle] = useState<string>("")

  useEffect(() => {
    const initializeReader = async () => {
      try {
        // Load book info
        const book = await fetchBook(id)
        setBookTitle(book.title)
        
        // 恢复上次阅读进度 (简单实现，可以使用 localStorage)
        const savedLocation = localStorage.getItem(`book-progress-${id}`)
        if (savedLocation) {
          setLocation(savedLocation)
        }
        
        // 检查 URL 中是否有特定的章节锚点
        const hash = window.location.hash
        if (hash) {
          // 如果有锚点，尝试导航到该位置
          console.log('Found hash in URL:', hash)
        }
        
        setUrl(`/api/download/book/${id}.epub`)
      } catch (error) {
        console.error('Failed to load book info:', error)
        setBookTitle('Unknown Book')
      } finally {
        setLoading(false)
      }
    }
    
    initializeReader()
  }, [id])

  const handleLocationChange = (epubcifi: string | number) => {
    console.log('Location changed:', epubcifi)
    setLocation(epubcifi)
    localStorage.setItem(`book-progress-${id}`, String(epubcifi))
  }

  const handleReaderError = (error: any) => {
    console.error('React Reader Error:', error)
    setError(error.message || 'Failed to load EPUB file')
    toast.error('Failed to load book content. This may be due to EPUB format compatibility issues.')
  }

  const testEpubAccess = async () => {
    try {
      const response = await fetch(`/api/download/book/${id}.epub`)
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }
      const blob = await response.blob()
      console.log('EPUB file size:', blob.size, 'bytes')
      console.log('EPUB content type:', blob.type)
      return true
    } catch (error) {
      console.error('EPUB access test failed:', error)
      return false
    }
  }

  if (loading) return <Skeleton className="h-screen w-full" />

  if (error) {
    return (
      <div className="h-[calc(100vh-4rem)] flex items-center justify-center">
        <div className="text-center max-w-md">
          <h2 className="text-xl font-semibold mb-2">Reader Error</h2>
          <p className="text-muted-foreground mb-4">{error}</p>
          <div className="space-y-2">
            <button 
              onClick={() => window.location.reload()} 
              className="px-4 py-2 bg-primary text-primary-foreground rounded-md mr-2"
            >
              Retry
            </button>
            <button 
              onClick={() => setUseAlternativeViewer(true)} 
              className="px-4 py-2 bg-secondary text-secondary-foreground rounded-md mr-2"
            >
              Use Chapter Viewer
            </button>
            <button 
              onClick={() => window.open(`/api/download/book/${id}.epub`, '_blank')} 
              className="px-4 py-2 bg-outline text-foreground rounded-md"
            >
              Download EPUB
            </button>
          </div>
          <p className="text-xs text-muted-foreground mt-4">
            Try the chapter viewer for a simpler reading experience, or download the EPUB file for external readers.
          </p>
        </div>
      </div>
    )
  }

  // Use alternative viewer if requested or if react-reader fails
  if (useAlternativeViewer) {
    return (
      <div className="h-[calc(100vh-4rem)] -m-4 md:-m-6 lg:-m-8">
        <EpubChapterViewer bookId={id} bookTitle={bookTitle} />
      </div>
    )
  }

  return (
    <div className="h-[calc(100vh-4rem)] -m-4 md:-m-6 lg:-m-8 relative">
      <ReactReader
        url={url}
        location={location}
        locationChanged={handleLocationChange}
        epubInitOptions={{
          openAs: 'epub',
          requestMethod: 'GET',
          requestCredentials: 'same-origin',
        }}
        epubOptions={{
          flow: 'paginated',
          manager: 'default',
          spread: 'auto',
        }}
        getRendition={(rendition: any) => {
          // 添加渲染器配置
          rendition.themes.default({
            body: {
              'font-family': 'Georgia, serif !important',
              'line-height': '1.6 !important',
              'color': theme === 'dark' ? '#e5e5e5 !important' : '#333 !important',
              'background': theme === 'dark' ? '#1a1a1a !important' : '#fff !important',
            }
          })
          
          // 监听渲染错误
          rendition.on('rendered', () => {
            console.log('EPUB rendered successfully')
          })
          
          rendition.on('error', (error: any) => {
            console.error('Rendition error:', error)
            handleReaderError(error)
          })
        }}
      />
    </div>
  )
}

// React Reader default theme structure for reference/override
const lightTheme = {
  body: {
    background: '#fff',
  },
  // ... other theme props
}

