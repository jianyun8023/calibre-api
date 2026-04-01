"use client"

import { useState } from "react"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Check, X, Edit2, Plus, Trash2 } from "lucide-react"

interface InlineEditFieldProps {
  label: string
  oldValue: unknown
  newValue: unknown
  onUpdate: (value: unknown) => void
  onRemove?: () => void
  type?: "text" | "array" | "textarea"
}

export function InlineEditField({
  label,
  oldValue,
  newValue,
  onUpdate,
  onRemove,
  type = "text",
}: InlineEditFieldProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [editValue, setEditValue] = useState(newValue)
  const [tempArrayValue, setTempArrayValue] = useState("")

  const formatValue = (v: unknown) => {
    if (v === undefined || v === null || v === "") {
      return <span className="text-muted-foreground italic">empty</span>
    }
    if (Array.isArray(v)) {
      if (v.length === 0) {
        return <span className="text-muted-foreground italic">empty</span>
      }
      return v.join(", ")
    }
    return String(v)
  }

  const handleSave = () => {
    onUpdate(editValue)
    setIsEditing(false)
  }

  const handleCancel = () => {
    setEditValue(newValue)
    setIsEditing(false)
  }

  const handleAddArrayItem = () => {
    if (!tempArrayValue.trim()) return
    const currentArray = Array.isArray(editValue) ? editValue : []
    setEditValue([...currentArray, tempArrayValue.trim()])
    setTempArrayValue("")
  }

  const handleRemoveArrayItem = (index: number) => {
    if (!Array.isArray(editValue)) return
    const newArray = editValue.filter((_, i) => i !== index)
    setEditValue(newArray)
  }

  if (type === "array") {
    const oldArr = Array.isArray(oldValue) ? oldValue : []
    const newArr = Array.isArray(editValue) ? editValue : []

    return (
      <div className="space-y-2">
        <div className="font-medium text-sm capitalize flex items-center justify-between">
          <span>{label}</span>
          <div className="flex items-center gap-1">
            {!isEditing && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setIsEditing(true)}
                className="h-6 px-2"
              >
                <Edit2 className="h-3 w-3 mr-1" />
                编辑
              </Button>
            )}
            {onRemove && (
              <Button
                variant="ghost"
                size="sm"
                onClick={onRemove}
                className="h-6 px-2 text-destructive hover:text-destructive hover:bg-destructive/10"
                title="移除此字段"
              >
                <Trash2 className="h-3 w-3" />
              </Button>
            )}
          </div>
        </div>

        {!isEditing ? (
          <div className="space-y-2">
            {oldArr.length > 0 && (
              <div className="text-xs text-muted-foreground">
                旧值: {oldArr.join(", ")}
              </div>
            )}
            <div className="flex flex-wrap gap-1">
              {newArr.map((item, idx) => (
                <Badge key={idx} variant="secondary">
                  {String(item)}
                </Badge>
              ))}
              {newArr.length === 0 && (
                <span className="text-muted-foreground italic text-sm">empty</span>
              )}
            </div>
          </div>
        ) : (
          <div className="space-y-2 p-3 border rounded-lg bg-yellow-50 dark:bg-yellow-950/20 border-yellow-200 dark:border-yellow-800">
            <div className="flex flex-wrap gap-1">
              {newArr.map((item, idx) => (
                <Badge
                  key={idx}
                  variant="secondary"
                  className="cursor-pointer hover:bg-destructive hover:text-destructive-foreground"
                  onClick={() => handleRemoveArrayItem(idx)}
                >
                  {String(item)}
                  <X className="h-3 w-3 ml-1" />
                </Badge>
              ))}
            </div>
            <div className="flex gap-2">
              <Input
                value={tempArrayValue}
                onChange={(e) => setTempArrayValue(e.target.value)}
                placeholder="添加新项"
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault()
                    handleAddArrayItem()
                  }
                }}
                className="flex-1"
              />
              <Button
                size="sm"
                variant="outline"
                onClick={handleAddArrayItem}
              >
                <Plus className="h-4 w-4" />
              </Button>
            </div>
            <div className="flex gap-2 justify-end">
              <Button size="sm" variant="outline" onClick={handleCancel}>
                <X className="h-3 w-3 mr-1" />
                取消
              </Button>
              <Button size="sm" onClick={handleSave}>
                <Check className="h-3 w-3 mr-1" />
                保存
              </Button>
            </div>
          </div>
        )}
      </div>
    )
  }

  // 文本类型
  return (
    <div className="space-y-2">
      <div className="font-medium text-sm capitalize flex items-center justify-between">
        <span>{label}</span>
        {onRemove && (
          <Button
            variant="ghost"
            size="sm"
            onClick={onRemove}
            className="h-6 px-2 text-destructive hover:text-destructive hover:bg-destructive/10"
            title="移除此字段"
          >
            <Trash2 className="h-3 w-3" />
          </Button>
        )}
      </div>
      <div className="grid grid-cols-2 gap-3 text-sm">
        <div className="space-y-1">
          <div className="text-xs text-muted-foreground">旧值</div>
          <div className="p-2 rounded bg-muted/30">
            {formatValue(oldValue)}
          </div>
        </div>
        <div className="space-y-1">
          <div className="text-xs text-muted-foreground flex items-center justify-between">
            <span>→ 新值</span>
            {!isEditing && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setIsEditing(true)}
                className="h-5 px-1"
              >
                <Edit2 className="h-3 w-3" />
              </Button>
            )}
          </div>
          {!isEditing ? (
            <div className="p-2 rounded bg-blue-50 dark:bg-blue-950/20 font-medium">
              {formatValue(newValue)}
            </div>
          ) : (
            <div className="space-y-2 p-2 rounded bg-yellow-50 dark:bg-yellow-950/20 border border-yellow-200 dark:border-yellow-800">
              {type === "textarea" ? (
                <Textarea
                  value={String(editValue || "")}
                  onChange={(e) => setEditValue(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Escape") {
                      e.preventDefault()
                      handleCancel()
                    }
                    // Ctrl+Enter 或 Cmd+Enter 保存
                    if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
                      e.preventDefault()
                      handleSave()
                    }
                  }}
                  rows={4}
                  autoFocus
                  className="resize-y"
                />
              ) : (
                <Input
                  value={String(editValue || "")}
                  onChange={(e) => setEditValue(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      e.preventDefault()
                      handleSave()
                    } else if (e.key === "Escape") {
                      e.preventDefault()
                      handleCancel()
                    }
                  }}
                  autoFocus
                />
              )}
              <div className="flex gap-2 justify-end">
                <Button size="sm" variant="outline" onClick={handleCancel}>
                  <X className="h-3 w-3 mr-1" />
                  取消
                </Button>
                <Button size="sm" onClick={handleSave}>
                  <Check className="h-3 w-3 mr-1" />
                  保存
                  {type === "textarea" && (
                    <span className="text-xs ml-1 opacity-70">(Ctrl+Enter)</span>
                  )}
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
