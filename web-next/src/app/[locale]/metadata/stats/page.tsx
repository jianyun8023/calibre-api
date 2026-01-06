"use client"

import { useEffect, useState, useCallback } from "react"
import { useTranslations } from "next-intl"
import { governanceService } from "@/lib/services/governance-service"
import { GovernanceStats } from "@/types/governance"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Progress } from "@/components/ui/progress"
import {
    RefreshCw,
    Clock,
    CheckCircle2,
    XCircle,
    FileCheck2,
    AlertTriangle,
    BookCopy,
    FileSearch,
    Edit3
} from "lucide-react"
import { toast } from "sonner"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

export default function MetadataStatsPage() {
    const t = useTranslations("governance")
    const common = useTranslations("common")

    const [stats, setStats] = useState<GovernanceStats | null>(null)
    const [loading, setLoading] = useState(true)

    const loadStats = useCallback(async () => {
        setLoading(true)
        try {
            const response = await governanceService.getStats()
            setStats(response)
        } catch (error) {
            toast.error(common("error"))
            console.error(error)
        } finally {
            setLoading(false)
        }
    }, [common])

    useEffect(() => {
        loadStats()
    }, [loadStats])

    const sourceLabels: Record<string, { label: string; icon: typeof Clock }> = {
        copyright_extract: { label: t("sourceCopyright") || "版权页抽取", icon: BookCopy },
        douban: { label: t("sourceDouban") || "豆瓣搜索", icon: FileSearch },
        manual: { label: t("sourceManual") || "手动编辑", icon: Edit3 },
    }

    const flagLabels: Record<string, string> = {
        collection_suspected: t("flagCollection") || "疑似合辑",
        multiple_isbn: t("flagMultipleIsbn") || "多个ISBN",
        isbn_invalid_checksum: t("flagInvalidChecksum") || "ISBN校验失败",
        title_too_long: t("flagTitleTooLong") || "标题过长",
        multiple_authors: t("flagMultipleAuthors") || "多作者",
        magazine_suspected: t("flagMagazine") || "疑似杂志",
    }

    const totalDrafts = stats
        ? stats.drafts.pending + stats.drafts.approved + stats.drafts.rejected + stats.drafts.applied
        : 0

    const totalConfidence = stats
        ? stats.confidence_distribution.high + stats.confidence_distribution.medium + stats.confidence_distribution.low
        : 0

    return (
        <div className="container mx-auto py-8 space-y-6">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">{t("stats") || "统计仪表盘"}</h1>
                    <p className="text-muted-foreground">
                        {t("statsDesc") || "元数据治理系统统计概览"}
                    </p>
                </div>

                <div className="flex items-center gap-2">
                    <Button variant="outline" size="sm" onClick={loadStats} disabled={loading}>
                        <RefreshCw className={cn("w-4 h-4 mr-2", loading && "animate-spin")} />
                        {common("refresh") || "刷新"}
                    </Button>
                </div>
            </div>

            {/* Status Cards */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                {loading ? (
                    Array.from({ length: 4 }).map((_, i) => (
                        <Card key={`stat-skeleton-${i}`} className="glass">
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <Skeleton className="h-4 w-24" />
                                <Skeleton className="h-4 w-4 rounded-full" />
                            </CardHeader>
                            <CardContent>
                                <Skeleton className="h-8 w-16" />
                            </CardContent>
                        </Card>
                    ))
                ) : (
                    <>
                        <Card className="glass border-yellow-500/20">
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">{t("pending") || "待审核"}</CardTitle>
                                <Clock className="h-4 w-4 text-yellow-500" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-3xl font-bold text-yellow-500">{stats?.drafts.pending || 0}</div>
                            </CardContent>
                        </Card>

                        <Card className="glass border-green-500/20">
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">{t("approved") || "已批准"}</CardTitle>
                                <CheckCircle2 className="h-4 w-4 text-green-500" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-3xl font-bold text-green-500">{stats?.drafts.approved || 0}</div>
                            </CardContent>
                        </Card>

                        <Card className="glass border-red-500/20">
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">{t("rejected") || "已拒绝"}</CardTitle>
                                <XCircle className="h-4 w-4 text-red-500" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-3xl font-bold text-red-500">{stats?.drafts.rejected || 0}</div>
                            </CardContent>
                        </Card>

                        <Card className="glass border-blue-500/20">
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">{t("applied") || "已应用"}</CardTitle>
                                <FileCheck2 className="h-4 w-4 text-blue-500" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-3xl font-bold text-blue-500">{stats?.drafts.applied || 0}</div>
                            </CardContent>
                        </Card>
                    </>
                )}
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                {/* Confidence Distribution */}
                <Card className="glass">
                    <CardHeader>
                        <CardTitle>{t("confidenceDistribution") || "置信度分布"}</CardTitle>
                        <CardDescription>{t("confidenceDesc") || "按置信度等级划分的草稿数量"}</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        {loading ? (
                            Array.from({ length: 3 }).map((_, i) => (
                                <div key={`conf-skeleton-${i}`} className="space-y-2">
                                    <Skeleton className="h-4 w-32" />
                                    <Skeleton className="h-3 w-full" />
                                </div>
                            ))
                        ) : stats ? (
                            <>
                                <div className="space-y-2">
                                    <div className="flex items-center justify-between text-sm">
                                        <span className="text-green-500 font-medium">{t("highConfidence") || "高置信度 (≥80%)"}</span>
                                        <span className="font-mono">{stats.confidence_distribution.high}</span>
                                    </div>
                                    <Progress
                                        value={totalConfidence > 0 ? (stats.confidence_distribution.high / totalConfidence) * 100 : 0}
                                        className="h-2 bg-green-500/10"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <div className="flex items-center justify-between text-sm">
                                        <span className="text-yellow-500 font-medium">{t("mediumConfidence") || "中置信度 (50-80%)"}</span>
                                        <span className="font-mono">{stats.confidence_distribution.medium}</span>
                                    </div>
                                    <Progress
                                        value={totalConfidence > 0 ? (stats.confidence_distribution.medium / totalConfidence) * 100 : 0}
                                        className="h-2 bg-yellow-500/10"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <div className="flex items-center justify-between text-sm">
                                        <span className="text-red-500 font-medium">{t("lowConfidence") || "低置信度 (<50%)"}</span>
                                        <span className="font-mono">{stats.confidence_distribution.low}</span>
                                    </div>
                                    <Progress
                                        value={totalConfidence > 0 ? (stats.confidence_distribution.low / totalConfidence) * 100 : 0}
                                        className="h-2 bg-red-500/10"
                                    />
                                </div>
                            </>
                        ) : null}
                    </CardContent>
                </Card>

                {/* Source Distribution */}
                <Card className="glass">
                    <CardHeader>
                        <CardTitle>{t("sourceDistribution") || "来源分布"}</CardTitle>
                        <CardDescription>{t("sourceDesc") || "按来源划分的草稿数量"}</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        {loading ? (
                            Array.from({ length: 3 }).map((_, i) => (
                                <div key={`source-skeleton-${i}`} className="flex items-center gap-3">
                                    <Skeleton className="h-8 w-8 rounded" />
                                    <div className="flex-1">
                                        <Skeleton className="h-4 w-24 mb-1" />
                                        <Skeleton className="h-3 w-12" />
                                    </div>
                                </div>
                            ))
                        ) : stats?.by_source ? (
                            Object.entries(stats.by_source).map(([source, count]) => {
                                const sourceInfo = sourceLabels[source] || { label: source, icon: BookCopy }
                                const Icon = sourceInfo.icon

                                return (
                                    <div key={source} className="flex items-center gap-3">
                                        <div className="p-2 rounded-lg bg-primary/10">
                                            <Icon className="h-4 w-4 text-primary" />
                                        </div>
                                        <div className="flex-1">
                                            <div className="text-sm font-medium">{sourceInfo.label}</div>
                                            <div className="text-2xl font-bold">{count}</div>
                                        </div>
                                    </div>
                                )
                            })
                        ) : (
                            <p className="text-muted-foreground text-sm">{t("noData") || "暂无数据"}</p>
                        )}
                    </CardContent>
                </Card>

                {/* Flags Distribution */}
                <Card className="glass md:col-span-2">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <AlertTriangle className="h-5 w-5 text-yellow-500" />
                            {t("flagsDistribution") || "标记分布"}
                        </CardTitle>
                        <CardDescription>{t("flagsDesc") || "需要关注的草稿标记统计"}</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {loading ? (
                            <div className="grid gap-4 sm:grid-cols-2 md:grid-cols-3">
                                {Array.from({ length: 6 }).map((_, i) => (
                                    <div key={`flag-skeleton-${i}`} className="flex items-center justify-between p-3 rounded-lg border">
                                        <Skeleton className="h-4 w-24" />
                                        <Skeleton className="h-6 w-8" />
                                    </div>
                                ))}
                            </div>
                        ) : stats?.flags_count && Object.keys(stats.flags_count).length > 0 ? (
                            <div className="grid gap-4 sm:grid-cols-2 md:grid-cols-3">
                                {Object.entries(stats.flags_count).map(([flag, count]) => (
                                    <div
                                        key={flag}
                                        className="flex items-center justify-between p-3 rounded-lg border bg-yellow-500/5 border-yellow-500/20"
                                    >
                                        <span className="text-sm">{flagLabels[flag] || flag}</span>
                                        <span className="text-lg font-bold text-yellow-600">{count}</span>
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <div className="text-center py-8 text-muted-foreground">
                                <CheckCircle2 className="h-8 w-8 mx-auto mb-2 opacity-30" />
                                <p>{t("noFlags") || "暂无标记"}</p>
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>
        </div>
    )
}
