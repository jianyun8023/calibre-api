# Implementation Tasks

## Overview
Implement Server-Sent Events (SSE) streaming for real-time task status updates to replace the current polling mechanism.

## Task Breakdown

### Task 1: Create SSE Manager Core Structure

**Description:** Implement the core SSE manager that handles client connections and message broadcasting.

**Dependencies:** None

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] Create `SSEManager` struct with client map and channels
- [ ] Create `SSEClient` struct with writer, flusher, and done channel
- [ ] Create `SSEMessage` struct with type, data, and timestamp fields
- [ ] Implement `NewSSEManager()` constructor
- [ ] Implement `NewSSEClient()` constructor

**Implementation Notes:**
- Use `sync.RWMutex` for thread-safe client map access
- Use channels for registration, unregistration, and broadcasting
- Store client ID, writer, flusher, and done channel in SSEClient

---

### Task 2: Implement Client Registration and Cleanup

**Description:** Implement client registration, unregistration, and connection cleanup logic.

**Dependencies:** Task 1

**Estimated Effort:** 1.5 hours

**Acceptance Criteria:**
- [ ] Implement `Register(client *SSEClient)` method with connection limit check
- [ ] Implement `Unregister(client *SSEClient)` method with resource cleanup
- [ ] Implement connection limit enforcement (max 100 connections)
- [ ] Implement automatic cleanup of stale connections
- [ ] Add logging for connection events

**Implementation Notes:**
- Check connection count before registration
- Clean up resources when unregistering
- Use goroutine for periodic stale connection cleanup

---

### Task 3: Implement Message Broadcasting

**Description:** Implement the broadcast mechanism to send messages to all connected clients.

**Dependencies:** Task 1, Task 2

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] Implement `Broadcast(message SSEMessage)` method
- [ ] Implement `Send(message SSEMessage)` method on SSEClient
- [ ] Handle write failures gracefully
- [ ] Remove failed connections automatically
- [ ] Log broadcast errors without blocking other clients

**Implementation Notes:**
- Use goroutine for each client send to avoid blocking
- Handle flusher.Flush() errors
- Continue broadcasting even if some clients fail

---

### Task 4: Create SSE HTTP Handler

**Description:** Create the HTTP handler for the `/api/tasks/stream` endpoint.

**Dependencies:** Task 1, Task 2, Task 3

**Estimated Effort:** 1.5 hours

**Acceptance Criteria:**
- [ ] Create `handleTaskStream` HTTP handler
- [ ] Set proper SSE headers (Content-Type, Cache-Control, Connection)
- [ ] Set CORS headers for cross-origin requests
- [ ] Register new SSE client on connection
- [ ] Send initial task list to new clients
- [ ] Keep connection alive until client disconnects

**Implementation Notes:**
- Check if ResponseWriter supports Flusher interface
- Send initial task list immediately after connection
- Use `<-client.Done()` to block until disconnection

---

### Task 5: Implement Heartbeat Mechanism

**Description:** Implement periodic heartbeat messages to keep connections alive.

**Dependencies:** Task 3

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] Create heartbeat goroutine in SSEManager
- [ ] Send heartbeat every 30 seconds
- [ ] Format heartbeat as `{"type": "heartbeat", "timestamp": "..."}`
- [ ] Update client lastSeen timestamp on successful send
- [ ] Stop heartbeat on manager shutdown

**Implementation Notes:**
- Use ticker for 30-second intervals
- Broadcast heartbeat to all clients
- Include timestamp in heartbeat message

---

### Task 6: Integrate with Task Manager

**Description:** Integrate SSE broadcasting with the existing task manager lifecycle.

**Dependencies:** Task 3

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] Add SSEManager reference to TaskManager
- [ ] Broadcast task updates when status changes
- [ ] Broadcast task updates when progress changes
- [ ] Broadcast new task creation
- [ ] Broadcast task completion/failure
- [ ] Format messages according to spec

**Implementation Notes:**
- Modify existing task update methods
- Check if SSEManager is nil before broadcasting
- Use proper message types (task_update, task_list)

---

### Task 7: Add Unit Tests

**Description:** Create unit tests for SSE manager and client functionality.

**Dependencies:** Task 1, Task 2, Task 3

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] Test SSE client registration and unregistration
- [ ] Test connection limit enforcement
- [ ] Test message broadcasting to single client
- [ ] Test message broadcasting to multiple clients
- [ ] Test client disconnection handling
- [ ] Test stale connection cleanup

**Implementation Notes:**
- Use httptest.ResponseRecorder for testing
- Mock Flusher interface
- Test concurrent operations

---

### Task 8: Add Property-Based Tests

**Description:** Create property-based tests for SSE correctness properties.

**Dependencies:** Task 7

**Estimated Effort:** 2.5 hours

**Acceptance Criteria:**
- [ ] Test Property 1: SSE Broadcast Consistency
- [ ] Test Property 2: Initial State Synchronization
- [ ] Test Property 4: Connection Resource Cleanup
- [ ] Test Property 6: Message Format Compliance
- [ ] Test Property 7: Connection Limit Enforcement
- [ ] Configure tests to run 100+ iterations

**Implementation Notes:**
- Use pgregory.net/rapid for property-based testing
- Generate random number of clients (1-100)
- Generate random task updates
- Verify all clients receive identical messages

---

### Task 9: Add Integration Tests

**Description:** Create integration tests for the complete SSE endpoint.

**Dependencies:** Task 4, Task 6

**Estimated Effort:** 1.5 hours

**Acceptance Criteria:**
- [ ] Test complete SSE connection lifecycle
- [ ] Test initial task list delivery
- [ ] Test task update broadcasting
- [ ] Test heartbeat delivery
- [ ] Test graceful disconnection

**Implementation Notes:**
- Use httptest.Server for integration testing
- Simulate real HTTP connections
- Test with actual task manager integration

---

### Task 10: Update Frontend to Use SSE

**Description:** Update the frontend to use SSE instead of polling.

**Dependencies:** Task 4

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] Create React hook for SSE connection (`useTaskStream`)
- [ ] Implement automatic reconnection with exponential backoff
- [ ] Implement fallback to polling if SSE fails
- [ ] Handle all message types (task_list, task_update, heartbeat)
- [ ] Update UI components to use SSE hook

**Implementation Notes:**
- Use EventSource API for SSE
- Implement reconnection logic with backoff
- Keep polling code as fallback
- Update task list state on messages

---

### Task 11: Add Monitoring and Logging

**Description:** Add monitoring metrics and logging for SSE operations.

**Dependencies:** Task 3, Task 5

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] Log client connections and disconnections
- [ ] Log broadcast operations
- [ ] Log errors with context
- [ ] Add metrics for active connections
- [ ] Add metrics for broadcast latency

**Implementation Notes:**
- Use structured logging
- Include client ID in logs
- Track connection count
- Measure broadcast time

---

### Task 12: Documentation and Examples

**Description:** Create documentation and usage examples for the SSE endpoint.

**Dependencies:** Task 10

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] Document SSE endpoint API
- [ ] Document message formats
- [ ] Provide JavaScript/TypeScript examples
- [ ] Document error handling
- [ ] Document fallback behavior

---

## Task Summary

| Task | Effort | Priority | Dependencies |
|------|--------|----------|--------------|
| Task 1: SSE Manager Core | 2h | High | None |
| Task 2: Client Registration | 1.5h | High | Task 1 |
| Task 3: Message Broadcasting | 2h | High | Task 1, 2 |
| Task 4: HTTP Handler | 1.5h | High | Task 1, 2, 3 |
| Task 5: Heartbeat Mechanism | 1h | Medium | Task 3 |
| Task 6: Task Manager Integration | 2h | High | Task 3 |
| Task 7: Unit Tests | 2h | High | Task 1, 2, 3 |
| Task 8: Property-Based Tests | 2.5h | Medium | Task 7 |
| Task 9: Integration Tests | 1.5h | Medium | Task 4, 6 |
| Task 10: Frontend SSE Hook | 2h | High | Task 4 |
| Task 11: Monitoring & Logging | 1h | Low | Task 3, 5 |
| Task 12: Documentation | 1h | Low | Task 10 |

**Total Estimated Effort:** 20 hours

## Implementation Order

1. **Phase 1: Core Infrastructure** (5.5 hours)
   - Task 1: SSE Manager Core
   - Task 2: Client Registration
   - Task 3: Message Broadcasting

2. **Phase 2: HTTP Integration** (3.5 hours)
   - Task 4: HTTP Handler
   - Task 5: Heartbeat Mechanism
   - Task 6: Task Manager Integration

3. **Phase 3: Testing** (6 hours)
   - Task 7: Unit Tests
   - Task 8: Property-Based Tests
   - Task 9: Integration Tests

4. **Phase 4: Frontend & Polish** (4 hours)
   - Task 10: Frontend SSE Hook
   - Task 11: Monitoring & Logging
   - Task 12: Documentation

## Success Criteria

- ✅ SSE endpoint handles 100+ concurrent connections
- ✅ Task updates delivered within 100ms
- ✅ All tests pass (unit, property-based, integration)
- ✅ Frontend successfully uses SSE with fallback
- ✅ No memory leaks or resource exhaustion
- ✅ Graceful handling of connection failures
