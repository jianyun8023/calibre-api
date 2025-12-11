// 豆瓣元数据结构
export interface DoubanBook {
  title: string
  sub_title?: string
  author: string[]
  publisher: string
  pubdate: string
  isbn13: string
  isbn10?: string
  summary: string
  image: string
  rating: {
    average: number
    numRaters: number
  }
  tags: Array<{ name: string; count: number }>
  pages?: number
  price?: string
}

export interface MetadataSearchResponse {
  success: boolean
  books: DoubanBook[]
  count: number
  start: number
  total: number
}

export interface MetadataISBNResponse extends DoubanBook {
  // Additional fields from ISBN API if needed
}

/**
 * 通过 ISBN 搜索元数据
 * 注意：豆瓣 API 返回格式与 Calibre API 不同，直接使用 fetch
 */
export async function searchMetadataByISBN(isbn: string): Promise<MetadataISBNResponse> {
  const cleanISBN = isbn.replace(/-/g, '')
  const response = await fetch(`/api/metadata/isbn/${cleanISBN}`)
  
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }
  
  return await response.json()
}

/**
 * 通过关键词搜索元数据（支持标题、作者、ISBN）
 * 注意：豆瓣 API 返回格式与 Calibre API 不同，直接使用 fetch
 */
export async function searchMetadata(query: string): Promise<MetadataSearchResponse> {
  const response = await fetch(`/api/metadata/search?query=${encodeURIComponent(query)}`)
  
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }
  
  return await response.json()
}

