# Testing Guide - Task SSE Streaming

## Prerequisites

1. **Backend must be running**: The Go server must be restarted to load the new SSE code
2. **Port 8080**: Backend should be accessible at `http://localhost:8080`

## Manual Testing Steps

### 1. Start the Backend Server

```bash
# Build and run the server
go build -o calibre-api ./main.go
./calibre-api
```

Or if using a different startup method, ensure the server is restarted to load the new code.

### 2. Test SSE Connection

```bash
# Connect to SSE stream
curl -N http://localhost:8080/api/tasks/stream
```

**Expected Output**:
```
data: {"type":"task_list","tasks":[],"timestamp":"2025-12-11T19:04:07Z"}

data: {"type":"heartbeat","timestamp":"2025-12-11T19:04:37Z"}

data: {"type":"heartbeat","timestamp":"2025-12-11T19:05:07Z"}
```

### 3. Test Task Updates

In another terminal, start a task:

```bash
# Start a Qdrant sync task
curl -X POST http://localhost:8080/api/tasks/start \
  -H "Content-Type: application/json" \
  -d '{"type":"qdrant_sync","mode":"full"}'
```

**Expected in SSE stream**:
```
data: {"type":"task_update","task":{"id":"...","type":"qdrant_sync","state":"running","progress":0,...},"timestamp":"..."}

data: {"type":"task_update","task":{"id":"...","type":"qdrant_sync","state":"running","progress":25,...},"timestamp":"..."}

data: {"type":"task_update","task":{"id":"...","type":"qdrant_sync","state":"completed","progress":100,...},"timestamp":"..."}
```

### 4. Test Multiple Clients

Open multiple terminals and connect to the stream:

```bash
# Terminal 1
curl -N http://localhost:8080/api/tasks/stream

# Terminal 2
curl -N http://localhost:8080/api/tasks/stream

# Terminal 3
curl -N http://localhost:8080/api/tasks/stream
```

All clients should receive the same updates simultaneously.

### 5. Test Connection Limits

Try connecting more than 100 clients (the configured limit):

```bash
# This should fail with a 503 error after 100 connections
for i in {1..101}; do
  curl -N http://localhost:8080/api/tasks/stream &
done
```

### 6. Test Stale Connection Cleanup

1. Connect to the stream
2. Suspend the curl process (Ctrl+Z)
3. Wait 60+ seconds
4. Check server logs - the connection should be cleaned up

### 7. Test Heartbeat

Connect and wait for 30+ seconds. You should see heartbeat messages every 30 seconds.

## Browser Testing

Open browser console and run:

```javascript
const eventSource = new EventSource('http://localhost:8080/api/tasks/stream');

eventSource.onmessage = (event) => {
  console.log('Received:', JSON.parse(event.data));
};

eventSource.onerror = (error) => {
  console.error('SSE Error:', error);
};
```

## Integration with Frontend

The frontend should connect to the SSE endpoint and update the UI in real-time:

```typescript
// In tasks page component
const eventSource = new EventSource('/api/tasks/stream');

eventSource.addEventListener('message', (event) => {
  const message = JSON.parse(event.data);
  
  switch (message.type) {
    case 'task_list':
      // Update task list
      tasks.value = message.tasks;
      break;
    case 'task_update':
      // Update specific task
      updateTask(message.task);
      break;
    case 'heartbeat':
      // Connection is alive
      console.log('Heartbeat received');
      break;
  }
});
```

## Troubleshooting

### Issue: Getting HTML instead of SSE

**Cause**: Backend server not running or not restarted after code changes

**Solution**: Restart the backend server

### Issue: Connection immediately closes

**Cause**: SSE headers not set correctly or client not supporting SSE

**Solution**: Check that `Content-Type: text/event-stream` header is present

### Issue: No task updates received

**Cause**: Task Manager not connected to SSE Manager

**Solution**: Verify `SetSSEManager()` is called in server initialization

### Issue: Memory leak with many connections

**Cause**: Clients not being cleaned up properly

**Solution**: Check that stale connection cleanup is running (should happen every 30s)

## Performance Metrics to Monitor

1. **Connection Count**: Should not exceed 100
2. **Message Latency**: Task updates should arrive within 100ms
3. **Memory Usage**: Should remain stable with active connections
4. **CPU Usage**: Heartbeat and cleanup should have minimal impact

## Next Steps After Testing

1. Add monitoring and logging (Task 11)
2. Integrate with frontend tasks page
3. Add error handling and retry logic in frontend
4. Performance testing with many concurrent connections
