"use client"

import { useTranslations } from "next-intl"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { MetadataDraft, DraftFlag } from "@/types/governance"
import { Check, X, AlertTriangle, Info, BookOpen, ExternalLink } from "lucide-react"
import { cn } from "@/lib/utils"
import Link from "next/link"

interface DraftCardProps {
  draft: MetadataDraft
  onApprove?: (id: number, version: number) => void
  onReject?: (id: number, version: number) => void
  onSkip?: (id: number, version: number) => void
  onUpdate?: (id: number, newValue: string, version: number) => void
}

export function DraftCard({ draft, onApprove, onReject, onSkip }: DraftCardProps) {
  const t = useTranslations("governance")
  const common = useTranslations("common")

  const getConfidenceColor = (score: number) => {
    if (score >= 0.8) return "text-green-500 bg-green-500/10 border-green-500/20"
    if (score >= 0.5) return "text-amber-500 bg-amber-500/10 border-amber-500/20"
    return "text-red-500 bg-red-500/10 border-red-500/20"
  }

  const getFlagColor = (flag: DraftFlag) => {
    switch (flag) {
      case "isbn_invalid_checksum":
      case "magazine_suspected":
        return "text-red-500 bg-red-500/10 border-red-500/20"
      case "collection_suspected":
      case "multiple_isbn":
        return "text-amber-500 bg-amber-500/10 border-amber-500/20"
      default:
        return "text-blue-500 bg-blue-500/10 border-blue-500/20"
    }
  }

  const sourceLabels: Record<string, string> = {
    copyright_extract: t("sourceCopyright") || "版权页抽取",
    douban: t("sourceDouban") || "豆瓣搜索",
    manual: t("sourceManual") || "手动编辑",
  }

  const flagLabels: Record<string, string> = {
    collection_suspected: t("flagCollection") || "疑似合辑",
    multiple_isbn: t("flagMultipleIsbn") || "多个ISBN",
    isbn_invalid_checksum: t("flagInvalidChecksum") || "ISBN校验失败",
    title_too_long: t("flagTitleTooLong") || "标题过长",
    multiple_authors: t("flagMultipleAuthors") || "多作者",
    magazine_suspected: t("flagMagazine") || "疑似杂志",
  }

  return (
    <Card className="glass overflow-hidden border-white/10">
      <CardHeader className="pb-3 bg-white/5">
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1">
            <CardTitle className="text-lg font-bold leading-tight line-clamp-1">
              {draft.book_title}
            </CardTitle>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <BookOpen className="w-3 h-3" />
              <span>ID: {draft.book_id}</span>
              <span>•</span>
              <Link
                href={`/detail/${draft.book_id}`}
                className="hover:text-primary flex items-center gap-1 transition-colors"
              >
                {t("details")} <ExternalLink className="w-3 h-3" />
              </Link>
            </div>
          </div>
          <Badge className={cn("font-mono", getConfidenceColor(draft.confidence))}>
            {Math.round(draft.confidence * 100)}%
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="pt-4 space-y-4">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div className="space-y-1">
            <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-semibold">
              {t("oldValue")}
            </span>
            <div className="p-2 rounded bg-red-500/5 border border-red-500/10 text-sm line-through opacity-70 break-all">
              {draft.old_value || "(Empty)"}
            </div>
          </div>
          <div className="space-y-1">
            <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-semibold">
              {t("newValue")}
            </span>
            <div className="p-2 rounded bg-green-500/10 border border-green-500/20 text-sm font-medium text-green-600 dark:text-green-400 break-all">
              {draft.new_value}
            </div>
          </div>
        </div>

        <div className="flex flex-wrap gap-4 text-xs">
          <div className="flex items-center gap-1.5">
            <span className="text-muted-foreground">{t("field")}:</span>
            <Badge variant="outline" className="text-[10px] uppercase font-bold">
              {draft.field}
            </Badge>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="text-muted-foreground">{t("source")}:</span>
            <span className="font-medium">{sourceLabels[draft.source] || draft.source}</span>
          </div>
        </div>

        {draft.flags && draft.flags.length > 0 && (
          <div className="flex flex-wrap gap-2 pt-1">
            {draft.flags.map((flag) => (
              <Badge
                key={flag}
                variant="outline"
                className={cn("text-[10px] py-0 px-2 flex items-center gap-1", getFlagColor(flag))}
              >
                <AlertTriangle className="w-3 h-3" />
                {flagLabels[flag] || flag}
              </Badge>
            ))}
          </div>
        )}

        {draft.confidence_breakdown && (
          <div className="p-2 rounded bg-black/5 dark:bg-white/5 text-[10px] text-muted-foreground grid grid-cols-3 gap-2 border border-white/5">
            <div>ISBN: {(draft.confidence_breakdown.isbn_score * 100).toFixed(0)}</div>
            <div>Ctx: {(draft.confidence_breakdown.context_score * 100).toFixed(0)}</div>
            <div>Pen: {(draft.confidence_breakdown.complexity_penalty * 100).toFixed(0)}</div>
          </div>
        )}
      </CardContent>

      <CardFooter className="bg-white/5 border-t border-white/5 p-3 gap-2">
        <Button
          variant="outline"
          size="sm"
          className="flex-1 text-xs h-8 hover:bg-green-500/10 hover:text-green-600 hover:border-green-500/20"
          onClick={() => onApprove?.(draft.id, draft.version)}
        >
          <Check className="w-3 h-3 mr-1" /> {t("approve")}
        </Button>
        <Button
          variant="outline"
          size="sm"
          className="flex-1 text-xs h-8 hover:bg-red-500/10 hover:text-red-600 hover:border-red-500/20"
          onClick={() => onReject?.(draft.id, draft.version)}
        >
          <X className="w-3 h-3 mr-1" /> {t("reject")}
        </Button>
      </CardFooter>
    </Card>
  )
}
