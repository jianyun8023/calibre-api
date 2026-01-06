"use client"

import { useEffect, useState, useCallback } from "react"
import { useTranslations } from "next-intl"
import { governanceService } from "@/lib/services/governance-service"
import { MetadataDraft, DraftFilter } from "@/types/governance"
import { DraftCard } from "@/components/governance/draft-card"
import { Button } from "@/components/ui/button"
import { RefreshCw, CheckCircle2, Check } from "lucide-react"
import { toast } from "sonner"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

export default function MetadataQueuePage() {
  const t = useTranslations("governance")
  const common = useTranslations("common")
  
  const [drafts, setDrafts] = useState<MetadataDraft[]>([])
  const [loading, setLoading] = useState(true)
  const [total, setTotal] = useState(0)
  const [filter, setFilter] = useState<DraftFilter>({
    status: 'pending',
    limit: 20,
    offset: 0
  })

  const loadDrafts = useCallback(async () => {
    setLoading(true)
    try {
      const response = await governanceService.listDrafts(filter)
      setDrafts(response.drafts)
      setTotal(response.total)
    } catch (error) {
      toast.error(common("error"))
      console.error(error)
    } finally {
      setLoading(false)
    }
  }, [filter, common])

  useEffect(() => {
    loadDrafts()
  }, [loadDrafts])

  const handleApprove = async (id: number, version: number) => {
    try {
      await governanceService.approveDraft(id, version)
      toast.success(t("approved"))
      loadDrafts()
    } catch (error) {
      toast.error(common("error"))
    }
  }

  const handleReject = async (id: number, version: number) => {
    try {
      await governanceService.rejectDraft(id, version)
      toast.success(t("rejected"))
      loadDrafts()
    } catch (error) {
      toast.error(common("error"))
    }
  }

  const handleApplyAll = async () => {
    try {
      const result = await governanceService.applyAll()
      toast.success(`${t("applied")}: ${result.success.length}`)
      loadDrafts()
    } catch (error) {
      toast.error(common("error"))
    }
  }

  return (
    <div className="container mx-auto py-8 space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">{t("queue")}</h1>
          <p className="text-muted-foreground">
            {total} {t("pending")}
          </p>
        </div>
        
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={loadDrafts} disabled={loading}>
            <RefreshCw className={cn("w-4 h-4 mr-2", loading && "animate-spin")} />
            {common("refresh") || "Refresh"}
          </Button>
          <Button variant="default" size="sm" onClick={handleApplyAll} className="bg-green-600 hover:bg-green-700">
            <CheckCircle2 className="w-4 h-4 mr-2" />
            {t("applyAll")}
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {loading ? (
          [1, 2, 3, 4, 5, 6].map((item) => (
            <Skeleton key={`queue-skeleton-${item}`} className="h-[300px] w-full rounded-xl" />
          ))
        ) : drafts.length > 0 ? (
          drafts.map((draft) => (
            <DraftCard 
              key={`draft-${draft.id}`} 
              draft={draft} 
              onApprove={handleApprove}
              onReject={handleReject}
            />
          ))
        ) : (
          <div className="col-span-full py-20 text-center space-y-4">
            <Check className="w-12 h-12 mx-auto text-muted-foreground opacity-20" />
            <p className="text-xl font-medium text-muted-foreground">{t("noDrafts")}</p>
          </div>
        )}
      </div>
    </div>
  )
}
