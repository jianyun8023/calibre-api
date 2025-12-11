"use client"

import { useState, useEffect } from "react"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { toast } from "sonner"
import { updateBook } from "@/lib/api/books"
import type { Book } from "@/types/book"
import { Save, X } from "lucide-react"

interface MetadataEditDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  book: Book
  onSuccess: () => void
}

export function MetadataEditDialog({ open, onOpenChange, book, onSuccess }: MetadataEditDialogProps) {
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
  const [saving, setSaving] = useState(false)

  // Initialize form data when dialog opens
  useEffect(() => {
    if (open && book) {
      setFormData({
        title: book.title || "",
        authors: book.authors?.join(", ") || "",
        publisher: book.publisher || "",
        pubdate: book.pubdate ? new Date(book.pubdate).toISOString().split('T')[0] : "",
        isbn: book.isbn || "",
        tags: book.tags?.join(", ") || "",
        rating: book.rating ? String(book.rating / 2) : "",
        comments: book.comments || "",
      })
    }
  }, [open, book])

  const handleSave = async () => {
    setSaving(true)
    try {
      const updateData: Partial<Book> = {}
      
      // Only include changed fields
      if (formData.title !== book.title) {
        updateData.title = formData.title
      }
      
      const authorsArray = formData.authors.split(",").map(a => a.trim()).filter(Boolean)
      if (JSON.stringify(authorsArray) !== JSON.stringify(book.authors)) {
        updateData.authors = authorsArray
      }
      
      if (formData.publisher !== book.publisher) {
        updateData.publisher = formData.publisher
      }
      
      if (formData.pubdate) {
        const newDate = new Date(formData.pubdate).toISOString()
        const oldDate = book.pubdate ? new Date(book.pubdate).toISOString() : ""
        if (newDate !== oldDate) {
          updateData.pubdate = newDate
        }
      }
      
      if (formData.isbn !== book.isbn) {
        updateData.isbn = formData.isbn
      }
      
      const tagsArray = formData.tags.split(",").map(t => t.trim()).filter(Boolean)
      if (JSON.stringify(tagsArray) !== JSON.stringify(book.tags)) {
        updateData.tags = tagsArray
      }
      
      if (formData.rating) {
        const newRating = parseFloat(formData.rating) * 2 // Convert to 0-10 scale
        if (newRating !== book.rating) {
          updateData.rating = newRating
        }
      }
      
      if (formData.comments !== book.comments) {
        updateData.comments = formData.comments
      }

      if (Object.keys(updateData).length === 0) {
        toast.info("No changes to save")
        onOpenChange(false)
        return
      }

      await updateBook(book.id, updateData)
      toast.success("Metadata updated successfully")
      onSuccess()
      onOpenChange(false)
    } catch (error) {
      console.error(error)
      toast.error("Failed to update metadata")
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Metadata</DialogTitle>
          <DialogDescription>
            Update the book&apos;s metadata. Changes will be saved to the library.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="title">Title</Label>
            <Input
              id="title"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              placeholder="Book title"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="authors">Authors</Label>
            <Input
              id="authors"
              value={formData.authors}
              onChange={(e) => setFormData({ ...formData, authors: e.target.value })}
              placeholder="Author1, Author2"
            />
            <p className="text-xs text-muted-foreground">Separate multiple authors with commas</p>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="publisher">Publisher</Label>
              <Input
                id="publisher"
                value={formData.publisher}
                onChange={(e) => setFormData({ ...formData, publisher: e.target.value })}
                placeholder="Publisher name"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="pubdate">Published Date</Label>
              <Input
                id="pubdate"
                type="date"
                value={formData.pubdate}
                onChange={(e) => setFormData({ ...formData, pubdate: e.target.value })}
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="isbn">ISBN</Label>
              <Input
                id="isbn"
                value={formData.isbn}
                onChange={(e) => setFormData({ ...formData, isbn: e.target.value })}
                placeholder="ISBN"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="rating">Rating</Label>
              <Input
                id="rating"
                type="number"
                min="0"
                max="5"
                step="0.5"
                value={formData.rating}
                onChange={(e) => setFormData({ ...formData, rating: e.target.value })}
                placeholder="0-5"
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="tags">Tags</Label>
            <Input
              id="tags"
              value={formData.tags}
              onChange={(e) => setFormData({ ...formData, tags: e.target.value })}
              placeholder="tag1, tag2, tag3"
            />
            <p className="text-xs text-muted-foreground">Separate multiple tags with commas</p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="comments">Description / Comments</Label>
            <Textarea
              id="comments"
              value={formData.comments}
              onChange={(e) => setFormData({ ...formData, comments: e.target.value })}
              placeholder="Book description or comments"
              rows={6}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={saving}>
            <X className="w-4 h-4 mr-2" />
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={saving}>
            <Save className="w-4 h-4 mr-2" />
            {saving ? "Saving..." : "Save Changes"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

