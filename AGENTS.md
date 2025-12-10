# AI Agent Instructions

## Project: calibre-api

Calibre API 是一个基于 Go 的书籍管理系统，集成了语义搜索、智能问答和 MCP 协议支持。

## 📚 项目文档导航

**核心指南**:
- `CLAUDE.md` - 完整的 AI 助手开发指南（架构、代码规范、开发流程）
- `specs/` - 功能规格文档（使用 LeanSpec 管理）
- `app/AGENTS.md` - 前端开发指南（Vue.js 3）

**规格文档**:
- `specs/007-000-project-overview/` - 项目概览和架构
- `specs/001-book-management/` - 书籍管理功能
- `specs/002-search-functionality/` - 搜索功能（混合搜索策略）
- `specs/003-mcp-integration/` - MCP 协议集成
- `specs/004-chat-agent/` - 智能问答
- `specs/005-qdrant-vector-search/` - 向量搜索
- `specs/006-task-management/` - 异步任务管理

## 🚨 CRITICAL: Before ANY Task

**STOP and check these first:**

1. **Discover context** → Use `board` tool to see project state
2. **Search for related work** → Use `search` tool before creating new specs
3. **Never create files manually** → Always use `create` tool for new specs

> **Why?** Skipping discovery creates duplicate work. Manual file creation breaks LeanSpec tooling.

## 🔧 Managing Specs

### MCP Tools (Preferred) with CLI Fallback

| Action | MCP Tool | CLI Fallback |
|--------|----------|--------------|
| Project status | `board` | `lean-spec board` |
| List specs | `list` | `lean-spec list` |
| Search specs | `search` | `lean-spec search "query"` |
| View spec | `view` | `lean-spec view <spec>` |
| Create spec | `create` | `lean-spec create <name>` |
| Update spec | `update` | `lean-spec update <spec> --status <status>` |
| Link specs | `link` | `lean-spec link <spec> --depends-on <other>` |
| Unlink specs | `unlink` | `lean-spec unlink <spec> --depends-on <other>` |
| Dependencies | `deps` | `lean-spec deps <spec>` |
| Token count | `tokens` | `lean-spec tokens <spec>` |

## ⚠️ Core Rules

| Rule | Details |
|------|---------|
| **NEVER edit frontmatter manually** | Use `update`, `link`, `unlink` for: `status`, `priority`, `tags`, `assignee`, `transitions`, timestamps, `depends_on` |
| **ALWAYS link spec references** | Content mentions another spec → `lean-spec link <spec> --depends-on <other>` |
| **Track status transitions** | `planned` → `in-progress` (before coding) → `complete` (after done) |
| **No nested code blocks** | Use indentation instead |

### 🚫 Common Mistakes

| ❌ Don't | ✅ Do Instead |
|----------|---------------|
| Create spec files manually | Use `create` tool |
| Skip discovery | Run `board` and `search` first |
| Leave status as "planned" | Update to `in-progress` before coding |
| Edit frontmatter manually | Use `update` tool |

## 📋 SDD Workflow

```
BEFORE: board → search → check existing specs
DURING: update status to in-progress → code → document decisions → link dependencies
AFTER:  update status to complete → document learnings
```

**Status tracks implementation, NOT spec writing.**

## Spec Dependencies

Use `depends_on` to express blocking relationships between specs:
- **`depends_on`** = True blocker, work order matters, directional (A depends on B)

Link dependencies when one spec builds on another:
```bash
lean-spec link <spec> --depends-on <other-spec>
```

## When to Use Specs

| ✅ Write spec | ❌ Skip spec |
|---------------|--------------|
| Multi-part features | Bug fixes |
| Breaking changes | Trivial changes |
| Design decisions | Self-explanatory refactors |

## Token Thresholds

| Tokens | Status |
|--------|--------|
| <2,000 | ✅ Optimal |
| 2,000-3,500 | ✅ Good |
| 3,500-5,000 | ⚠️ Consider splitting |
| >5,000 | 🔴 Must split |

## First Principles (Priority Order)

1. **Context Economy** - <2,000 tokens optimal, >3,500 needs splitting
2. **Signal-to-Noise** - Every word must inform a decision
3. **Intent Over Implementation** - Capture why, let how emerge
4. **Bridge the Gap** - Both human and AI must understand
5. **Progressive Disclosure** - Add complexity only when pain is felt

---

**Remember:** LeanSpec tracks what you're building. Keep specs in sync with your work!