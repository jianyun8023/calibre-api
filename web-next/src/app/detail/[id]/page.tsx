"use client"

import { useEffect, useState, useCallback } from "react"
import { useParams, useRouter } from "next/navigation"
import dynamic from "next/dynamic"
import { fetchBook, deleteBook } from "@/lib/api/books"
import type { Book } from "@/types/book"
import { Skeleton } from "@/components/ui/skeleton"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { 
  ArrowLeft, 
  Trash2, 
  Download, 
  BookOpen, 
  Edit, 
  RefreshCw, 
  List, 
  Share2, 
  User, 
  Building2, 
  Calendar, 
  Star, 
  FileText,
  Tag as TagIcon,
  Search
} from "lucide-react"
import { formatFileSize, copyToClipboard } from "@/lib/utils"
import { toast } from "sonner"
import Image from "next/image"
import type { DoubanBook } from "@/lib/api/metadata"

// 懒加载对话框组件（非关键路径）
const MetadataEditDialog = dynamic(() => import("@/components/metadata-edit-dialog").then(mod => ({ default: mod.MetadataEditDialog })), {
  loading: () => <div>Loading...</div>,
})

const MetadataSearchDialog = dynamic(() => import("@/components/metadata-search-dialog").then(mod => ({ default: mod.MetadataSearchDialog })), {
  loading: () => <div>Loading...</div>,
})

const MetadataCompareDialog = dynamic(() => import("@/components/metadata-compare-dialog").then(mod => ({ default: mod.MetadataCompareDialog })), {
  loading: () => <div>Loading...</div>,
})

const BookTocDialog = dynamic(() => import("@/components/book-toc-dialog").then(mod => ({ default: mod.BookTocDialog })), {
  loading: () => <div>Loading...</div>,
  ssr: false,
})

export default function BookDetailPage() {
  const params = useParams()
  const router = useRouter()
  const id = params.id as string
  
  const [book, setBook] = useState<Book | null>(null)
  const [loading, setLoading] = useState(true)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [searchDialogOpen, setSearchDialogOpen] = useState(false)
  const [compareDialogOpen, setCompareDialogOpen] = useState(false)
  const [tocDialogOpen, setTocDialogOpen] = useState(false)
  const [doubanMetadata, setDoubanMetadata] = useState<DoubanBook | null>(null)

  const loadBook = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchBook(id)
      setBook(data)
    } catch (e) {
      console.error(e)
      toast.error("Failed to load book details")
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    if (id) {
      loadBook()
    }
  }, [id, loadBook])

  const handleDelete = async () => {
    try {
      await deleteBook(parseInt(id))
      toast.success("Book deleted successfully")
      router.push("/books")
    } catch (e) {
      console.error(e)
      toast.error("Failed to delete book")
    }
  }

  const handleEdit = () => {
    setEditDialogOpen(true)
  }

  const handleSearch = () => {
    setSearchDialogOpen(true)
  }

  const handleToc = () => {
    setTocDialogOpen(true)
  }

  const handleMetadataSelect = (metadata: DoubanBook) => {
    setDoubanMetadata(metadata)
    setCompareDialogOpen(true)
  }

  const handleRefresh = async () => {
    toast.info("Refreshing metadata...")
    await loadBook()
    toast.success("Metadata refreshed")
  }

  const handleDownload = () => {
    if (!book?.file_path) {
      toast.error("No download link available")
      return
    }
    
    // 从文件路径提取文件格式
    const pathParts = book.file_path.split('/')
    const lastPart = pathParts[pathParts.length - 1]
    const fileExt = lastPart.split('.').pop() || 'epub'
    const fileName = `${book.title}.${fileExt.toLowerCase()}`
    
    // 创建隐藏的下载链接
    const link = document.createElement('a')
    link.href = book.file_path
    link.download = fileName
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    
    toast.success(`Downloading ${fileName}`)
  }

  if (loading) {
    return (
      <div className="max-w-5xl mx-auto space-y-8">
        <Button variant="ghost" disabled><ArrowLeft className="mr-2 h-4 w-4"/> Back</Button>
        <div className="grid md:grid-cols-3 gap-8">
          <Skeleton className="h-[500px] rounded-lg" />
          <div className="md:col-span-2 space-y-4">
            <Skeleton className="h-12 w-3/4" />
            <Skeleton className="h-6 w-1/2" />
            <div className="space-y-2 mt-8">
              {Array.from({length: 6}, (_, i) => `skeleton-${i}`).map((key) => (
                <Skeleton key={key} className="h-12 w-full" />
              ))}
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (!book) {
    return <div className="text-center py-20">Book not found</div>
  }

  return (
    <div className="max-w-5xl mx-auto space-y-6 pb-20">
      <Button variant="ghost" onClick={() => router.back()} className="mb-4">
        <ArrowLeft className="mr-2 h-4 w-4"/> Back
      </Button>

      <div className="grid md:grid-cols-12 gap-8">
        {/* Left Column: Cover */}
        <div className="md:col-span-4 flex flex-col items-center">
          <div className="relative w-full max-w-[300px] aspect-2/3 rounded-lg overflow-hidden shadow-2xl mb-6">
            <Image 
              src={book.cover} 
              alt={book.title}
              fill
              className="object-cover"
              sizes="(max-width: 768px) 100vw, 300px"
              unoptimized
            />
          </div>
        </div>

        {/* Right Column: Info */}
        <div className="md:col-span-8 space-y-6">
          <div className="space-y-2">
            <div className="flex justify-between items-start">
              <h1 className="text-3xl font-bold">{book.title}</h1>
              <div className="flex gap-2">
                <Button variant="outline" size="icon" title="Search Metadata" onClick={handleSearch}>
                  <Search className="h-4 w-4" />
                </Button>
                <Button variant="outline" size="icon" title="Edit Metadata" onClick={handleEdit}>
                  <Edit className="h-4 w-4" />
                </Button>
                <Button variant="outline" size="icon" title="Refresh Metadata" onClick={handleRefresh}>
                  <RefreshCw className="h-4 w-4" />
                </Button>
              </div>
            </div>
            
            {/* Metadata List */}
            <Card>
              <CardContent className="p-0 divide-y">
                <MetaItem icon={Share2} label="ID" value={
                  <Button variant="link" className="h-auto p-0" onClick={() => copyToClipboard(String(book.id))}>
                    {book.id} (Copy)
                  </Button>
                } />
                <MetaItem icon={User} label="Authors" value={
                  <div className="flex flex-wrap gap-2">
                    {book.authors?.map(author => (
                      <Badge key={author} variant="secondary" className="cursor-pointer hover:bg-secondary/80">
                        {author}
                      </Badge>
                    ))}
                  </div>
                } />
                <MetaItem icon={Building2} label="Publisher" value={book.publisher} />
                <MetaItem icon={Calendar} label="Published" value={book.pubdate ? new Date(book.pubdate).toLocaleDateString() : '-'} />
                <MetaItem icon={Star} label="Rating" value={`${book.rating / 2} / 5`} />
                <MetaItem icon={TagIcon} label="Tags" value={
                  <div className="flex flex-wrap gap-2">
                    {book.tags?.map(tag => (
                      <Badge key={tag} variant="outline" className="cursor-pointer">
                        {tag}
                      </Badge>
                    ))}
                  </div>
                } />
                <MetaItem icon={FileText} label="Size" value={formatFileSize(book.size || 0)} />
              </CardContent>
            </Card>

            {/* Action Buttons */}
            <div className="flex flex-wrap gap-4 pt-4">
              <Button className="flex-1" onClick={() => router.push(`/read/${book.id}`)}>
                <BookOpen className="mr-2 h-4 w-4" /> Read
              </Button>
              <Button variant="secondary" className="flex-1" onClick={handleToc}>
                <List className="mr-2 h-4 w-4" /> TOC
              </Button>
              {book.file_path && (
                <Button variant="secondary" className="flex-1" onClick={handleDownload}>
                  <Download className="mr-2 h-4 w-4" /> Download
                </Button>
              )}
              
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button variant="destructive" size="icon">
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                    <AlertDialogDescription>
                      This action cannot be undone. This will permanently delete the book from your library.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction onClick={handleDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                      Delete
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </div>
          </div>
        </div>
      </div>

      {/* Comments / Description */}
      {book.comments && (
        <Card>
          <CardHeader>
            <CardTitle>Description</CardTitle>
          </CardHeader>
          <CardContent>
            {/* eslint-disable-next-line react/no-danger */}
            <div 
              className="prose dark:prose-invert max-w-none"
              dangerouslySetInnerHTML={{ __html: book.comments }}
            />
          </CardContent>
        </Card>
      )}

      {/* Metadata Edit Dialog */}
      {book && (
        <>
          <MetadataEditDialog
            open={editDialogOpen}
            onOpenChange={setEditDialogOpen}
            book={book}
            onSuccess={loadBook}
          />
          <MetadataSearchDialog
            open={searchDialogOpen}
            onOpenChange={setSearchDialogOpen}
            book={book}
            onSelect={handleMetadataSelect}
          />
          <BookTocDialog
            open={tocDialogOpen}
            onOpenChange={setTocDialogOpen}
            bookId={id}
            bookTitle={book.title}
          />
          {doubanMetadata && (
            <MetadataCompareDialog
              open={compareDialogOpen}
              onOpenChange={setCompareDialogOpen}
              book={book}
              doubanMetadata={doubanMetadata}
              onSuccess={loadBook}
            />
          )}
        </>
      )}
    </div>
  )
}

function MetaItem({ icon: Icon, label, value }: { icon: React.ComponentType<{ className?: string }>, label: string, value: React.ReactNode }) {
  return (
    <div className="flex items-center py-3 px-4">
      <div className="flex items-center w-1/3 min-w-[120px] text-muted-foreground">
        <Icon className="mr-2 h-4 w-4" />
        <span className="text-sm font-medium">{label}</span>
      </div>
      <div className="flex-1 text-sm font-medium">{value || '-'}</div>
    </div>
  )
}

