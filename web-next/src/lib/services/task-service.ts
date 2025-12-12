/**
 * Task Service
 * 
 * This service handles task management operations:
 * - Getting task list
 * - Starting and stopping tasks
 * - Subscribing to task updates via SSE
 * 
 * @example
 * ```typescript
 * import { taskService } from '@/lib/services/task-service'
 * 
 * // Get tasks
 * const tasks = await taskService.getTasks()
 * 
 * // Start a task
 * const task = await taskService.startTask('qdrant_sync', 'full')
 * ```
 */

import { BaseApiService } from './base-service'
import { UnifiedApiClient, apiClient } from '../api-client-v2'
import { ErrorHandler, errorHandler } from '../error-handler'

// ============================================================================
// Type Definitions
// ============================================================================

/**
 * Task Status
 */
export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'

/**
 * Task Type
 */
export type TaskType = 
  | 'qdrant_sync' 
  | 'toc_extract' 
  | 'check_missing'
  | 'book_operation'
  | string

/**
 * Task Mode
 */
export type TaskMode = 'full' | 'incremental' | 'single' | string

/**
 * Task
 */
export interface Task {
  id: string
  type: TaskType
  status: TaskStatus
  progress: number
  message: string
  created_at: string
  updated_at: string
  started_at?: string
  completed_at?: string
  error?: string
  metadata?: Record<string, any>
}

/**
 * Start Task Request
 */
export interface StartTaskRequest {
  type: TaskType
  mode?: TaskMode
  options?: Record<string, any>
}

/**
 * Task Update Event
 */
export interface TaskUpdateEvent {
  task_id: string
  status: TaskStatus
  progress: number
  message: string
  timestamp: string
}

// ============================================================================
// Task Service Class
// ============================================================================

export class TaskService extends BaseApiService {
  constructor(client: UnifiedApiClient, errorHandler: ErrorHandler) {
    super(client, errorHandler)
  }

  // ==========================================================================
  // Task Operations
  // ==========================================================================

  /**
   * Get all tasks
   * 
   * @returns Array of tasks
   * 
   * @example
   * ```typescript
   * const tasks = await taskService.getTasks()
   * console.log(`Found ${tasks.length} tasks`)
   * ```
   */
  async getTasks(): Promise<Task[]> {
    return this.handleRequest(async () => {
      return this.client.get<Task[]>('/api/tasks')
    })
  }

  /**
   * Get task by ID
   * 
   * @param id - Task ID
   * @returns Task details
   * 
   * @example
   * ```typescript
   * const task = await taskService.getTask('task-123')
   * ```
   */
  async getTask(id: string): Promise<Task> {
    return this.handleRequest(async () => {
      return this.client.get<Task>(`/api/tasks/${id}`)
    })
  }

  /**
   * Start a new task
   * 
   * @param type - Task type
   * @param mode - Task mode (optional)
   * @param options - Additional options (optional)
   * @returns Started task
   * 
   * @example
   * ```typescript
   * // Start full Qdrant sync
   * const task = await taskService.startTask('qdrant_sync', 'full')
   * 
   * // Start incremental sync
   * const task2 = await taskService.startTask('qdrant_sync', 'incremental')
   * ```
   */
  async startTask(type: TaskType, mode?: TaskMode, options?: Record<string, any>): Promise<Task> {
    return this.handleRequest(async () => {
      return this.client.post<Task>('/api/tasks', {
        type,
        mode,
        options,
      })
    })
  }

  /**
   * Stop a running task
   * 
   * @param id - Task ID
   * 
   * @example
   * ```typescript
   * await taskService.stopTask('task-123')
   * ```
   */
  async stopTask(id: string): Promise<void> {
    return this.handleRequest(async () => {
      return this.client.post<void>(`/api/tasks/${id}/stop`, {})
    })
  }

  /**
   * Delete a task
   * 
   * @param id - Task ID
   * 
   * @example
   * ```typescript
   * await taskService.deleteTask('task-123')
   * ```
   */
  async deleteTask(id: string): Promise<void> {
    return this.handleRequest(async () => {
      return this.client.delete<void>(`/api/tasks/${id}`)
    })
  }

  // ==========================================================================
  // SSE Streaming Operations
  // ==========================================================================

  /**
   * Subscribe to task updates via Server-Sent Events
   * 
   * @returns EventSource for receiving task updates
   * 
   * @example
   * ```typescript
   * const eventSource = taskService.subscribeToTaskUpdates()
   * 
   * eventSource.addEventListener('message', (event) => {
   *   const update: TaskUpdateEvent = JSON.parse(event.data)
   *   console.log(`Task ${update.task_id}: ${update.message}`)
   * })
   * 
   * eventSource.addEventListener('error', (error) => {
   *   console.error('SSE error:', error)
   *   eventSource.close()
   * })
   * ```
   */
  subscribeToTaskUpdates(): EventSource {
    return new EventSource('/api/tasks/stream')
  }

  /**
   * Subscribe to specific task updates
   * 
   * @param taskId - Task ID to monitor
   * @returns EventSource for receiving task updates
   * 
   * @example
   * ```typescript
   * const eventSource = taskService.subscribeToTask('task-123')
   * 
   * eventSource.addEventListener('message', (event) => {
   *   const update: TaskUpdateEvent = JSON.parse(event.data)
   *   console.log(`Progress: ${update.progress}%`)
   * })
   * ```
   */
  subscribeToTask(taskId: string): EventSource {
    return new EventSource(`/api/tasks/${taskId}/stream`)
  }

  // ==========================================================================
  // Helper Methods
  // ==========================================================================

  /**
   * Wait for task completion
   * 
   * @param taskId - Task ID
   * @param pollInterval - Polling interval in milliseconds (default: 1000)
   * @param timeout - Timeout in milliseconds (default: 300000 = 5 minutes)
   * @returns Completed task
   * 
   * @example
   * ```typescript
   * const task = await taskService.startTask('qdrant_sync', 'full')
   * const completedTask = await taskService.waitForCompletion(task.id)
   * console.log('Task completed:', completedTask.status)
   * ```
   */
  async waitForCompletion(
    taskId: string,
    pollInterval: number = 1000,
    timeout: number = 300000
  ): Promise<Task> {
    const startTime = Date.now()

    while (true) {
      const task = await this.getTask(taskId)

      if (task.status === 'completed' || task.status === 'failed' || task.status === 'cancelled') {
        return task
      }

      if (Date.now() - startTime > timeout) {
        throw new Error(`Task ${taskId} timeout after ${timeout}ms`)
      }

      await this.sleep(pollInterval)
    }
  }

  /**
   * Get tasks by status
   * 
   * @param status - Task status to filter
   * @returns Array of tasks with matching status
   * 
   * @example
   * ```typescript
   * const runningTasks = await taskService.getTasksByStatus('running')
   * ```
   */
  async getTasksByStatus(status: TaskStatus): Promise<Task[]> {
    const tasks = await this.getTasks()
    return tasks.filter(task => task.status === status)
  }

  /**
   * Get tasks by type
   * 
   * @param type - Task type to filter
   * @returns Array of tasks with matching type
   * 
   * @example
   * ```typescript
   * const syncTasks = await taskService.getTasksByType('qdrant_sync')
   * ```
   */
  async getTasksByType(type: TaskType): Promise<Task[]> {
    const tasks = await this.getTasks()
    return tasks.filter(task => task.type === type)
  }
}

// ============================================================================
// Default Service Instance
// ============================================================================

/**
 * Default task service instance
 * 
 * @example
 * ```typescript
 * import { taskService } from '@/lib/services/task-service'
 * 
 * const tasks = await taskService.getTasks()
 * ```
 */
export const taskService = new TaskService(apiClient, errorHandler)

