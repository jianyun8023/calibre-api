"use client"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useState } from "react"
import { Play, Square, RefreshCw, Clock, CheckCircle, XCircle, AlertCircle, Wifi, WifiOff } from "lucide-react"
import { useTaskStream } from "@/hooks/use-task-stream"

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

const TASK_TYPES = [
  { value: "semantic_sync", label: "Semantic Sync", description: "Sync books to search index" },
  { value: "toc_extract", label: "TOC Extract", description: "Extract table of contents" },
  { value: "check_missing", label: "Check Missing", description: "Check missing vectors" },
  { value: "copyright_extract", label: "Copyright Extract", description: "Extract ISBN and metadata from copyright pages" },
  { value: "cleanup_orphans", label: "Cleanup Orphans", description: "Remove orphaned books from search index" },
]

const TASK_MODES = [
  { value: "full", label: "Full", description: "Process all books" },
  { value: "incremental", label: "Incremental", description: "Process only new books" },
]

export default function TasksPage() {
  const [selectedType, setSelectedType] = useState("semantic_sync")
  const [selectedMode, setSelectedMode] = useState("incremental")

  // Use the new task stream hook
  const {
    tasks,
    loading,
    connected,
    useFallback,
    refresh,
    startTask: startTaskStream,
    stopTask: stopTaskStream
  } = useTaskStream({
    onTaskComplete: (task) => {
      console.log('Task completed:', task.type)
    },
    onError: (error) => {
      console.error('Task stream error:', error)
    }
  })

  const handleStartTask = async () => {
    try {
      await startTaskStream(selectedType, selectedMode)
    } catch (error) {
      // Error handling is done in the hook
    }
  }

  const handleStopTask = async (id: string) => {
    try {
      await stopTaskStream(id)
    } catch (error) {
      // Error handling is done in the hook
    }
  }

  const getStateIcon = (state: string) => {
    switch (state) {
      case "running":
        return <Clock className="w-4 h-4 animate-spin" />
      case "completed":
        return <CheckCircle className="w-4 h-4" />
      case "error":
        return <XCircle className="w-4 h-4" />
      case "stopped":
        return <Square className="w-4 h-4" />
      default:
        return <AlertCircle className="w-4 h-4" />
    }
  }

  const getStateBadgeVariant = (state: string): "default" | "secondary" | "destructive" | "outline" => {
    switch (state) {
      case "running":
        return "default"
      case "completed":
        return "secondary"
      case "error":
        return "destructive"
      default:
        return "outline"
    }
  }

  const formatDuration = (start: string, end: string) => {
    if (!end) return "Running..."
    const duration = new Date(end).getTime() - new Date(start).getTime()
    const seconds = Math.floor(duration / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)

    if (hours > 0) return `${hours}h ${minutes % 60}m`
    if (minutes > 0) return `${minutes}m ${seconds % 60}s`
    return `${seconds}s`
  }

  return (
    <div className="container max-w-6xl">
      <div className="flex justify-between items-center mb-8">
        <div className="flex items-center gap-4">
          <h1 className="text-3xl font-bold">Task Management</h1>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            {useFallback ? (
              <>
                <WifiOff className="w-4 h-4" />
                <span>Polling Mode</span>
              </>
            ) : connected ? (
              <>
                <Wifi className="w-4 h-4 text-green-500" />
                <span>Live Updates</span>
              </>
            ) : (
              <>
                <WifiOff className="w-4 h-4 text-yellow-500" />
                <span>Connecting...</span>
              </>
            )}
          </div>
        </div>
        <Button onClick={refresh} variant="outline" size="icon">
          <RefreshCw className="w-4 h-4" />
        </Button>
      </div>

      {/* Start New Task */}
      <Card className="glass mb-6">
        <CardHeader>
          <CardTitle>Start New Task</CardTitle>
          <CardDescription>
            Launch background tasks for book processing
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="flex-1">
              <label className="text-sm font-medium mb-2 block">Task Type</label>
              <Select value={selectedType} onValueChange={setSelectedType}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {TASK_TYPES.map((type) => (
                    <SelectItem key={type.value} value={type.value}>
                      <div>
                        <div className="font-medium">{type.label}</div>
                        <div className="text-xs text-muted-foreground">{type.description}</div>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex-1">
              <label className="text-sm font-medium mb-2 block">Mode</label>
              <Select value={selectedMode} onValueChange={setSelectedMode}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {TASK_MODES.map((mode) => (
                    <SelectItem key={mode.value} value={mode.value}>
                      <div>
                        <div className="font-medium">{mode.label}</div>
                        <div className="text-xs text-muted-foreground">{mode.description}</div>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-end">
              <Button onClick={handleStartTask} className="w-full sm:w-auto">
                <Play className="w-4 h-4 mr-2" />
                Start Task
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Task List */}
      <div className="space-y-4">
        <h2 className="text-xl font-semibold">Active & Recent Tasks</h2>

        {loading ? (
          <Card className="glass">
            <CardContent className="py-8 text-center text-muted-foreground">
              Loading tasks...
            </CardContent>
          </Card>
        ) : !tasks || tasks.length === 0 ? (
          <Card className="glass">
            <CardContent className="py-8 text-center text-muted-foreground">
              No tasks found. Start a new task above.
            </CardContent>
          </Card>
        ) : (
          tasks.map((task) => (
            <Card key={task.id} className="glass">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div>{getStateIcon(task.state)}</div>
                    <div>
                      <CardTitle className="text-lg">
                        {TASK_TYPES.find(t => t.value === task.type)?.label || task.type}
                      </CardTitle>
                      <CardDescription>
                        {task.mode} mode · Started {new Date(task.start_time).toLocaleString()}
                      </CardDescription>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant={getStateBadgeVariant(task.state)}>
                      {task.state}
                    </Badge>
                    {task.state === "running" && (
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleStopTask(task.id)}
                      >
                        <Square className="w-3 h-3 mr-1" />
                        Stop
                      </Button>
                    )}
                  </div>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                {/* Progress Bar */}
                {task.state === "running" && (
                  <div className="space-y-1">
                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">Progress</span>
                      <span className="font-medium">{Math.round(task.progress * 100)}%</span>
                    </div>
                    <Progress value={task.progress * 100} />
                  </div>
                )}

                {/* Message */}
                {task.message && (
                  <p className="text-sm text-muted-foreground">{task.message}</p>
                )}

                {/* Error */}
                {task.error && (
                  <div className="bg-destructive/10 text-destructive px-3 py-2 rounded-md text-sm">
                    <strong>Error:</strong> {task.error}
                  </div>
                )}

                {/* Duration */}
                {task.state !== "running" && task.end_time && (
                  <div className="text-sm text-muted-foreground">
                    Duration: {formatDuration(task.start_time, task.end_time)}
                  </div>
                )}
              </CardContent>
            </Card>
          ))
        )}
      </div>
    </div>
  )
}

