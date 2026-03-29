"use client"

import { useState, useEffect, useCallback } from "react"
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Progress } from "@/components/ui/progress"
import { toast } from "sonner"
import { X, ChevronLeft, ChevronRight, Sparkles } from "lucide-react"
import { WorkbenchCard } from "./workbench-card"

interface Draft {
  id: number
  book_id: string
  action: "delete" | "update"
  data: string
  status: string
  created_at: string
}

interface DraftWorkbenchDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onRefresh: () => void
}

export function DraftWorkbenchDialog({
  open,
  onOpenChange,
  onRefresh,
}: DraftWorkbenchDialogProps) {
  const [queue, setQueue] = useState<Draft[]>([])
  const [currentIndex, setCurrentIndex] = useState(0)
  const [isProcessing, setIsProcessing] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [processedCount, setProcessedCount] = useState(0)
  const [showCompletion, setShowCompletion] = useState(false)

  // 加载草稿队列
  const loadQueue = useCallback(async () => {
    setIsLoading(true)
    try {
      // 加载所有待处理草稿
      const res = await fetch("/api/drafts?limit=1000&offset=0")
      if (!res.ok) throw new Error("Failed to fetch drafts")
      const data = await res.json()
      const drafts = data.data || []
      setQueue(drafts)
      setCurrentIndex(0)
      setProcessedCount(0)
    } catch (error) {
      console.error("Failed to load drafts:", error)
      toast.error("加载草稿失败")
    } finally {
      setIsLoading(false)
    }
  }, [])

  // 打开时加载队列
  useEffect(() => {
    if (open) {
      loadQueue()
      setShowCompletion(false)
    }
  }, [open, loadQueue])

  // 处理完成时显示提示
  useEffect(() => {
    if (queue.length === 0 && !isLoading && processedCount > 0) {
      setShowCompletion(true)
      const timer = setTimeout(() => {
        onOpenChange(false)
        onRefresh()
      }, 3000)
      return () => clearTimeout(timer)
    }
  }, [queue.length, isLoading, processedCount, onOpenChange, onRefresh])

  // 当队列变化时，调整当前索引以避免越界
  useEffect(() => {
    if (currentIndex >= queue.length && queue.length > 0) {
      setCurrentIndex(queue.length - 1)
    } else if (queue.length === 0) {
      setCurrentIndex(0)
    }
  }, [queue.length, currentIndex])

  // 应用草稿
  const handleApply = async (draftId: number, editedData?: Record<string, unknown>) => {
    setIsProcessing(true)
    try {
      // 如果有编辑数据，先更新草稿
      if (editedData && Object.keys(editedData).length > 0) {
        const updateRes = await fetch("/api/drafts/update", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            updates: [{
              id: String(draftId),
              data: editedData,
            }]
          }),
        })
        if (!updateRes.ok) {
          throw new Error("Failed to update draft")
        }
      }
      
      // 应用草稿
      const res = await fetch("/api/drafts/apply", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ids: [draftId] }),
      })
      if (!res.ok) throw new Error("Failed to apply draft")
      
      toast.success("草稿已应用")
      setQueue(prev => prev.filter(d => d.id !== draftId))
      setProcessedCount(prev => prev + 1)
    } catch (error) {
      console.error("Failed to apply draft:", error)
      toast.error("应用草稿失败")
    } finally {
      setIsProcessing(false)
    }
  }

  // 拒绝草稿
  const handleReject = async (draftId: number) => {
    setIsProcessing(true)
    try {
      const res = await fetch("/api/drafts/reject", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ids: [draftId] }),
      })
      if (!res.ok) throw new Error("Failed to reject draft")
      
      toast.success("草稿已拒绝")
      setQueue(prev => prev.filter(d => d.id !== draftId))
      setProcessedCount(prev => prev + 1)
    } catch (error) {
      console.error("Failed to reject draft:", error)
      toast.error("拒绝草稿失败")
    } finally {
      setIsProcessing(false)
    }
  }

  // 跳过草稿
  const handleSkip = () => {
    if (currentIndex < queue.length - 1) {
      setCurrentIndex(prev => prev + 1)
    }
  }

  // 上一本
  const handlePrevious = () => {
    if (currentIndex > 0) {
      setCurrentIndex(prev => prev - 1)
    }
  }

  // 下一本
  const handleNext = () => {
    if (currentIndex < queue.length - 1) {
      setCurrentIndex(prev => prev + 1)
    }
  }

  // 键盘快捷键
  useEffect(() => {
    if (!open) return

    const handleKeyDown = (e: KeyboardEvent) => {
      // 如果正在输入，忽略快捷键
      const target = e.target as HTMLElement
      if (target.tagName === "INPUT" || target.tagName === "TEXTAREA") {
        return
      }

      switch (e.key) {
        case "ArrowLeft":
        case "a":
        case "A":
          e.preventDefault()
          handlePrevious()
          break
        case "ArrowRight":
        case "d":
        case "D":
          e.preventDefault()
          handleNext()
          break
        case "Enter":
        case "y":
        case "Y":
          e.preventDefault()
          if (!isProcessing && currentDraft) {
            handleApply(currentDraft.id, undefined)
          }
          break
        case "Delete":
        case "n":
        case "N":
          e.preventDefault()
          if (!isProcessing && currentDraft) {
            handleReject(currentDraft.id)
          }
          break
        case " ":
          e.preventDefault()
          handleSkip()
          break
        case "Escape":
          e.preventDefault()
          onOpenChange(false)
          break
      }
    }

    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [open, currentIndex, isProcessing, queue, onOpenChange])

  const currentDraft = queue[currentIndex]
  const progress = queue.length > 0 ? ((processedCount) / (queue.length + processedCount)) * 100 : 0

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="w-screen h-screen max-w-none sm:max-w-none p-0 gap-0 rounded-none border-0 top-0 left-0 translate-x-0 translate-y-0">
        <DialogTitle className="sr-only">草稿工作台</DialogTitle>
        
        {/* 头部导航栏 */}
        <div className="flex items-center justify-between px-6 py-4 border-b">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onOpenChange(false)}
          >
            <X className="h-5 w-5" />
          </Button>
          
          <div className="flex items-center gap-2">
            <Sparkles className="h-5 w-5 text-primary" />
            <h2 className="text-lg font-semibold">
              草稿工作台 ({currentIndex + 1}/{queue.length})
            </h2>
          </div>

          <div className="w-10" />
        </div>

        {/* 主内容区 */}
        <div className="flex-1 flex items-center justify-center p-6 overflow-hidden">
          {isLoading ? (
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4" />
              <p className="text-muted-foreground">加载草稿队列...</p>
            </div>
          ) : showCompletion ? (
            <div className="text-center space-y-4 animate-in fade-in zoom-in duration-500">
              <div className="text-6xl">✅</div>
              <h3 className="text-2xl font-bold">所有草稿已处理完成</h3>
              <p className="text-muted-foreground">
                已处理 {processedCount} 个草稿
              </p>
            </div>
          ) : queue.length === 0 ? (
            <div className="text-center space-y-4">
              <div className="text-6xl">📭</div>
              <h3 className="text-xl font-semibold">暂无待处理草稿</h3>
              <p className="text-muted-foreground">
                所有草稿都已处理完毕
              </p>
            </div>
          ) : !currentDraft ? (
            <div className="text-center space-y-4">
              <div className="text-6xl">⚠️</div>
              <h3 className="text-xl font-semibold">加载出错</h3>
              <p className="text-muted-foreground">
                无法加载当前草稿，请刷新重试
              </p>
            </div>
          ) : (
            <div className="flex items-center gap-4 w-full max-w-[90vw]">
              {/* 左箭头 */}
              <Button
                variant="outline"
                size="icon"
                className="h-12 w-12 shrink-0"
                onClick={handlePrevious}
                disabled={currentIndex === 0 || isProcessing}
              >
                <ChevronLeft className="h-6 w-6" />
              </Button>

              {/* 当前草稿卡片 */}
              <div className="flex-1 animate-in fade-in slide-in-from-right-5 duration-300">
                <WorkbenchCard
                  draft={currentDraft}
                  onApply={(editedData?: Record<string, unknown>) => handleApply(currentDraft.id, editedData)}
                  onReject={() => handleReject(currentDraft.id)}
                  onSkip={handleSkip}
                  isProcessing={isProcessing}
                />
              </div>

              {/* 右箭头 */}
              <Button
                variant="outline"
                size="icon"
                className="h-12 w-12 shrink-0"
                onClick={handleNext}
                disabled={currentIndex >= queue.length - 1 || isProcessing}
              >
                <ChevronRight className="h-6 w-6" />
              </Button>
            </div>
          )}
        </div>

        {/* 底部进度条 */}
        {!isLoading && !showCompletion && queue.length > 0 && (
          <div className="px-6 py-4 border-t space-y-2">
            <Progress value={progress} className="h-2" />
            <div className="flex items-center justify-between text-sm text-muted-foreground">
              <span>已处理 {processedCount} 个</span>
              <span>剩余 {queue.length} 个</span>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
