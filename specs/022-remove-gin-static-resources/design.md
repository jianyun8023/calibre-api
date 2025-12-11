# Design Document

## Overview

This design outlines the removal of static file serving functionality from the Gin backend. Since the frontend has been migrated to Next.js (spec 009-frontend-migration), the backend should focus exclusively on API endpoints. This change simplifies the architecture, reduces configuration complexity, and enforces clear separation of concerns between frontend and backend services.

## Architecture

### Current Architecture
```
┌─────────────────────────────────────┐
│         Gin Backend                 │
│  ┌──────────────────────────────┐  │
│  │   API Routes                 │  │
│  │   /api/books, /api/search    │  │
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │   MCP Routes                 │  │
│  │   /mcp/sse, /mcp/message     │  │
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │   Static File Serving        │  │ ← TO BE REMOVED
│  │   /assets/*, /, /index       │  │
│  │   NoRoute → index.html       │  │
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘
```

### Target Architecture
```
┌─────────────────────────────────────┐
│         Gin Backend                 │
│  ┌──────────────────────────────┐  │
│  │   API Routes                 │  │
│  │   /api/books, /api/search    │  │
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │   MCP Routes                 │  │
│  │   /mcp/sse, /mcp/message     │  │
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │   NoRoute → 404 JSON         │  │ ← SIMPLIFIED
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│      Next.js Frontend (web-next)    │
│  Serves all static assets & pages   │
└─────────────────────────────────────┘
```

## Components and Interfaces

### 1. main.go Changes

**Current Implementation:**
- `setPages()` function configures static file serving
- Registers `/assets` static directory
- Registers `/`, `/index`, `/favico.ico` routes
- NoRoute handler serves index.html for SPA routing

**Target Implementation:**
- Remove or simplify `setPages()` function
- Remove all static file route registrations
- Replace NoRoute handler with JSON 404 response
- Remove StaticDir dependency

### 2. Configuration Changes

**File:** `internal/calibre/types.go`

**Current Config Struct:**
```go
type Config struct {
    Address   string
    Debug     bool
    StaticDir string  // ← TO BE REMOVED
    TmpDir    string
    // ... other fields
}
```

**Target Config Struct:**
```go
type Config struct {
    Address   string
    Debug     bool
    TmpDir    string
    // ... other fields
}
```

### 3. Configuration Initialization Changes

**File:** `main.go` - `initConfig()` function

**Removals:**
- `viper.SetDefault("staticDir", "./static")` - Remove default value
- StaticDir validation in `validateConfig()`

## Data Models

No data model changes required. This is purely a routing and configuration change.

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property Reflection

After analyzing all acceptance criteria, I identified the following consolidations:
- Criteria 1.1, 1.2, 1.3 can be combined into a single comprehensive example testing router initialization
- Criteria 1.5 and 2.4 are identical and should be consolidated
- Criteria 2.1 and 2.2 both test config loading and can be combined
- Criteria 1.4 and 3.3 are true properties that apply across multiple inputs

### Properties

Property 1: NoRoute returns JSON 404
*For any* unmatched route path, when a request is made to that path, the response should be JSON format with HTTP 404 status code, not HTML content
**Validates: Requirements 1.4**

Property 2: All registered routes are API or MCP routes
*For any* route in the Gin router's route list, the route path should start with `/api/`, `/mcp/`, or be `/ping` (health check)
**Validates: Requirements 3.3**

### Examples

Example 1: Router initialization excludes static routes
When the Gin router is initialized, verify that:
- No routes match the pattern `/assets/*`
- Routes `/`, `/index`, `/favico.ico` are not registered
- The route list does not contain static file serving routes
**Validates: Requirements 1.1, 1.2, 1.3**

Example 2: Server starts without StaticDir configuration
When initializing the server with a config that omits the StaticDir field, the server should start successfully without validation errors
**Validates: Requirements 1.5, 2.4**

Example 3: Config loads without StaticDir
When loading a configuration file that does not include StaticDir:
- No default value should be set for StaticDir
- No validation errors should occur
- The Config struct's StaticDir field should be empty
**Validates: Requirements 2.1, 2.2**

## Error Handling

### NoRoute Handler

**Current Behavior:**
- Serves index.html for any unmatched route
- Enables SPA client-side routing

**New Behavior:**
- Returns JSON 404 response
- Format: `{"error": "route not found", "path": "<requested-path>"}`
- HTTP Status: 404 Not Found

### Configuration Validation

**Removals:**
- Remove StaticDir existence check
- Remove StaticDir empty string check

**Retained:**
- Address validation (required)
- Other existing validations remain unchanged

## Testing Strategy

### Unit Tests

Unit tests will verify specific examples and integration points:

1. **Router Initialization Test**
   - Initialize Gin router with production configuration
   - Verify no static file routes are registered
   - Verify specific routes (/, /index, /favico.ico) are absent

2. **NoRoute Handler Test**
   - Make request to non-existent route
   - Verify response is JSON with 404 status
   - Verify response contains error message and path

3. **Configuration Loading Test**
   - Load config without StaticDir field
   - Verify no errors during loading
   - Verify no default value is set
   - Verify server initialization succeeds

### Property-Based Tests

Property-based tests will verify universal properties across many inputs:

1. **Property Test: NoRoute JSON Response**
   - Generate random invalid route paths
   - For each path, make HTTP request
   - Verify response is always JSON with 404 status
   - Verify response never contains HTML content

2. **Property Test: Route List Composition**
   - Get complete list of registered routes
   - For each route, verify path starts with `/api/`, `/mcp/`, or equals `/ping`
   - Verify no routes match static file patterns

### Testing Framework

- **Unit Testing**: Go's built-in `testing` package
- **Property-Based Testing**: [gopter](https://github.com/leanovate/gopter) - Go property testing library
- **HTTP Testing**: `net/http/httptest` for request/response testing
- **Minimum Iterations**: 100 iterations per property test

## Implementation Notes

### Migration Path

1. Remove `setPages()` function or simplify to only set NoRoute handler
2. Remove StaticDir from Config struct
3. Remove StaticDir from viper defaults and validation
4. Update NoRoute handler to return JSON 404
5. Update any documentation referencing static file serving

### Backward Compatibility

This is a **breaking change** for deployments that:
- Rely on Gin to serve the frontend
- Have not migrated to Next.js frontend

**Mitigation:**
- This change should only be deployed after confirming Next.js frontend is running
- Update deployment documentation to reflect separate frontend/backend services
- Update docker-compose or deployment scripts to run both services

### Configuration Migration

Existing `config.yaml` files may still contain `staticDir` field. This is acceptable:
- Viper will load the field but it won't be used
- No validation errors will occur
- Users can remove it at their convenience

## Dependencies

**Depends on:**
- Spec 009-frontend-migration (completed) - Next.js frontend must be operational

**Blocks:**
- None - This is a cleanup task

## References

- Spec 009-frontend-migration: Frontend migration to Next.js
- Gin Static Files Documentation: https://gin-gonic.com/docs/examples/serving-static-files/
- Gin NoRoute Handler: https://gin-gonic.com/docs/examples/custom-http-config/
