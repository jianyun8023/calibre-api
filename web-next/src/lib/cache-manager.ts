/**
 * Cache Manager
 * 
 * This module provides caching capabilities for API responses.
 * It supports multiple eviction strategies and TTL-based expiration.
 * 
 * @example
 * ```typescript
 * import { cacheManager } from '@/lib/cache-manager'
 * 
 * // Set cache
 * await cacheManager.set('book:1', bookData, 60000) // 1 minute TTL
 * 
 * // Get cache
 * const cachedBook = await cacheManager.get<Book>('book:1')
 * 
 * // Delete cache
 * await cacheManager.delete('book:1')
 * 
 * // Clear all cache
 * await cacheManager.clear()
 * ```
 */

import { CacheConfig, CacheEntry, RequestConfig } from '@/types/api-v2'

// ============================================================================
// Cache Storage Interface
// ============================================================================

interface CacheStorage {
  get(key: string): Promise<string | null>
  set(key: string, value: string): Promise<void>
  delete(key: string): Promise<void>
  clear(): Promise<void>
  keys(): Promise<string[]>
}

// ============================================================================
// In-Memory Cache Storage
// ============================================================================

class InMemoryCacheStorage implements CacheStorage {
  private storage = new Map<string, string>()

  async get(key: string): Promise<string | null> {
    return this.storage.get(key) || null
  }

  async set(key: string, value: string): Promise<void> {
    this.storage.set(key, value)
  }

  async delete(key: string): Promise<void> {
    this.storage.delete(key)
  }

  async clear(): Promise<void> {
    this.storage.clear()
  }

  async keys(): Promise<string[]> {
    return Array.from(this.storage.keys())
  }
}

// ============================================================================
// LocalStorage Cache Storage
// ============================================================================

class LocalStorageCacheStorage implements CacheStorage {
  private prefix = 'api_cache_'

  async get(key: string): Promise<string | null> {
    if (typeof window === 'undefined') return null
    return localStorage.getItem(this.prefix + key)
  }

  async set(key: string, value: string): Promise<void> {
    if (typeof window === 'undefined') return
    localStorage.setItem(this.prefix + key, value)
  }

  async delete(key: string): Promise<void> {
    if (typeof window === 'undefined') return
    localStorage.removeItem(this.prefix + key)
  }

  async clear(): Promise<void> {
    if (typeof window === 'undefined') return
    const keys = await this.keys()
    keys.forEach(key => localStorage.removeItem(this.prefix + key))
  }

  async keys(): Promise<string[]> {
    if (typeof window === 'undefined') return []
    const keys: string[] = []
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i)
      if (key && key.startsWith(this.prefix)) {
        keys.push(key.replace(this.prefix, ''))
      }
    }
    return keys
  }
}

// ============================================================================
// Cache Manager Class
// ============================================================================

export class CacheManager {
  private config: CacheConfig
  private storage: CacheStorage
  private accessOrder: string[] = [] // For LRU strategy

  constructor(config: Partial<CacheConfig> = {}, storage?: CacheStorage) {
    this.config = {
      ttl: config.ttl || 60000, // Default 1 minute
      maxSize: config.maxSize || 100,
      strategy: config.strategy || 'lru',
    }

    // Use provided storage or default to in-memory
    this.storage = storage || new InMemoryCacheStorage()
  }

  // ==========================================================================
  // Cache Operations
  // ==========================================================================

  /**
   * Get cached data
   * 
   * @template T - The type of cached data
   * @param key - Cache key
   * @returns Cached data or null if not found/expired
   */
  async get<T>(key: string): Promise<T | null> {
    try {
      const data = await this.storage.get(key)
      if (!data) return null

      const entry: CacheEntry<T> = JSON.parse(data)

      // Check if expired
      if (this.isExpired(entry)) {
        await this.delete(key)
        return null
      }

      // Update access order for LRU
      if (this.config.strategy === 'lru') {
        this.updateAccessOrder(key)
      }

      return entry.data
    } catch (error) {
      console.error('[Cache Manager] Get error:', error)
      return null
    }
  }

  /**
   * Set cached data
   * 
   * @template T - The type of data to cache
   * @param key - Cache key
   * @param value - Data to cache
   * @param ttl - Time to live in milliseconds (optional)
   */
  async set<T>(key: string, value: T, ttl?: number): Promise<void> {
    try {
      // Check cache size and evict if necessary
      await this.evictIfNecessary()

      const entry: CacheEntry<T> = {
        data: value,
        timestamp: Date.now(),
        ttl: ttl || this.config.ttl,
        key,
      }

      await this.storage.set(key, JSON.stringify(entry))

      // Update access order for LRU
      if (this.config.strategy === 'lru') {
        this.updateAccessOrder(key)
      }
    } catch (error) {
      console.error('[Cache Manager] Set error:', error)
    }
  }

  /**
   * Delete cached data
   * 
   * @param key - Cache key
   */
  async delete(key: string): Promise<void> {
    try {
      await this.storage.delete(key)
      
      // Remove from access order
      const index = this.accessOrder.indexOf(key)
      if (index !== -1) {
        this.accessOrder.splice(index, 1)
      }
    } catch (error) {
      console.error('[Cache Manager] Delete error:', error)
    }
  }

  /**
   * Clear all cached data
   */
  async clear(): Promise<void> {
    try {
      await this.storage.clear()
      this.accessOrder = []
    } catch (error) {
      console.error('[Cache Manager] Clear error:', error)
    }
  }

  // ==========================================================================
  // Cache Strategy Methods
  // ==========================================================================

  /**
   * Check if cache entry is expired
   */
  isExpired(entry: CacheEntry<any>): boolean {
    return Date.now() - entry.timestamp > entry.ttl
  }

  /**
   * Determine if request should be cached
   */
  shouldCache(request: RequestConfig): boolean {
    // Only cache GET requests by default
    if (request.method && request.method !== 'GET') {
      return false
    }

    // Check if caching is explicitly disabled
    if (request.useCache === false) {
      return false
    }

    return true
  }

  /**
   * Generate cache key from request config
   */
  generateKey(request: RequestConfig): string {
    const url = request.url || ''
    const method = request.method || 'GET'
    const body = request.body ? JSON.stringify(request.body) : ''
    
    // Simple hash function
    const hash = this.simpleHash(`${method}:${url}:${body}`)
    return `req_${hash}`
  }

  // ==========================================================================
  // Private Helper Methods
  // ==========================================================================

  /**
   * Update access order for LRU strategy
   */
  private updateAccessOrder(key: string): void {
    // Remove key if it exists
    const index = this.accessOrder.indexOf(key)
    if (index !== -1) {
      this.accessOrder.splice(index, 1)
    }

    // Add key to the end (most recently used)
    this.accessOrder.push(key)
  }

  /**
   * Evict entries if cache size exceeds max size
   */
  private async evictIfNecessary(): Promise<void> {
    const keys = await this.storage.keys()
    
    if (keys.length >= this.config.maxSize) {
      const keyToEvict = this.selectKeyToEvict(keys)
      if (keyToEvict) {
        await this.delete(keyToEvict)
      }
    }
  }

  /**
   * Select key to evict based on strategy
   */
  private selectKeyToEvict(keys: string[]): string | null {
    if (keys.length === 0) return null

    switch (this.config.strategy) {
      case 'lru':
        // Evict least recently used
        return this.accessOrder[0] || keys[0]

      case 'fifo':
        // Evict first in (oldest)
        return keys[0]

      case 'custom':
        // Custom strategy (can be overridden)
        return keys[0]

      default:
        return keys[0]
    }
  }

  /**
   * Simple hash function for generating cache keys
   */
  private simpleHash(str: string): string {
    let hash = 0
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i)
      hash = ((hash << 5) - hash) + char
      hash = hash & hash // Convert to 32-bit integer
    }
    return Math.abs(hash).toString(36)
  }

  // ==========================================================================
  // Cache Statistics
  // ==========================================================================

  /**
   * Get cache statistics
   */
  async getStats(): Promise<CacheStats> {
    const keys = await this.storage.keys()
    let validEntries = 0
    let expiredEntries = 0

    for (const key of keys) {
      const data = await this.storage.get(key)
      if (data) {
        try {
          const entry = JSON.parse(data)
          if (this.isExpired(entry)) {
            expiredEntries++
          } else {
            validEntries++
          }
        } catch {
          expiredEntries++
        }
      }
    }

    return {
      totalEntries: keys.length,
      validEntries,
      expiredEntries,
      maxSize: this.config.maxSize,
      strategy: this.config.strategy,
    }
  }

  /**
   * Clean up expired entries
   */
  async cleanupExpired(): Promise<number> {
    const keys = await this.storage.keys()
    let cleanedCount = 0

    for (const key of keys) {
      const data = await this.storage.get(key)
      if (data) {
        try {
          const entry = JSON.parse(data)
          if (this.isExpired(entry)) {
            await this.delete(key)
            cleanedCount++
          }
        } catch {
          await this.delete(key)
          cleanedCount++
        }
      }
    }

    return cleanedCount
  }
}

// ============================================================================
// Cache Statistics Interface
// ============================================================================

export interface CacheStats {
  totalEntries: number
  validEntries: number
  expiredEntries: number
  maxSize: number
  strategy: string
}

// ============================================================================
// Default Cache Manager Instance
// ============================================================================

/**
 * Default cache manager instance (in-memory)
 * 
 * This instance can be used throughout the application for caching.
 * 
 * @example
 * ```typescript
 * import { cacheManager } from '@/lib/cache-manager'
 * 
 * await cacheManager.set('key', data)
 * const cached = await cacheManager.get('key')
 * ```
 */
export const cacheManager = new CacheManager({
  ttl: 300000, // 5 minutes
  maxSize: 100,
  strategy: 'lru',
})

/**
 * Persistent cache manager instance (localStorage)
 * 
 * Use this for data that should persist across page reloads.
 */
export const persistentCacheManager = typeof window !== 'undefined' 
  ? new CacheManager(
      {
        ttl: 3600000, // 1 hour
        maxSize: 50,
        strategy: 'lru',
      },
      new LocalStorageCacheStorage()
    )
  : cacheManager // Fallback to in-memory on server

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Create a cached version of an async function
 * 
 * @template T - The return type of the function
 * @param fn - The async function to cache
 * @param keyFn - Function to generate cache key from arguments
 * @param ttl - Cache TTL in milliseconds
 * @returns Cached version of the function
 * 
 * @example
 * ```typescript
 * const cachedFetchBook = withCache(
 *   (id: string) => apiClient.get<Book>(`/book/${id}`),
 *   (id) => `book:${id}`,
 *   60000 // 1 minute
 * )
 * 
 * const book = await cachedFetchBook('1')
 * ```
 */
export function withCache<T extends (...args: any[]) => Promise<any>>(
  fn: T,
  keyFn: (...args: Parameters<T>) => string,
  ttl?: number
): T {
  return (async (...args: Parameters<T>) => {
    const key = keyFn(...args)

    // Try to get from cache
    const cached = await cacheManager.get(key)
    if (cached !== null) {
      return cached
    }

    // Execute function and cache result
    const result = await fn(...args)
    await cacheManager.set(key, result, ttl)

    return result
  }) as T
}

// Clean up expired entries periodically (every 5 minutes)
if (typeof window !== 'undefined') {
  setInterval(() => {
    cacheManager.cleanupExpired()
    persistentCacheManager.cleanupExpired()
  }, 300000)
}

