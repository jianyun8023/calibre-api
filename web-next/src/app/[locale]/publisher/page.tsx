"use client"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Pagination } from "@/components/pagination"
import { bookService } from "@/lib/services"
import { useEffect, useState, useMemo, useCallback } from "react"
import { Search, Building2 } from "lucide-react"
import Link from "next/link"
import { toast } from "sonner"

const ITEMS_PER_PAGE = 50

export default function PublishersPage() {
  const [publishers, setPublishers] = useState<string[]>([])
  const [filteredPublishers, setFilteredPublishers] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState("")
  const [currentPage, setCurrentPage] = useState(1)

  const fetchPublishers = useCallback(async () => {
    try {
      const data = await bookService.getPublishers()
      // Sort alphabetically
      const sorted = data.sort((a, b) => a.localeCompare(b))
      setPublishers(sorted)
      setFilteredPublishers(sorted)
    } catch (error) {
      console.error("Failed to fetch publishers:", error)
      toast.error("Failed to load publishers", {
        description: error instanceof Error ? error.message : "Unknown error"
      })
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchPublishers()
  }, [fetchPublishers])

  useEffect(() => {
    if (searchQuery.trim() === "") {
      setFilteredPublishers(publishers)
    } else {
      const query = searchQuery.toLowerCase()
      setFilteredPublishers(
        publishers.filter((p) => p.toLowerCase().includes(query))
      )
    }
    // Reset to first page when search changes
    setCurrentPage(1)
  }, [searchQuery, publishers])

  // Calculate pagination
  const totalPages = useMemo(() => {
    return Math.ceil(filteredPublishers.length / ITEMS_PER_PAGE)
  }, [filteredPublishers.length])

  const paginatedPublishers = useMemo(() => {
    const startIndex = (currentPage - 1) * ITEMS_PER_PAGE
    const endIndex = startIndex + ITEMS_PER_PAGE
    return filteredPublishers.slice(startIndex, endIndex)
  }, [filteredPublishers, currentPage])

  const handleNextPage = () => {
    if (currentPage < totalPages) {
      setCurrentPage(prev => prev + 1)
      window.scrollTo({ top: 0, behavior: 'smooth' })
    }
  }

  const handlePrevPage = () => {
    if (currentPage > 1) {
      setCurrentPage(prev => prev - 1)
      window.scrollTo({ top: 0, behavior: 'smooth' })
    }
  }

  return (
    <div className="container max-w-6xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Publishers</h1>
        <p className="text-muted-foreground">
          Browse books by publisher and view statistics
        </p>
      </div>

      {/* Statistics Card */}
      <div className="grid gap-4 md:grid-cols-1 mb-8">
        <Card className="glass">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Publishers</CardTitle>
            <Building2 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{publishers.length}</div>
            <p className="text-xs text-muted-foreground">
              Unique publishers in library
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Search and List */}
      <Card className="glass">
        <CardHeader>
          <CardTitle>All Publishers</CardTitle>
          <CardDescription>
            Search and filter through all publishers
          </CardDescription>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              type="text"
              placeholder="Search publishers..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9"
            />
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="py-8 text-center text-muted-foreground">
              Loading publishers...
            </div>
          ) : filteredPublishers.length === 0 ? (
            <div className="py-8 text-center text-muted-foreground">
              {searchQuery ? "No publishers found matching your search" : "No publishers found"}
            </div>
          ) : (
            <>
              <div className="grid gap-2">
                {paginatedPublishers.map((publisher, index) => (
                  <Link
                    key={`publisher-${index}-${publisher || 'unknown'}`}
                    href={`/search?publisher=${encodeURIComponent(publisher)}`}
                    className="flex items-center gap-3 p-3 rounded-lg hover:bg-accent transition-colors"
                  >
                    <Building2 className="w-4 h-4 text-muted-foreground" />
                    <span className="font-medium">{publisher || "Unknown"}</span>
                  </Link>
                ))}
              </div>

              {/* Pagination */}
              {totalPages > 1 && (
                <Pagination
                  currentPage={currentPage}
                  totalPages={totalPages}
                  hasNext={currentPage < totalPages}
                  hasPrev={currentPage > 1}
                  onNext={handleNextPage}
                  onPrev={handlePrevPage}
                />
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Show count */}
      {filteredPublishers.length > 0 && (
        <div className="mt-4 text-center text-sm text-muted-foreground">
          {searchQuery ? (
            <>Showing {filteredPublishers.length} of {publishers.length} publishers</>
          ) : (
            <>
              Showing {((currentPage - 1) * ITEMS_PER_PAGE) + 1} - {Math.min(currentPage * ITEMS_PER_PAGE, filteredPublishers.length)} of {filteredPublishers.length} publishers
            </>
          )}
        </div>
      )}
    </div>
  )
}

