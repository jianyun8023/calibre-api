"use client"

import { useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"
import { 
  History, 
  TrendingUp, 
  X, 
  ChevronDown, 
  ChevronUp,
  Clock,
  Search
} from "lucide-react"
import { useSearchHistory, type SearchHistoryItem } from "../hooks/use-search-history"

interface SearchHistoryProps {
  onSearchSelect: (query: string, mode: string) => void
  isOpen: boolean
  onToggle: () => void
}

export function SearchHistory({ onSearchSelect, isOpen, onToggle }: SearchHistoryProps) {
  const {
    history,
    removeFromHistory,
    clearHistory,
    getRecentSearches,
    getPopularSearches
  } = useSearchHistory()

  const recentSearches = getRecentSearches()
  const popularSearches = getPopularSearches()

  const formatTimestamp = (timestamp: number) => {
    const now = Date.now()
    const diff = now - timestamp
    
    if (diff < 60000) return "刚刚"
    if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
    if (diff < 604800000) return `${Math.floor(diff / 86400000)}天前`
    return new Date(timestamp).toLocaleDateString()
  }

  const getModeLabel = (mode: string) => {
    switch (mode) {
      case 'keyword': return '关键词'
      case 'semantic': return '语义'
      case 'hybrid': return '混合'
      default: return mode
    }
  }

  const getModeColor = (mode: string) => {
    switch (mode) {
      case 'keyword': return 'default'
      case 'semantic': return 'secondary'
      case 'hybrid': return 'outline'
      default: return 'default'
    }
  }

  if (history.length === 0) {
    return null
  }

  return (
    <Card className="w-80 h-fit glass">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-lg">
            <History className="w-5 h-5" />
            搜索历史
          </CardTitle>
          <Button
            variant="ghost"
            size="sm"
            onClick={onToggle}
            className="p-1 h-8 w-8"
          >
            {isOpen ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </Button>
        </div>
        {history.length > 0 && (
          <Button
            variant="outline"
            size="sm"
            onClick={clearHistory}
            className="w-full mt-2"
          >
            <X className="w-4 h-4 mr-2" />
            清除历史记录
          </Button>
        )}
      </CardHeader>

      <Collapsible open={isOpen} onOpenChange={onToggle}>
        <CollapsibleContent>
          <CardContent className="space-y-6">
            {/* Recent Searches */}
            {recentSearches.length > 0 && (
              <div className="space-y-3">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Clock className="w-4 h-4" />
                  最近搜索
                </div>
                <div className="space-y-2">
                  {recentSearches.map((item) => (
                    <div
                      key={item.id}
                      className="group flex items-center justify-between p-2 rounded-lg hover:bg-accent/50 transition-colors cursor-pointer"
                      onClick={() => onSearchSelect(item.query, item.mode)}
                    >
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <Search className="w-3 h-3 text-muted-foreground shrink-0" />
                          <span className="text-sm font-medium truncate">
                            {item.query}
                          </span>
                        </div>
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                          <Badge 
                            variant={getModeColor(item.mode) as any} 
                            className="text-xs px-1 py-0 h-4"
                          >
                            {getModeLabel(item.mode)}
                          </Badge>
                          <span>{formatTimestamp(item.timestamp)}</span>
                          {item.resultsCount !== undefined && (
                            <span>{item.resultsCount} 本书</span>
                          )}
                        </div>
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="opacity-0 group-hover:opacity-100 h-6 w-6 p-0 shrink-0"
                        onClick={(e) => {
                          e.stopPropagation()
                          removeFromHistory(item.id)
                        }}
                      >
                        <X className="w-3 h-3" />
                      </Button>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Popular Searches */}
            {popularSearches.length > 0 && recentSearches.length > 0 && (
              <Separator />
            )}

            {popularSearches.length > 0 && (
              <div className="space-y-3">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <TrendingUp className="w-4 h-4" />
                  热门搜索
                </div>
                <div className="flex flex-wrap gap-2">
                  {popularSearches.map(({ query, count }) => (
                    <Button
                      key={query}
                      variant="outline"
                      size="sm"
                      className="h-7 text-xs"
                      onClick={() => onSearchSelect(query, 'hybrid')}
                    >
                      {query}
                      <Badge variant="secondary" className="ml-1 text-xs px-1 py-0 h-4">
                        {count}
                      </Badge>
                    </Button>
                  ))}
                </div>
              </div>
            )}

            {/* All History */}
            {history.length > 5 && (
              <>
                <Separator />
                <div className="space-y-3">
                  <div className="text-sm font-medium text-muted-foreground">
                    全部历史 ({history.length})
                  </div>
                  <div className="max-h-48 overflow-y-auto space-y-1">
                    {history.slice(5).map((item) => (
                      <div
                        key={item.id}
                        className="group flex items-center justify-between p-1 rounded hover:bg-accent/50 transition-colors cursor-pointer"
                        onClick={() => onSearchSelect(item.query, item.mode)}
                      >
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="text-sm truncate">{item.query}</span>
                            <Badge 
                              variant={getModeColor(item.mode) as any} 
                              className="text-xs px-1 py-0 h-4 shrink-0"
                            >
                              {getModeLabel(item.mode)}
                            </Badge>
                          </div>
                          <div className="text-xs text-muted-foreground">
                            {formatTimestamp(item.timestamp)}
                          </div>
                        </div>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="opacity-0 group-hover:opacity-100 h-5 w-5 p-0 shrink-0"
                          onClick={(e) => {
                            e.stopPropagation()
                            removeFromHistory(item.id)
                          }}
                        >
                          <X className="w-3 h-3" />
                        </Button>
                      </div>
                    ))}
                  </div>
                </div>
              </>
            )}
          </CardContent>
        </CollapsibleContent>
      </Collapsible>
    </Card>
  )
}