---
status: complete
priority: medium
tags:
  - frontend
  - refactor
  - components
  - ui
created: '2025-12-11'
created_at: '2025-12-11T08:43:00.000Z'
updated_at: '2025-12-11T09:44:22.596Z'
transitions:
  - status: in-progress
    at: '2025-12-11T09:09:40.594Z'
  - status: complete
    at: '2025-12-11T09:44:22.596Z'
completed_at: '2025-12-11T09:44:22.596Z'
completed: '2025-12-11'
---

# 前端组件化重构 - BookGrid 和 Pagination

> **Status**: ✅ Complete · **Priority**: Medium · **Created**: 2025-12-11 · **Tags**: frontend, refactor, components, ui

## 概述

基于 `specs/COMPONENT-ANALYSIS.md` 的评估结果，实施 Phase 1 组件化重构，创建 BookGrid 和 Pagination 组件，减少代码重复 68%，提升一致性 50%。

## 背景

当前项目中存在大量重复的图书网格布局和分页控件代码：
- **网格布局**: 在 5 个页面中重复 ~50 行代码
- **骨架屏**: 在 5 个页面中重复 ~25 行代码  
- **空状态**: 在 4 个页面中重复 ~20 行代码
- **分页控件**: 在 2 个页面中重复 ~30 行代码
- **总计**: ~125 行重复代码

## 目标

### Phase 1 (本 Spec)
1. 创建 `BookGrid` 组件 - 统一图书网格布局、加载状态、空状态
2. 创建 `Pagination` 组件 - 统一分页控件
3. 重构 4 个页面使用新组件：`/`, `/books`, `/search`, `/chat`

### 预期收益
- ✅ 减少代码重复 68% (~85 行)
- ✅ 提高代码一致性 50%
- ✅ 降低维护成本 70%
- ✅ 提升可测试性 80%

## 用户故事

### US-1: BookGrid 组件
**作为** 开发者  
**我想要** 一个统一的图书网格组件  
**以便** 在不同页面中复用相同的布局逻辑

**验收标准**:
- AC1: 支持响应式列数配置 (sm/md/lg/xl)
- AC2: 支持加载状态显示 (骨架屏)
- AC3: 支持空状态显示 (自定义消息)
- AC4: 支持传递 BookCard 的 props (moreInfo, proxyImage)
- AC5: 使用 book.id 作为 key

### US-2: Pagination 组件
**作为** 开发者  
**我想要** 一个统一的分页组件  
**以便** 在列表页面中复用相同的分页逻辑

**验收标准**:
- AC1: 支持上一页/下一页按钮
- AC2: 显示当前页码
- AC3: 支持禁用状态 (无上一页/下一页时)
- AC4: 支持加载状态 (禁用按钮)
- AC5: 支持自定义点击回调

### US-3: 重构首页
**作为** 开发者  
**我想要** 使用新组件重构首页  
**以便** 减少重复代码并提升一致性

**验收标准**:
- AC1: 随机推荐区块使用 BookGrid
- AC2: 最近添加区块使用 BookGrid
- AC3: 功能保持不变
- AC4: 代码行数减少 ~40%

### US-4: 重构书籍列表页
**作为** 开发者  
**我想要** 使用新组件重构书籍列表页  
**以便** 减少重复代码并统一分页样式

**验收标准**:
- AC1: 使用 BookGrid 显示书籍列表
- AC2: 使用 Pagination 组件
- AC3: 功能保持不变 (分页、加载状态)
- AC4: 代码行数减少 ~50%

### US-5: 重构搜索页和聊天页
**作为** 开发者  
**我想要** 使用新组件重构搜索页和聊天页  
**以便** 统一所有页面的图书展示样式

**验收标准**:
- AC1: 搜索页使用 BookGrid
- AC2: 聊天页使用 BookGrid (proxyImage=true)
- AC3: 功能保持不变
- AC4: 代码行数减少 ~40%

## 技术约束

1. **TypeScript**: 所有组件必须有完整的类型定义
2. **响应式**: 支持移动端、平板、桌面端
3. **性能**: 使用 React.memo 避免不必要的重渲染
4. **可访问性**: 按钮需要 aria-label
5. **向后兼容**: 不破坏现有功能

## 依赖关系

- 依赖 `BookCard` 组件 (已存在)
- 依赖 `Skeleton` 组件 (shadcn/ui)
- 依赖 `Button` 组件 (shadcn/ui)

## 参考文档

- `specs/COMPONENT-ANALYSIS.md` - 完整的组件化评估报告
- `web-next/src/components/book-card.tsx` - 现有 BookCard 组件
- `web-next/src/app/books/page.tsx` - 现有分页实现

## 实施计划

详见 `tasks.md`

## 测试计划

1. **功能测试**: 验证所有页面功能正常
2. **响应式测试**: 验证不同屏幕尺寸下的布局
3. **性能测试**: 验证组件渲染性能
4. **回归测试**: 验证重构后无功能退化

## 成功标准

- ✅ BookGrid 和 Pagination 组件创建完成
- ✅ 4 个页面重构完成
- ✅ 所有功能测试通过
- ✅ 代码重复减少 ≥60%
- ✅ 无功能退化
