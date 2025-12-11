"use client"

import { useState, useEffect, useCallback } from "react"

export interface SearchHistoryItem {
  id: string
  query: string
  mode: string
  timestamp: number
  resultsCount?: number
}

const STORAGE_KEY = "search-history"
const MAX_HISTORY_ITEMS = 20

export function useSearchHistory() {
  const [history, setHistory] = useState<SearchHistoryItem[]>([])

  // Load history from localStorage on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored) {
        const parsed = JSON.parse(stored)
        setHistory(Array.isArray(parsed) ? parsed : [])
      }
    } catch (error) {
      console.error("Failed to load search history:", error)
    }
  }, [])

  // Save history to localStorage
  const saveHistory = useCallback((newHistory: SearchHistoryItem[]) => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(newHistory))
      setHistory(newHistory)
    } catch (error) {
      console.error("Failed to save search history:", error)
    }
  }, [])

  // Add new search to history
  const addToHistory = useCallback((query: string, mode: string, resultsCount?: number) => {
    if (!query.trim()) return

    const newItem: SearchHistoryItem = {
      id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      query: query.trim(),
      mode,
      timestamp: Date.now(),
      resultsCount
    }

    setHistory(currentHistory => {
      // Remove duplicate queries (keep most recent)
      const filtered = currentHistory.filter(item => 
        item.query.toLowerCase() !== query.toLowerCase().trim()
      )
      
      // Add new item at the beginning
      const newHistory = [newItem, ...filtered]
      
      // Limit to MAX_HISTORY_ITEMS
      const limited = newHistory.slice(0, MAX_HISTORY_ITEMS)
      
      // Save to localStorage
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(limited))
      } catch (error) {
        console.error("Failed to save search history:", error)
      }
      
      return limited
    })
  }, [])

  // Remove item from history
  const removeFromHistory = useCallback((id: string) => {
    setHistory(currentHistory => {
      const newHistory = currentHistory.filter(item => item.id !== id)
      saveHistory(newHistory)
      return newHistory
    })
  }, [saveHistory])

  // Clear all history
  const clearHistory = useCallback(() => {
    saveHistory([])
  }, [saveHistory])

  // Get recent searches (last 5)
  const getRecentSearches = useCallback(() => {
    return history.slice(0, 5)
  }, [history])

  // Get popular searches (most frequent)
  const getPopularSearches = useCallback(() => {
    const queryCount = new Map<string, number>()
    
    history.forEach(item => {
      const query = item.query.toLowerCase()
      queryCount.set(query, (queryCount.get(query) || 0) + 1)
    })
    
    return Array.from(queryCount.entries())
      .sort((a, b) => b[1] - a[1])
      .slice(0, 5)
      .map(([query, count]) => ({ query, count }))
  }, [history])

  return {
    history,
    addToHistory,
    removeFromHistory,
    clearHistory,
    getRecentSearches,
    getPopularSearches
  }
}