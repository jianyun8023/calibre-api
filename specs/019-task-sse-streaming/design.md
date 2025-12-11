# Design Document

## Overview

This design implements Server-Sent Events (SSE) streaming for real-time task status updates in the calibre-api system. The solution replaces the current polling mechanism with a push-based approach, providing instant updates when task status changes while maintaining backward compatibility through intelligent fallback.

**Key Benefits:**
- **Real-time Updates**: Instant notification of task status changes (< 100ms latency)
- **Reduced Server Load**: Eliminates continuous polling requests (up to 90% reduction in API calls)
- **Better User Experience**: Live progress updates without manual refresh
- **Resource Efficiency**: Single persistent connection vs multiple polling requests

## Architecture

### High-Level Architecture

```
┌─────────────────┐    SSE Stream     ┌─────────────────┐
│   Frontend      │◄─────────────────►│   SSE Handler   │
│   (React Hook)  │                   │   (Go Gorilla)  │
└─────────────────┘                   └─────────────────┘
         │                                       │
         │ Fallback Polling                      │
         ▼                                       ▼
┌─────────────────┐    REST API       ┌─────────────────┐
│   Polling API   │◄─────────────────►│  Task Manager   │
│   (/api/tasks)  │                   │   (Existing)    │
└─────────────────┘                   └─────────────────┘
```

### SSE Connection Flow

```
Client                    Server
  │                         │
  │──── GET /api/tasks/stream ────►│
  │                         │
  │◄─── SSE Headers ────────│
  │                         │
  │◄─── Initial Task List ──│
  │                         │
  │◄─── Task Updates ───────│ (Real-time)
  │                         │
  │◄─── Heartbeat ──────────│ (Every 30s)
  │                         │
```

## Components and Interfaces

### 1. SSE Connection Manager

**Purpose**: Manages multiple SSE connections and broadcasts updates

```go
type SSEManager struct {
    clients    map[string]*SSEClient
    register   chan *SSEClient
    unregister chan *SSEClient
    broadcast  chan SSEMessage
    mu         sync.RWMutex
}

type SSEClient struct {
    id       string
    writer   http.ResponseWriter
    flusher  http.Flusher
    done     chan struct{}
    lastSeen time.Time
}

type SSEMessage struct {
    Type string      `json:"type"`
    Data interface{} `json:"data,omitempty"`
    Task *TaskStatus `json:"task,omitempty"`
    Tasks []TaskStatus `json:"tasks,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}
```

### 2. HTTP Handler

**Purpose**: Handles SSE endpoint and connection establishment

```go
func (s *Server) handleTaskStream(w http.ResponseWriter, r *http.Request) {
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    // Create SSE client and register
    client := NewSSEClient(w)
    s.sseManager.Register(client)
    
    // Send initial task list
    tasks := s.taskManager.GetTasks()
    client.Send(SSEMessage{
        Type: "task_list",
        Tasks: tasks,
        Timestamp: time.Now(),
    })
    
    // Keep connection alive
    <-client.Done()
}
```

### 3. Task Manager Integration

**Purpose**: Integrate SSE broadcasting with existing task lifecycle

```go
// Modify existing TaskManager to broadcast updates
func (m *Manager) updateTaskStatus(id string, status TaskStatus) {
    m.mu.Lock()
    // Update internal state
    m.mu.Unlock()
    
    // Broadcast to SSE clients
    if m.sseManager != nil {
        m.sseManager.Broadcast(SSEMessage{
            Type: "task_update",
            Task: &status,
            Timestamp: time.Now(),
        })
    }
}
```

## Data Models

### SSE Message Types

**1. Initial Task List**
```json
{
  "type": "task_list",
  "tasks": [
    {
      "id": "uuid-123",
      "type": "qdrant_sync",
      "mode": "incremental",
      "state": "running",
      "progress": 0.45,
      "message": "Processing book 450/1000",
      "start_time": "2025-12-11T10:00:00Z",
      "end_time": null
    }
  ],
  "timestamp": "2025-12-11T10:30:00Z"
}
```

**2. Task Update**
```json
{
  "type": "task_update",
  "task": {
    "id": "uuid-123",
    "type": "qdrant_sync",
    "mode": "incremental", 
    "state": "running",
    "progress": 0.67,
    "message": "Processing book 670/1000",
    "start_time": "2025-12-11T10:00:00Z",
    "end_time": null
  },
  "timestamp": "2025-12-11T10:35:00Z"
}
```

**3. Heartbeat**
```json
{
  "type": "heartbeat",
  "timestamp": "2025-12-11T10:35:30Z"
}
```

### Connection State Management

```go
type ConnectionState struct {
    ConnectedClients int           `json:"connected_clients"`
    TotalConnections int64         `json:"total_connections"`
    LastBroadcast    time.Time     `json:"last_broadcast"`
    ActiveTasks      int           `json:"active_tasks"`
}
```
## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property Reflection

After reviewing all testable properties from the prework analysis, I identified several areas where properties can be consolidated:

- Properties 1.1, 1.3, 1.4, and 2.4 all test broadcasting to multiple clients and can be combined into a comprehensive broadcast property
- Properties 4.1, 4.2, 4.3, and 4.4 all test message format consistency and can be combined into a message format property
- Properties 2.2 and 3.3 both test heartbeat behavior and can be combined

**Property 1: SSE Broadcast Consistency**
*For any* SSE manager with multiple connected clients, when a task status change occurs, all active clients should receive the same update message simultaneously
**Validates: Requirements 1.1, 1.3, 1.4, 2.4**

**Property 2: Initial State Synchronization**
*For any* new SSE client connection, the client should receive the complete current task list as the first message
**Validates: Requirements 1.5**

**Property 3: Update Latency Guarantee**
*For any* task progress update, all connected SSE clients should receive the update within 100ms of the change
**Validates: Requirements 1.2**

**Property 4: Connection Resource Cleanup**
*For any* SSE client that disconnects, the server should automatically remove the client from the active connections list and free associated resources
**Validates: Requirements 2.3**

**Property 5: Heartbeat Consistency**
*For any* established SSE connection with no task updates, the server should send heartbeat messages at regular 30-second intervals
**Validates: Requirements 2.2, 3.3**

**Property 6: Message Format Compliance**
*For any* SSE message sent to clients, the message should be valid JSON with the required type field and appropriate data structure
**Validates: Requirements 3.1, 4.1**

**Property 7: Connection Limit Enforcement**
*For any* attempt to establish SSE connections beyond the configured limit, the server should reject new connections while maintaining existing ones
**Validates: Requirements 3.2**

**Property 8: Stale Connection Detection**
*For any* SSE connection that becomes unresponsive, the server should detect and remove it within 60 seconds
**Validates: Requirements 3.4**

**Property 9: Error Isolation**
*For any* error occurring in one SSE connection, other active connections should continue to function normally
**Validates: Requirements 4.5**

**Property 10: Reconnection State Sync**
*For any* SSE client that reconnects after a failure, the server should send the current complete task list to resynchronize state
**Validates: Requirements 5.2**

## Error Handling

### Connection Management Errors

**1. Client Disconnection**
```go
func (c *SSEClient) handleDisconnection() {
    select {
    case <-c.done:
        return // Already handled
    default:
        close(c.done)
        log.Infof("SSE client %s disconnected", c.id)
    }
}
```

**2. Write Failures**
```go
func (c *SSEClient) Send(message SSEMessage) error {
    data, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }
    
    _, err = fmt.Fprintf(c.writer, "data: %s\n\n", data)
    if err != nil {
        c.handleDisconnection()
        return fmt.Errorf("failed to write SSE message: %w", err)
    }
    
    c.flusher.Flush()
    return nil
}
```

### Broadcasting Errors

**1. Partial Broadcast Failure**
- Continue broadcasting to healthy connections
- Log failed connections for monitoring
- Remove failed connections from active list

**2. Message Serialization Failure**
- Log the error with message details
- Skip the problematic message
- Continue with next message in queue

### Resource Management

**1. Connection Limits**
```go
const MaxSSEConnections = 100

func (m *SSEManager) Register(client *SSEClient) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if len(m.clients) >= MaxSSEConnections {
        return fmt.Errorf("maximum SSE connections reached: %d", MaxSSEConnections)
    }
    
    m.clients[client.id] = client
    return nil
}
```

**2. Memory Leak Prevention**
- Periodic cleanup of stale connections
- Timeout-based connection removal
- Graceful shutdown handling

## Testing Strategy

### Unit Testing Approach

**Connection Management Tests:**
- Test SSE client registration and deregistration
- Verify connection limit enforcement
- Test stale connection cleanup

**Message Broadcasting Tests:**
- Test single client message delivery
- Test multi-client broadcast consistency
- Test message format validation

**Error Handling Tests:**
- Test client disconnection scenarios
- Test write failure recovery
- Test partial broadcast failure handling

### Property-Based Testing Approach

The testing strategy will use Go's testing framework with custom property-based testing utilities for generating random test scenarios.

**Property Test Framework:**
```go
// Use rapid for property-based testing in Go
import "pgregory.net/rapid"

func TestSSEBroadcastConsistency(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random number of clients (1-10)
        numClients := rapid.IntRange(1, 10).Draw(t, "numClients")
        
        // Generate random task update
        taskUpdate := generateRandomTaskUpdate(t)
        
        // Test that all clients receive the same message
        // Implementation details...
    })
}
```

**Test Configuration:**
- Minimum 100 iterations per property test
- Random seed logging for reproducibility
- Timeout handling for long-running tests

### Integration Testing

**SSE Endpoint Tests:**
- Test complete SSE connection lifecycle
- Test integration with existing task manager
- Test frontend compatibility

**Load Testing:**
- Test with maximum concurrent connections
- Test broadcast performance under load
- Test memory usage with long-running connections

### Manual Testing Scenarios

1. **Multi-tab Testing**: Open multiple browser tabs and verify all receive updates
2. **Network Interruption**: Disconnect network and verify reconnection behavior
3. **Server Restart**: Restart server and verify graceful connection termination
4. **Long-running Tasks**: Monitor SSE during extended task execution