"use client"

import { useState, useEffect, useMemo } from "react"
import dynamic from "next/dynamic"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { HoverCard, HoverCardContent, HoverCardTrigger } from "@/components/ui/hover-card"
import { toast } from "sonner"
import { ExternalLink, CheckCircle, XCircle, SkipForward, AlertTriangle, RotateCcw, Search, Edit, Book } from "lucide-react"
import Image from "next/image"
import { InlineEditField } from "./inline-edit-field"
import type { DoubanBook } from "@/lib/api/metadata"

// 懒加载元数据对话框
const MetadataSearchDialog = dynamic(() => import("./metadata-search-dialog").then(mod => ({ default: mod.MetadataSearchDialog })), {
  ssr: false,
})

const QuickEditDialog = dynamic(() => import("./quick-edit-dialog").then(mod => ({ default: mod.QuickEditDialog })), {
  ssr: false,
})

const BookPreviewDialog = dynamic(() => import("./book-preview-dialog").then(mod => ({ default: mod.BookPreviewDialog })), {
  ssr: false,
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
  [key: string]: unknown
}

interface WorkbenchCardProps {
  draft: Draft
  onApply: (editedData?: Record<string, unknown>) => void
  onReject: () => void
  onSkip: () => void
  isProcessing: boolean
}

export function WorkbenchCard({
  draft,
  onApply,
  onReject,
  onSkip,
  isProcessing,
}: WorkbenchCardProps) {
  const [book, setBook] = useState<BookMetadata | null>(null)
  const [loading, setLoading] = useState(true)
  const [editedData, setEditedData] = useState<Record<string, unknown>>({})
  const [searchDialogOpen, setSearchDialogOpen] = useState(false)
  const [quickEditDialogOpen, setQuickEditDialogOpen] = useState(false)
  const [previewDialogOpen, setPreviewDialogOpen] = useState(false)

  // 解析更新数据（使用 useMemo 避免每次渲染都重新计算）
  const updates = useMemo(() => {
    const result: Record<string, unknown> = {}
    if (!draft || draft.action !== "update") return result
    
    try {
      const parsed = JSON.parse(draft.data)
      for (const [k, v] of Object.entries(parsed)) {
        if (v === null || v === "0001-01-01T00:00:00Z") continue
        const oldV = book ? book[k] : undefined
        const isNewEmpty = v === "" || (Array.isArray(v) && v.length === 0) || v === 0
        const isOldEmpty = oldV === undefined || oldV === null || oldV === "" || (Array.isArray(oldV) && oldV.length === 0) || oldV === 0
        
        // 允许清空标签
        const isTagsClearOperation = k === 'tags' && Array.isArray(v) && v.length === 0 && !isOldEmpty
        
        if (isNewEmpty && !isTagsClearOperation) continue
        if (v === oldV) continue
        
        // 数组对比
        if (Array.isArray(v) && Array.isArray(oldV)) {
          const sortedV = [...v].sort()
          const sortedOldV = [...oldV].sort()
          if (JSON.stringify(sortedV) === JSON.stringify(sortedOldV)) continue
        }
        
        result[k] = v
      }
    } catch {}
    return result
  }, [draft, book])

  // 计算所有需要显示的字段（合并原始草稿和编辑数据，过滤无意义变更）
  const displayFields = useMemo(() => {
    const allFields = new Set<string>([...Object.keys(updates), ...Object.keys(editedData)])
    
    // 过滤掉与旧值相同或空值的字段
    return Array.from(allFields).filter(key => {
      const newValue = editedData[key] !== undefined ? editedData[key] : updates[key]
      const oldValue = book ? book[key] : undefined
      
      // 如果新值为空且不是清空操作，跳过
      const isNewEmpty = newValue === undefined || newValue === null || newValue === "" || (Array.isArray(newValue) && newValue.length === 0) || newValue === 0
      const isOldEmpty = oldValue === undefined || oldValue === null || oldValue === "" || (Array.isArray(oldValue) && oldValue.length === 0) || oldValue === 0
      const isTagsClearOperation = key === 'tags' && Array.isArray(newValue) && newValue.length === 0 && !isOldEmpty
      
      if (isNewEmpty && !isTagsClearOperation) return false
      
      // 如果新值与旧值相同，跳过
      if (newValue === oldValue) return false
      
      // 数组字段特殊处理
      if (Array.isArray(newValue) && Array.isArray(oldValue)) {
        const sortedNew = [...newValue].sort()
        const sortedOld = [...oldValue].sort()
        if (JSON.stringify(sortedNew) === JSON.stringify(sortedOld)) return false
      }
      
      return true
    })
  }, [updates, editedData, book])

  // 加载书籍数据
  useEffect(() => {
    if (!draft?.book_id) {
      setLoading(false)
      return
    }
    
    const loadBook = async () => {
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
    }
    loadBook()
  }, [draft?.book_id])

  // 当切换到新草稿时，重置编辑数据和对话框状态
  useEffect(() => {
    setEditedData({ ...updates })
    setSearchDialogOpen(false)
    setQuickEditDialogOpen(false)
  }, [draft?.id, updates])

  // 防御性检查：如果 draft 不存在，显示错误状态
  if (!draft) {
    return (
      <Card className="w-full shadow-lg">
        <CardContent className="py-12">
          <div className="text-center text-muted-foreground">
            <AlertTriangle className="h-12 w-12 mx-auto mb-4 text-destructive" />
            <p>无法加载草稿数据</p>
          </div>
        </CardContent>
      </Card>
    )
  }
  
  // 获取当前显示的值（优先使用编辑后的值）
  const getCurrentValue = (key: string) => {
    return editedData[key] !== undefined ? editedData[key] : updates[key]
  }
  
  // 重置编辑
  const handleReset = () => {
    setEditedData({ ...updates })
  }
  
  // 更新字段值
  const handleFieldUpdate = (key: string, value: unknown) => {
    setEditedData(prev => ({ ...prev, [key]: value }))
  }
  
  // 检查是否有未保存的修改
  const hasUnsavedChanges = JSON.stringify(editedData) !== JSON.stringify(updates)
  
  // 处理豆瓣元数据选择
  const handleMetadataSelect = (metadata: DoubanBook) => {
    // 将豆瓣元数据转换为可编辑的格式，填充到 editedData
    const newEditedData: Record<string, unknown> = { ...editedData }
    
    if (metadata.title) {
      newEditedData.title = metadata.title
    }
    if (metadata.author && metadata.author.length > 0) {
      newEditedData.authors = metadata.author
    }
    if (metadata.publisher) {
      newEditedData.publisher = metadata.publisher
    }
    if (metadata.pubdate) {
      newEditedData.pubdate = metadata.pubdate
    }
    if (metadata.isbn13) {
      newEditedData.isbn = metadata.isbn13
    }
    if (metadata.rating && metadata.rating.average > 0) {
      newEditedData.rating = metadata.rating.average
    }
    if (metadata.tags && metadata.tags.length > 0) {
      newEditedData.tags = metadata.tags.map(t => t.name)
    }
    if (metadata.summary) {
      newEditedData.comments = metadata.summary
    }
    
    setEditedData(newEditedData)
    setSearchDialogOpen(false)
    toast.success("豆瓣元数据已填充到编辑区域，请检查后再应用")
  }
  
  // 处理快速编辑应用
  const handleQuickEditApply = (data: Record<string, unknown>) => {
    setEditedData(prev => ({ ...prev, ...data }))
  }

  const formatValue = (v: unknown) => {
    if (v === undefined || v === null || v === "") return <span className="text-muted-foreground italic">empty</span>
    if (Array.isArray(v)) {
      if (v.length === 0) return <span className="text-muted-foreground italic">empty</span>
      return v.join(", ")
    }
    if (typeof v === 'object') return JSON.stringify(v)
    return String(v)
  }

  const renderFieldDiff = (key: string, oldValue: unknown, newValue: unknown) => {
    const isArrayField = key === 'tags' || key === 'authors'
    
    if (isArrayField) {
      const oldArr = Array.isArray(oldValue) ? oldValue : []
      const newArr = Array.isArray(newValue) ? newValue : []
      
      // 清空操作
      if (newArr.length === 0 && oldArr.length > 0) {
        return (
          <div className="space-y-2">
            <div className="font-medium text-sm capitalize">{key}</div>
            <div className="text-xs text-muted-foreground mb-1">清空所有项目</div>
            <div className="flex flex-wrap gap-1">
              {oldArr.map((item, idx) => (
                <Badge key={idx} variant="destructive" className="line-through">
                  {String(item)}
                </Badge>
              ))}
            </div>
          </div>
        )
      }

      const removed = oldArr.filter(item => !newArr.includes(item))
      const added = newArr.filter(item => !oldArr.includes(item))
      const unchanged = oldArr.filter(item => newArr.includes(item))

      if (removed.length === 0 && added.length === 0) return null

      return (
        <div className="space-y-2">
          <div className="font-medium text-sm capitalize">{key}</div>
          {unchanged.length > 0 && (
            <div className="text-xs text-muted-foreground">
              保持: {unchanged.join(", ")}
            </div>
          )}
          {removed.length > 0 && (
            <div className="flex flex-wrap gap-1">
              {removed.map((item, idx) => (
                <Badge key={idx} variant="destructive">
                  - {String(item)}
                </Badge>
              ))}
            </div>
          )}
          {added.length > 0 && (
            <div className="flex flex-wrap gap-1">
              {added.map((item, idx) => (
                <Badge key={idx} className="bg-green-500 hover:bg-green-600">
                  + {String(item)}
                </Badge>
              ))}
            </div>
          )}
        </div>
      )
    }

    // 普通字段
    return (
      <div className="space-y-2">
        <div className="font-medium text-sm capitalize">{key}</div>
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div className="space-y-1">
            <div className="text-xs text-muted-foreground">旧值</div>
            <div className="p-2 rounded bg-muted/30">
              {formatValue(oldValue)}
            </div>
          </div>
          <div className="space-y-1">
            <div className="text-xs text-muted-foreground">→ 新值</div>
            <div className="p-2 rounded bg-blue-50 dark:bg-blue-950/20 font-medium">
              {formatValue(newValue)}
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (loading) {
    return (
      <Card className="w-full">
        <CardHeader>
          <Skeleton className="h-6 w-2/3" />
          <Skeleton className="h-4 w-1/3" />
        </CardHeader>
        <CardContent>
          <div className="flex gap-4">
            <Skeleton className="h-40 w-28 shrink-0" />
            <div className="flex-1 space-y-3">
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-3/4" />
              <Skeleton className="h-4 w-1/2" />
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="w-full shadow-lg flex flex-col max-h-[85vh]">
      <CardHeader className="pb-4 shrink-0">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            <Badge variant={draft.action === "delete" ? "destructive" : "secondary"}>
              {draft.action === "delete" ? "删除" : "更新"}
            </Badge>
            <span className="text-sm text-muted-foreground">
              书籍 ID: {draft.book_id}
            </span>
          </div>
          <a
            href={`/zh-CN/detail/${draft.book_id}`}
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary hover:underline flex items-center gap-1 text-sm"
          >
            查看详情
            <ExternalLink className="h-3 w-3" />
          </a>
        </div>
      </CardHeader>

      <CardContent className="space-y-6 overflow-y-auto flex-1 scroll-smooth [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-gray-300 [&::-webkit-scrollbar-thumb]:rounded-full hover:[&::-webkit-scrollbar-thumb]:bg-gray-400 dark:[&::-webkit-scrollbar-thumb]:bg-gray-600 dark:hover:[&::-webkit-scrollbar-thumb]:bg-gray-500">
        {/* 书籍信息 */}
        {book && (
          <div className="flex gap-4">
            {book.cover ? (
              <HoverCard openDelay={200}>
                <HoverCardTrigger asChild>
                  <div className="shrink-0 cursor-pointer">
                    <Image
                      src={book.cover}
                      alt={book.title}
                      width={112}
                      height={160}
                      className="rounded shadow-md object-cover transition-opacity hover:opacity-90"
                      onError={(e: React.SyntheticEvent<HTMLImageElement>) => {
                        const target = e.target as HTMLImageElement
                        target.style.display = 'none'
                      }}
                    />
                  </div>
                </HoverCardTrigger>
                <HoverCardContent side="right" align="start" className="w-auto p-1 border-2">
                  <Image
                    src={book.cover}
                    alt={`${book.title} - 大图`}
                    width={400}
                    height={600}
                    className="rounded shadow-xl object-contain max-h-[80vh]"
                    onError={(e: React.SyntheticEvent<HTMLImageElement>) => {
                      const target = e.target as HTMLImageElement
                      target.style.display = 'none'
                    }}
                  />
                </HoverCardContent>
              </HoverCard>
            ) : (
              <div className="h-40 w-28 shrink-0 bg-muted rounded flex items-center justify-center text-muted-foreground text-xs">
                无封面
              </div>
            )}

            <div className="flex-1 space-y-2">
              <h3 className="text-xl font-bold">{book.title}</h3>
              {book.authors && book.authors.length > 0 && (
                <p className="text-sm text-muted-foreground">
                  作者: {book.authors.join(", ")}
                </p>
              )}
              {book.publisher && (
                <p className="text-sm text-muted-foreground">
                  出版社: {book.publisher}
                </p>
              )}
              {book.isbn && (
                <p className="text-sm text-muted-foreground font-mono">
                  ISBN: {book.isbn}
                </p>
              )}
            </div>
          </div>
        )}

        {/* DELETE 类型 */}
        {draft.action === "delete" && (
          <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4">
            <div className="flex items-center gap-2 text-destructive mb-2">
              <AlertTriangle className="h-5 w-5" />
              <span className="font-semibold">将删除此书籍</span>
            </div>
            <p className="text-sm text-muted-foreground">
              此操作不可撤销，书籍文件和所有元数据将被永久删除。
            </p>
          </div>
        )}

        {/* UPDATE 类型 - 显示变更 */}
        {draft.action === "update" && displayFields.length > 0 && (
          <div className="space-y-4">
            <div className="flex items-center gap-2 text-sm font-semibold text-primary">
              <div className="h-px flex-1 bg-border" />
              <span>草稿修改内容</span>
              <div className="h-px flex-1 bg-border" />
            </div>
            
            {displayFields.map((key) => {
              const oldValue = book ? book[key] : undefined
              const currentValue = getCurrentValue(key)
              const isArrayField = key === 'tags' || key === 'authors'
              const isTextareaField = key === 'comments'
              
              // 可内联编辑的字段
              const editableFields = ['title', 'publisher', 'isbn', 'comments', 'tags', 'authors']
              
              if (editableFields.includes(key)) {
                let fieldType: "text" | "array" | "textarea" = "text"
                if (isArrayField) fieldType = "array"
                else if (isTextareaField) fieldType = "textarea"
                
                return (
                  <div key={key}>
                    <InlineEditField
                      label={key}
                      oldValue={oldValue}
                      newValue={currentValue}
                      onUpdate={(value) => handleFieldUpdate(key, value)}
                      type={fieldType}
                    />
                  </div>
                )
              }
              
              // 其他字段使用只读显示（如 rating、pubdate 等）
              return (
                <div key={key}>
                  {renderFieldDiff(key, oldValue, currentValue)}
                </div>
              )
            })}
            
            {/* 快捷操作 */}
            <div className="flex items-center gap-2 pt-2 border-t">
              <div className="h-px flex-1 bg-border" />
              <span className="text-xs text-muted-foreground">快捷操作</span>
              <div className="h-px flex-1 bg-border" />
            </div>
            <div className="flex items-center justify-center gap-2 flex-wrap">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setSearchDialogOpen(true)}
                disabled={isProcessing}
              >
                <Search className="h-3 w-3 mr-1" />
                搜索豆瓣元数据
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setQuickEditDialogOpen(true)}
                disabled={isProcessing}
              >
                <Edit className="h-3 w-3 mr-1" />
                手动编辑
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPreviewDialogOpen(true)}
                disabled={!draft?.book_id || isProcessing}
              >
                <Book className="h-3 w-3 mr-1" />
                预览内容
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={handleReset}
                disabled={!hasUnsavedChanges || isProcessing}
              >
                <RotateCcw className="h-3 w-3 mr-1" />
                重置修改
              </Button>
            </div>
            <div className="text-xs text-center text-muted-foreground">
              💡 提示：搜索和手动编辑的元数据会自动填充到上方编辑区域，您可以继续调整后再应用
            </div>
            
            {hasUnsavedChanges && (
              <div className="bg-yellow-50 dark:bg-yellow-950/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-3 text-sm text-yellow-800 dark:text-yellow-200">
                ⚠️ 您有未保存的修改，点击「应用草稿」将使用编辑后的数据
              </div>
            )}
          </div>
        )}

        {/* UPDATE 类型但无变更 */}
        {draft.action === "update" && displayFields.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            <p>无可见变更</p>
          </div>
        )}
      </CardContent>

      {/* 操作按钮 - 固定在底部 */}
      <div className="border-t px-6 py-4 shrink-0 bg-background">
        <div className="flex items-center justify-center gap-3">
          <Button
            variant="outline"
            onClick={onSkip}
            disabled={isProcessing}
          >
            <SkipForward className="h-4 w-4 mr-2" />
            跳过
          </Button>
          
          <Button
            variant="outline"
            className="text-destructive hover:text-destructive hover:bg-destructive/10"
            onClick={onReject}
            disabled={isProcessing}
          >
            <XCircle className="h-4 w-4 mr-2" />
            拒绝草稿
          </Button>
          
          <Button
            className="bg-green-600 hover:bg-green-700"
            onClick={() => onApply(hasUnsavedChanges ? editedData : undefined)}
            disabled={isProcessing}
          >
            <CheckCircle className="h-4 w-4 mr-2" />
            应用草稿
          </Button>
        </div>
      </div>

      {/* 元数据搜索对话框（仅更新操作显示） */}
      {draft.action === "update" && book && (
        <>
          <MetadataSearchDialog
            open={searchDialogOpen}
            onOpenChange={setSearchDialogOpen}
            book={book}
            onSelect={handleMetadataSelect}
          />
          <QuickEditDialog
            open={quickEditDialogOpen}
            onOpenChange={setQuickEditDialogOpen}
            bookId={Number(draft.book_id)}
            initialData={{ ...book, ...editedData }}
            onApply={handleQuickEditApply}
          />
        </>
      )}

      {/* 书籍预览对话框 */}
      {book && (
        <BookPreviewDialog
          open={previewDialogOpen}
          onOpenChange={setPreviewDialogOpen}
          bookId={draft.book_id}
          bookTitle={book.title || `书籍 ${draft.book_id}`}
        />
      )}
    </Card>
  )
}
