export interface ApiResponse<T> {
  data: T
  message?: string
  error?: string
}

export interface PaginatedResponse<T> {
  records: T[]
  total: number
  next_cursor?: string
}
