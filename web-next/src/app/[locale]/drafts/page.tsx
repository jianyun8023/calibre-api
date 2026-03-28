"use client"

import { useEffect, useState, useCallback } from "react"
import { useTranslations } from "next-intl"
import dynamic from "next/dynamic"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { toast } from "sonner"
import { RefreshCcw, CheckCircle, XCircle, ArrowRight, Minus, Plus, Search, Edit } from "lucide-react"
import { Pagination } from "@/components/pagination"
import type { DoubanBook } from "@/lib/api/metadata"

// 懒加载元数据对话框组件
const MetadataEditDialog = dynamic(() => import("@/components/metadata-edit-dialog").then(mod => ({ default: mod.MetadataEditDialog })), {
  loading: () => <div>Loading...</div>,
})

const MetadataSearchDialog = dynamic(() => import("@/components/metadata-search-dialog").then(mod => ({ default: mod.MetadataSearchDialog })), {
  loading: () => <div>Loading...</div>,
})

const MetadataCompareDialog = dynamic(() => import("@/components/metadata-compare-dialog").then(mod => ({ default: mod.MetadataCompareDialog })), {
  loading: () => <div>Loading...</div>,
})

interface Draft {
  id: number
  book_id: string
  action: "delete" | "update"
  data: string
  status: string
  created_at: string
}

interface BookMetadata {
  id: number
  title: string
  authors: string[]
  publisher: string
  pubdate: string
  isbn: string
  tags: string[]
  rating: number
  comments: string
  cover: string
  file_path?: string
  size?: number
  identifiers?: Record<string, string>
  [key: string]: unknown | string[] | number | null | undefined
}

function DraftItem({ 
  draft, 
  isSelected, 
  onSelect,
  onApply,
  onReject,
  onBookUpdate
}: { 
  draft: Draft, 
  isSelected: boolean, 
  onSelect: (id: number, checked: boolean) => void,
  onApply: (id: number) => void,
  onReject: (id: number) => void,
  onBookUpdate: () => void
}) {
  const [book, setBook] = useState<BookMetadata | null>(null)
  const [loading, setLoading] = useState(true)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [searchDialogOpen, setSearchDialogOpen] = useState(false)
  const [compareDialogOpen, setCompareDialogOpen] = useState(false)
  const [doubanMetadata, setDoubanMetadata] = useState<DoubanBook | null>(null)

  const loadBook = useCallback(async () => {
    setLoading(true)
    try {
      const res = await fetch(`/api/book/${draft.book_id}`)
      if (res.ok) {
        const data = await res.json()
        setBook(data.data)
      }
    } catch (err) {
      console.error("Failed to fetch book", err)
    } finally {
      setLoading(false)
    }
  }, [draft.book_id])

  useEffect(() => {
    loadBook()
  }, [loadBook])

  const handleMetadataSelect = (metadata: DoubanBook) => {
    setDoubanMetadata(metadata)
    setCompareDialogOpen(true)
  }

  const handleMetadataUpdateSuccess = () => {
    loadBook()
    onBookUpdate()
  }

  // Parse update data
  const updates: Record<string, unknown> = {}
  // We need to wait for book data to be loaded before we can accurately diff
  if (draft.action === "update" && !loading) {
    try {
      const parsed = JSON.parse(draft.data)
      const sensitiveKeys = ['token', 'password', 'secret', 'apikey', 'authorization']
      
      for (const [k, v] of Object.entries(parsed)) {
        // Skip empty/zero values to only show actual changes
        if (v === null || v === "0001-01-01T00:00:00Z") {
          continue;
        }
        
        const oldV = book ? book[k] : undefined;

        // If the value is empty, but the original book also has an empty value, skip it
        const isNewEmpty = v === "" || (Array.isArray(v) && v.length === 0) || v === 0;
        const isOldEmpty = oldV === undefined || oldV === null || oldV === "" || (Array.isArray(oldV) && oldV.length === 0) || oldV === 0;
        
        if (isNewEmpty && isOldEmpty) {
          continue;
        }
        
        // **SPECIAL CASE**: Tags is the ONLY field that can be cleared (tags: [])
        // If this is a tags field with empty array and the old value is not empty, show it as a clear operation
        const isTagsClearOperation = k === 'tags' && Array.isArray(v) && v.length === 0 && !isOldEmpty;
        
        // **CRITICAL FIX**: The API sends the entire struct with empty fields for things that weren't changed.
        // We must completely ignore any empty new value to prevent false "clear" diffs.
        // EXCEPT for tags clearing operation
        if (isNewEmpty && !isTagsClearOperation) {
            continue;
        }
        
        // If the new value is exactly the same as the old value, skip it
        if (v === oldV) continue;
        
        // Handle array comparisons more robustly
        if (Array.isArray(v)) {
          if (Array.isArray(oldV)) {
            // Sort both arrays before comparing to ignore order differences
            const sortedV = [...v].sort();
            const sortedOldV = [...oldV].sort();
            if (JSON.stringify(sortedV) === JSON.stringify(sortedOldV)) {
              continue;
            }
          }
        }
        
        // Handle case where draft has string but book has array (like tags, authors)
        if (typeof v === 'string') {
          if (Array.isArray(oldV)) {
            // Some fields might have different formatting but mean the same thing
            // e.g. "author1, author2" vs ["author1", "author2"]
            const vArr = v.split(',').map(s => s.trim()).filter(Boolean).sort();
            const oldVArr = [...oldV].map(s => String(s).trim()).filter(Boolean).sort();
            if (JSON.stringify(vArr) === JSON.stringify(oldVArr)) {
              continue;
            }
          }
        }
        
        // Handle case where draft has array but book has string
        if (Array.isArray(v)) {
          if (typeof oldV === 'string') {
            const vArr = [...v].map(s => String(s).trim()).filter(Boolean).sort();
            const oldVArr = oldV.split(',').map(s => s.trim()).filter(Boolean).sort();
            if (JSON.stringify(vArr) === JSON.stringify(oldVArr)) {
              continue;
            }
          }
        }
        
        // Special handling for strings (trim before comparing)
        if (typeof v === 'string') {
          if (typeof oldV === 'string' && v.trim() === oldV.trim()) {
            continue;
          }
        }
        
        // Special handling for dates
        if (k === 'pubdate' && typeof v === 'string') {
          try {
            const d1 = new Date(v);
            if (!Number.isNaN(d1.getTime()) && d1.getFullYear() <= 1900) {
              continue; // Skip zero dates
            }
            if (typeof oldV === 'string') {
              const d2 = new Date(oldV);
              if (!Number.isNaN(d1.getTime()) && !Number.isNaN(d2.getTime())) {
                if (d1.toLocaleDateString() === d2.toLocaleDateString()) {
                  continue;
                }
              }
            }
          } catch {}
        }
        
        // Handle numeric 0 vs undefined/null
        if (v === 0 && (oldV === undefined || oldV === null || oldV === 0 || oldV === "")) {
           continue;
        }
        
        // Handle undefined vs null vs empty string
        if ((v === "" || v === null || v === undefined) && (oldV === "" || oldV === null || oldV === undefined)) {
            continue;
        }

        if (sensitiveKeys.some(sk => k.toLowerCase().includes(sk))) {
          updates[k] = '********';
        } else {
          updates[k] = v;
        }
      }
    } catch {}
  }

  // 格式化单个值用于显示
  const formatValue = (key: string, v: unknown) => {
    if (v === undefined || v === null || v === "") return <span className="text-muted-foreground italic">empty</span>
    if (Array.isArray(v)) {
      if (v.length === 0) return <span className="text-muted-foreground italic">empty</span>
      return v.join(", ")
    }
    if (typeof v === 'object') return JSON.stringify(v)
    if (key === 'pubdate' && typeof v === 'string') {
      try {
        const d = new Date(v)
        if (!Number.isNaN(d.getTime())) {
          if (d.getFullYear() <= 1900) return <span className="text-muted-foreground italic">empty</span>
          return d.toLocaleDateString()
        }
      } catch {}
    }
    if (v === 0 && (key === 'rating' || key === 'series_index')) {
       return <span className="text-muted-foreground italic">empty</span>
    }
    return String(v)
  }

  // 渲染数组类型的详细对比（tags、authors）
  const renderArrayDiff = (key: string, oldValue: unknown, newValue: unknown) => {
    const oldArr = Array.isArray(oldValue) ? oldValue : []
    const newArr = Array.isArray(newValue) ? newValue : []
    
    // 如果新值是空数组，表示清空操作
    if (newArr.length === 0 && oldArr.length > 0) {
      return (
        <div className="space-y-2">
          <div className="text-xs text-muted-foreground mb-1">清空所有项目</div>
          {oldArr.map((item, idx) => (
            <div key={`old-${idx}`} className="flex items-center gap-2 text-sm bg-red-50 dark:bg-red-950/20 text-red-700 dark:text-red-400 px-3 py-1.5 rounded border border-red-200 dark:border-red-800">
              <Minus className="h-3 w-3 shrink-0" />
              <span className="line-through">{String(item)}</span>
            </div>
          ))}
        </div>
      )
    }

    const removed = oldArr.filter(item => !newArr.includes(item))
    const added = newArr.filter(item => !oldArr.includes(item))
    const unchanged = oldArr.filter(item => newArr.includes(item))

    if (removed.length === 0 && added.length === 0) {
      return <span className="text-muted-foreground italic text-sm">无变化</span>
    }

    return (
      <div className="space-y-2">
        {/* 保持不变的项 */}
        {unchanged.length > 0 && (
          <div className="text-xs text-muted-foreground mb-1">
            保持不变 ({unchanged.length} 项): {unchanged.join(", ")}
          </div>
        )}
        
        {/* 移除的项 */}
        {removed.map((item, idx) => (
          <div key={`removed-${idx}`} className="flex items-center gap-2 text-sm bg-red-50 dark:bg-red-950/20 text-red-700 dark:text-red-400 px-3 py-1.5 rounded border border-red-200 dark:border-red-800">
            <Minus className="h-3 w-3 shrink-0" />
            <span className="line-through">{String(item)}</span>
          </div>
        ))}
        
        {/* 新增的项 */}
        {added.map((item, idx) => (
          <div key={`added-${idx}`} className="flex items-center gap-2 text-sm bg-green-50 dark:bg-green-950/20 text-green-700 dark:text-green-400 px-3 py-1.5 rounded border border-green-200 dark:border-green-800">
            <Plus className="h-3 w-3 shrink-0" />
            <span className="font-medium">{String(item)}</span>
          </div>
        ))}
      </div>
    )
  }

  // 渲染单个字段的对比
  const renderFieldDiff = (key: string, oldValue: unknown, newValue: unknown) => {
    const isArrayField = key === 'tags' || key === 'authors'
    
    if (isArrayField) {
      return (
        <div className="space-y-2">
          <div className="font-medium capitalize text-sm">{key}</div>
          {renderArrayDiff(key, oldValue, newValue)}
        </div>
      )
    }

    // 普通字段的对比
    const oldIsEmpty = oldValue === undefined || oldValue === null || oldValue === "" || oldValue === 0
    const newIsEmpty = newValue === undefined || newValue === null || newValue === "" || (Array.isArray(newValue) && newValue.length === 0)

    return (
      <div className="space-y-2">
        <div className="font-medium capitalize text-sm text-muted-foreground">{key}</div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {/* 旧值 */}
          <div className="space-y-1">
            <div className="text-xs text-muted-foreground">当前值</div>
            <div className={`p-3 rounded-md border text-sm ${
              oldIsEmpty 
                ? "bg-muted/30 border-muted" 
                : "bg-red-50 dark:bg-red-950/20 border-red-200 dark:border-red-800"
            }`}>
              {oldIsEmpty ? (
                <span className="text-muted-foreground italic">empty</span>
              ) : (
                <span className={oldIsEmpty ? "" : "text-red-700 dark:text-red-400"}>
                  {formatValue(key, oldValue)}
                </span>
              )}
            </div>
          </div>

          {/* 新值 */}
          <div className="space-y-1">
            <div className="text-xs text-muted-foreground flex items-center gap-1">
              <ArrowRight className="h-3 w-3" />
              更新为
            </div>
            <div className={`p-3 rounded-md border text-sm ${
              newIsEmpty
                ? "bg-muted/30 border-muted"
                : "bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800"
            }`}>
              {newIsEmpty ? (
                <span className="text-muted-foreground italic">empty</span>
              ) : (
                <span className="text-green-700 dark:text-green-400 font-medium">
                  {formatValue(key, newValue)}
                </span>
              )}
            </div>
          </div>
        </div>
      </div>
    )
  }

  const renderChanges = () => {
    if (draft.action === "delete") {
      return (
        <div className="space-y-3">
          <div className="text-destructive font-medium flex items-center gap-2">
            <XCircle className="h-4 w-4" /> This book will be deleted
          </div>
          {loading ? (
            <Skeleton className="h-16 w-full" />
          ) : book ? (
            <div className="bg-muted/30 p-3 rounded-md text-sm">
              <div className="grid grid-cols-[100px_1fr] gap-y-2 gap-x-4">
                <span className="text-muted-foreground">Title:</span>
                <span className="font-medium">{book.title}</span>
                <span className="text-muted-foreground">Authors:</span>
                <span>{book.authors?.join(", ")}</span>
                {book.publisher && (
                  <>
                    <span className="text-muted-foreground">Publisher:</span>
                    <span>{book.publisher}</span>
                  </>
                )}
              </div>
            </div>
          ) : (
            <div className="text-muted-foreground text-sm italic">Book metadata not available</div>
          )}
        </div>
      )
    }

    if (Object.keys(updates).length === 0) {
      if (loading) {
        return <Skeleton className="h-24 w-full" />
      }
      return <span className="text-muted-foreground italic">No visible changes</span>
    }

    return (
      <div className="space-y-4">
        {loading ? (
          <Skeleton className="h-5 w-1/2 mb-4" />
        ) : book ? (
          <div className="bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-800 p-3 rounded-md">
            <div className="text-sm text-blue-900 dark:text-blue-200">
              <span className="font-medium">当前书籍:</span> {book.title}
              {book.authors && book.authors.length > 0 && (
                <span className="ml-2 text-blue-700 dark:text-blue-300">
                  - {book.authors.join(", ")}
                </span>
              )}
            </div>
          </div>
        ) : null}
        
        <div className="space-y-6">
          {Object.entries(updates).map(([key, newValue]) => {
            const oldValue = book ? book[key] : undefined
            return (
              <div key={key} className="border-l-2 border-blue-500 pl-4">
                {renderFieldDiff(key, oldValue, newValue)}
              </div>
            )
          })}
        </div>
      </div>
    )
  }

  return (
    <Card className={isSelected ? "border-primary" : ""}>
      <CardHeader className="py-4 flex flex-row items-center space-y-0 gap-4">
        <Checkbox
          checked={isSelected}
          onCheckedChange={(checked) => onSelect(draft.id, !!checked)}
        />
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <CardTitle className="text-lg">Book ID: {draft.book_id}</CardTitle>
            <Badge variant={draft.action === "delete" ? "destructive" : "secondary"}>
              {draft.action.toUpperCase()}
            </Badge>
          </div>
          <CardDescription>
            Created at: {new Date(draft.created_at).toLocaleString()}
          </CardDescription>
        </div>
        <div className="flex gap-2">
          {/* 元数据操作按钮（仅更新操作显示） */}
          {draft.action === "update" && book && (
            <>
              <Button 
                variant="outline" 
                size="icon" 
                title="搜索元数据"
                onClick={() => setSearchDialogOpen(true)}
              >
                <Search className="h-4 w-4" />
              </Button>
              <Button 
                variant="outline" 
                size="icon" 
                title="手动编辑元数据"
                onClick={() => setEditDialogOpen(true)}
              >
                <Edit className="h-4 w-4" />
              </Button>
            </>
          )}
          <Button 
            variant="outline" 
            size="sm" 
            className="text-green-600 hover:text-green-700 hover:bg-green-50"
            onClick={() => onApply(draft.id)}
          >
            <CheckCircle className="h-4 w-4 mr-1" /> Apply
          </Button>
          <Button 
            variant="outline" 
            size="sm" 
            className="text-destructive hover:text-destructive hover:bg-destructive/10"
            onClick={() => onReject(draft.id)}
          >
            <XCircle className="h-4 w-4 mr-1" /> Reject
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {renderChanges()}
      </CardContent>

      {/* 元数据对话框（仅更新操作显示） */}
      {draft.action === "update" && book && (
        <>
          <MetadataEditDialog
            open={editDialogOpen}
            onOpenChange={setEditDialogOpen}
            book={book}
            onSuccess={handleMetadataUpdateSuccess}
          />
          <MetadataSearchDialog
            open={searchDialogOpen}
            onOpenChange={setSearchDialogOpen}
            book={book}
            onSelect={handleMetadataSelect}
          />
          {doubanMetadata && (
            <MetadataCompareDialog
              open={compareDialogOpen}
              onOpenChange={setCompareDialogOpen}
              book={book}
              doubanMetadata={doubanMetadata}
              onSuccess={handleMetadataUpdateSuccess}
            />
          )}
        </>
      )}
    </Card>
  )
}

export default function DraftsPage() {
  const t = useTranslations("drafts")
  const [drafts, setDrafts] = useState<Draft[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set())
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const limit = 10

  const fetchDrafts = useCallback(async (currentPage: number) => {
    setLoading(true)
    try {
      const offset = (currentPage - 1) * limit
      const res = await fetch(`/api/drafts?limit=${limit}&offset=${offset}`)
      if (!res.ok) throw new Error("Failed to fetch drafts")
      const data = await res.json()
      setDrafts(data.data || [])
      setTotalPages(data.total_pages || 1)
      setPage(data.page || 1)
    } catch (err) {
      console.error(err)
      toast.error(t("fetchError"))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    fetchDrafts(1)
  }, [fetchDrafts])

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedIds(new Set(drafts.map(d => d.id)))
    } else {
      setSelectedIds(new Set())
    }
  }

  const handleSelect = (id: number, checked: boolean) => {
    const newSelected = new Set(selectedIds)
    if (checked) {
      newSelected.add(id)
    } else {
      newSelected.delete(id)
    }
    setSelectedIds(newSelected)
  }

  const handleApply = async (ids?: number[]) => {
    const targetIds = ids || Array.from(selectedIds)
    if (targetIds.length === 0) return
    try {
      const res = await fetch("/api/drafts/apply", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ids: targetIds }),
      })
      if (!res.ok) throw new Error("Failed to apply drafts")
      toast.success(t("applySuccess"))
      if (!ids) {
        setSelectedIds(new Set())
      } else {
        const newSelected = new Set(selectedIds)
        ids.forEach(id => { newSelected.delete(id) })
        setSelectedIds(newSelected)
      }
      fetchDrafts(page)
    } catch (err) {
      console.error(err)
      toast.error(t("applyError"))
    }
  }

  const handleReject = async (ids?: number[]) => {
    const targetIds = ids || Array.from(selectedIds)
    if (targetIds.length === 0) return
    try {
      const res = await fetch("/api/drafts/reject", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ids: targetIds }),
      })
      if (!res.ok) throw new Error("Failed to reject drafts")
      toast.success(t("rejectSuccess"))
      if (!ids) {
        setSelectedIds(new Set())
      } else {
        const newSelected = new Set(selectedIds)
        ids.forEach(id => { newSelected.delete(id) })
        setSelectedIds(newSelected)
      }
      fetchDrafts(page)
    } catch (err) {
      console.error(err)
      toast.error(t("rejectError"))
    }
  }

  return (
    <div className="container mx-auto py-8 px-4 max-w-5xl">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">{t("title")}</h1>
          <p className="text-muted-foreground mt-2">
            {t("description")}
          </p>
        </div>
        <Button variant="outline" size="icon" onClick={() => fetchDrafts(page)} disabled={loading}>
          <RefreshCcw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
        </Button>
      </div>

      <div className="flex items-center space-x-4 mb-4 bg-muted/50 p-4 rounded-lg">
        <Checkbox
          checked={drafts.length > 0 && selectedIds.size === drafts.length}
          onCheckedChange={(checked) => handleSelectAll(!!checked)}
          disabled={drafts.length === 0}
        />
        <span className="text-sm font-medium">
          {selectedIds.size} {t("selected")}
        </span>
        <div className="flex-1" />
        <Button
          variant="default"
          size="sm"
          disabled={selectedIds.size === 0}
          onClick={() => handleApply()}
          className="bg-green-600 hover:bg-green-700"
        >
          <CheckCircle className="mr-2 h-4 w-4" />
          {t("apply")}
        </Button>
        <Button
          variant="destructive"
          size="sm"
          disabled={selectedIds.size === 0}
          onClick={() => handleReject()}
        >
          <XCircle className="mr-2 h-4 w-4" />
          {t("reject")}
        </Button>
      </div>

      <div className="space-y-4">
        {loading ? (
          Array.from({ length: 3 }).map((_, i) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: skeleton items
            <Card key={i}>
              <CardHeader className="py-4">
                <Skeleton className="h-6 w-1/3 mb-2" />
                <Skeleton className="h-4 w-1/4" />
              </CardHeader>
            </Card>
          ))
        ) : drafts.length === 0 ? (
          <div className="text-center py-12 text-muted-foreground">
            {t("noDrafts")}
          </div>
        ) : (
          drafts.map((draft) => (
            <DraftItem 
              key={draft.id} 
              draft={draft} 
              isSelected={selectedIds.has(draft.id)} 
              onSelect={handleSelect}
              onApply={(id) => handleApply([id])}
              onReject={(id) => handleReject([id])}
              onBookUpdate={() => fetchDrafts(page)}
            />
          ))
        )}

        {drafts.length > 0 && totalPages > 1 && (
          <div className="mt-8">
            <Pagination
              currentPage={page}
              totalPages={totalPages}
              hasNext={page < totalPages}
              hasPrev={page > 1}
              onNext={() => fetchDrafts(page + 1)}
              onPrev={() => fetchDrafts(page - 1)}
            />
          </div>
        )}
      </div>
    </div>
  )
}
