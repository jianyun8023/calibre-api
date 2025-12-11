"use client"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { apiRequest } from "@/lib/api-client"
import { useEffect, useState } from "react"
import { Search, Building2, BookOpen, TrendingUp, BarChart } from "lucide-react"
import Link from "next/link"

interface Publisher {
  name: string
  count: number
}

export default function PublishersPage() {
  const [publishers, setPublishers] = useState<Publisher[]>([])
  const [filteredPublishers, setFilteredPublishers] = useState<Publisher[]>([])
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState("")

  useEffect(() => {
    fetchPublishers()
  }, [])

  useEffect(() => {
    if (searchQuery.trim() === "") {
      setFilteredPublishers(publishers)
    } else {
      const query = searchQuery.toLowerCase()
      setFilteredPublishers(
        publishers.filter((p) => p.name.toLowerCase().includes(query))
      )
    }
  }, [searchQuery, publishers])

  const fetchPublishers = async () => {
    try {
      const data = await apiRequest<Publisher[]>("/api/publisher")
      // Sort by book count (descending)
      const sorted = data.sort((a, b) => b.count - a.count)
      setPublishers(sorted)
      setFilteredPublishers(sorted)
    } catch (error) {
      console.error("Failed to fetch publishers:", error)
    } finally {
      setLoading(false)
    }
  }

  const totalBooks = publishers.reduce((sum, p) => sum + p.count, 0)
  const topPublishers = filteredPublishers.slice(0, 5)

  return (
    <div className="container max-w-6xl py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Publishers</h1>
        <p className="text-muted-foreground">
          Browse books by publisher and view statistics
        </p>
      </div>

      {/* Statistics Cards */}
      <div className="grid gap-4 md:grid-cols-3 mb-8">
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

        <Card className="glass">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Books</CardTitle>
            <BookOpen className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalBooks}</div>
            <p className="text-xs text-muted-foreground">
              Books across all publishers
            </p>
          </CardContent>
        </Card>

        <Card className="glass">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Avg Books/Publisher</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {publishers.length > 0 ? Math.round(totalBooks / publishers.length) : 0}
            </div>
            <p className="text-xs text-muted-foreground">
              Average books per publisher
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Top Publishers */}
      {topPublishers.length > 0 && (
        <Card className="glass mb-8">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart className="w-5 h-5" />
              Top 5 Publishers
            </CardTitle>
            <CardDescription>
              Publishers with the most books in your library
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {topPublishers.map((publisher, index) => {
                const percentage = (publisher.count / totalBooks) * 100
                return (
                  <div key={publisher.name} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <Badge variant="outline" className="w-8 justify-center">
                          {index + 1}
                        </Badge>
                        <Link
                          href={`/search?publisher=${encodeURIComponent(publisher.name)}`}
                          className="font-medium hover:underline"
                        >
                          {publisher.name || "Unknown"}
                        </Link>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="text-sm text-muted-foreground">
                          {publisher.count} books
                        </span>
                        <span className="text-sm font-medium">
                          {percentage.toFixed(1)}%
                        </span>
                      </div>
                    </div>
                    <div className="h-2 bg-secondary rounded-full overflow-hidden">
                      <div
                        className="h-full bg-primary transition-all"
                        style={{ width: `${percentage}%` }}
                      />
                    </div>
                  </div>
                )
              })}
            </div>
          </CardContent>
        </Card>
      )}

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
            <div className="grid gap-2">
              {filteredPublishers.map((publisher) => (
                <Link
                  key={publisher.name}
                  href={`/search?publisher=${encodeURIComponent(publisher.name)}`}
                  className="flex items-center justify-between p-3 rounded-lg hover:bg-accent transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <Building2 className="w-4 h-4 text-muted-foreground" />
                    <span className="font-medium">{publisher.name || "Unknown"}</span>
                  </div>
                  <Badge variant="secondary">
                    {publisher.count} {publisher.count === 1 ? "book" : "books"}
                  </Badge>
                </Link>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Show count */}
      {filteredPublishers.length > 0 && searchQuery && (
        <div className="mt-4 text-center text-sm text-muted-foreground">
          Showing {filteredPublishers.length} of {publishers.length} publishers
        </div>
      )}
    </div>
  )
}

