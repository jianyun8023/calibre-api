# Implementation Tasks

## Overview

实现 EPUB 阅读器增强功能，包括阅读进度保存、书签管理、文本高亮、笔记功能，以及应用用户设置。

## Task Breakdown

### Task 1: Create Reader Data Types

**Description:** 定义阅读器数据的 TypeScript 类型。

**Dependencies:** None

**Estimated Effort:** 30 minutes

**Acceptance Criteria:**
- [ ] 创建 `web-next/src/types/reader.ts` 文件
- [ ] 定义 `ReadingProgress`, `Bookmark`, `Highlight`, `Note` 接口
- [ ] 定义 `ReaderData` 接口
- [ ] 添加辅助类型和常量

**Validates: Requirements 1.2, 2.2, 3.2, 4.2**

---

### Task 2: Implement Reader Context and Provider

**Description:** 创建 Reader Context 管理阅读数据。

**Dependencies:** Task 1

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] 创建 `web-next/src/contexts/reader-context.tsx`
- [ ] 实现 `ReaderContext` 和 `ReaderProvider`
- [ ] 实现 `loadReaderData()` 从 localStorage 加载
- [ ] 实现 `saveReaderData()` 保存到 localStorage
- [ ] 实现错误处理和默认值回退

**Validates: Requirements 1.1, 1.2, 1.4**

---

### Task 3: Implement Progress Tracking

**Description:** 实现阅读进度自动保存功能。

**Dependencies:** Task 2

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] 实现 `updateProgress()` 方法
- [ ] 添加防抖逻辑（2 秒延迟）
- [ ] 计算阅读进度百分比
- [ ] 自动保存到 localStorage

**Validates: Requirements 1.1, 1.2, 1.3, 9.1, 9.2**

---

### Task 4: Implement Bookmark Management

**Description:** 实现书签的添加、删除和列表功能。

**Dependencies:** Task 2

**Estimated Effort:** 1.5 hours

**Acceptance Criteria:**
- [ ] 实现 `addBookmark()` 方法
- [ ] 实现 `removeBookmark()` 方法
- [ ] 实现 `getBookmarks()` 方法
- [ ] 生成唯一 ID（使用 crypto.randomUUID()）
- [ ] 提取预览文本和章节信息
- [ ] 自动保存到 localStorage

**Validates: Requirements 2.1, 2.2, 2.3, 2.4, 2.5, 2.6**

---

### Task 5: Implement Highlight Management

**Description:** 实现文本高亮的添加、删除和列表功能。

**Dependencies:** Task 2

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] 实现 `addHighlight()` 方法
- [ ] 实现 `removeHighlight()` 方法
- [ ] 实现 `getHighlights()` 方法
- [ ] 支持 4 种颜色（yellow, green, blue, pink）
- [ ] 应用高亮到 rendition
- [ ] 自动保存到 localStorage

**Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7**

---

### Task 6: Implement Note Management

**Description:** 实现笔记的添加、编辑、删除和列表功能。

**Dependencies:** Task 2

**Estimated Effort:** 1.5 hours

**Acceptance Criteria:**
- [ ] 实现 `addNote()` 方法
- [ ] 实现 `updateNote()` 方法
- [ ] 实现 `removeNote()` 方法
- [ ] 实现 `getNotes()` 方法
- [ ] 记录创建和修改时间
- [ ] 自动保存到 localStorage

**Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 4.8**

---

### Task 7: Apply Reader Settings

**Description:** 将用户设置应用到阅读器。

**Dependencies:** Task 2

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] 使用 `useSettings` hook 获取阅读器设置
- [ ] 应用字体大小到 rendition
- [ ] 应用字体系列到 rendition
- [ ] 应用行高到 rendition
- [ ] 应用阅读器主题（light/dark/sepia）
- [ ] 监听设置变化并实时更新

**Validates: Requirements 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7**

---

### Task 8: Create Reader Sidebar Component

**Description:** 创建侧边栏组件显示书签、高亮和笔记。

**Dependencies:** Task 4, Task 5, Task 6

**Estimated Effort:** 3 hours

**Acceptance Criteria:**
- [ ] 创建 `web-next/src/components/reader-sidebar.tsx`
- [ ] 实现三个标签页（Bookmarks, Highlights, Notes）
- [ ] 显示每个类型的列表
- [ ] 实现点击跳转功能
- [ ] 实现删除功能
- [ ] 添加空状态提示
- [ ] 响应式设计（移动端全屏）

**Validates: Requirements 6.1, 6.2, 6.3, 6.4, 6.6**

---

### Task 9: Create Reader Toolbar Component

**Description:** 创建工具栏组件提供快捷操作。

**Dependencies:** Task 3

**Estimated Effort:** 1.5 hours

**Acceptance Criteria:**
- [ ] 创建 `web-next/src/components/reader-toolbar.tsx`
- [ ] 添加书签按钮
- [ ] 添加侧边栏切换按钮
- [ ] 显示阅读进度
- [ ] 实现自动隐藏逻辑（3 秒）
- [ ] 响应式设计

**Validates: Requirements 10.1, 10.2, 10.3, 10.4, 10.5**

---

### Task 10: Integrate Components into Reader Page

**Description:** 将所有组件集成到阅读器页面。

**Dependencies:** Task 7, Task 8, Task 9

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] 在阅读器页面添加 ReaderProvider
- [ ] 集成 Reader Sidebar
- [ ] 集成 Reader Toolbar
- [ ] 实现组件间的交互
- [ ] 处理文本选择事件
- [ ] 应用设置到 rendition

**Validates: Requirements 5.6, 6.1, 10.1**

---

### Task 11: Implement Keyboard Shortcuts

**Description:** 添加键盘快捷键支持。

**Dependencies:** Task 10

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] 实现左右箭头翻页
- [ ] 实现 B 键添加书签
- [ ] 实现 H 键添加高亮
- [ ] 实现 N 键添加笔记
- [ ] 实现 S 键切换侧边栏
- [ ] 实现 ESC 键关闭对话框
- [ ] 添加快捷键提示

**Validates: Requirements 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7**

---

### Task 12: Implement Export/Import Functionality

**Description:** 实现阅读数据的导出和导入。

**Dependencies:** Task 2

**Estimated Effort:** 1.5 hours

**Acceptance Criteria:**
- [ ] 实现 `exportData()` 方法
- [ ] 生成 JSON 文件并触发下载
- [ ] 实现 `importData()` 方法
- [ ] 验证导入文件格式
- [ ] 合并导入的数据
- [ ] 处理数据冲突

**Validates: Requirements 7.1, 7.2, 7.3, 7.4, 7.5**

---

### Task 13: Add Note Dialog Component

**Description:** 创建笔记输入对话框。

**Dependencies:** Task 6

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] 创建 `web-next/src/components/note-dialog.tsx`
- [ ] 实现笔记输入表单
- [ ] 显示关联的文本
- [ ] 实现保存和取消按钮
- [ ] 支持编辑现有笔记

**Validates: Requirements 4.1, 4.6**

---

### Task 14: Add Highlight Color Picker

**Description:** 创建高亮颜色选择器。

**Dependencies:** Task 5

**Estimated Effort:** 45 minutes

**Acceptance Criteria:**
- [ ] 创建颜色选择器组件
- [ ] 支持 4 种颜色
- [ ] 显示颜色预览
- [ ] 集成到高亮功能

**Validates: Requirements 3.6**

---

### Task 15: Add Unit Tests

**Description:** 为阅读器功能添加单元测试。

**Dependencies:** Task 2, Task 3, Task 4, Task 5, Task 6

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] 测试 Reader Context 初始化
- [ ] 测试进度保存和加载
- [ ] 测试书签 CRUD 操作
- [ ] 测试高亮 CRUD 操作
- [ ] 测试笔记 CRUD 操作
- [ ] 测试 localStorage 集成
- [ ] 测试错误处理

**Validates: Requirements 1.1, 2.1, 3.1, 4.1**

---

## Task Summary

| Task | Effort | Priority | Dependencies |
|------|--------|----------|--------------|
| Task 1: Reader Types | 0.5h | High | None |
| Task 2: Reader Context | 2h | High | Task 1 |
| Task 3: Progress Tracking | 1h | High | Task 2 |
| Task 4: Bookmark Management | 1.5h | High | Task 2 |
| Task 5: Highlight Management | 2h | High | Task 2 |
| Task 6: Note Management | 1.5h | High | Task 2 |
| Task 7: Apply Settings | 1h | Medium | Task 2 |
| Task 8: Reader Sidebar | 3h | High | Task 4, 5, 6 |
| Task 9: Reader Toolbar | 1.5h | Medium | Task 3 |
| Task 10: Integration | 2h | High | Task 7, 8, 9 |
| Task 11: Keyboard Shortcuts | 1h | Low | Task 10 |
| Task 12: Export/Import | 1.5h | Low | Task 2 |
| Task 13: Note Dialog | 1h | Medium | Task 6 |
| Task 14: Color Picker | 0.75h | Low | Task 5 |
| Task 15: Unit Tests | 2h | Medium | Task 2-6 |

**Total Estimated Effort:** 22.25 hours

## Implementation Order

1. **Phase 1: Core Infrastructure** (4 hours)
   - Task 1: Reader Types
   - Task 2: Reader Context
   - Task 3: Progress Tracking

2. **Phase 2: Data Management** (5 hours)
   - Task 4: Bookmark Management
   - Task 5: Highlight Management
   - Task 6: Note Management

3. **Phase 3: UI Components** (5.5 hours)
   - Task 7: Apply Settings
   - Task 8: Reader Sidebar
   - Task 9: Reader Toolbar

4. **Phase 4: Integration** (4.75 hours)
   - Task 10: Integration
   - Task 11: Keyboard Shortcuts
   - Task 13: Note Dialog
   - Task 14: Color Picker

5. **Phase 5: Polish** (3 hours)
   - Task 12: Export/Import
   - Task 15: Unit Tests

## Success Criteria

- ✅ 阅读进度自动保存和恢复
- ✅ 书签、高亮、笔记功能完整
- ✅ 侧边栏显示所有阅读数据
- ✅ 应用用户设置到阅读器
- ✅ 键盘快捷键支持
- ✅ 导出/导入功能正常
- ✅ 所有测试通过
- ✅ 响应式设计适配移动端
