import { useState, useEffect, useRef, useCallback } from 'react'
import { toast } from 'sonner'

interface TaskStatus {
  id: string
  type: string
  mode: string
  state: "running" | "completed" | "error" | "stopped"
  progress: number
  message: string
  start_time: string
  end_time: string
  error?: string
}

interface UseTaskStreamOptions {
  onTaskUpdate?: (task: TaskStatus) => void
  onTaskComplete?: (task: TaskStatus) => void
  onError?: (error: Error) => void
}

export function useTaskStream(options: UseTaskStreamOptions = {}) {
  const [tasks, setTasks] = useState<TaskStatus[]>([])
  const [loading, setLoading] = useState(true)
  const [connected, setConnected] = useState(false)
  const eventSourceRef = useRef<EventSource | null>(null)
  const fallbackIntervalRef = useRef<NodeJS.Timeout | null>(null)
  const [useFallback, setUseFallback] = useState(false)

  // Fetch tasks using traditional API (fallback method)
  const fetchTasks = useCallback(async () => {
    try {
      const response = await fetch('/api/tasks')
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }
      const data = await response.json()
      const taskList = data.data || data || []
      setTasks(Array.isArray(taskList) ? taskList : [])
      return taskList
    } catch (error) {
      console.error('Failed to fetch tasks:', error)
      options.onError?.(error instanceof Error ? error : new Error('Failed to fetch tasks'))
      setTasks([])
      return []
    } finally {
      setLoading(false)
    }
  }, [options])

  // Initialize SSE connection
  const connectSSE = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    try {
      const eventSource = new EventSource('/api/tasks/stream')
      eventSourceRef.current = eventSource

      eventSource.onopen = () => {
        console.log('Task stream connected')
        setConnected(true)
        setUseFallback(false)
        setLoading(false)
      }

      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          
          if (data.type === 'task_list') {
            // Full task list update
            const taskList = data.tasks || []
            setTasks(Array.isArray(taskList) ? taskList : [])
          } else if (data.type === 'task_update') {
            // Single task update
            const updatedTask = data.task
            setTasks(prevTasks => {
              const newTasks = prevTasks.map(task => 
                task.id === updatedTask.id ? updatedTask : task
              )
              // If task not found, add it
              if (!prevTasks.find(task => task.id === updatedTask.id)) {
                newTasks.push(updatedTask)
              }
              return newTasks
            })
            
            options.onTaskUpdate?.(updatedTask)
            
            // Notify on completion
            if (updatedTask.state === 'completed') {
              options.onTaskComplete?.(updatedTask)
              toast.success(`Task completed: ${updatedTask.type}`)
            } else if (updatedTask.state === 'error') {
              toast.error(`Task failed: ${updatedTask.type}`)
            }
          }
        } catch (error) {
          console.error('Failed to parse SSE message:', error)
        }
      }

      eventSource.onerror = () => {
        setConnected(false)
        
        // Fall back to polling after SSE failure
        if (!useFallback) {
          console.log('SSE endpoint unavailable, using polling mode')
          setUseFallback(true)
          eventSource.close()
          startFallbackPolling()
        }
      }

    } catch (error) {
      console.error('Failed to create SSE connection:', error)
      setUseFallback(true)
      startFallbackPolling()
    }
  }, [options, useFallback])

  // Start fallback polling
  const startFallbackPolling = useCallback(() => {
    if (fallbackIntervalRef.current) {
      clearInterval(fallbackIntervalRef.current)
    }

    // Initial fetch
    fetchTasks()

    // Set up polling interval
    fallbackIntervalRef.current = setInterval(async () => {
      const taskList = await fetchTasks()
      
      // Only poll if there are running tasks
      if (!taskList.some((task: TaskStatus) => task.state === 'running')) {
        // No running tasks, reduce polling frequency
        if (fallbackIntervalRef.current) {
          clearInterval(fallbackIntervalRef.current)
          fallbackIntervalRef.current = setInterval(fetchTasks, 10000) // 10 seconds
        }
      }
    }, 2000) // 2 seconds for active tasks
  }, [fetchTasks])

  // Initialize connection
  useEffect(() => {
    // Try SSE first, fall back to polling if it fails
    connectSSE()

    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
      }
      if (fallbackIntervalRef.current) {
        clearInterval(fallbackIntervalRef.current)
      }
    }
  }, [connectSSE])

  // Manual refresh function
  const refresh = useCallback(() => {
    if (useFallback) {
      fetchTasks()
    } else {
      // Reconnect SSE
      connectSSE()
    }
  }, [useFallback, fetchTasks, connectSSE])

  // Start task function
  const startTask = useCallback(async (type: string, mode: string) => {
    try {
      const response = await fetch('/api/tasks/start', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ type, mode }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || `HTTP ${response.status}`)
      }

      const data = await response.json()
      toast.success('Task started successfully')
      
      // Refresh task list
      if (useFallback) {
        setTimeout(fetchTasks, 500) // Small delay to let backend update
      }
      
      return data
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to start task'
      toast.error(message)
      throw error
    }
  }, [useFallback, fetchTasks])

  // Stop task function
  const stopTask = useCallback(async (id: string) => {
    try {
      const response = await fetch(`/api/tasks/${id}/stop`, {
        method: 'POST',
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || `HTTP ${response.status}`)
      }

      toast.success('Task stopped')
      
      // Refresh task list
      if (useFallback) {
        setTimeout(fetchTasks, 500) // Small delay to let backend update
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to stop task'
      toast.error(message)
      throw error
    }
  }, [useFallback, fetchTasks])

  return {
    tasks,
    loading,
    connected,
    useFallback,
    refresh,
    startTask,
    stopTask,
  }
}