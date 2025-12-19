---
status: planned
phase: PREWORK
phase_history: []
complexity: high
current_action: "Starting PREWORK phase - gathering project context"
created: '{date}'
tags: []
priority: high
---

# {name}

> **Status**: {status} · **Phase**: {phase} · **Complexity**: {complexity} · **Priority**: {priority} · **Created**: {date}

## 概述 (Overview)

<!-- English: Brief description of what problem this spec solves and why it matters now.
     This spec uses the detailed PDPI-spec workflow with strict phase gates.
     
     中文：简要描述本规格解决什么问题，为什么现在需要解决。
     本规格使用详细的 PDPI-spec 工作流，包含严格的阶段门控。
-->

## 文档索引 (Documents)

<!-- English: This spec follows the PDPI-spec workflow. Documents are created in phase order.
     All phase documents use English instructions with bilingual content.
     
     中文：本规格遵循 PDPI-spec 工作流。文档按阶段顺序创建。
     所有阶段文档使用英文指令和双语内容。
-->

### 阶段文档 (Phase Documents)

**Phase 0: PREWORK** (Context Gathering)
- [prework.md](prework.md) - 项目上下文分析 (Project Context Analysis)

**Phase 1: REQUIREMENTS** (Requirements Definition)
- [requirements.md](requirements.md) - 需求定义和验收标准 (Requirements and Acceptance Criteria)

**Phase 2: DESIGN** (Architecture Design)
- [design.md](design.md) - 架构设计和技术决策 (Architecture Design and Technical Decisions)

**Phase 3: PLAN** (Implementation Plan)
- [plan.md](plan.md) - 可执行的实施计划 (Executable Implementation Plan)

**Phase 4: IMPLEMENTATION** (Execution)
- [IMPLEMENTATION.md](IMPLEMENTATION.md) - 实施笔记和进度 (Implementation Notes and Progress)

**Phase 5: ACCEPTANCE** (QA and Sign-off)
- Acceptance testing tracked in phase_history

## 阶段工作流 (Phase Workflow)

<!-- English: This section tracks the current phase and workflow state.
     AI should update this as spec progresses through phases.
     
     中文：本部分追踪当前阶段和工作流状态。
     AI应随着规格在各阶段推进而更新此部分。
-->

**当前阶段 (Current Phase)**: {phase}

**下一步行动 (Next Action)**: {current_action}

### 阶段进度表 (Phase Progress)

<!-- English: Updated automatically as spec progresses. 
     Status: ⚪ Not Started | 🟡 In Progress | 🟢 Approved | 🔴 Rejected
     
     中文：随着规格推进自动更新。
     状态：⚪ 未开始 | 🟡 进行中 | 🟢 已批准 | 🔴 已拒绝
-->

| 阶段 (Phase) | 状态 (Status) | 开始日期 (Started) | 完成日期 (Completed) | QA 结果 (QA Result) | 备注 (Notes) |
|-------------|--------------|-------------------|---------------------|-------------------|-------------|
| **Phase 0: PREWORK** | 🟡 待办 (Pending) | - | - | - | Context gathering |
| **Phase 1: REQUIREMENTS** | ⚪ 未开始 | - | - | - | Requirements definition |
| **Phase 2: DESIGN** | ⚪ 未开始 | - | - | - | Architecture design |
| **Phase 3: PLAN** | ⚪ 未开始 | - | - | - | Implementation planning |
| **Phase 4: IMPLEMENTATION** | ⚪ 未开始 | - | - | - | Code execution |
| **Phase 5: ACCEPTANCE** | ⚪ 未开始 | - | - | - | Final QA and sign-off |

## 阶段历史 (Phase History)

<!-- English: Automatically populated from frontmatter phase_history field.
     Each phase completion is recorded with QA results.
     
     中文：自动从 frontmatter 的 phase_history 字段填充。
     每个阶段完成时记录QA结果。
-->

_No phase history yet. History will be recorded as spec progresses through phases._

_尚无阶段历史。历史将随着规格在各阶段推进而记录。_

## PREWORK 总结 (PREWORK Summary)

<!-- English: High-level summary of project context from prework.md
     中文：来自 prework.md 的项目上下文高层总结
-->

### 项目技术栈 (Project Tech Stack)

<!-- English: Framework, languages, key libraries
     中文：框架、语言、关键库
-->

- Framework: 
- Languages: 
- Key Dependencies: 

### 关键约束 (Key Constraints)

<!-- English: Important constraints from existing codebase
     中文：来自现有代码库的重要约束
-->

## REQUIREMENTS 总结 (REQUIREMENTS Summary)

<!-- English: High-level summary from requirements.md
     中文：来自 requirements.md 的高层总结
-->

### 核心问题 (Core Problem)

<!-- English: What problem are we solving?
     中文：我们要解决什么问题？
-->

### 关键用户故事 (Key User Stories)

<!-- English: Top 2-3 user stories with priorities
     Format: [Priority] As a [role], I want to [action], so that [benefit]
     中文：2-3个优先级最高的用户故事
     格式：[优先级] 作为[角色]，我想要[行动]，以便[收益]
-->

### 验收标准 (Acceptance Criteria)

<!-- English: High-level acceptance criteria
     Detailed Gherkin scenarios are in requirements.md
     中文：高层验收标准
     详细的 Gherkin 场景在 requirements.md 中
-->

- [ ] 

## DESIGN 总结 (DESIGN Summary)

<!-- English: High-level architecture summary from design.md
     中文：来自 design.md 的架构总结
-->

### 架构决策 (Architecture Decisions)

<!-- English: Key ADRs (Architecture Decision Records)
     中文：关键架构决策记录
-->

### 文件清单 (File Manifest)

<!-- English: Files to create/modify (from design.md Section 9)
     This is critical for PLAN phase
     中文：需要创建/修改的文件（来自 design.md 第9节）
     这对 PLAN 阶段至关重要
-->

**创建的文件 (Files to Create)**:
- 

**修改的文件 (Files to Modify)**:
- 

## PLAN 总结 (PLAN Summary)

<!-- English: Implementation plan summary from plan.md
     中文：来自 plan.md 的实施计划总结
-->

### 里程碑 (Milestones)

<!-- English: Major milestones from plan.md
     中文：来自 plan.md 的主要里程碑
-->

- [ ] Milestone 1: 
- [ ] Milestone 2: 
- [ ] Milestone 3: 

### 估算工作量 (Estimated Effort)

<!-- English: Total estimated effort from plan.md
     中文：来自 plan.md 的总估算工作量
-->

**Total**: _hours

## 风险和阻塞项 (Risks and Blockers)

<!-- English: Track risks and blocking issues
     中文：跟踪风险和阻塞问题
-->

### 当前风险 (Current Risks)

<!-- English: Identified risks with mitigation strategies
     中文：已识别的风险及缓解策略
-->

### 阻塞项 (Blockers)

<!-- English: Current blockers preventing progress
     中文：当前阻止进展的阻塞项
-->

- [ ] No blockers currently

## 相关链接 (Related Links)

<!-- English: Links to related specs, dependencies, issues
     中文：相关规格、依赖项、问题的链接
-->

- **依赖的规格 (Depends On)**: 
- **被依赖的规格 (Required By)**: 
- **相关问题 (Related Issues)**: 

---

## 给 AI 的指令 (Instructions for AI)

<!-- English: These instructions guide AI through the PDPI-spec workflow.
     DO NOT REMOVE THIS SECTION.
-->

### Phase Workflow Protocol

**Current Phase**: Check the `phase` field in frontmatter above.

**Phase Transition Rules**:
1. **PREWORK → REQUIREMENTS**: After PREWORK QA passes
2. **REQUIREMENTS → DESIGN**: After REQUIREMENTS QA passes
3. **DESIGN → PLAN**: After DESIGN QA passes
4. **PLAN → IMPLEMENTATION**: After PLAN QA passes
5. **IMPLEMENTATION → ACCEPTANCE**: After all implementation complete
6. **ACCEPTANCE → COMPLETE**: After acceptance QA passes

**QA Gate Protocol**:
- At each phase completion, run corresponding QA checklist from AGENTS.md
- Generate QA report with verdict: 🔴 Rejected | 🟡 Needs Revision | 🟢 Approved
- Update `phase_history` in frontmatter with QA result
- Only advance phase if QA approved AND user confirms

**Phase Documentation**:
- Each phase creates its corresponding document (prework.md, requirements.md, etc.)
- Follow templates in `.lean-spec/templates/phases/`
- Use English for instructions, bilingual for section titles, Chinese for user-facing content

**References**:
- Phase instructions: See AGENTS.md sections for each phase
- QA checklists: See AGENTS.md QA sections
- Templates: See `.lean-spec/templates/phases/`

