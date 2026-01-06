# Calibre API - Next.js Frontend

现代化的书籍管理前端，基于 Next.js 15 和 Shadcn/UI 构建。

## ✨ 特性

### 核心功能
- 📚 **书籍管理** - 浏览、搜索、查看书籍详情
- 📖 **EPUB 阅读器** - 基于 react-reader，支持阅读进度保存
- 🔍 **智能搜索** - 关键词/语义/混合三种搜索模式
- 💬 **AI 对话** - Vercel AI SDK 流式对话，支持书籍推荐
- 📊 **出版社统计** - 出版社列表、统计信息、书籍筛选
- ⚙️ **批量编辑** - 元数据批量更新、预览变更
- 🔧 **任务管理** - 异步任务监控、实时进度显示
- 🎨 **主题切换** - 亮色/暗色/系统主题

### 技术亮点
- **Next.js 16** - App Router, Static Export
- **React 19** - 最新的 React 特性
- **Shadcn/UI** - 25+ 高质量组件
- **Tailwind CSS 4** - Glassmorphism 样式系统
- **Vercel AI SDK** - 无缝处理 SSE 流式响应
- **TypeScript 5** - 完整的类型安全
- **Zustand** - 轻量级状态管理
- **next-themes** - 主题切换支持

## 🚀 快速开始

### 前置要求
- Node.js 18+
- pnpm (推荐) 或 npm

### 安装依赖
```bash
cd web-next
pnpm install
```

### 开发模式
```bash
pnpm dev
```

访问 http://localhost:3000

### 构建生产版本
```bash
# 静态导出（推荐，用于 Go 后端托管）
pnpm build

# 输出目录: out/
```

### 集成到 Go 后端
```bash
# 1. 构建前端
cd web-next
pnpm build

# 2. 复制到 Go 项目
# 输出已经在 out/ 目录，Go 后端会自动托管

# 3. 启动 Go 后端
cd ..
./calibre-api
```

访问 http://localhost:8080

## 📁 项目结构

```
web-next/
├── src/
│   ├── app/                    # Next.js App Router
│   │   ├── page.tsx           # 首页（随机推荐 + 最近添加）
│   │   ├── books/             # 书籍列表页
│   │   ├── detail/[id]/       # 书籍详情页
│   │   ├── read/[id]/         # EPUB 阅读器
│   │   ├── search/            # 搜索页面
│   │   ├── chat/              # AI 对话页面
│   │   ├── settings/          # 设置页面
│   │   ├── tasks/             # 任务管理
│   │   ├── publisher/         # 出版社统计
│   │   ├── metadata/manager/  # 批量元数据编辑
│   │   ├── layout.tsx         # 全局布局
│   │   └── globals.css        # 全局样式
│   ├── components/
│   │   ├── ui/                # Shadcn/UI 组件（25+ 组件）
│   │   ├── app-header.tsx     # 全局头部
│   │   ├── app-sidebar.tsx    # 侧边导航
│   │   ├── book-card.tsx      # 书籍卡片（3D 倾斜效果）
│   │   ├── markdown.tsx       # Markdown 渲染器
│   │   ├── mode-toggle.tsx    # 主题切换按钮
│   │   └── theme-provider.tsx # 主题提供者
│   ├── lib/
│   │   ├── api/               # API 客户端封装
│   │   ├── api-client.ts      # Fetch 封装（统一错误处理）
│   │   ├── utils.ts           # 工具函数
│   │   └── cn.ts              # Tailwind 类名合并
│   ├── stores/
│   │   ├── theme-store.ts     # 主题状态
│   │   └── ui-store.ts        # UI 状态（侧边栏等）
│   └── types/
│       ├── book.ts            # 书籍类型定义
│       └── api.ts             # API 响应类型
├── public/                     # 静态资源
├── tailwind.config.ts         # Tailwind 配置（含 Glassmorphism 插件）
├── next.config.ts             # Next.js 配置（静态导出）
├── tsconfig.json              # TypeScript 配置
└── package.json               # 依赖管理
```

## 🎨 核心组件

### BookCard
书籍卡片组件，包含：
- 3D 倾斜效果（鼠标悬停）
- Glassmorphism 样式
- 封面图片、标题、作者、评分
- 标签展示

```tsx
import { BookCard } from "@/components/book-card"

<BookCard book={book} />
```

### Markdown 渲染器
支持 GitHub Flavored Markdown 和代码高亮：
```tsx
import { MarkdownRenderer } from "@/components/markdown"

<MarkdownRenderer content={markdown} />
```

### AppSidebar
侧边导航栏，包含：
- 响应式设计（桌面固定，移动端抽屉）
- 导航菜单
- 主题切换
- 用户信息

## 🎯 页面功能

### 首页 (`/`)
- 随机推荐 12 本书籍
- 最近添加的书籍列表
- 快速搜索入口

### 书籍列表 (`/books`)
- 基于游标的无限分页
- 骨架屏加载状态
- 书籍卡片展示

### 书籍详情 (`/detail/[id]`)
- 完整元数据展示
- 标签管理
- 文件下载
- 在线阅读入口
- 删除/编辑功能

### 阅读器 (`/read/[id]`)
- react-reader EPUB 阅读器
- 阅读进度自动保存（localStorage）
- 主题适配（亮色/暗色）
- 目录导航

### 搜索 (`/search`)
- 三种搜索模式：
  - **关键词搜索** - 精确匹配标题、作者、ISBN
  - **语义搜索** - AI 理解自然语言查询
  - **混合搜索** - 结合关键词和语义
- 高级过滤（作者、出版社、标签）
- 实时搜索建议

### AI 对话 (`/chat`)
- Vercel AI SDK 流式响应
- Markdown 渲染（代码高亮）
- 书籍推荐和搜索
- 对话历史管理

### 设置 (`/settings`)
- **外观**: 主题切换（亮色/暗色/系统）
- **API 配置**: 后端 API 端点设置
- **阅读器**: 字体大小、字体系列、行高
- **搜索**: 默认搜索模式、每页结果数

### 任务管理 (`/tasks`)
- 启动新任务（Qdrant 同步、TOC 提取、检查缺失）
- 实时状态更新（每 2 秒轮询）
- 进度条显示
- 任务历史记录

### 出版社 (`/publisher`)
- 统计信息（总数、书籍数、平均值）
- Top 5 出版社排行榜
- 搜索和过滤
- 点击跳转到书籍列表

### 元数据管理 (`/metadata/manager`)
- 搜索和选择书籍
- 批量编辑元数据
- 预览变更
- 批量提交更新

## 🛠️ 开发指南

### 添加新页面
```bash
# 创建页面目录
mkdir -p src/app/my-page

# 创建页面文件
touch src/app/my-page/page.tsx
```

### 添加 Shadcn/UI 组件
```bash
# 使用 shadcn-ui CLI（如果已安装）
npx shadcn-ui@latest add [component-name]

# 或者手动创建组件
touch src/components/ui/my-component.tsx
```

### 调用 API
```tsx
import { apiRequest } from "@/lib/api-client"

// 自动处理标准响应格式和错误
const books = await apiRequest<Book[]>("/api/books/all")
```

### 状态管理
```tsx
import { useUIStore } from "@/stores/ui-store"

const { sidebarOpen, setSidebarOpen } = useUIStore()
```

### 样式规范
- 使用 Tailwind CSS 工具类
- Glassmorphism 效果使用 `.glass` 类
- 响应式前缀: `sm:`, `md:`, `lg:`, `xl:`
- 暗黑模式: `dark:` 前缀

## 📦 依赖说明

### 核心依赖
```json
{
  "next": "16.0.8",
  "react": "19.2.1",
  "typescript": "5.x"
}
```

### UI 库
```json
{
  "@radix-ui/react-*": "最新版本",  // Shadcn/UI 底层组件
  "tailwindcss": "4.x",              // CSS 框架
  "lucide-react": "最新版本",        // 图标库
  "next-themes": "最新版本"          // 主题管理
}
```

### 功能库
```json
{
  "ai": "最新版本",                  // Vercel AI SDK
  "@ai-sdk/react": "最新版本",       // React Hooks
  "react-reader": "最新版本",        // EPUB 阅读器
  "epubjs": "最新版本",              // EPUB 解析库
  "react-markdown": "最新版本",      // Markdown 渲染
  "remark-gfm": "最新版本",          // GitHub Flavored Markdown
  "rehype-highlight": "最新版本",    // 代码高亮
  "zustand": "最新版本",             // 状态管理
  "sonner": "最新版本"               // Toast 通知
}
```

## 🎨 设计系统

### 颜色
- **Primary**: 主色调（可在 `tailwind.config.ts` 中配置）
- **Secondary**: 次要颜色
- **Accent**: 强调色
- **Muted**: 弱化文本
- **Destructive**: 危险操作

### 间距
- `p-*` / `m-*`: 4px 的倍数
- `gap-*`: Flex/Grid 间距
- `space-*`: 子元素间距

### 字体
- **Sans**: 系统默认无衬线字体
- **Serif**: 衬线字体（阅读器）
- **Mono**: 等宽字体（代码）

### 动画
- `transition-*`: 过渡效果
- `animate-*`: 关键帧动画
- `hover:` / `focus:`: 交互状态

## 🚀 性能优化

### 已实现
- ✅ Next.js 自动代码分割
- ✅ 图片懒加载（next/image）
- ✅ 基于游标的分页（减少数据传输）
- ✅ React Server Components（部分页面）
- ✅ 静态导出（快速加载）

### 待优化
- [ ] 虚拟滚动（大列表）
- [ ] Service Worker（离线支持）
- [ ] 图片优化（WebP/AVIF）
- [ ] Bundle 分析和优化

## 🐛 已知问题

1. **Next.js Dev Server 网络错误**
   - 症状: `uv_interface_addresses` 错误
   - 解决: 不影响功能，可忽略
   - 生产构建正常

2. **EPUB 阅读器初次加载慢**
   - 原因: epub.js 库较大
   - 解决: 已使用动态导入 (`next/dynamic`)

## 📝 开发笔记

### 迁移自 Vue 3
本项目从 Vue 3 + Element Plus 迁移而来，主要变更：
- **状态管理**: Pinia → Zustand
- **UI 库**: Element Plus → Shadcn/UI
- **路由**: Vue Router → Next.js App Router
- **阅读器**: vue-reader → react-reader
- **Markdown**: markdown-it → react-markdown

### 技术选型理由
- **Next.js**: React 生态最成熟的框架，SEO 友好
- **Shadcn/UI**: 高度可定制，组件即代码，无 npm 依赖
- **Vercel AI SDK**: 一流的 AI 体验，自动处理 SSE
- **Tailwind CSS**: 快速开发，一致的设计系统

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交变更 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 代码规范
- 使用 ESLint 和 Prettier
- 遵循 TypeScript 严格模式
- 组件文件名使用 PascalCase
- 工具函数使用 camelCase

## 📄 许可证

MIT License - 详见 [LICENSE](../LICENSE) 文件

## 🔗 相关链接

- [后端仓库](https://github.com/jianyun8023/calibre-api)
- [Vue.js 前端](../app/calibre-pages)
- [Next.js 文档](https://nextjs.org/docs)
- [Shadcn/UI 文档](https://ui.shadcn.com)
- [Vercel AI SDK](https://sdk.vercel.ai/docs)
- [react-reader](https://github.com/gerhardsletten/react-reader)

---

**维护者**: jianyun8023  
**当前版本**: 2.0.0 (Next.js 重构)  
**最后更新**: 2025-12-10
