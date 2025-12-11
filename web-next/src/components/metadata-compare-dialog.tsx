"use client"

import { useState, useEffect } from "react"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { Checkbox } from "@/components/ui/checkbox"
import { toast } from "sonner"
import { updateBook } from "@/lib/api/books"
import type { Book } from "@/types/book"
import type { DoubanBook } from "@/lib/api/metadata"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { Save } from "lucide-react"

interface MetadataCompareDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  book: Book
  doubanMetadata: DoubanBook
  onSuccess: () => void
}

type FieldSource = "new" | "old"

interface FieldSelection {
  title: FieldSource
  authors: FieldSource
  publisher: FieldSource
  pubdate: FieldSource
  isbn: FieldSource
  rating: FieldSource
  tags: FieldSource
  comments: FieldSource
}

export function MetadataCompareDialog({ open, onOpenChange, book, doubanMetadata, onSuccess }: MetadataCompareDialogProps) {
  const [selection, setSelection] = useState<FieldSelection>({
    title: "new",
    authors: "new",
    publisher: "new",
    pubdate: "new",
    isbn: "new",
    rating: "new",
    tags: "new",
    comments: "new",
  })
  const [selectedAuthors, setSelectedAuthors] = useState<string[]>([])
  const [selectedTags, setSelectedTags] = useState<string[]>([])
  const [saving, setSaving] = useState(false)
  const [useSubTitle, setUseSubTitle] = useState(true)

  // Initialize selections when dialog opens
  useEffect(() => {
    if (open && doubanMetadata) {
      setSelectedAuthors(doubanMetadata.author || [])
      setSelectedTags(doubanMetadata.tags?.map(t => t.name) || [])
    }
  }, [open, doubanMetadata])

  // Update selectedAuthors when source changes
  useEffect(() => {
    if (selection.authors === "new") {
      setSelectedAuthors(doubanMetadata.author || [])
    } else {
      setSelectedAuthors(book.authors || [])
    }
  }, [selection.authors, book.authors, doubanMetadata.author])

  // Update selectedTags when source changes
  useEffect(() => {
    if (selection.tags === "new") {
      setSelectedTags(doubanMetadata.tags?.map(t => t.name) || [])
    } else {
      setSelectedTags(book.tags || [])
    }
  }, [selection.tags, book.tags, doubanMetadata.tags])

  const getTitle = () => {
    if (selection.title === "new") {
      return useSubTitle && doubanMetadata.sub_title
        ? `${doubanMetadata.title}：${doubanMetadata.sub_title}`
        : doubanMetadata.title
    }
    return book.title
  }

  const parseDateString = (dateString?: string) => {
    if (!dateString) return new Date(0).toISOString()
    const parts = dateString.split("-")
    const year = parseInt(parts[0], 10)
    const month = parts.length > 1 ? parseInt(parts[1], 10) - 1 : 0
    const day = parts.length > 2 ? parseInt(parts[2], 10) : 1
    return new Date(year, month, day).toISOString()
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const updateData: Partial<Book> = {
        title: getTitle(),
        authors: selectedAuthors,
        publisher: selection.publisher === "new" ? doubanMetadata.publisher : book.publisher,
        pubdate: selection.pubdate === "new" ? parseDateString(doubanMetadata.pubdate) : book.pubdate,
        isbn: selection.isbn === "new" ? doubanMetadata.isbn13 : book.isbn,
        rating: selection.rating === "new" ? (doubanMetadata.rating?.average || 0) * 2 : book.rating, // Convert to 0-10 scale
        tags: selectedTags,
        comments: selection.comments === "new" ? doubanMetadata.summary?.replace(/class=".*?"/g, '') : book.comments,
      }

      await updateBook(book.id, updateData)
      toast.success("元数据更新成功")
      onSuccess()
      onOpenChange(false)
    } catch (error) {
      console.error(error)
      toast.error("元数据更新失败")
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>元数据对比</DialogTitle>
          <DialogDescription>
            选择要使用的元数据。每个字段都可以选择使用新数据（豆瓣）或保留旧数据。
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Title */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label className="font-semibold">标题</Label>
              <RadioGroup
                value={selection.title}
                onValueChange={(value) => setSelection({ ...selection, title: value as FieldSource })}
                className="flex gap-2"
              >
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="new" id="title-new" />
                  <Label htmlFor="title-new" className="cursor-pointer">新</Label>
                </div>
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="old" id="title-old" />
                  <Label htmlFor="title-old" className="cursor-pointer">旧</Label>
                </div>
              </RadioGroup>
            </div>
            <div className="p-3 bg-accent rounded text-sm">
              {getTitle()}
            </div>
            {selection.title === "new" && doubanMetadata.sub_title && (
              <div className="flex items-center gap-2">
                <Checkbox
                  checked={useSubTitle}
                  onCheckedChange={(checked) => setUseSubTitle(checked as boolean)}
                  id="use-subtitle"
                />
                <Label htmlFor="use-subtitle" className="text-sm cursor-pointer">
                  包含副标题
                </Label>
              </div>
            )}
          </div>

          <Separator />

          {/* Authors */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label className="font-semibold">作者</Label>
              <RadioGroup
                value={selection.authors}
                onValueChange={(value) => setSelection({ ...selection, authors: value as FieldSource })}
                className="flex gap-2"
              >
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="new" id="authors-new" />
                  <Label htmlFor="authors-new" className="cursor-pointer">新</Label>
                </div>
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="old" id="authors-old" />
                  <Label htmlFor="authors-old" className="cursor-pointer">旧</Label>
                </div>
              </RadioGroup>
            </div>
            <div className="flex flex-wrap gap-2">
              {(selection.authors === "new" ? doubanMetadata.author : book.authors || []).map((author) => (
                <div key={author} className="flex items-center gap-2">
                  <Checkbox
                    checked={selectedAuthors.includes(author)}
                    onCheckedChange={(checked) => {
                      if (checked) {
                        setSelectedAuthors([...selectedAuthors, author])
                      } else {
                        setSelectedAuthors(selectedAuthors.filter(a => a !== author))
                      }
                    }}
                    id={`author-${author}`}
                  />
                  <Label htmlFor={`author-${author}`} className="cursor-pointer">{author}</Label>
                </div>
              ))}
            </div>
          </div>

          <Separator />

          {/* Publisher */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label className="font-semibold">出版社</Label>
              <RadioGroup
                value={selection.publisher}
                onValueChange={(value) => setSelection({ ...selection, publisher: value as FieldSource })}
                className="flex gap-2"
              >
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="new" id="publisher-new" />
                  <Label htmlFor="publisher-new" className="cursor-pointer">新</Label>
                </div>
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="old" id="publisher-old" />
                  <Label htmlFor="publisher-old" className="cursor-pointer">旧</Label>
                </div>
              </RadioGroup>
            </div>
            <div className="p-3 bg-accent rounded text-sm">
              {selection.publisher === "new" ? doubanMetadata.publisher : book.publisher || "-"}
            </div>
          </div>

          <Separator />

          {/* Pubdate & ISBN */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label className="font-semibold">出版日期</Label>
                <RadioGroup
                  value={selection.pubdate}
                  onValueChange={(value) => setSelection({ ...selection, pubdate: value as FieldSource })}
                  className="flex gap-2"
                >
                  <div className="flex items-center gap-2">
                    <RadioGroupItem value="new" id="pubdate-new" />
                    <Label htmlFor="pubdate-new" className="cursor-pointer text-xs">新</Label>
                  </div>
                  <div className="flex items-center gap-2">
                    <RadioGroupItem value="old" id="pubdate-old" />
                    <Label htmlFor="pubdate-old" className="cursor-pointer text-xs">旧</Label>
                  </div>
                </RadioGroup>
              </div>
              <div className="p-3 bg-accent rounded text-sm">
                {selection.pubdate === "new" ? doubanMetadata.pubdate : (book.pubdate ? new Date(book.pubdate).toLocaleDateString() : "-")}
              </div>
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label className="font-semibold">ISBN</Label>
                <RadioGroup
                  value={selection.isbn}
                  onValueChange={(value) => setSelection({ ...selection, isbn: value as FieldSource })}
                  className="flex gap-2"
                >
                  <div className="flex items-center gap-2">
                    <RadioGroupItem value="new" id="isbn-new" />
                    <Label htmlFor="isbn-new" className="cursor-pointer text-xs">新</Label>
                  </div>
                  <div className="flex items-center gap-2">
                    <RadioGroupItem value="old" id="isbn-old" />
                    <Label htmlFor="isbn-old" className="cursor-pointer text-xs">旧</Label>
                  </div>
                </RadioGroup>
              </div>
              <div className="p-3 bg-accent rounded text-sm">
                {selection.isbn === "new" ? doubanMetadata.isbn13 : book.isbn || "-"}
              </div>
            </div>
          </div>

          <Separator />

          {/* Rating */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label className="font-semibold">评分</Label>
              <RadioGroup
                value={selection.rating}
                onValueChange={(value) => setSelection({ ...selection, rating: value as FieldSource })}
                className="flex gap-2"
              >
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="new" id="rating-new" />
                  <Label htmlFor="rating-new" className="cursor-pointer">新</Label>
                </div>
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="old" id="rating-old" />
                  <Label htmlFor="rating-old" className="cursor-pointer">旧</Label>
                </div>
              </RadioGroup>
            </div>
            <div className="p-3 bg-accent rounded text-sm">
              {selection.rating === "new"
                ? `${doubanMetadata.rating?.average || 0} / 5 (${doubanMetadata.rating?.numRaters || 0} 人评价)`
                : `${book.rating ? book.rating / 2 : 0} / 5`}
            </div>
          </div>

          <Separator />

          {/* Tags */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label className="font-semibold">标签</Label>
              <RadioGroup
                value={selection.tags}
                onValueChange={(value) => setSelection({ ...selection, tags: value as FieldSource })}
                className="flex gap-2"
              >
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="new" id="tags-new" />
                  <Label htmlFor="tags-new" className="cursor-pointer">新</Label>
                </div>
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="old" id="tags-old" />
                  <Label htmlFor="tags-old" className="cursor-pointer">旧</Label>
                </div>
              </RadioGroup>
            </div>
            <div className="flex flex-wrap gap-2">
              {(selection.tags === "new" ? doubanMetadata.tags?.map(t => t.name) : book.tags || []).map((tag) => (
                <div key={tag} className="flex items-center gap-2">
                  <Checkbox
                    checked={selectedTags.includes(tag)}
                    onCheckedChange={(checked) => {
                      if (checked) {
                        setSelectedTags([...selectedTags, tag])
                      } else {
                        setSelectedTags(selectedTags.filter(t => t !== tag))
                      }
                    }}
                    id={`tag-${tag}`}
                  />
                  <Label htmlFor={`tag-${tag}`} className="cursor-pointer">
                    <Badge variant="outline">{tag}</Badge>
                  </Label>
                </div>
              ))}
            </div>
          </div>

          <Separator />

          {/* Comments */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label className="font-semibold">简介</Label>
              <RadioGroup
                value={selection.comments}
                onValueChange={(value) => setSelection({ ...selection, comments: value as FieldSource })}
                className="flex gap-2"
              >
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="new" id="comments-new" />
                  <Label htmlFor="comments-new" className="cursor-pointer">新</Label>
                </div>
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="old" id="comments-old" />
                  <Label htmlFor="comments-old" className="cursor-pointer">旧</Label>
                </div>
              </RadioGroup>
            </div>
            <div className="p-3 bg-accent rounded text-sm max-h-40 overflow-y-auto">
              {selection.comments === "new"
                ? doubanMetadata.summary?.replace(/class=".*?"/g, '') || "-"
                : book.comments || "-"}
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={saving}>
            取消
          </Button>
          <Button onClick={handleSave} disabled={saving}>
            {saving ? (
              "保存中..."
            ) : (
              <>
                <Save className="w-4 h-4 mr-2" />
                保存更新
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

