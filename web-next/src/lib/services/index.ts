/**
 * Services Index
 * 
 * This module exports all API services for easy importing.
 * 
 * @example
 * ```typescript
 * import { bookService, metadataService, taskService } from '@/lib/services'
 * 
 * const book = await bookService.getBook('1')
 * const metadata = await metadataService.searchByISBN('9787111633082')
 * ```
 */

// Export service classes
export { BaseApiService } from './base-service'
export { BookService, bookService } from './book-service'
export { MetadataService, metadataService } from './metadata-service'
export { TaskService, taskService } from './task-service'

// Export type definitions
export type {
  // Book Service
  BookListResponse,
  TocResponse,
  TocChapter,
  BookUpdateData,
} from './book-service'

export type {
  // Metadata Service
  DoubanBook,
  MetadataSearchResponse,
  MetadataISBNResponse,
} from './metadata-service'

export type {
  // Task Service
  Task,
  TaskStatus,
  TaskType,
  TaskMode,
  StartTaskRequest,
  TaskUpdateEvent,
} from './task-service'

