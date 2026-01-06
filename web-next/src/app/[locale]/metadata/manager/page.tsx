"use client"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Checkbox } from "@/components/ui/checkbox"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { bookService } from "@/lib/services"
import type { Book } from "@/types/book"
import { useState } from "react"
import { Search, Save, RefreshCw, FileEdit, Info, AlertCircle } from "lucide-react"
import { toast } from "sonner"

interface BatchUpdate {
  title?: string
  authors?: string[]
  publisher?: string
  isbn?: string
  tags?: string[]
  rating?: number
  comments?: string
}

export default function MetadataManagerPage() {
  const [searchQuery, setSearchQuery] = useState("")
  const [searchResults, setSearchResults] = useState<Book[]>([])
  const [selectedBooks, setSelectedBooks] = useState<Set<number>>(new Set())
  const [searching, setSearching] = useState(false)
  const [batchUpdate, setBatchUpdate] = useState<BatchUpdate>({})
  const [updating, setUpdating] = useState(false)

  const searchBooks = async () => {
    if (!searchQuery.trim()) {
      toast.error("Please enter a search query")
      return
    }

    setSearching(true)
    try {
      const response = await bookService.searchBooks({
        q: searchQuery,
        mode: 'text',
        pagination: { page: 1, page_size: 50 }
      })
      setSearchResults(response.data || [])
      if (response.data && response.data.length > 0) {
        toast.success(`Found ${response.data.length} books`)
      } else {
        toast.info("No books found")
      }
    } catch (error) {
      console.error("Search failed:", error)
      toast.error("Search failed", {
        description: error instanceof Error ? error.message : "Unknown error"
      })
    } finally {
      setSearching(false)
    }
  }

  const toggleBookSelection = (bookId: number) => {
    const newSelection = new Set(selectedBooks)
    if (newSelection.has(bookId)) {
      newSelection.delete(bookId)
    } else {
      newSelection.add(bookId)
    }
    setSelectedBooks(newSelection)
  }

  const selectAll = () => {
    setSelectedBooks(new Set(searchResults.map(b => b.id)))
  }

  const deselectAll = () => {
    setSelectedBooks(new Set())
  }

  const applyBatchUpdate = async () => {
    if (selectedBooks.size === 0) {
      toast.error("Please select at least one book")
      return
    }

    if (Object.keys(batchUpdate).length === 0) {
      toast.error("Please specify at least one field to update")
      return
    }

    setUpdating(true)
    let successCount = 0
    let errorCount = 0

    for (const bookId of selectedBooks) {
      try {
        await bookService.updateBook(String(bookId), batchUpdate)
        successCount++
      } catch (error) {
        errorCount++
        console.error(`Failed to update book ${bookId}:`, error)
      }
    }

    setUpdating(false)

    if (errorCount === 0) {
      toast.success(`Successfully updated ${successCount} books`)
      // Reset state
      setBatchUpdate({})
      setSelectedBooks(new Set())
      setSearchResults([])
      setSearchQuery("")
    } else {
      toast.warning(`Updated ${successCount} books, ${errorCount} failed`)
    }
  }

  return (
    <div className="container max-w-6xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Metadata Manager</h1>
        <p className="text-muted-foreground">
          Batch update metadata for multiple books at once
        </p>
      </div>

      <Tabs defaultValue="search" className="space-y-6">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="search">
            <Search className="w-4 h-4 mr-2" />
            Search Books
          </TabsTrigger>
          <TabsTrigger value="edit" disabled={selectedBooks.size === 0}>
            <FileEdit className="w-4 h-4 mr-2" />
            Edit Metadata ({selectedBooks.size})
          </TabsTrigger>
          <TabsTrigger value="preview" disabled={selectedBooks.size === 0}>
            <Info className="w-4 h-4 mr-2" />
            Preview Changes
          </TabsTrigger>
        </TabsList>

        {/* Search Tab */}
        <TabsContent value="search" className="space-y-4">
          <Card className="glass">
            <CardHeader>
              <CardTitle>Find Books to Update</CardTitle>
              <CardDescription>
                Search for books you want to batch update
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex gap-2">
                <Input
                  type="text"
                  placeholder="Search by title, author, ISBN..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && searchBooks()}
                />
                <Button onClick={searchBooks} disabled={searching}>
                  <Search className="w-4 h-4 mr-2" />
                  {searching ? "Searching..." : "Search"}
                </Button>
              </div>

              {searchResults.length > 0 && (
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <p className="text-sm text-muted-foreground">
                      Found {searchResults.length} books
                    </p>
                    <div className="flex gap-2">
                      <Button size="sm" variant="outline" onClick={selectAll}>
                        Select All
                      </Button>
                      <Button size="sm" variant="outline" onClick={deselectAll}>
                        Deselect All
                      </Button>
                    </div>
                  </div>

                  <div className="space-y-2 max-h-96 overflow-y-auto">
                    {searchResults.map((book) => (
                      <div
                        key={book.id}
                        className="flex items-center gap-3 p-3 rounded-lg border hover:bg-accent/50 transition-colors"
                      >
                        <Checkbox
                          checked={selectedBooks.has(book.id)}
                          onCheckedChange={() => toggleBookSelection(book.id)}
                        />
                        <div className="flex-1 min-w-0">
                          <div className="font-medium truncate">{book.title}</div>
                          <div className="text-sm text-muted-foreground truncate">
                            {book.authors?.join(", ") || "Unknown Author"}
                          </div>
                        </div>
                        <Badge variant="secondary">{book.publisher || "No Publisher"}</Badge>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Edit Tab */}
        <TabsContent value="edit" className="space-y-4">
          <Card className="glass">
            <CardHeader>
              <CardTitle>Batch Edit Metadata</CardTitle>
              <CardDescription>
                Fields you leave empty will not be updated
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="title">Title (Leave empty to keep original)</Label>
                <Input
                  id="title"
                  type="text"
                  placeholder="New title..."
                  value={batchUpdate.title || ""}
                  onChange={(e) => setBatchUpdate({ ...batchUpdate, title: e.target.value })}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="authors">Authors (Comma-separated)</Label>
                <Input
                  id="authors"
                  type="text"
                  placeholder="Author1, Author2..."
                  value={batchUpdate.authors?.join(", ") || ""}
                  onChange={(e) =>
                    setBatchUpdate({
                      ...batchUpdate,
                      authors: e.target.value.split(",").map((a) => a.trim()).filter(Boolean),
                    })
                  }
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="publisher">Publisher</Label>
                <Input
                  id="publisher"
                  type="text"
                  placeholder="Publisher name..."
                  value={batchUpdate.publisher || ""}
                  onChange={(e) => setBatchUpdate({ ...batchUpdate, publisher: e.target.value })}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="isbn">ISBN</Label>
                <Input
                  id="isbn"
                  type="text"
                  placeholder="ISBN..."
                  value={batchUpdate.isbn || ""}
                  onChange={(e) => setBatchUpdate({ ...batchUpdate, isbn: e.target.value })}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="tags">Tags (Comma-separated)</Label>
                <Input
                  id="tags"
                  type="text"
                  placeholder="tag1, tag2..."
                  value={batchUpdate.tags?.join(", ") || ""}
                  onChange={(e) =>
                    setBatchUpdate({
                      ...batchUpdate,
                      tags: e.target.value.split(",").map((t) => t.trim()).filter(Boolean),
                    })
                  }
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="rating">Rating (0-5)</Label>
                <Input
                  id="rating"
                  type="number"
                  min="0"
                  max="5"
                  step="0.5"
                  placeholder="Rating..."
                  value={batchUpdate.rating || ""}
                  onChange={(e) =>
                    setBatchUpdate({ ...batchUpdate, rating: parseFloat(e.target.value) })
                  }
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="comments">Comments/Description</Label>
                <Textarea
                  id="comments"
                  placeholder="Book description..."
                  value={batchUpdate.comments || ""}
                  onChange={(e) => setBatchUpdate({ ...batchUpdate, comments: e.target.value })}
                  rows={4}
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Preview Tab */}
        <TabsContent value="preview" className="space-y-4">
          <Card className="glass">
            <CardHeader>
              <CardTitle>Preview Changes</CardTitle>
              <CardDescription>
                Review the changes before applying them
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="bg-blue-500/10 text-blue-500 dark:text-blue-400 px-4 py-3 rounded-lg flex items-start gap-3">
                <AlertCircle className="w-5 h-5 mt-0.5 shrink-0" />
                <div>
                  <p className="font-medium">You are about to update {selectedBooks.size} books</p>
                  <p className="text-sm mt-1">
                    Only non-empty fields will be updated. This action cannot be undone.
                  </p>
                </div>
              </div>

              <div className="space-y-2">
                <h3 className="font-semibold">Fields to Update:</h3>
                <div className="grid gap-2">
                  {Object.keys(batchUpdate).length === 0 ? (
                    <p className="text-sm text-muted-foreground">
                      No fields to update. Go to &quot;Edit Metadata&quot; tab to specify changes.
                    </p>
                  ) : (
                    Object.entries(batchUpdate).map(([key, value]) => {
                      if (!value || (Array.isArray(value) && value.length === 0)) return null
                      return (
                        <div key={key} className="flex items-center gap-2 p-2 bg-accent/50 rounded">
                          <Badge variant="outline">{key}</Badge>
                          <span className="text-sm">
                            {Array.isArray(value) ? value.join(", ") : String(value)}
                          </span>
                        </div>
                      )
                    })
                  )}
                </div>
              </div>

              <div className="flex justify-end gap-2 pt-4">
                <Button variant="outline" onClick={() => setBatchUpdate({})}>
                  <RefreshCw className="w-4 h-4 mr-2" />
                  Reset Changes
                </Button>
                <Button
                  onClick={applyBatchUpdate}
                  disabled={updating || Object.keys(batchUpdate).length === 0}
                >
                  <Save className="w-4 h-4 mr-2" />
                  {updating ? "Updating..." : `Apply to ${selectedBooks.size} Books`}
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}

