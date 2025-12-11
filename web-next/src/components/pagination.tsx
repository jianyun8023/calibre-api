"use client"

import { Button } from '@/components/ui/button'
import { ArrowLeft, ArrowRight } from 'lucide-react'
import { cn } from '@/lib/utils'

interface PaginationProps {
  currentPage: number
  totalPages: number
  hasNext: boolean
  hasPrev?: boolean
  onNext: () => void
  onPrev: () => void
  loading?: boolean
  className?: string
}

export function Pagination({
  currentPage,
  totalPages,
  hasNext,
  hasPrev,
  onNext,
  onPrev,
  loading = false,
  className
}: PaginationProps) {
  const isPrevDisabled = !hasPrev || loading
  const isNextDisabled = !hasNext || loading

  return (
    <div className={cn("flex items-center justify-center gap-4 py-8", className)}>
      <Button
        variant="outline"
        onClick={onPrev}
        disabled={isPrevDisabled}
        aria-label="Previous page"
      >
        <ArrowLeft className="h-4 w-4 mr-2" />
        Previous
      </Button>
      <span className="text-sm font-medium">
        Page {currentPage} of {totalPages}
      </span>
      <Button
        variant="outline"
        onClick={onNext}
        disabled={isNextDisabled}
        aria-label="Next page"
      >
        Next
        <ArrowRight className="h-4 w-4 ml-2" />
      </Button>
    </div>
  )
}
