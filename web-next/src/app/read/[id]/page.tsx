"use client"

import { useState, useEffect } from "react"
import { useParams } from "next/navigation"
import { ReactReader } from "react-reader"
import { useTheme } from "next-themes"
import { Skeleton } from "@/components/ui/skeleton"

export default function ReadBookPage() {
  const params = useParams()
  const id = params.id as string
  const { theme } = useTheme()
  
  const [location, setLocation] = useState<string | number>(0)
  const [url, setUrl] = useState<string>("")

  useEffect(() => {
    // 恢复上次阅读进度 (简单实现，可以使用 localStorage)
    const savedLocation = localStorage.getItem(`book-progress-${id}`)
    if (savedLocation) {
      setLocation(savedLocation)
    }
    setUrl(`/api/download/book/${id}.epub`)
  }, [id])

  const handleLocationChange = (epubcifi: string | number) => {
    setLocation(epubcifi)
    localStorage.setItem(`book-progress-${id}`, String(epubcifi))
  }

  if (!url) return <Skeleton className="h-screen w-full" />

  return (
    <div className="h-[calc(100vh-4rem)] -m-4 md:-m-6 lg:-m-8 relative">
      <ReactReader
        url={url}
        location={location}
        locationChanged={handleLocationChange}
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

