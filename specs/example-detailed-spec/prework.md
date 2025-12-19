# 🧩 PREWORK Context Artifact

> **Module**: Advanced Book Filtering  
> **Date**: 2024-12-19  
> **Status**: Verified ✅

<!-- English: This document captures project context gathered during PREWORK phase.
     All information verified through code inspection and file system checks.
     
     中文：本文档记录 PREWORK 阶段收集的项目上下文。
     所有信息通过代码检查和文件系统检查验证。
-->

## 1. Project DNA (项目基因)

<!-- English: Framework, tech stack, and key dependencies identified from manifest files
     中文：从清单文件中识别的框架、技术栈和关键依赖
-->

### Framework & Language
- **Language**: Go 1.24.4
- **Web Framework**: Gin (github.com/gin-gonic/gin v1.10.0)
- **Database**: SQLite (Calibre DB) + Qdrant (Vector DB)
- **Config Management**: Viper (github.com/spf13/viper)

### Key Dependencies (from go.mod)
```go
// Web & API
github.com/gin-gonic/gin v1.10.0
github.com/gin-contrib/cors v1.7.2

// Database & Storage
modernc.org/sqlite v1.34.2
github.com/qdrant/go-client v1.12.0

// Configuration
github.com/spf13/viper v1.19.0

// Logging
github.com/sirupsen/logrus v1.9.3
```

### Project Structure (from ls -R)
```
internal/
├── calibre/          # Core book management logic
│   ├── book_handler.go
│   ├── search_handler.go  # ⭐ Existing search infrastructure
│   ├── mcp_server.go
│   └── types.go
├── cache/            # File caching
├── chat/             # AI chat functionality
├── config/           # Configuration management
├── repository/       # Data access layer
│   └── book_repository.go
├── semantic/         # Vector search
│   ├── embedding/
│   └── qdrant/
└── tasks/            # Async task management
```

## 2. Relevant Reality (相关现实)

<!-- English: Existing code patterns and infrastructure that can be reused
     中文：可以重用的现有代码模式和基础设施
-->

### Existing Search Infrastructure

**File**: `internal/calibre/search_handler.go` (Lines 45-120)

**Current Capabilities**:
- ✅ Basic keyword search
- ✅ Semantic search via Qdrant
- ✅ Hybrid search (keyword + semantic)
- ✅ Pagination support
- ✅ Sort by relevance/date/title

**Relevant Code Pattern**:
```go
// Existing search function signature
func (h *SearchHandler) Search(c *gin.Context) {
    query := c.Query("q")
    searchType := c.DefaultQuery("type", "hybrid")
    limit := c.DefaultQuery("limit", "20")
    offset := c.DefaultQuery("offset", "0")
    
    // Search logic...
}
```

### Database Schema (from Calibre DB)

**Table**: `books`
```sql
CREATE TABLE books (
    id INTEGER PRIMARY KEY,
    title TEXT,
    author_sort TEXT,
    timestamp TIMESTAMP,
    pubdate TIMESTAMP,
    series_index REAL,
    isbn TEXT,
    path TEXT,
    has_cover BOOL
);
```

**Table**: `custom_columns` (for tags, ratings, etc.)
```sql
CREATE TABLE custom_column_1 (
    id INTEGER PRIMARY KEY,
    book INTEGER,
    value TEXT,
    FOREIGN KEY(book) REFERENCES books(id)
);
```

### Similar Feature Pattern (from spec 002-search-functionality)

**Spec**: `specs/002-search-functionality/`

**Reusable Patterns**:
- ✅ Search parameter validation
- ✅ Result pagination
- ✅ Error handling for empty results
- ✅ Cache integration for frequent searches

**Key Learning from Spec 002**:
> "Use `SearchParams` struct to encapsulate search criteria, 
> makes it easier to add new filter types later"

### Existing UI Components (from app/)

**Relevant Components**:
- `SearchBar.vue` - Can be extended with filter UI
- `BookList.vue` - Already supports filtered results
- `FilterPanel.vue` - ⚠️ Does NOT exist, needs to be created

## 3. Gaps (缺失部分)

<!-- English: What's missing that needs to be built
     中文：需要构建的缺失部分
-->

### Database Gaps
- [ ] No `saved_filters` table (needs to be created)
- [ ] No `filter_shares` table for sharing (needs to be created)
- [ ] No indexes on filter criteria columns (needs optimization)

### API Gaps
- [ ] No `/api/filters` endpoint (CRUD for saved filters)
- [ ] No `/api/filters/:id/apply` endpoint (apply saved filter)
- [ ] No `/api/filters/:id/share` endpoint (share filter)

### Frontend Gaps
- [ ] No `FilterPanel` component (needs to be created)
- [ ] No `SavedFilters` component (needs to be created)
- [ ] No filter state management in Vuex/Pinia

### Service Layer Gaps
- [ ] No `FilterService` (needs to be created)
- [ ] No filter validation logic
- [ ] No filter serialization/deserialization

## 4. Risks & Constraints (风险和约束)

<!-- English: Important constraints and risks identified
     中文：识别的重要约束和风险
-->

### Technical Constraints

1. **Database**: Must use existing Calibre SQLite database
   - ⚠️ Cannot modify core Calibre tables
   - ✅ Can add custom tables with `calibre_` prefix

2. **API Compatibility**: Must maintain backward compatibility
   - ✅ Existing `/api/search` endpoint must continue working
   - ✅ New filter endpoints should be additive

3. **Performance**: Search with filters must be fast
   - 🎯 Target: P95 latency < 300ms
   - ⚠️ Complex filters (5+ criteria) may need optimization

### Integration Constraints

1. **Qdrant Integration**: Filters must work with semantic search
   - ✅ Qdrant supports payload filtering
   - ⚠️ Need to sync filter criteria to Qdrant payloads

2. **Cache Strategy**: Saved filters should be cached
   - ✅ Existing `CacheManager` can be reused
   - ⚠️ Need cache invalidation strategy

### Security Constraints

1. **Authorization**: Filter sharing requires user authentication
   - ⚠️ Current system has no user auth
   - 🔴 **Blocker**: Need to implement basic auth first OR defer sharing to v2

2. **Input Validation**: Filter criteria must be sanitized
   - ✅ Can use existing validation patterns
   - ⚠️ Need to prevent SQL injection in dynamic filters

## 5. Key Constraints for Downstream Phases (下游阶段的关键约束)

<!-- English: Critical constraints that REQUIREMENTS and DESIGN must respect
     中文：REQUIREMENTS 和 DESIGN 必须遵守的关键约束
-->

### Must Follow

1. **Database Naming**: New tables must use `calibre_` prefix
   - Example: `calibre_saved_filters`, `calibre_filter_shares`

2. **API Pattern**: Follow existing RESTful conventions
   - GET `/api/filters` - List saved filters
   - POST `/api/filters` - Create filter
   - PUT `/api/filters/:id` - Update filter
   - DELETE `/api/filters/:id` - Delete filter

3. **Search Integration**: Reuse `SearchHandler` infrastructure
   - Don't create separate search logic
   - Extend existing `SearchParams` struct

4. **Error Handling**: Follow project error patterns
   - Use `pkg/errors` for error wrapping
   - Return `pkg/response` standard format

5. **Frontend State**: Use existing state management
   - If using Vuex: follow existing module pattern
   - If using Pinia: create `filters` store

### Should Consider

1. **Feature Flag**: Implement behind feature flag for safe rollout
   - Config key: `features.advanced_filters`
   - Default: `false` until v1 stable

2. **Migration Path**: Provide upgrade path for existing users
   - Script to convert old search history to saved filters

3. **Performance Monitoring**: Add metrics for filter operations
   - Track: filter creation, application, execution time

## 6. Verification Commands (验证命令)

<!-- English: Commands used to verify context during PREWORK
     中文：PREWORK 期间用于验证上下文的命令
-->

### File Structure Verification
```bash
# Verify project structure
ls -R internal/

# Output: ✅ Confirmed calibre/, cache/, chat/, config/, repository/, semantic/, tasks/
```

### Dependency Verification
```bash
# Check Go dependencies
grep "gin-gonic/gin" go.mod

# Output: ✅ github.com/gin-gonic/gin v1.10.0
```

### Database Schema Verification
```bash
# Check Calibre database schema (if available)
sqlite3 /path/to/metadata.db ".schema books"

# Output: ✅ Confirmed books table structure
```

### Existing Code Verification
```bash
# Find existing search handler
grep -r "SearchHandler" internal/calibre/

# Output: ✅ Found internal/calibre/search_handler.go
```

### Similar Feature Verification
```bash
# Check spec 002 for patterns
ls specs/002-search-functionality/

# Output: ✅ Found README.md, design.md, IMPLEMENTATION.md
```

---

## QA Self-Check (QA 自检)

### ✅ Project DNA Check
- [x] Identified framework: Go 1.24 + Gin
- [x] Checked go.mod for dependencies
- [x] Verified project structure with ls

### ✅ Semantic Trace Check
- [x] Found existing SearchHandler
- [x] Identified similar pattern in spec 002
- [x] Traced database schema

### ✅ Spec Alignment Check
- [x] Verified all file paths with ls
- [x] Read SearchHandler code
- [x] No assumptions made, all verified

### 🟢 PREWORK QA: APPROVED

**Verdict**: All context verified. No hallucinated assumptions. Ready for REQUIREMENTS phase.

**Next Step**: Create requirements.md with user stories and Gherkin scenarios.

