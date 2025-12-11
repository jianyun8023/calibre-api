# 实施任务

## Phase 1: 创建组件 (30 分钟)

- [x] 1.1 创建 BookGrid 组件
  - 文件: `web-next/src/components/book-grid.tsx`
  - 实现加载状态、空状态、网格布局
  - 添加响应式列数配置
  - 使用 React.memo 优化
  - _验收: US-1 所有 AC_

- [x] 1.2 创建 Pagination 组件
  - 文件: `web-next/src/components/pagination.tsx`
  - 实现上一页/下一页按钮
  - 添加禁用和加载状态
  - 添加可访问性属性
  - _验收: US-2 所有 AC_

## Phase 2: 重构页面 (60 分钟)

- [x] 2.1 重构首页
  - 文件: `web-next/src/app/page.tsx`
  - 替换随机推荐和最近添加区块
  - 删除重复代码
  - _验收: US-3 所有 AC_

- [x] 2.2 重构书籍列表页
  - 文件: `web-next/src/app/books/page.tsx`
  - 使用 BookGrid 和 Pagination
  - 测试分页功能
  - _验收: US-4 所有 AC_

- [x] 2.3 重构搜索页
  - 文件: `web-next/src/app/search/page.tsx`
  - 使用 BookGrid
  - 测试搜索功能
  - _验收: US-5 AC1_

- [x] 2.4 重构聊天页
  - 文件: `web-next/src/app/chat/page.tsx`
  - 使用 BookGrid (proxyImage=true)
  - 测试推荐功能
  - _验收: US-5 AC2_

## Phase 3: 测试验证 (30 分钟)

- [x] 3.1 功能测试
  - 测试所有页面功能正常
  - 验证加载状态、空状态
  - 验证分页功能
  - 检查控制台无错误

- [x] 3.2 响应式测试
  - 测试移动端 (< 640px): 2 列
  - 测试平板端 (640-768px): 2 列
  - 测试桌面端 (768-1024px): 3 列
  - 测试大屏幕 (1024-1280px): 4 列
  - 测试超大屏 (> 1280px): 5 列

- [x] 3.3 布局测试
  - 验证所有页面不超出屏幕高度
  - 验证 Header 固定在顶部
  - 验证 Sidebar 固定宽度
  - 验证 Main 区域独立滚动
  - 创建布局验证文档

## 完成标准

- ✅ 所有任务完成
- ✅ 所有测试通过
- ✅ 代码重复减少 ≥60%
- ✅ 无功能退化
