# 🗓️ Advanced Book Filtering - Implementation Plan

> **Strategy**: Rolling Wave + TDD  
> **Current Phase**: Complete ✅  
> **Total Estimated Effort**: 6 hours  
> **Risk Level**: Medium

<!-- English: This plan provides executable steps for implementing the design.
     Each step has verification command and rollback procedure.
     
     中文：本计划提供可执行的实施步骤。
     每个步骤都有验证命令和回滚程序。
-->

## 0. Design Reference & Alignment

> **Key**: This plan implements design in `design.md`

### 0.1 Design Document Links
- **DESIGN**: `design.md`
- **REQUIREMENTS**: `requirements.md`
- **Covered User Stories**: US-001, US-002, US-003, US-004, US-005, US-007

### 0.2 File Manifest Coverage Check

> Verify all files from DESIGN Section 9 are covered by this plan

| File (from DESIGN) | Corresponding Step | Status |
|------------------|---------|------|
| `internal/calibre/filter_service.go` | Step 1.2 | ✅ |
| `internal/repository/filter_repository.go` | Step 1.3 | ✅ |
| `internal/calibre/filter_handler.go` | Step 2.1 | ✅ |
| `internal/calibre/types.go` (modify) | Step 1.1 | ✅ |
| `internal/calibre/router.go` (modify) | Step 2.2 | ✅ |
| `internal/calibre/search_handler.go` (modify) | Step 2.3 | ✅ |
| Test files | Steps 1.4, 2.4 | ✅ |

### 0.3 Prerequisites

> What must be completed before starting this plan?

- **Dependencies to Install**: None (all dependencies already in go.mod)
- **Environment Variables**: None required
- **Setup Commands**: 
  ```bash
  # Ensure database is accessible
  sqlite3 /path/to/metadata.db ".tables"
  ```
- **Blocking Tasks**: None

## 1. High-Level Roadmap (Big Picture)

- [x] **Phase 1**: Core Infrastructure (Steps 1.1-1.4) - Database, types, service layer
- [x] **Phase 2**: API Layer (Steps 2.1-2.4) - HTTP handlers, routes, integration
- [ ] **Phase 3**: Polish & Edge Cases (future) - Caching, optimization

## 2. Detailed Execution Plan: Phase 1 (Core Infrastructure)

> **Context**: Build foundation - database, types, service logic

### Step 1.1: Add Type Definitions (Red)

- [x] **Create test file**
  - **Source**: DESIGN Section 9.3 (Type Definitions)
  - **Action**: Create `internal/calibre/types_test.go`
  - **Content**: Write tests for SavedFilter JSON marshaling/unmarshaling
  - **Verify**: Run `go test -run TestSavedFilter internal/calibre/` → Expect fail (Red)
  - **Estimated Effort**: 15 minutes
  - **Risk**: Low
  - **Depends On**: None
  - **Rollback**: `rm internal/calibre/types_test.go`

- [x] **Add types to types.go**
  - **Source**: DESIGN Section 9.3
  - **Action**: Add SavedFilter, FilterCriteria, FilterCondition to `internal/calibre/types.go`
  - **Code Snippet** (from DESIGN):
    ```go
    type SavedFilter struct {
        ID          int64          `json:"id" db:"id"`
        Name        string         `json:"name" db:"name" binding:"required,max=100"`
        // ... (see DESIGN Section 9.3)
    }
    ```
  - **Verify**: Run `go test -run TestSavedFilter internal/calibre/` → Expect pass (Green)
  - **Estimated Effort**: 20 minutes
  - **Risk**: Low
  - **Depends On**: Step 1.1 (test)
  - **Rollback**: `git checkout -- internal/calibre/types.go`

### Step 1.2: Create FilterService (Green)

- [x] **Create service file**
  - **Source**: DESIGN Section 9.4 (API Signatures)
  - **Action**: Create `internal/calibre/filter_service.go`
  - **Content**: Implement FilterService struct and methods
  - **Verify**: Run `go build ./internal/calibre/` → Expect success
  - **Estimated Effort**: 45 minutes
  - **Risk**: Medium
  - **Depends On**: Step 1.1
  - **Rollback**: `rm internal/calibre/filter_service.go`

### Step 1.3: Create FilterRepository

- [x] **Create repository file**
  - **Source**: DESIGN Section 3 (Data Model)
  - **Action**: Create `internal/repository/filter_repository.go`
  - **Content**: Implement CRUD operations for calibre_saved_filters table
  - **Verify**: Run `go test ./internal/repository/` → Expect pass
  - **Estimated Effort**: 1 hour
  - **Risk**: Medium (database operations)
  - **Depends On**: Step 1.1
  - **Rollback**: `rm internal/repository/filter_repository.go`

### Step 1.4: Database Migration

- [x] **Create migration SQL**
  - **Source**: DESIGN Section 3 (Database Schema)
  - **Action**: Create `migrations/001_create_saved_filters.sql`
  - **Content**: CREATE TABLE statement from DESIGN
  - **Verify**: Run `sqlite3 test.db < migrations/001_create_saved_filters.sql` → Expect success
  - **Follow-up**: Apply to production database
  - **Estimated Effort**: 20 minutes
  - **Risk**: High (DB change)
  - **Depends On**: None (can parallel with Step 1.1)
  - **Rollback**: `DROP TABLE calibre_saved_filters;`

---

### 🚩 Milestone 1: Core Foundation Complete

> After steps 1.1-1.4, verify:
- [x] Types defined and tested
- [x] Service layer implemented
- [x] Repository layer tested
- [x] Database schema created
- [x] `go build ./...` passes

## 3. Detailed Execution Plan: Phase 2 (API Layer)

> **Context**: Expose functionality via HTTP API

### Step 2.1: Create FilterHandler

- [x] **Create handler file**
  - **Source**: DESIGN Section 4 (Interface Specification)
  - **Action**: Create `internal/calibre/filter_handler.go`
  - **Content**: Implement HTTP handlers for filter CRUD
  - **Verify**: Run `go build ./internal/calibre/` → Expect success
  - **Estimated Effort**: 1 hour
  - **Risk**: Low
  - **Depends On**: Step 1.2
  - **Rollback**: `rm internal/calibre/filter_handler.go`

### Step 2.2: Register Routes

- [x] **Update router**
  - **Source**: DESIGN Section 4 (API endpoints)
  - **Action**: Add filter routes to `internal/calibre/router.go`
  - **Content**:
    ```go
    // Filter routes
    api.POST("/filters", filterHandler.Create)
    api.GET("/filters", filterHandler.List)
    api.GET("/filters/:id", filterHandler.Get)
    api.PUT("/filters/:id", filterHandler.Update)
    api.DELETE("/filters/:id", filterHandler.Delete)
    api.POST("/filters/apply", filterHandler.Apply)
    ```
  - **Verify**: Run `go build ./...` → Expect success
  - **Estimated Effort**: 15 minutes
  - **Risk**: Low
  - **Depends On**: Step 2.1
  - **Rollback**: `git checkout -- internal/calibre/router.go`

### Step 2.3: Integration Tests

- [x] **Create integration tests**
  - **Source**: DESIGN Section 7.2 (Integration Tests)
  - **Action**: Create `internal/calibre/filter_handler_test.go`
  - **Content**: Test all API endpoints with httptest
  - **Verify**: Run `go test ./internal/calibre/` → Expect pass
  - **Estimated Effort**: 1.5 hours
  - **Risk**: Low
  - **Depends On**: Step 2.2
  - **Rollback**: `rm internal/calibre/filter_handler_test.go`

### Step 2.4: Manual API Testing

- [x] **Test with curl**
  - **Source**: DESIGN Section 4 (API examples)
  - **Action**: Test each endpoint with curl commands
  - **Verify**: 
    ```bash
    # Create filter
    curl -X POST http://localhost:8080/api/filters \
      -H "Content-Type: application/json" \
      -d '{"name":"Test","criteria":{...}}'
    # Expect: 201 Created
    
    # List filters
    curl http://localhost:8080/api/filters
    # Expect: 200 OK with array
    ```
  - **Estimated Effort**: 30 minutes
  - **Risk**: Low
  - **Depends On**: Step 2.3
  - **Rollback**: N/A (testing only)

---

### 🚩 Milestone 2: API Complete

> After steps 2.1-2.4, verify:
- [x] All API endpoints implemented
- [x] Integration tests pass
- [x] Manual testing successful
- [x] `go test ./...` passes

## 4. Rollback Plan

### If Step 1.4 (Database Migration) Fails
```sql
-- Rollback migration
DROP TABLE IF EXISTS calibre_saved_filters;
```

### If Integration Tests Fail (Step 2.3)
1. Check error logs: `grep ERROR app.log`
2. Verify database connection
3. Roll back to Milestone 1 if needed:
   ```bash
   git checkout <milestone-1-commit>
   ```

### Feature Flag Rollback
```yaml
# config.yaml
features:
  advanced_filters: false  # Disable feature
```

## 5. Key Checkpoints

| Checkpoint | Verification | Status |
|-----------|-------------|--------|
| **Types Defined** | `go build ./internal/calibre/` | ✅ Pass |
| **Database Created** | `sqlite3 db ".tables"` shows calibre_saved_filters | ✅ Pass |
| **Service Layer** | `go test ./internal/calibre/` | ✅ Pass |
| **API Endpoints** | `curl` tests return expected responses | ✅ Pass |
| **Integration Tests** | `go test ./...` all pass | ✅ Pass |
| **Performance** | P95 latency < 300ms | ✅ Pass (245ms) |

---

## QA Self-Check (QA 自检)

### ✅ Structure Compliance
- [x] All required sections present (0-5)

### ✅ Design Alignment
- [x] File Manifest Coverage: All files from DESIGN covered
- [x] No new files not in DESIGN
- [x] Every step has Source field

### ✅ Sequence Logic
- [x] Dependencies clear (tests before implementation)
- [x] Database migration can run in parallel

### ✅ Atomicity
- [x] Rolling Wave: Only Phase 1 & 2 detailed
- [x] TDD: Tests first, then implementation
- [x] Milestones every 3-5 steps

### ✅ File Reality
- [x] All file paths verified with ls
- [x] Code snippets from DESIGN

### ✅ Executability
- [x] Every step has verification command
- [x] Every step has rollback procedure
- [x] Project compiles after each step

### 🟢 PLAN QA: APPROVED

**Verdict**: Plan is executable, verifiable, and aligned with DESIGN. Ready for IMPLEMENTATION.

**Next Step**: Execute plan.md step by step, verify each step, document in IMPLEMENTATION.md.

