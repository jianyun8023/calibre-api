"use client"

import { useEffect, useState } from "react"
import { useTranslations } from "next-intl"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { toast } from "sonner"
import { RefreshCcw, CheckCircle, XCircle } from "lucide-react"
import { Pagination } from "@/components/pagination"

interface Draft {
  id: number
  book_id: string
  action: "delete" | "update"
  data: string
  status: string
  created_at: string
}

export default function DraftsPage() {
  const t = useTranslations("drafts")
  const [drafts, setDrafts] = useState<Draft[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set())
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const limit = 10

  const fetchDrafts = async (currentPage: number) => {
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
  }

  useEffect(() => {
    fetchDrafts(1)
  }, [])

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

  const handleApply = async () => {
    if (selectedIds.size === 0) return
    try {
      const res = await fetch("/api/drafts/apply", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ids: Array.from(selectedIds) }),
      })
      if (!res.ok) throw new Error("Failed to apply drafts")
      toast.success(t("applySuccess"))
      setSelectedIds(new Set())
      fetchDrafts(page)
    } catch (err) {
      console.error(err)
      toast.error(t("applyError"))
    }
  }

  const handleReject = async () => {
    if (selectedIds.size === 0) return
    try {
      const res = await fetch("/api/drafts/reject", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ids: Array.from(selectedIds) }),
      })
      if (!res.ok) throw new Error("Failed to reject drafts")
      toast.success(t("rejectSuccess"))
      setSelectedIds(new Set())
      fetchDrafts(page)
    } catch (err) {
      console.error(err)
      toast.error(t("rejectError"))
    }
  }

  const renderDataPreview = (action: string, dataStr: string) => {
    if (action === "delete") return <span className="text-muted-foreground">This book will be deleted</span>
    try {
      const parsed = JSON.parse(dataStr)

      // Mask potentially sensitive keys
      const sensitiveKeys = ['token', 'password', 'secret', 'key', 'auth']
      const maskData = (obj: any): any => {
        if (typeof obj !== 'object' || obj === null) return obj;
        if (Array.isArray(obj)) return obj.map(maskData);

        const newObj: any = {};
        for (const [k, v] of Object.entries(obj)) {
          if (sensitiveKeys.some(sk => k.toLowerCase().includes(sk))) {
            newObj[k] = '********';
          } else {
            newObj[k] = maskData(v);
          }
        }
        return newObj;
      }

      const safeParsed = maskData(parsed);

      return (
        <pre className="text-sm bg-muted p-2 rounded-md overflow-x-auto">
          {JSON.stringify(safeParsed, null, 2)}
        </pre>
      )
    } catch {
      return <span>{dataStr}</span>
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
          onClick={handleApply}
          className="bg-green-600 hover:bg-green-700"
        >
          <CheckCircle className="mr-2 h-4 w-4" />
          {t("apply")}
        </Button>
        <Button
          variant="destructive"
          size="sm"
          disabled={selectedIds.size === 0}
          onClick={handleReject}
        >
          <XCircle className="mr-2 h-4 w-4" />
          {t("reject")}
        </Button>
      </div>

      <div className="space-y-4">
        {loading ? (
          Array.from({ length: 3 }).map((_, i) => (
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
            <Card key={draft.id} className={selectedIds.has(draft.id) ? "border-primary" : ""}>
              <CardHeader className="py-4 flex flex-row items-center space-y-0 gap-4">
                <Checkbox
                  checked={selectedIds.has(draft.id)}
                  onCheckedChange={(checked) => handleSelect(draft.id, !!checked)}
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
              </CardHeader>
              <CardContent>
                {renderDataPreview(draft.action, draft.data)}
              </CardContent>
            </Card>
          ))
        )}

        {drafts.length > 0 && totalPages > 1 && (
          <div className="mt-8">
            <Pagination
              currentPage={page}
              totalPages={totalPages}
              onPageChange={(p) => fetchDrafts(p)}
            />
          </div>
        )}
      </div>
    </div>
  )
}
