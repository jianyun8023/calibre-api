"use client"

import { useState, useEffect } from "react"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { toast } from "sonner"
import { Check, X, FileSearch, Loader2 } from "lucide-react"
import { extractBookMetadata } from "@/lib/api/books"

interface QuickEditDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  bookId: number
  initialData: Record<string, unknown>
  onApply: (data: Record<string, unknown>) => void
}

export function QuickEditDialog({ open, onOpenChange, bookId, initialData, onApply }: QuickEditDialogProps) {
  const [formData, setFormData] = useState({
    title: "",
    authors: "",
    publisher: "",
    pubdate: "",
    isbn: "",
    tags: "",
    rating: "",
    comments: "",
  })
  const [extracting, setExtracting] = useState(false)

  // 初始化表单数据
  useEffect(() => {
    if (open && initialData) {
      setFormData({
        title: String(initialData.title || ""),
        authors: Array.isArray(initialData.authors) ? initialData.authors.join(", ") : "",
        publisher: String(initialData.publisher || ""),
        pubdate: String(initialData.pubdate || ""),
        isbn: String(initialData.isbn || ""),
        tags: Array.isArray(initialData.tags) ? initialData.tags.join(", ") : "",
        rating: initialData.rating ? String(initialData.rating) : "",
        comments: String(initialData.comments || ""),
      })
    }
  }, [open, initialData])

  // 从书籍文件中提取元数据
  const handleExtract = async () => {
    setExtracting(true)
    try {
      const result = await extractBookMetadata(String(bookId))
      if (result.success && result.data) {
        // 将提取的数据填充到表单中（只填充空字段）
        const extractedDate = result.data.publish_date || ''
        
        setFormData(prev => ({
          ...prev,
          isbn: prev.isbn || result.data?.isbn || '',
          authors: prev.authors || result.data?.author || '',
          publisher: prev.publisher || result.data?.publisher || '',
          pubdate: prev.pubdate || extractedDate,
        }))
        
        toast.success(`从文件提取成功！${result.data.isbn ? `ISBN: ${result.data.isbn}` : ''}`)
      } else {
        toast.warning(result.message || '未能从书籍文件中提取到元数据')
      }
    } catch (error) {
      console.error(error)
      toast.error('提取元数据失败，请确保书籍文件为 EPUB 格式')
    } finally {
      setExtracting(false)
    }
  }

  const handleApply = () => {
    const result: Record<string, unknown> = {}
    
    // 只添加有值的字段
    if (formData.title.trim()) result.title = formData.title.trim()
    if (formData.authors.trim()) {
      result.authors = formData.authors.split(",").map(a => a.trim()).filter(Boolean)
    }
    if (formData.publisher.trim()) result.publisher = formData.publisher.trim()
    if (formData.pubdate.trim()) result.pubdate = formData.pubdate.trim()
    if (formData.isbn.trim()) result.isbn = formData.isbn.trim()
    if (formData.tags.trim()) {
      result.tags = formData.tags.split(",").map(t => t.trim()).filter(Boolean)
    }
    if (formData.rating.trim()) {
      const rating = parseFloat(formData.rating)
      if (!Number.isNaN(rating)) result.rating = rating
    }
    if (formData.comments.trim()) result.comments = formData.comments.trim()

    onApply(result)
    onOpenChange(false)
    toast.success("编辑内容已填充到工作台")
  }

  const handleCancel = () => {
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>快速编辑元数据</DialogTitle>
          <DialogDescription>
            编辑元数据字段，点击应用后会填充到工作台的编辑区域
          </DialogDescription>
        </DialogHeader>

        {/* 从文件提取按钮 */}
        <div className="flex justify-end -mt-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleExtract}
            disabled={extracting}
          >
            {extracting ? (
              <>
                <Loader2 className="h-3 w-3 mr-1 animate-spin" />
                提取中...
              </>
            ) : (
              <>
                <FileSearch className="h-3 w-3 mr-1" />
                从文件读取元数据
              </>
            )}
          </Button>
        </div>

        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="title">标题</Label>
            <Input
              id="title"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              placeholder="书籍标题"
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="authors">作者</Label>
            <Input
              id="authors"
              value={formData.authors}
              onChange={(e) => setFormData({ ...formData, authors: e.target.value })}
              placeholder="多个作者用逗号分隔"
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="publisher">出版社</Label>
            <Input
              id="publisher"
              value={formData.publisher}
              onChange={(e) => setFormData({ ...formData, publisher: e.target.value })}
              placeholder="出版社名称"
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="pubdate">出版日期</Label>
            <Input
              id="pubdate"
              value={formData.pubdate}
              onChange={(e) => setFormData({ ...formData, pubdate: e.target.value })}
              placeholder="YYYY-MM-DD"
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="isbn">ISBN</Label>
            <Input
              id="isbn"
              value={formData.isbn}
              onChange={(e) => setFormData({ ...formData, isbn: e.target.value })}
              placeholder="ISBN-10 或 ISBN-13"
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="tags">标签</Label>
            <Input
              id="tags"
              value={formData.tags}
              onChange={(e) => setFormData({ ...formData, tags: e.target.value })}
              placeholder="多个标签用逗号分隔"
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="rating">评分</Label>
            <Input
              id="rating"
              type="number"
              step="0.1"
              min="0"
              max="10"
              value={formData.rating}
              onChange={(e) => setFormData({ ...formData, rating: e.target.value })}
              placeholder="0-10"
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="comments">简介</Label>
            <Textarea
              id="comments"
              value={formData.comments}
              onChange={(e) => setFormData({ ...formData, comments: e.target.value })}
              placeholder="书籍简介或备注"
              rows={6}
              className="resize-y"
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleCancel}>
            <X className="h-4 w-4 mr-1" />
            取消
          </Button>
          <Button onClick={handleApply}>
            <Check className="h-4 w-4 mr-1" />
            应用到工作台
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
