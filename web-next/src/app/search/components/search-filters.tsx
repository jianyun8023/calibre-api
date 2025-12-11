"use client"

import { useState, useEffect } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Slider } from "@/components/ui/slider"
import { Checkbox } from "@/components/ui/checkbox"
import { 
  ChevronDown, 
  ChevronUp, 
  X, 
  Filter,
  Building2,
  User,
  Tag,
  Star,
  Calendar,
  SortAsc
} from "lucide-react"
import { fetchPublishers } from "@/lib/api/books"

export interface SearchFiltersState {
  publishers: string[]
  authors: string[]
  tags: string[]
  ratingRange: [number, number]
  yearRange: [number, number]
  sortBy: string
  sortOrder: 'asc' | 'desc'
}

interface SearchFiltersProps {
  filters: SearchFiltersState
  onFiltersChange: (filters: SearchFiltersState) => void
  availablePublishers?: string[]
  availableAuthors?: string[]
  availableTags?: string[]
  isOpen: boolean
  onToggle: () => void
}

export function SearchFilters({
  filters,
  onFiltersChange,
  availablePublishers = [],
  availableAuthors = [],
  availableTags = [],
  isOpen,
  onToggle
}: SearchFiltersProps) {
  const [publishers, setPublishers] = useState<string[]>([])
  const [publisherSearch, setPublisherSearch] = useState("")
  const [authorSearch, setAuthorSearch] = useState("")
  const [tagSearch, setTagSearch] = useState("")

  // Load publishers from API
  useEffect(() => {
    const loadPublishers = async () => {
      try {
        const data = await fetchPublishers()
        setPublishers(data || [])
      } catch (error) {
        console.error("Failed to load publishers:", error)
      }
    }
    loadPublishers()
  }, [])

  const updateFilters = (updates: Partial<SearchFiltersState>) => {
    onFiltersChange({ ...filters, ...updates })
  }

  const clearAllFilters = () => {
    onFiltersChange({
      publishers: [],
      authors: [],
      tags: [],
      ratingRange: [0, 5],
      yearRange: [1900, new Date().getFullYear()],
      sortBy: 'relevance',
      sortOrder: 'desc'
    })
  }

  const getActiveFiltersCount = () => {
    let count = 0
    if (filters.publishers.length > 0) count++
    if (filters.authors.length > 0) count++
    if (filters.tags.length > 0) count++
    if (filters.ratingRange[0] > 0 || filters.ratingRange[1] < 5) count++
    if (filters.yearRange[0] > 1900 || filters.yearRange[1] < new Date().getFullYear()) count++
    return count
  }

  const filteredPublishers = publishers.filter(pub => 
    pub.toLowerCase().includes(publisherSearch.toLowerCase())
  ).slice(0, 10)

  const filteredAuthors = availableAuthors.filter(author => 
    author.toLowerCase().includes(authorSearch.toLowerCase())
  ).slice(0, 10)

  const filteredTags = availableTags.filter(tag => 
    tag.toLowerCase().includes(tagSearch.toLowerCase())
  ).slice(0, 10)

  return (
    <Card className="w-80 h-fit glass">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-lg">
            <Filter className="w-5 h-5" />
            高级过滤器
            {getActiveFiltersCount() > 0 && (
              <Badge variant="secondary" className="ml-2">
                {getActiveFiltersCount()}
              </Badge>
            )}
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
        {getActiveFiltersCount() > 0 && (
          <Button
            variant="outline"
            size="sm"
            onClick={clearAllFilters}
            className="w-full mt-2"
          >
            <X className="w-4 h-4 mr-2" />
            清除所有过滤器
          </Button>
        )}
      </CardHeader>

      <Collapsible open={isOpen} onOpenChange={onToggle}>
        <CollapsibleContent>
          <CardContent className="space-y-6">
            {/* 出版社过滤器 */}
            <div className="space-y-3">
              <Label className="flex items-center gap-2 text-sm font-medium">
                <Building2 className="w-4 h-4" />
                出版社
              </Label>
              <Input
                placeholder="搜索出版社..."
                value={publisherSearch}
                onChange={(e) => setPublisherSearch(e.target.value)}
                className="h-8"
              />
              <div className="max-h-32 overflow-y-auto space-y-2">
                {filteredPublishers.map((publisher) => (
                  <div key={publisher} className="flex items-center space-x-2">
                    <Checkbox
                      id={`publisher-${publisher}`}
                      checked={filters.publishers.includes(publisher)}
                      onCheckedChange={(checked) => {
                        if (checked) {
                          updateFilters({
                            publishers: [...filters.publishers, publisher]
                          })
                        } else {
                          updateFilters({
                            publishers: filters.publishers.filter(p => p !== publisher)
                          })
                        }
                      }}
                    />
                    <Label
                      htmlFor={`publisher-${publisher}`}
                      className="text-sm cursor-pointer truncate flex-1"
                    >
                      {publisher}
                    </Label>
                  </div>
                ))}
              </div>
              {filters.publishers.length > 0 && (
                <div className="flex flex-wrap gap-1">
                  {filters.publishers.map((publisher) => (
                    <Badge key={publisher} variant="secondary" className="text-xs">
                      {publisher}
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-4 w-4 p-0 ml-1 hover:bg-transparent"
                        onClick={() => updateFilters({
                          publishers: filters.publishers.filter(p => p !== publisher)
                        })}
                      >
                        <X className="w-3 h-3" />
                      </Button>
                    </Badge>
                  ))}
                </div>
              )}
            </div>

            <Separator />

            {/* 作者过滤器 */}
            <div className="space-y-3">
              <Label className="flex items-center gap-2 text-sm font-medium">
                <User className="w-4 h-4" />
                作者
              </Label>
              <Input
                placeholder="搜索作者..."
                value={authorSearch}
                onChange={(e) => setAuthorSearch(e.target.value)}
                className="h-8"
              />
              <div className="max-h-32 overflow-y-auto space-y-2">
                {filteredAuthors.map((author) => (
                  <div key={author} className="flex items-center space-x-2">
                    <Checkbox
                      id={`author-${author}`}
                      checked={filters.authors.includes(author)}
                      onCheckedChange={(checked) => {
                        if (checked) {
                          updateFilters({
                            authors: [...filters.authors, author]
                          })
                        } else {
                          updateFilters({
                            authors: filters.authors.filter(a => a !== author)
                          })
                        }
                      }}
                    />
                    <Label
                      htmlFor={`author-${author}`}
                      className="text-sm cursor-pointer truncate flex-1"
                    >
                      {author}
                    </Label>
                  </div>
                ))}
              </div>
              {filters.authors.length > 0 && (
                <div className="flex flex-wrap gap-1">
                  {filters.authors.map((author) => (
                    <Badge key={author} variant="secondary" className="text-xs">
                      {author}
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-4 w-4 p-0 ml-1 hover:bg-transparent"
                        onClick={() => updateFilters({
                          authors: filters.authors.filter(a => a !== author)
                        })}
                      >
                        <X className="w-3 h-3" />
                      </Button>
                    </Badge>
                  ))}
                </div>
              )}
            </div>

            <Separator />

            {/* 标签过滤器 */}
            <div className="space-y-3">
              <Label className="flex items-center gap-2 text-sm font-medium">
                <Tag className="w-4 h-4" />
                标签
              </Label>
              <Input
                placeholder="搜索标签..."
                value={tagSearch}
                onChange={(e) => setTagSearch(e.target.value)}
                className="h-8"
              />
              <div className="max-h-32 overflow-y-auto space-y-2">
                {filteredTags.map((tag) => (
                  <div key={tag} className="flex items-center space-x-2">
                    <Checkbox
                      id={`tag-${tag}`}
                      checked={filters.tags.includes(tag)}
                      onCheckedChange={(checked) => {
                        if (checked) {
                          updateFilters({
                            tags: [...filters.tags, tag]
                          })
                        } else {
                          updateFilters({
                            tags: filters.tags.filter(t => t !== tag)
                          })
                        }
                      }}
                    />
                    <Label
                      htmlFor={`tag-${tag}`}
                      className="text-sm cursor-pointer truncate flex-1"
                    >
                      {tag}
                    </Label>
                  </div>
                ))}
              </div>
              {filters.tags.length > 0 && (
                <div className="flex flex-wrap gap-1">
                  {filters.tags.map((tag) => (
                    <Badge key={tag} variant="secondary" className="text-xs">
                      {tag}
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-4 w-4 p-0 ml-1 hover:bg-transparent"
                        onClick={() => updateFilters({
                          tags: filters.tags.filter(t => t !== tag)
                        })}
                      >
                        <X className="w-3 h-3" />
                      </Button>
                    </Badge>
                  ))}
                </div>
              )}
            </div>

            <Separator />

            {/* 评分过滤器 */}
            <div className="space-y-3">
              <Label className="flex items-center gap-2 text-sm font-medium">
                <Star className="w-4 h-4" />
                评分范围
              </Label>
              <div className="px-2">
                <Slider
                  value={filters.ratingRange}
                  onValueChange={(value) => updateFilters({ ratingRange: value as [number, number] })}
                  max={5}
                  min={0}
                  step={0.5}
                  className="w-full"
                />
                <div className="flex justify-between text-xs text-muted-foreground mt-1">
                  <span>{filters.ratingRange[0]} 星</span>
                  <span>{filters.ratingRange[1]} 星</span>
                </div>
              </div>
            </div>

            <Separator />

            {/* 年份过滤器 */}
            <div className="space-y-3">
              <Label className="flex items-center gap-2 text-sm font-medium">
                <Calendar className="w-4 h-4" />
                出版年份
              </Label>
              <div className="px-2">
                <Slider
                  value={filters.yearRange}
                  onValueChange={(value) => updateFilters({ yearRange: value as [number, number] })}
                  max={new Date().getFullYear()}
                  min={1900}
                  step={1}
                  className="w-full"
                />
                <div className="flex justify-between text-xs text-muted-foreground mt-1">
                  <span>{filters.yearRange[0]}</span>
                  <span>{filters.yearRange[1]}</span>
                </div>
              </div>
            </div>

            <Separator />

            {/* 排序选项 */}
            <div className="space-y-3">
              <Label className="flex items-center gap-2 text-sm font-medium">
                <SortAsc className="w-4 h-4" />
                排序方式
              </Label>
              <Select
                value={`${filters.sortBy}-${filters.sortOrder}`}
                onValueChange={(value) => {
                  const [sortBy, sortOrder] = value.split('-')
                  updateFilters({ sortBy, sortOrder: sortOrder as 'asc' | 'desc' })
                }}
              >
                <SelectTrigger className="h-8">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="relevance-desc">相关性（默认）</SelectItem>
                  <SelectItem value="pubdate-desc">出版时间（新→旧）</SelectItem>
                  <SelectItem value="pubdate-asc">出版时间（旧→新）</SelectItem>
                  <SelectItem value="rating-desc">评分（高→低）</SelectItem>
                  <SelectItem value="rating-asc">评分（低→高）</SelectItem>
                  <SelectItem value="title-asc">标题（A→Z）</SelectItem>
                  <SelectItem value="title-desc">标题（Z→A）</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </CollapsibleContent>
      </Collapsible>
    </Card>
  )
}