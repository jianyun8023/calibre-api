# 📝 Phase 1: REQUIREMENTS - Requirements Definition

> **Role**: Technical Product Manager  
> **Objective**: Transform vague ideas into testable, unambiguous requirements  
> **Language**: English instructions, bilingual section titles, Gherkin scenarios in English, explanations in Chinese

---

## Instructions for AI

**Core Principles**:
1. **Problem > Solution**: Focus on WHAT and WHY, not HOW
2. **Zero Ambiguity**: Every requirement must be verifiable
3. **Testability**: Use Gherkin (Given-When-Then) for acceptance criteria
4. **Explicit Boundaries**: Define what's NOT included

---

## 0. 上下文参考 (Context Reference)

> **Link to PREWORK**: `prework.md`

### 0.1 继承的关键约束 (Inherited Key Constraints)

**From PREWORK Section 6**:
1. 
2. 
3. 

---

## 1. 问题空间分析 (Problem Space Analysis)

### 1.1 核心痛点 (Core Pain Point)

**What friction does the user face? What task can they NOT do today?**

当前问题 (Current Problem):
- 

用户影响 (User Impact):
- 

### 1.2 要完成的任务 (Job To Be Done - JTBD)

**Format**: When [context], I want to [motivation], so that [expected outcome]

**中文**: 当[上下文]时，我想要[动机]，以便[预期结果]

- **When**: 
- **I want to**: 
- **So that**: 

### 1.3 战略价值 (Strategic Value)

**Why solve this NOW?**

选择理由 (Rationale):
- [ ] **Compliance**: Legal or regulatory requirement
- [ ] **Revenue**: Direct revenue impact
- [ ] **Retention**: Prevents user churn
- [ ] **Efficiency**: Saves significant time/cost
- [ ] **Foundation**: Enables future features

**Priority Justification**:


---

## 2. 领域术语表 (Domain Glossary)

**Ubiquitous Language**: Define domain terms used throughout this spec

| 术语 (Term) | 定义 (Definition) | 同义词 (Synonyms) | 备注 (Notes) |
|------------|------------------|------------------|-------------|
|  |  |  |  |

---

## 3. 用户故事 (User Stories)

**Format**: As a [role], I want to [action], so that [benefit]

**中文格式**: 作为[角色]，我想要[行动]，以便[收益]

### 3.1 User Story Table

| ID | 角色 (Role) | 功能 (Feature) | 收益 (Benefit) | 优先级 (Priority) | 工作量 (Effort) |
|----|------------|---------------|---------------|------------------|---------------|
| US-001 |  |  |  | Must Have | M |

### 3.2 Priority Legend (MoSCoW)

- **Must Have** (必须有): MVP critical, cannot release without
- **Should Have** (应该有): Important but not critical
- **Could Have** (可以有): Nice to have
- **Won't Have** (不包含): Explicitly NOT in this iteration

### 3.3 Effort Legend

- **XS**: < 1 hour
- **S**: 1-4 hours
- **M**: 4-8 hours (half day to 1 day)
- **L**: 1-2 days
- **XL**: > 2 days (needs breakdown)

---

## 4. 功能需求和验收标准 (Functional Requirements & Acceptance Criteria)

**Format**: Gherkin (Given-When-Then) - Use ENGLISH for Gherkin scenarios

### Feature: [US-001] [Feature Name]

#### Scenario 1: Happy Path

```gherkin
Given [initial state/precondition]
When [user action or system event]
Then [expected result/outcome]
```

**中文说明 (Chinese Explanation)**:
- 

#### Scenario 2: Error Handling

```gherkin
Given [error condition]
When [action that triggers error]
Then [graceful error handling]
  And [system remains stable]
```

**中文说明 (Chinese Explanation)**:
- 

#### Scenario 3: Edge Case

```gherkin
Given [edge case condition]
When [action]
Then [expected behavior]
```

**中文说明 (Chinese Explanation)**:
- 

### Feature: [US-002] [Feature Name]

[Repeat format above]

---

## 5. 数据需求 (Data Requirements)

**Business perspective**: What data entities and relationships are needed?

### 5.1 核心实体 (Core Entities)

**Entity**: [EntityName]
- **属性 (Attributes)**:
  - 
- **业务规则 (Business Rules)**:
  - 
- **验证规则 (Validation Rules)**:
  - 

### 5.2 实体关系 (Entity Relationships)

| Entity A | Relationship | Entity B | 说明 (Notes) |
|----------|-------------|----------|-------------|
|  | One-to-Many |  |  |

---

## 6. 非功能性需求 (Non-Functional Requirements - NFRs)

### 6.1 性能 (Performance)

- **Response Time**: P95 latency < [X]ms
- **Throughput**: Support [X] requests per second
- **Concurrency**: Handle [X] concurrent users

### 6.2 安全性 (Security)

- **Authentication**: 
- **Authorization**: 
- **Data Encryption**: 
- **Input Validation**: 

### 6.3 可靠性 (Reliability)

- **Uptime**: [X]% availability
- **Data Retention**: 
- **Backup**: 
- **Error Recovery**: 

### 6.4 可用性 (Usability)

- **Mobile Responsive**: Yes/No
- **Accessibility**: WCAG 2.1 AA compliance
- **Browser Support**: 
- **Internationalization**: 

### 6.5 可维护性 (Maintainability)

- **Logging**: 
- **Monitoring**: 
- **Documentation**: 

---

## 7. 范围外 / 未来工作 (Out of Scope / Future Work)

**Explicitly define what we are NOT building**

- ❌ 
- ❌ 
- ❌ 

**Future Enhancements** (not in this iteration):
- 🔮 
- 🔮 

---

## 8. 跨模块依赖 (Cross-Module Dependencies)

| 依赖项 (Dependency) | 所属模块 (Owner Module) | 接口/契约 (Interface/Contract) | 状态 (Status) | 预计时间 (ETA) |
|-------------------|----------------------|-------------------------------|--------------|---------------|
|  |  |  | ✅ Available |  |

**Blocking Dependencies** (must exist before we start):
- 

**Soft Dependencies** (nice to have):
- 

---

## 9. 验收标准总结 (Acceptance Criteria Summary)

> **For Phase 5 (ACCEPTANCE)**: High-level checklist for stakeholder sign-off

### Must-Have Criteria (Must all pass)

- [ ] **Core Function**: User can [main capability]
- [ ] **Error Handling**: System gracefully handles [error case]
- [ ] **Performance**: Meets NFR targets (response time, throughput)
- [ ] **Security**: Authentication and authorization working
- [ ] **Data Integrity**: No data loss or corruption

### Should-Have Criteria (Nice to have)

- [ ] **User Experience**: UI intuitive and responsive
- [ ] **Accessibility**: WCAG 2.1 AA compliant
- [ ] **Documentation**: User docs and API docs complete

---

## QA Checklist: REQUIREMENTS Critical Review

### S-DEEP-CT Model

#### 0. **S**trategic Effectiveness ("Why" Test) 🔴 CRITICAL

- [ ] **XY Problem Detection**: Does requirement specify UI solution (like "add modal") instead of solving user problem?
- [ ] **Occam's Razor**: Is proposed solution the simplest way?
- [ ] **Value Alignment**: Does this feature serve project's core user persona?

#### 1. **D**efinition Completeness (Testability)

- [ ] **Gherkin Compliance**: All scenarios in strict `Given-When-Then` format?
- [ ] **Quantifiable**: Vague terms ("fast", "reliable") replaced with metrics?
- [ ] **Negative Scenarios**: Each feature has at least one error/sad path?

#### 2. **E**dge Cases (Robustness)

- [ ] **Concurrency**: What happens if two users act simultaneously?
- [ ] **Connectivity**: What if network fails mid-operation?
- [ ] **Data Limits**: Covered empty input, max-length input, special characters?

#### 3. **E**xplicit Scope (Boundaries)

- [ ] **Out of Scope**: Section 7 explicitly states what we're NOT building?
- [ ] **Permissions**: RBAC rules for each operation defined?

#### 4. **P**recision (Ambiguity)

- [ ] **Glossary Check**: Domain terms defined in Section 2?
- [ ] **ID References**: All user stories have unique IDs?
- [ ] **Priority Check**: All user stories marked with MoSCoW priority?
- [ ] **Effort Check**: All user stories have effort estimates?

#### 5. **C**ompleteness (MECE - Mutually Exclusive, Collectively Exhaustive)

- [ ] **Missing Flows**: Any "dead ends" in user flow?
- [ ] **NFRs Complete**: Section 6 covers performance, security, reliability?
- [ ] **Acceptance Criteria**: Section 9 exists and complete?

#### 6. **T**raceability (Context & Dependencies)

- [ ] **Context Reference**: Section 0 exists and references PREWORK?
- [ ] **Inherited Constraints**: Section 0.1 lists constraints from PREWORK Section 6?
- [ ] **Cross-Module Dependencies**: Section 8 identifies all external dependencies?

### QA Verdict

**Status**: ⚪ Not Started | 🟡 In Progress | 🔴 Rejected | 🟡 Needs Revision | 🟢 Approved

**Issues Found**:
- 

**Approval Criteria**:
- All S-DEEP-CT checks pass
- No critical ambiguities remain
- All Gherkin scenarios testable

---

## Phase Completion

**REQUIREMENTS Status**: ⚪ Not Started | 🟡 In Progress | 🟢 Complete | 🔴 Needs Revision

**Next Action**: After QA passes, advance to DESIGN phase

**Verification**: 
- ✅ All user stories have acceptance criteria
- ✅ All Gherkin scenarios testable
- ✅ No ambiguous requirements
- ✅ Constraints from PREWORK incorporated

---

**Phase Transition**: When complete and QA approved, AI should:
1. Update spec frontmatter: add REQUIREMENTS to `phase_history` with status APPROVED
2. Update spec frontmatter: set `phase` to DESIGN
3. Ask user: "REQUIREMENTS approved. Proceed to DESIGN phase?"

