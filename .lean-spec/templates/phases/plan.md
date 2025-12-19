# 📅 Phase 3: PLAN - Implementation Plan

> **Role**: Senior Engineering Manager  
> **Objective**: Transform design into linear, safe, verifiable execution steps  
> **Language**: English instructions, step instructions in English, comments can be bilingual

---

## Instructions for AI

**Core Principles**:
1. **Atomic Steps**: Each step completable in one AI turn or one commit
2. **Dependency Order**: Satisfy dependencies (Schema → Service → Handler → UI)
3. **Verifiable Checkpoints**: Each step has clear pass/fail verification
4. **Green-to-Green**: Project compiles and runs after every step
5. **TDD Workflow**: Write test → Implement → Verify

---

## 0. 设计参考和对齐 (Design Reference & Alignment)

> **Key**: This plan implements design in `design.md`

### 0.1 设计文档链接 (Design Document Links)

- **DESIGN**: `design.md`
- **REQUIREMENTS**: `requirements.md`
- **PREWORK**: `prework.md`

**Covered User Stories**: US-001, US-002, ...

### 0.2 文件清单覆盖检查 (File Manifest Coverage Check)

**Verify all files from DESIGN Section 9 are covered by this plan**

| File (from DESIGN Section 9) | Corresponding Step | Status |
|------------------------------|-------------------|--------|
| `internal/domain/entity.go` | Step 1.1 | ✅ Covered |
| `internal/service/entity_service.go` | Step 1.3 | ✅ Covered |

### 0.3 前置条件 (Prerequisites)

**What must be completed before starting this plan?**

#### Dependencies to Install

```bash
# Example for Go
go get github.com/some/package

# Example for Node.js
npm install some-package
```

#### Environment Variables

Add to `.env` or environment:
```bash
VARIABLE_NAME=value
```

#### Setup Commands

```bash
# Database migrations
make migrate-up

# Build project
make build
```

#### Blocking Tasks

- ✅ None / ❌ [Link to blocking task or spec]

---

## 1. 高层路线图 (High-Level Roadmap)

- [ ] **Phase 1**: [Name, e.g., "Core API"] (detailed below)
- [ ] **Phase 2**: [Name, e.g., "Frontend Integration"] (pending Phase 1)
- [ ] **Phase 3**: [Name, e.g., "Polish & Edge Cases"] (future)

**Current Focus**: Only Phase 1 is planned in detail below

---

## 2. 详细执行计划：Phase 1 (Detailed Execution Plan: Phase 1)

**Context**: Only plan atomic steps for this phase. Follow TDD workflow.

### Step 1.1: Setup & Test (Red) - [Feature Name]

- [ ] **Create test file**

**Source**: DESIGN Section 7.1 (Unit Tests)

**Action**: Create `[path/to/test/file]`

**Content**: Write failing test cases for `[feature_name]`

**Test Code**:
```go
// Example for Go
// File: internal/service/entity_service_test.go
package service

import "testing"

func TestEntityService_Create(t *testing.T) {
    // Test implementation
    t.Fatal("not implemented")
}
```

**Verify**: 
```bash
go test ./internal/service/... -run TestEntityService_Create
# Expected: FAIL (Red - test not implemented yet)
```

**Estimated Effort**: 15 minutes

**Risk**: Low

**Depends On**: None

**Rollback**: 
```bash
git checkout -- internal/service/entity_service_test.go
```

---

### Step 1.2: Create Data Model (Green)

- [ ] **Implement entity model**

**Source**: DESIGN Section 3.1 (Data Model), Section 9.3 (Type Definitions)

**Action**: Create `[path/to/model/file]`

**Code Snippet** (from DESIGN):
```go
// Copy from DESIGN Section 9.3
package domain

type Entity struct {
    ID        string    `json:"id"`
    Name      string    `json:"name" binding:"required,max=100"`
    CreatedAt time.Time `json:"created_at"`
}

func (e *Entity) Validate() error {
    // Validation logic
    return nil
}
```

**Verify**: 
```bash
go build ./internal/domain/...
# Expected: Success (compiles without errors)
```

**Estimated Effort**: 20 minutes

**Risk**: Low

**Depends On**: None (can run in parallel with Step 1.1)

**Rollback**: 
```bash
git checkout -- internal/domain/entity.go
```

---

### Step 1.3: Create Service Interface

- [ ] **Implement service layer**

**Source**: DESIGN Section 4.2 (Function Signatures), Section 9.4 (API Signatures)

**Action**: Create `[path/to/service/file]`

**Code Snippet** (from DESIGN):
```go
// Copy from DESIGN Section 9.4
package service

type EntityService struct {
    db Database
}

func NewEntityService(db Database) *EntityService {
    return &EntityService{db: db}
}

func (s *EntityService) Create(ctx context.Context, entity *domain.Entity) error {
    // Validation
    if err := entity.Validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    // Persist to database
    return s.db.Insert(ctx, entity)
}
```

**Verify**: 
```bash
go test ./internal/service/... -run TestEntityService_Create
# Expected: PASS (Green - test now passes)
```

**Estimated Effort**: 30 minutes

**Risk**: Medium

**Depends On**: Step 1.1 (test file), Step 1.2 (model)

**Rollback**: 
```bash
git checkout -- internal/service/entity_service.go
```

---

### Step 1.4: Database Schema Update

- [ ] **Update database schema**

**Source**: DESIGN Section 3 (Data Model)

**Action**: Create migration file `[path/to/migration]`

**Code Snippet** (from DESIGN):
```sql
-- Migration: 001_create_entity_table.up.sql
CREATE TABLE IF NOT EXISTS entity (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_entity_name ON entity(name);
```

**Rollback Migration**:
```sql
-- Migration: 001_create_entity_table.down.sql
DROP INDEX IF EXISTS idx_entity_name;
DROP TABLE IF EXISTS entity;
```

**Verify**: 
```bash
# Run migration
make migrate-up

# Check table exists
sqlite3 database.db "SELECT name FROM sqlite_master WHERE type='table' AND name='entity';"
# Expected: entity
```

**Follow-up**: Run migration command

**Estimated Effort**: 20 minutes

**Risk**: Medium (DB change)

**Depends On**: None

**Rollback**: 
```bash
make migrate-down
```

---

### Step 1.5: Create HTTP Handler

- [ ] **Implement HTTP handlers**

**Source**: DESIGN Section 4.1 (Public API)

**Action**: Create `[path/to/handler/file]`

**Code Snippet**:
```go
package handler

type EntityHandler struct {
    service *service.EntityService
}

func NewEntityHandler(service *service.EntityService) *EntityHandler {
    return &EntityHandler{service: service}
}

func (h *EntityHandler) Create(c *gin.Context) {
    var entity domain.Entity
    if err := c.ShouldBindJSON(&entity); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    if err := h.service.Create(c.Request.Context(), &entity); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(201, gin.H{"success": true, "data": entity})
}
```

**Verify**: 
```bash
go build ./internal/handler/...
# Expected: Success
```

**Estimated Effort**: 25 minutes

**Risk**: Low

**Depends On**: Step 1.3 (service)

**Rollback**: 
```bash
git checkout -- internal/handler/entity_handler.go
```

---

### Step 1.6: Register Routes

- [ ] **Register HTTP routes**

**Source**: DESIGN Section 9.2 (Files to Modify)

**Action**: Update `internal/router.go`

**Code Snippet**:
```go
// Add to router setup
func SetupRouter() *gin.Engine {
    r := gin.Default()
    
    // Existing routes...
    
    // New entity routes
    entityHandler := handler.NewEntityHandler(entityService)
    r.POST("/api/entity", entityHandler.Create)
    r.GET("/api/entity/:id", entityHandler.GetByID)
    
    return r
}
```

**Verify**: 
```bash
# Build and run server
go build && ./calibre-api

# Test endpoint
curl -X POST http://localhost:8080/api/entity \
  -H "Content-Type: application/json" \
  -d '{"name":"test"}'
# Expected: 201 Created
```

**Estimated Effort**: 15 minutes

**Risk**: Low

**Depends On**: Step 1.5 (handler)

**Rollback**: 
```bash
git checkout -- internal/router.go
```

---

### 🚩 Milestone 1: Core API Complete

**After steps 1.1-1.6, verify**:

- [ ] Tests exist and pass
- [ ] Database schema created
- [ ] API endpoints respond correctly
- [ ] Project builds successfully: `go build`
- [ ] Integration test passes

**Verification Command**:
```bash
# Run all tests
go test ./...

# Build project
go build

# Manual API test
curl http://localhost:8080/api/entity
```

---

## 3. 回滚计划 (Rollback Plan)

### Full Rollback Strategy

**If critical issues discovered**:

1. **Stop deployment** immediately
2. **Revert code changes**:
   ```bash
   git revert <commit-hash>
   # Or hard reset (if not pushed to main)
   git reset --hard HEAD~1
   ```
3. **Rollback database migrations**:
   ```bash
   make migrate-down
   ```
4. **Verify rollback**:
   ```bash
   go test ./...
   go build
   ```

### Partial Rollback

**If specific step fails**:
- Use step-specific rollback command from that step
- Re-run verification to ensure stability
- Fix issue and retry step

---

## 4. 关键检查点 (Key Checkpoints)

### After Each Step

- [ ] Step verification command passes
- [ ] Project still compiles: `go build`
- [ ] Existing tests still pass: `go test ./...`
- [ ] No new linter errors: `golangci-lint run`

### After Each Milestone

- [ ] All milestone verification criteria met
- [ ] Integration tests pass
- [ ] Manual testing confirms functionality
- [ ] Performance within acceptable range

---

## QA Checklist: PLAN Pre-Flight Check

### SAFE-RUN-D Model

#### 0. **S**tructure Compliance 🔴 CRITICAL

- [ ] Section 0: Design Reference & Alignment
- [ ] Section 0.1: Design Document Links
- [ ] Section 0.2: File Manifest Coverage Check
- [ ] Section 0.3: Prerequisites
- [ ] Section 1: High-Level Roadmap
- [ ] Section 2: Detailed Execution Plan (Phase 1 only)
- [ ] Section 3: Rollback Plan
- [ ] Section 4: Key Checkpoints

#### 1. **D**esign Alignment (Traceability) 🔴 CRITICAL

- [ ] **File Manifest Coverage**: Section 0.2 lists all files from DESIGN Section 9?
- [ ] **No New Files**: PLAN doesn't introduce files not in DESIGN?
- [ ] **Source Traceability**: Every step has `Source` field linking to DESIGN?
- [ ] **No Architecture Decisions**: PLAN doesn't make decisions belonging in DESIGN?

#### 2. **S**equence Logic (Dependency Check)

- [ ] **Dependency Order**: Schema before Service, Service before Handler, Handler before Routes?
- [ ] **Depends On Field**: Every step has `Depends On` field?
- [ ] **Prerequisites Check**: Section 0.3 lists all packages to install?

#### 3. **A**tomicity & Complexity

- [ ] **Rolling Wave**: Detailed plan limited to Phase 1?
- [ ] **TDD Enforcement**: Phase 1 starts with test creation (Red)?
- [ ] **Milestones**: Milestone checkpoint every 3-5 steps?
- [ ] **Step Size**: Each step completable in < 30 minutes?

#### 4. **F**ile Reality & Context (Hallucination Check)

- [ ] **Path Verification**: Files in "edit" steps actually exist?
- [ ] **Naming Conventions**: New files follow project conventions?
- [ ] **Code Snippets**: Complex steps have code snippets from DESIGN?

#### 5. **E**xecutability (Green-to-Green)

- [ ] **Compile Safety**: Project compiles after each step?
- [ ] **Migration Safety**: Plan includes migration command after schema changes?
- [ ] **Rollback Field**: Every step has `Rollback` command?

#### 6. **R**un & Verification (Testability)

- [ ] **Explicit Verification**: Every step has concrete verification command?
- [ ] **Test Coverage**: Verification commands actually test the implementation?
- [ ] **Effort Estimates**: Every step has `Estimated Effort`?

#### 7. **U**nambiguous Steps (Clarity Check)

- [ ] **No Vague Actions**: All actions specific and explicit?
- [ ] **No Missing Details**: Junior developer can execute without questions?

### QA Verdict

**Status**: ⚪ Not Started | 🟡 In Progress | 🔴 Rejected | 🟡 Needs Revision | 🟢 Approved

**Issues Found**:
- 

---

## Phase Completion

**PLAN Status**: ⚪ Not Started | 🟡 In Progress | 🟢 Complete | 🔴 Needs Revision

**Next Action**: After QA passes, advance to IMPLEMENTATION phase

**Verification**: 
- ✅ All files from DESIGN Section 9 covered
- ✅ Dependencies ordered correctly
- ✅ Each step has verification command
- ✅ Ready for implementation (no ambiguity)

---

**Phase Transition**: When complete and QA approved, AI should:
1. Update spec frontmatter: add PLAN to `phase_history` with status APPROVED
2. Update spec frontmatter: set `phase` to IMPLEMENTATION
3. Update spec frontmatter: set `current_action` to first step description
4. Ask user: "PLAN approved. Proceed to IMPLEMENTATION phase?"

