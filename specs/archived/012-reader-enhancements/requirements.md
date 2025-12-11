# Requirements Document

## Introduction

本规格定义了 EPUB 阅读器增强功能的需求。系统应支持阅读进度保存、书签管理、文本高亮、笔记功能，并应用用户的阅读器设置（字体、行高等），提供更好的阅读体验。

## Glossary

- **Reader System**: EPUB 阅读器系统
- **Reading Progress**: 用户在书籍中的当前阅读位置（EPUB CFI）
- **Bookmark**: 用户标记的书籍位置，用于快速返回
- **Highlight**: 用户选中并标记的文本片段
- **Note**: 用户在特定位置添加的文字注释
- **EPUB CFI**: EPUB Canonical Fragment Identifier，用于定位 EPUB 内容的标准
- **Rendition**: ePub.js 的渲染实例，控制书籍显示
- **localStorage**: 浏览器本地存储，用于持久化阅读数据

## Requirements

### Requirement 1

**User Story:** 作为读者，我希望我的阅读进度能够自动保存，这样我下次打开书籍时可以从上次的位置继续阅读。

#### Acceptance Criteria

1. WHEN 用户翻页或滚动 THEN Reader System SHALL 自动保存当前阅读位置到 localStorage
2. WHEN 用户重新打开书籍 THEN Reader System SHALL 恢复到上次保存的阅读位置
3. WHEN 阅读位置变化 THEN Reader System SHALL 在 2 秒内保存新位置（防抖）
4. WHEN localStorage 不可用 THEN Reader System SHALL 优雅降级，不影响阅读功能
5. WHEN 用户关闭浏览器 THEN Reader System SHALL 确保最后的阅读位置已保存

### Requirement 2

**User Story:** 作为读者，我希望能够添加书签，以便快速返回到重要的章节或段落。

#### Acceptance Criteria

1. WHEN 用户点击"添加书签"按钮 THEN Reader System SHALL 在当前位置创建书签
2. WHEN 书签创建 THEN Reader System SHALL 保存书签的位置、时间戳和页面预览文本
3. WHEN 用户查看书签列表 THEN Reader System SHALL 显示所有书签及其预览
4. WHEN 用户点击书签 THEN Reader System SHALL 跳转到该书签位置
5. WHEN 用户删除书签 THEN Reader System SHALL 从列表中移除该书签
6. WHEN 书签数据变化 THEN Reader System SHALL 立即保存到 localStorage

### Requirement 3

**User Story:** 作为读者，我希望能够高亮重要的文本，以便后续复习和参考。

#### Acceptance Criteria

1. WHEN 用户选中文本并点击"高亮"按钮 THEN Reader System SHALL 标记该文本为高亮
2. WHEN 高亮创建 THEN Reader System SHALL 保存高亮的位置、文本内容和颜色
3. WHEN 用户查看高亮列表 THEN Reader System SHALL 显示所有高亮及其文本
4. WHEN 用户点击高亮 THEN Reader System SHALL 跳转到该高亮位置
5. WHEN 用户删除高亮 THEN Reader System SHALL 从书籍和列表中移除该高亮
6. WHEN 用户选择高亮颜色 THEN Reader System SHALL 支持至少 4 种颜色（黄色、绿色、蓝色、粉色）
7. WHEN 高亮数据变化 THEN Reader System SHALL 立即保存到 localStorage

### Requirement 4

**User Story:** 作为读者，我希望能够在书籍中添加笔记，记录我的想法和评论。

#### Acceptance Criteria

1. WHEN 用户选中文本并点击"添加笔记"按钮 THEN Reader System SHALL 打开笔记输入框
2. WHEN 用户输入笔记内容并保存 THEN Reader System SHALL 创建笔记并关联到选中位置
3. WHEN 笔记创建 THEN Reader System SHALL 保存笔记的位置、内容、时间戳和关联文本
4. WHEN 用户查看笔记列表 THEN Reader System SHALL 显示所有笔记及其内容
5. WHEN 用户点击笔记 THEN Reader System SHALL 跳转到该笔记位置并显示笔记内容
6. WHEN 用户编辑笔记 THEN Reader System SHALL 更新笔记内容和修改时间
7. WHEN 用户删除笔记 THEN Reader System SHALL 从列表中移除该笔记
8. WHEN 笔记数据变化 THEN Reader System SHALL 立即保存到 localStorage

### Requirement 5

**User Story:** 作为读者，我希望阅读器能够应用我在设置中配置的字体和样式偏好。

#### Acceptance Criteria

1. WHEN 阅读器加载 THEN Reader System SHALL 从 Settings Context 读取阅读器设置
2. WHEN 应用字体大小设置 THEN Reader System SHALL 将字体大小应用到书籍内容
3. WHEN 应用字体系列设置 THEN Reader System SHALL 将字体系列应用到书籍内容
4. WHEN 应用行高设置 THEN Reader System SHALL 将行高应用到书籍内容
5. WHEN 应用阅读器主题设置 THEN Reader System SHALL 应用对应的颜色主题（light/dark/sepia）
6. WHEN 设置变化 THEN Reader System SHALL 立即更新阅读器样式
7. WHEN 设置不可用 THEN Reader System SHALL 使用默认样式

### Requirement 6

**User Story:** 作为读者，我希望有一个侧边栏来管理书签、高亮和笔记，方便我快速访问和管理这些内容。

#### Acceptance Criteria

1. WHEN 用户点击侧边栏按钮 THEN Reader System SHALL 打开/关闭侧边栏
2. WHEN 侧边栏打开 THEN Reader System SHALL 显示三个标签页（书签、高亮、笔记）
3. WHEN 用户切换标签页 THEN Reader System SHALL 显示对应类型的内容列表
4. WHEN 列表为空 THEN Reader System SHALL 显示友好的空状态提示
5. WHEN 用户在列表中搜索 THEN Reader System SHALL 过滤显示匹配的项目
6. WHEN 侧边栏在移动设备上 THEN Reader System SHALL 使用全屏模式显示

### Requirement 7

**User Story:** 作为读者，我希望能够导出我的书签、高亮和笔记，以便备份或在其他设备上使用。

#### Acceptance Criteria

1. WHEN 用户点击"导出"按钮 THEN Reader System SHALL 生成包含所有阅读数据的 JSON 文件
2. WHEN 导出文件生成 THEN Reader System SHALL 包含书籍信息、书签、高亮和笔记
3. WHEN 导出完成 THEN Reader System SHALL 触发浏览器下载该文件
4. WHEN 用户导入文件 THEN Reader System SHALL 验证文件格式并合并数据
5. WHEN 导入数据冲突 THEN Reader System SHALL 保留最新的数据

### Requirement 8

**User Story:** 作为读者，我希望阅读器有键盘快捷键，以便更高效地导航和操作。

#### Acceptance Criteria

1. WHEN 用户按左箭头键 THEN Reader System SHALL 翻到上一页
2. WHEN 用户按右箭头键 THEN Reader System SHALL 翻到下一页
3. WHEN 用户按 B 键 THEN Reader System SHALL 添加书签到当前位置
4. WHEN 用户按 H 键 THEN Reader System SHALL 高亮选中的文本
5. WHEN 用户按 N 键 THEN Reader System SHALL 为选中的文本添加笔记
6. WHEN 用户按 S 键 THEN Reader System SHALL 打开/关闭侧边栏
7. WHEN 用户按 ESC 键 THEN Reader System SHALL 关闭打开的对话框或侧边栏

### Requirement 9

**User Story:** 作为读者，我希望能够看到阅读进度百分比，了解我在书籍中的位置。

#### Acceptance Criteria

1. WHEN 阅读器显示内容 THEN Reader System SHALL 计算并显示当前阅读进度百分比
2. WHEN 用户翻页 THEN Reader System SHALL 更新进度百分比
3. WHEN 进度显示 THEN Reader System SHALL 同时显示当前页码和总页数（如果可用）
4. WHEN 用户点击进度条 THEN Reader System SHALL 跳转到对应位置

### Requirement 10

**User Story:** 作为读者，我希望阅读器界面简洁，不会分散我的注意力。

#### Acceptance Criteria

1. WHEN 用户开始阅读 THEN Reader System SHALL 自动隐藏工具栏（3 秒后）
2. WHEN 用户移动鼠标或触摸屏幕 THEN Reader System SHALL 显示工具栏
3. WHEN 工具栏显示 THEN Reader System SHALL 在 3 秒无操作后自动隐藏
4. WHEN 用户点击内容区域 THEN Reader System SHALL 切换工具栏显示/隐藏
5. WHEN 侧边栏打开 THEN Reader System SHALL 保持工具栏可见
