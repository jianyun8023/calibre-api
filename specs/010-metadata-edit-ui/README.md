---
status: complete
created: '2025-12-10'
tags:
  - frontend
  - metadata
  - ui
  - crud
priority: high
created_at: '2025-12-10T15:38:25.349Z'
depends_on:
  - 009-frontend-migration
updated_at: '2025-12-11T03:35:13.305Z'
transitions:
  - status: in-progress
    at: '2025-12-10T15:45:12.848Z'
  - status: complete
    at: '2025-12-11T03:35:13.305Z'
completed_at: '2025-12-11T03:35:13.305Z'
completed: '2025-12-11'
---

# 书籍元数据编辑 UI

> **Status**: ✅ Complete · **Priority**: High · **Created**: 2025-12-10 · **Tags**: frontend, metadata, ui, crud

## Overview

实现书籍详情页的完整元数据编辑功能，包括单本书籍的元数据编辑表单、验证、API 调用和批量元数据管理功能。

**功能范围**：
- 单本书籍元数据编辑对话框
- 批量元数据管理页面
- 表单验证和错误处理
- API 集成和数据同步

## Design

### 技术架构

**组件结构**：
```
web-next/src/
├── components/
│   └── metadata-edit-dialog.tsx     # 单本书籍编辑对话框
├── app/
│   ├── detail/[id]/page.tsx         # 书籍详情页（集成编辑对话框）
│   └── metadata/manager/page.tsx    # 批量元数据管理页面
└── lib/api/books.ts                  # API 调用函数
```

### 单本书籍编辑

**MetadataEditDialog 组件**：
- 使用 Shadcn/UI Dialog 组件
- 表单字段：
  - Title（标题）
  - Authors（作者，逗号分隔）
  - Publisher（出版社）
  - Published Date（出版日期，日期选择器）
  - ISBN
  - Rating（评分，0-5）
  - Tags（标签，逗号分隔）
  - Comments（描述）

**功能特性**：
- 自动填充当前书籍数据
- 实时表单验证
- 仅提交更改的字段
- 保存成功后自动刷新页面

### 批量元数据管理

**三步流程（使用 Tabs）**：

1. **Search Books**：
   - 搜索输入框
   - 搜索结果列表（复选框选择）
   - Select All / Deselect All 按钮

2. **Edit Metadata**：
   - 批量编辑表单（所有字段可选）
   - 空字段不会更新

3. **Preview Changes**：
   - 显示将要更新的字段
   - 显示受影响的书籍数量
   - 确认提示

## Plan

- [x] 创建 MetadataEditDialog 组件
- [x] 集成对话框到书籍详情页
- [x] 实现单本书籍元数据编辑
- [x] 修复批量管理页面的搜索 API
- [x] 测试单本书籍编辑功能
- [x] 测试批量元数据管理功能
- [x] 更新文档

## Test

### 单本书籍编辑

- [x] 打开编辑对话框，验证表单预填充
- [x] 修改标题、作者、标签、评分
- [x] 保存更改，验证 API 调用
- [x] 验证页面刷新并显示更新后的数据
- [x] 测试取消按钮关闭对话框

### 批量元数据管理

- [x] 搜索书籍，验证返回结果
- [x] 选择多本书籍
- [x] 填写批量编辑表单
- [x] 预览更改
- [x] 应用更改并验证成功

### 测试结果

✅ **单本书籍编辑**：
- 成功修改《驽马》的元数据：
  - 评分：0/5 → 4.5/5 ✅
  - 标签：无 → "间谍小说, 英国文学, 悬疑" ✅
  - 页面自动刷新并显示更新后的数据 ✅

✅ **批量元数据管理**：
- 搜索功能正常，返回 50 本书 ✅
- 书籍列表显示正确（标题、作者、出版社）✅
- 复选框选择功能正常 ✅

## Notes

### API 端点

- **单本书籍更新**：`POST /api/book/{id}/update`
  - Body: `{ title?, authors?, publisher?, pubdate?, isbn?, tags?, rating?, comments? }`
  - 后端会触发 Qdrant 向量数据库的更新任务

- **搜索书籍**：`GET /api/search?query={query}&limit={limit}&offset={offset}&mode=keyword`
  - 返回：`{ total: number, records: Book[] }`

### 数据格式转换

- **评分**：前端使用 0-5 范围，后端使用 0-10 范围（需要 × 2）
- **作者/标签**：前端使用逗号分隔字符串，后端使用数组
- **日期**：前端使用 ISO 8601 格式

### 已知限制

- 批量更新不支持撤销操作
- 批量更新时如果部分失败，会显示成功/失败计数
- 元数据管理页面目前没有实现分页（搜索限制为 50 本书）

### 新增功能（2025-12-10）

**元数据网络搜索功能**：
- ✅ 通过 ISBN、书名或作者搜索豆瓣元数据
- ✅ 显示搜索结果列表（封面、标题、作者、出版社、日期、ISBN、评分）
- ✅ 新旧元数据对比界面
- ✅ 每个字段支持选择"新"或"旧"数据
- ✅ 作者和标签支持多选
- ✅ 标题支持包含/不包含副标题

**实现的组件**：
- `MetadataSearchDialog`: 元数据搜索对话框
- `MetadataCompareDialog`: 新旧元数据对比对话框
- `/lib/api/metadata.ts`: 豆瓣 API 调用函数

**API 端点**：
- `GET /api/metadata/search?query={query}` - 搜索豆瓣元数据
- `GET /api/metadata/isbn/{isbn}` - 通过 ISBN 获取元数据

### 未来优化

- 添加元数据历史记录
- 批量操作支持进度条
- 支持更多元数据源（不仅限于豆瓣）
