"use client"

import { useEffect, useState, useCallback } from "react"
import { useTranslations } from "next-intl"
import { governanceService } from "@/lib/services/governance-service"
import { MetadataChangelog, ChangelogFilter } from "@/types/governance"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/table"
import {
  RefreshCw,
  Undo2,
  BookOpen,
  History,
  CheckCircle2,
  AlertCircle
} from "lucide-react"
import { toast } from "sonner"
import { Skeleton } from "@/components/ui/skeleton"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { format } from "date-fns"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export default function MetadataHistoryPage() {
  const t = useTranslations("governance")
  const common = useTranslations("common")

  const [changelogs, setChangelogs] = useState<MetadataChangelog[]>([])
  const [loading, setLoading] = useState(true)
  const [total, setTotal] = useState(0)
  const [filter, setFilter] = useState<ChangelogFilter>({
    limit: 50,
    offset: 0
  })

  const [revertId, setRevertId] = useState<number | null>(null)
  const [revertReason, setRevertReason] = useState("")
  const [isRevertOpen, setIsRevertOpen] = useState(false)

  const loadHistory = useCallback(async () => {
    setLoading(true)
    try {
      const response = await governanceService.listChangelogs(filter)
      setChangelogs(response.changelogs)
      setTotal(response.total)
    } catch (error) {
      toast.error(common("error"))
      console.error(error)
    } finally {
      setLoading(false)
    }
  }, [filter, common])

  useEffect(() => {
    loadHistory()
  }, [loadHistory])

  const handleRevert = (id: number) => {
    setRevertId(id)
    setRevertReason("")
    setIsRevertOpen(true)
  }

  const confirmRevert = async () => {
    if (!revertId) return

    try {
      await governanceService.revertChangelog(revertId, revertReason)
      toast.success(t("applied"))
      setIsRevertOpen(false)
      loadHistory()
    } catch (error) {
      toast.error(common("error"))
    }
  }

  return (
    <div className="container mx-auto py-8 space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">{t("history")}</h1>
          <p className="text-muted-foreground">
            {total} {t("applied")}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={loadHistory} disabled={loading}>
            <RefreshCw className={cn("w-4 h-4 mr-2", loading && "animate-spin")} />
            {common("refresh") || "Refresh"}
          </Button>
        </div>
      </div>

      <div className="rounded-md border glass">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[150px]">{t("date")}</TableHead>
              <TableHead>{t("field")}</TableHead>
              <TableHead className="max-w-[200px]">{t("oldValue")}</TableHead>
              <TableHead className="max-w-[200px]">{t("newValue")}</TableHead>
              <TableHead>{t("status")}</TableHead>
              <TableHead className="text-right">{t("actions")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              Array.from({ length: 10 }).map((_, i) => (
                <TableRow key={`history-skeleton-${i}`}>
                  <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                  <TableCell className="text-right"><Skeleton className="h-8 w-16 ml-auto" /></TableCell>
                </TableRow>
              ))
            ) : changelogs.length > 0 ? (
              changelogs.map((log) => (
                <TableRow key={log.id} className={cn(log.reverted_at && "opacity-50 grayscale bg-muted/5")}>
                  <TableCell className="text-xs font-mono">
                    {format(new Date(log.applied_at), "yyyy-MM-dd HH:mm")}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className="uppercase text-[10px]">{log.field}</Badge>
                  </TableCell>
                  <TableCell className="max-w-[200px] truncate text-xs line-through opacity-70">
                    {log.old_value || "-"}
                  </TableCell>
                  <TableCell className="max-w-[200px] truncate text-xs font-medium text-green-600 dark:text-green-400">
                    {log.new_value}
                  </TableCell>
                  <TableCell>
                    {log.reverted_at ? (
                      <Badge variant="destructive" className="text-[10px]">{t("revert")}</Badge>
                    ) : (
                      <Badge variant="secondary" className="text-[10px] bg-green-500/10 text-green-600 border-green-500/20">
                        {t("applied")}
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    {!log.reverted_at && (
                      <Button variant="ghost" size="sm" onClick={() => handleRevert(log.id)} className="h-8 text-xs">
                        <Undo2 className="w-3 h-3 mr-1" /> {t("revert")}
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={6} className="h-24 text-center">
                  <History className="w-8 h-8 mx-auto text-muted-foreground opacity-20 mb-2" />
                  <p className="text-muted-foreground">No history found</p>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <Dialog open={isRevertOpen} onOpenChange={setIsRevertOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("revert")}</DialogTitle>
            <DialogDescription>
              {t("revertDesc") || "Please provide a reason for reverting this change."}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="reason">{common("reason") || "Reason"}</Label>
              <Input
                id="reason"
                value={revertReason}
                onChange={(e) => setRevertReason(e.target.value)}
                placeholder="Enter reason..."
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsRevertOpen(false)}>
              {common("cancel")}
            </Button>
            <Button onClick={confirmRevert} disabled={!revertReason.trim()}>
              {common("confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
