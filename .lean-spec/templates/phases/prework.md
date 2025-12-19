# 🕵️ Phase 0: PREWORK - Context Gathering

> **Role**: Context Detective  
> **Objective**: Gather project context to prevent hallucinations and ensure design fits reality  
> **Language**: English instructions, bilingual section titles, Chinese content

---

## Instructions for AI

**Before generating requirements, design, or code, you MUST understand the existing codebase.**

Follow this protocol: **DNA → Trace → Align**

---

## 1. 项目 DNA 分析 (Project DNA Analysis)

### 1.1 框架和技术栈 (Framework & Tech Stack)

**Action**: Read manifest files (`package.json`, `go.mod`, `requirements.txt`, `Cargo.toml`)

**Framework**:
- Main framework: 
- Version: 

**Core Libraries & Dependencies**:
- 
- 
- 

**Language & Runtime**:
- Language: 
- Version: 

### 1.2 配置和约束 (Configuration & Constraints)

**Action**: Read config files (`tsconfig.json`, `config.yaml`, `.env.example`)

**Configuration Files Analyzed**:
- 
- 

**Key Constraints**:
- Build configuration: 
- Path aliases: 
- Environment variables: 
- 

### 1.3 目录结构 (Directory Structure)

**Action**: Use `ls` to scan directory structure (depth 2-3)

**Project Layout**:
```
[Paste output of ls command showing key directories]
```

**Code Organization Pattern**:
- Source code location: 
- Test location: 
- Configuration location: 

---

## 2. 语义追踪 (Semantic Tracing)

### 2.1 关键词扩展 (Keyword Expansion)

**User Request Keywords**:
- Primary: 
- Related terms: 
- Domain concepts: 

**Search Strategy**:
```bash
# Commands used for semantic search
grep -r "..." 
ls -la ...
```

### 2.2 依赖追踪 (Dependency Tracing)

**Target Component/Module**:
- Component name: 
- File location: 

**Upstream Dependencies** (Who calls this?):
- 
- 

**Downstream Dependencies** (What does this call?):
- 
- 

**Blast Radius** (What breaks if we change this?):
- 

### 2.3 模式匹配 (Pattern Matching)

**Similar Existing Features**:

| Feature | Location | Pattern Worth Copying |
|---------|----------|---------------------|
|  |  |  |

**Anti-Patterns to Avoid**:
- 

---

## 3. 规格对齐 (Spec Alignment)

### 3.1 集成点分析 (Integration Points)

**Database Schema** (if applicable):
- Schema file: 
- Relevant tables/models: 
- Relationships: 

**API Registry** (if applicable):
- Router/handler location: 
- Existing endpoints: 

**UI Components** (if applicable):
- Component library: 
- Reusable components: 

### 3.2 可重用性检查 (Reusability Check)

**Components We Can Reuse**:

| Component | Location | Purpose |
|-----------|----------|---------|
|  |  |  |

**Components We Need to Create**:
- 

### 3.3 文件路径验证 (File Path Verification)

**Verification Commands Used**:
```bash
ls ...
grep -r "..." ...
```

**Verified File Paths**:
- ✅ 
- ✅ 

**Missing/Unverified Paths**:
- ❌ 

---

## 4. 缺口分析 (Gap Analysis)

### 4.1 现有功能 (What Exists)

- ✅ 
- ✅ 

### 4.2 缺失功能 (What's Missing)

- [ ] 
- [ ] 

### 4.3 需要修改的部分 (What Needs Modification)

- 🔧 
- 🔧 

---

## 5. 风险和约束 (Risks & Constraints)

### 5.1 技术约束 (Technical Constraints)

**MUST Follow**:
- 
- 

**MUST NOT**:
- 
- 

### 5.2 风险 (Risks)

| Risk | Impact | Mitigation |
|------|--------|-----------|
|  | High/Medium/Low |  |

### 5.3 性能考虑 (Performance Considerations)

- 
- 

---

## 6. 给下游阶段的关键约束 (Key Constraints for Downstream Phases)

**CRITICAL**: These constraints MUST be referenced in REQUIREMENTS and DESIGN phases.

1. **Architecture Constraint**: 
   - Example: Must use existing Modal component from `src/components/ui/dialog.tsx`

2. **Authentication Pattern**: 
   - Example: Use `ctx.session.user.id` (no custom auth)

3. **State Management**: 
   - Example: Zustand (not Redux or Context API)

4. **API Pattern**: 
   - Example: RESTful endpoints in `internal/calibre/` handlers

5. **Code Style**: 
   - Example: Follow existing error handling pattern with `pkg/errors`

---

## 7. 验证命令记录 (Verification Commands Log)

**Document all commands used to verify context**:

```bash
# Framework verification
cat go.mod
cat package.json

# Directory structure
ls -la
tree -L 2

# Search for similar features
grep -r "similar_pattern" src/

# Verify file existence
ls -la path/to/expected/file
```

---

## QA Checklist: PREWORK Fact Check

### Self-Review Before Submitting

- [ ] **Project DNA Check**: Identified framework, version, and key dependencies?
- [ ] **No Hallucinations**: Did NOT suggest libraries not in project dependencies?
- [ ] **File Verification**: Used `ls` or `grep` to verify ALL mentioned file paths?
- [ ] **Pattern Discovery**: Found at least 1 similar existing feature to copy patterns from?
- [ ] **Semantic Trace**: Traced upstream callers AND downstream dependencies?
- [ ] **Reusability**: Checked for reusable components before suggesting new ones?
- [ ] **Constraints Listed**: Section 6 has at least 3 specific constraints for downstream phases?
- [ ] **Commands Logged**: Section 7 documents all verification commands used?

### Rejection Criteria

**Automatic Rejection If**:
- ❌ Suggests using `axios` but `package.json` only has `fetch` or `ky`
- ❌ Assumes `src/utils/api.ts` exists without running `ls` to verify
- ❌ Plans to create new Table component when `components/ui/table.tsx` already exists
- ❌ Didn't read database schema before suggesting schema changes

---

## Phase Completion

**PREWORK Status**: ⚪ Not Started | 🟡 In Progress | 🟢 Complete | 🔴 Needs Revision

**Next Action**: After QA passes, advance to REQUIREMENTS phase

**Verification**: 
- ✅ All file paths verified
- ✅ Project constraints documented
- ✅ Similar patterns identified
- ✅ No unverified assumptions

---

**Phase Transition**: When complete and QA approved, AI should:
1. Update spec frontmatter: add PREWORK to `phase_history` with status APPROVED
2. Update spec frontmatter: set `phase` to REQUIREMENTS
3. Ask user: "PREWORK complete. Proceed to REQUIREMENTS phase?"

