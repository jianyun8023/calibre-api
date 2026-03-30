"use client"

import dynamic from "next/dynamic"
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { X } from "lucide-react"

// 动态导入 EpubChapterViewer
const EpubChapterViewer = dynamic(
  () => import("./epub-chapter-viewer").then(mod => ({ default: mod.EpubChapterViewer })),
  {
    ssr: false,
    loading: () => (
      <div className="h-[80vh] flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4" />
          <p className="text-muted-foreground">加载阅读器...</p>
        </div>
      </div>
    ),
  }
)

interface BookPreviewDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  bookId: string
  bookTitle: string
}

export function BookPreviewDialog({
  open,
  onOpenChange,
  bookId,
  bookTitle,
}: BookPreviewDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="w-[80vw] h-[80vh] max-w-none sm:max-w-none p-0 gap-0" showCloseButton={false}>
        <DialogTitle className="sr-only">书籍预览 - {bookTitle}</DialogTitle>
        
        {/* 头部标题栏 */}
        <div className="flex items-center justify-between px-3 py-1 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 shrink-0">
          <div className="flex-1 min-w-0">
            <h2 className="text-xs font-semibold truncate leading-none mb-0.5">
              书籍预览
            </h2>
            <p className="text-xs text-muted-foreground truncate leading-none">
              {bookTitle}
            </p>
          </div>
          
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onOpenChange(false)}
            className="shrink-0 ml-2 h-6 w-6 p-0"
          >
            <X className="h-3 w-3" />
          </Button>
        </div>

        {/* 内容区域 - 嵌入 EpubChapterViewer */}
        <div className="flex-1 overflow-hidden">
          {open && (
            <EpubChapterViewer bookId={bookId} bookTitle={bookTitle} />
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
