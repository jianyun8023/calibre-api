# Implementation Tasks

## Overview

实现用户设置持久化系统，包括 Settings Context、Hook、localStorage 集成、导入/导出功能，以及增强的设置页面 UI。

## Task Breakdown

### Task 1: Create Settings Types and Constants

**Description:** 定义设置的 TypeScript 类型和默认值常量。

**Dependencies:** None

**Estimated Effort:** 30 minutes

**Acceptance Criteria:**
- [ ] 创建 `web-next/src/types/settings.ts` 文件
- [ ] 定义 `AppSettings` 接口（包含 API、Reader、Search、UI 配置）
- [ ] 定义 `SettingsExport` 接口（包含版本和时间戳）
- [ ] 定义 `DEFAULT_SETTINGS` 常量
- [ ] 添加设置验证的辅助类型

**Implementation Notes:**
- 使用严格的类型定义（literal types for enums）
- 包含 JSDoc 注释说明每个字段的用途和范围
- 版本号使用 semver 格式

**Validates: Requirements 8.1, 10.3**

---

### Task 2: Implement Settings Context and Provider

**Description:** 创建 Settings Context 和 Provider 组件，管理全局设置状态。

**Dependencies:** Task 1

**Estimated Effort:** 1.5 hours

**Acceptance Criteria:**
- [ ] 创建 `web-next/src/contexts/settings-context.tsx`
- [ ] 实现 `SettingsContext` 和 `SettingsProvider`
- [ ] 实现 `loadSettings()` 从 localStorage 加载
- [ ] 实现 `saveSettings()` 保存到 localStorage
- [ ] 实现错误处理和默认值回退
- [ ] 添加 loading 状态管理

**Implementation Notes:**
- 使用 `useEffect` 在组件挂载时加载设置
- 使用 `useEffect` 在设置变化时自动保存
- 添加 try-catch 处理 localStorage 错误
- 使用 `JSON.parse/stringify` 进行序列化

**Validates: Requirements 1.1, 1.2, 1.3, 1.4**

---

### Task 3: Create useSettings Hook

**Description:** 创建自定义 Hook 以便组件访问和修改设置。

**Dependencies:** Task 2

**Estimated Effort:** 30 minutes

**Acceptance Criteria:**
- [ ] 创建 `web-next/src/hooks/use-settings.ts`
- [ ] 实现 `useSettings()` hook
- [ ] 添加 context 未找到的错误处理
- [ ] 导出 hook 供其他组件使用

**Implementation Notes:**
- 检查 context 是否存在，否则抛出有意义的错误
- 返回完整的 context 值（settings, updateSettings, etc.）

**Validates: Requirements 8.1, 8.2**

---

### Task 4: Implement Settings Update Logic

**Description:** 实现设置更新、重置和验证逻辑。

**Dependencies:** Task 2, Task 3

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] 实现 `updateSettings()` 方法（支持部分更新）
- [ ] 实现 `resetSettings()` 方法
- [ ] 实现设置值验证（范围检查、类型检查）
- [ ] 添加成功/失败的 toast 通知
- [ ] 确保更新后立即保存到 localStorage

**Implementation Notes:**
- 使用 spread operator 进行部分更新
- 验证数值范围（fontSize: 12-24, lineHeight: 1.2-2.0）
- 验证枚举值（fontFamily, searchMode, etc.）
- 使用 `sonner` toast 显示通知

**Validates: Requirements 1.1, 1.5, 2.1, 2.2, 2.3, 6.1, 6.2, 6.3**

---

### Task 5: Implement Export Settings Functionality

**Description:** 实现设置导出功能，生成 JSON 文件供用户下载。

**Dependencies:** Task 2, Task 3

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] 实现 `exportSettings()` 方法
- [ ] 生成包含版本和时间戳的导出对象
- [ ] 创建 Blob 并触发浏览器下载
- [ ] 使用格式化的文件名（包含时间戳）
- [ ] 添加导出成功/失败通知

**Implementation Notes:**
- 文件名格式: `calibre-settings-YYYY-MM-DD-HHmmss.json`
- 使用 `JSON.stringify(data, null, 2)` 格式化输出
- 使用 `URL.createObjectURL` 和 `<a>` 元素触发下载
- 清理创建的 Object URL

**Validates: Requirements 3.1, 3.2, 3.3, 3.4**

---

### Task 6: Implement Import Settings Functionality

**Description:** 实现设置导入功能，从 JSON 文件恢复配置。

**Dependencies:** Task 2, Task 3

**Estimated Effort:** 1.5 hours

**Acceptance Criteria:**
- [ ] 实现 `importSettings()` 方法
- [ ] 验证导入文件的 JSON 格式
- [ ] 验证设置数据的结构和类型
- [ ] 处理版本兼容性（可选的迁移逻辑）
- [ ] 应用导入的设置并保存到 localStorage
- [ ] 添加导入成功/失败通知

**Implementation Notes:**
- 使用 FileReader API 读取文件
- 验证必需字段存在
- 验证设置值的有效性
- 如果版本不匹配，尝试迁移或拒绝
- 使用 try-catch 处理所有可能的错误

**Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5**

---

### Task 7: Integrate Settings Provider into App

**Description:** 将 Settings Provider 集成到应用的根布局中。

**Dependencies:** Task 2

**Estimated Effort:** 30 minutes

**Acceptance Criteria:**
- [ ] 在 `web-next/src/app/[locale]/layout.tsx` 中添加 SettingsProvider
- [ ] 确保 Provider 包裹所有需要访问设置的组件
- [ ] 验证 Provider 在 ThemeProvider 之后（避免主题冲突）
- [ ] 测试设置在不同页面间的共享

**Implementation Notes:**
- Provider 顺序: ThemeProvider → SettingsProvider → children
- 确保 SSR 兼容性（使用 'use client' 如果需要）

**Validates: Requirements 8.3**

---

### Task 8: Enhance Settings Page UI

**Description:** 增强设置页面，添加导入/导出按钮和更好的 UI 组织。

**Dependencies:** Task 4, Task 5, Task 6, Task 7

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] 重构 `web-next/src/app/[locale]/settings/page.tsx` 使用 useSettings hook
- [ ] 添加导出设置按钮
- [ ] 添加导入设置按钮（带文件选择器）
- [ ] 改进设置分类和布局
- [ ] 添加设置说明和帮助文本
- [ ] 添加当前值显示（如滑块的数值）
- [ ] 改进响应式设计

**Implementation Notes:**
- 使用 Card 组件组织不同类别的设置
- 使用 Label 和描述文本提高可读性
- 导入按钮使用隐藏的 `<input type="file">`
- 添加 loading 状态指示器
- 使用 Lucide icons 增强视觉效果

**Validates: Requirements 10.1, 10.2, 10.3**

---

### Task 9: Apply Settings in Reader Page

**Description:** 在阅读器页面应用用户的阅读器设置。

**Dependencies:** Task 3, Task 7

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] 在阅读器页面使用 useSettings hook
- [ ] 应用字体大小设置到阅读器内容
- [ ] 应用字体系列设置到阅读器内容
- [ ] 应用行高设置到阅读器内容
- [ ] 应用阅读器主题设置（light/dark/sepia）
- [ ] 确保设置变化时阅读器立即更新

**Implementation Notes:**
- 使用 CSS 变量或内联样式应用设置
- 考虑使用 `useEffect` 监听设置变化
- 阅读器主题可能需要额外的 CSS 类

**Validates: Requirements 6.1, 6.2, 6.3, 6.4**

---

### Task 10: Apply Settings in Search Page

**Description:** 在搜索页面应用用户的搜索偏好设置。

**Dependencies:** Task 3, Task 7

**Estimated Effort:** 45 minutes

**Acceptance Criteria:**
- [ ] 在搜索页面使用 useSettings hook
- [ ] 使用默认搜索模式作为初始值
- [ ] 使用每页结果数设置
- [ ] 应用自动完成设置（如果实现）
- [ ] 确保设置变化在下次搜索时生效

**Implementation Notes:**
- 从 settings.search 读取默认值
- 用户可以在搜索时临时覆盖默认值
- 考虑在 URL 参数中保存当前搜索模式

**Validates: Requirements 7.1, 7.2, 7.3, 7.4**

---

### Task 11: Add Unit Tests for Settings Logic

**Description:** 为设置逻辑添加单元测试。

**Dependencies:** Task 1, Task 2, Task 3, Task 4

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] 测试 DEFAULT_SETTINGS 的完整性
- [ ] 测试 loadSettings() 从 localStorage 加载
- [ ] 测试 saveSettings() 保存到 localStorage
- [ ] 测试 updateSettings() 部分更新
- [ ] 测试 resetSettings() 重置功能
- [ ] 测试错误处理（localStorage 失败、数据损坏）
- [ ] 测试设置验证逻辑

**Implementation Notes:**
- 使用 Vitest 作为测试框架
- Mock localStorage API
- 测试边界条件和错误情况
- 使用 `@testing-library/react` 测试 Context

**Validates: Requirements 1.1, 1.2, 1.3, 1.4, 2.1, 2.2**

---

### Task 12: Add Property-Based Tests

**Description:** 添加基于属性的测试以验证设置系统的正确性属性。

**Dependencies:** Task 11

**Estimated Effort:** 2 hours

**Acceptance Criteria:**
- [ ] 测试 Property 1: Settings Persistence Round Trip
- [ ] 测试 Property 2: Default Settings Fallback
- [ ] 测试 Property 3: Export-Import Round Trip
- [ ] 测试 Property 6: Settings Validation
- [ ] 配置测试运行 100+ 迭代

**Implementation Notes:**
- 使用 `@fast-check/vitest` 进行属性测试
- 创建 settings arbitraries（随机生成器）
- 测试各种边界情况和随机输入
- 验证不变量在所有情况下都成立

**Validates: Requirements 1.1, 1.2, 3.1, 4.1, 6.1**

---

### Task 13: Add Integration Tests

**Description:** 添加端到端集成测试验证完整的设置流程。

**Dependencies:** Task 8, Task 9, Task 10

**Estimated Effort:** 1.5 hours

**Acceptance Criteria:**
- [ ] 测试完整的设置保存和加载流程
- [ ] 测试导出和导入流程
- [ ] 测试跨页面的设置应用
- [ ] 测试设置重置流程
- [ ] 测试错误恢复场景

**Implementation Notes:**
- 使用 Playwright 或 Cypress 进行 E2E 测试
- 测试真实的用户交互流程
- 验证 localStorage 的实际读写
- 测试页面刷新后的持久化

**Validates: Requirements 1.1, 1.2, 3.1, 4.1, 6.4, 7.4**

---

### Task 14: Add UI Settings Section

**Description:** 添加 UI 相关的设置选项（紧凑模式、封面显示、网格列数）。

**Dependencies:** Task 8

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] 添加紧凑模式切换开关
- [ ] 添加显示封面图片切换开关
- [ ] 添加网格列数选择器
- [ ] 在书籍列表页面应用这些设置
- [ ] 确保设置立即生效

**Implementation Notes:**
- 使用 Switch 组件实现切换开关
- 使用 Select 组件选择网格列数
- 在 book-grid 组件中应用这些设置
- 考虑使用 CSS Grid 的 auto-fit 功能

**Validates: Requirements 10.1, 10.2**

---

### Task 15: Documentation and Examples

**Description:** 创建设置系统的文档和使用示例。

**Dependencies:** All previous tasks

**Estimated Effort:** 1 hour

**Acceptance Criteria:**
- [ ] 在 spec 中添加使用示例
- [ ] 文档化 Settings Context API
- [ ] 文档化 useSettings Hook API
- [ ] 添加常见问题解答
- [ ] 添加故障排除指南

**Implementation Notes:**
- 包含代码示例
- 说明如何在新组件中使用设置
- 说明如何添加新的设置字段
- 说明版本迁移策略

---

## Task Summary

| Task | Effort | Priority | Dependencies |
|------|--------|----------|--------------|
| Task 1: Settings Types | 0.5h | High | None |
| Task 2: Settings Context | 1.5h | High | Task 1 |
| Task 3: useSettings Hook | 0.5h | High | Task 2 |
| Task 4: Update Logic | 1h | High | Task 2, 3 |
| Task 5: Export Functionality | 1h | Medium | Task 2, 3 |
| Task 6: Import Functionality | 1.5h | Medium | Task 2, 3 |
| Task 7: Integrate Provider | 0.5h | High | Task 2 |
| Task 8: Enhance Settings Page | 2h | High | Task 4, 5, 6, 7 |
| Task 9: Apply in Reader | 1h | Medium | Task 3, 7 |
| Task 10: Apply in Search | 0.75h | Medium | Task 3, 7 |
| Task 11: Unit Tests | 2h | High | Task 1-4 |
| Task 12: Property Tests | 2h | Medium | Task 11 |
| Task 13: Integration Tests | 1.5h | Medium | Task 8-10 |
| Task 14: UI Settings | 1h | Low | Task 8 |
| Task 15: Documentation | 1h | Low | All |

**Total Estimated Effort:** 17.75 hours

## Implementation Order

1. **Phase 1: Core Infrastructure** (4 hours)
   - Task 1: Settings Types
   - Task 2: Settings Context
   - Task 3: useSettings Hook
   - Task 4: Update Logic

2. **Phase 2: Import/Export** (2.5 hours)
   - Task 5: Export Functionality
   - Task 6: Import Functionality

3. **Phase 3: Integration** (4.25 hours)
   - Task 7: Integrate Provider
   - Task 8: Enhance Settings Page
   - Task 9: Apply in Reader
   - Task 10: Apply in Search

4. **Phase 4: Testing** (5.5 hours)
   - Task 11: Unit Tests
   - Task 12: Property Tests
   - Task 13: Integration Tests

5. **Phase 5: Polish** (2 hours)
   - Task 14: UI Settings
   - Task 15: Documentation

## Success Criteria

- ✅ 所有设置自动保存到 localStorage
- ✅ 页面刷新后设置正确恢复
- ✅ 导出/导入功能正常工作
- ✅ 设置在所有页面间正确应用
- ✅ 所有测试通过（单元测试、属性测试、集成测试）
- ✅ 错误处理健壮（localStorage 失败、数据损坏）
- ✅ UI 清晰易用，有适当的帮助文本
- ✅ 性能良好，无不必要的重新渲染
