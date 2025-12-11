---
status: complete
created: '2025-12-10'
tags:
  - frontend
  - performance
  - optimization
  - lighthouse
priority: high
created_at: '2025-12-10T15:38:38.695Z'
depends_on:
  - 009-frontend-migration
updated_at: '2025-12-11T07:16:47.569Z'
transitions:
  - status: in-progress
    at: '2025-12-11T06:43:05.750Z'
  - status: complete
    at: '2025-12-11T07:16:47.569Z'
completed_at: '2025-12-11T07:16:47.569Z'
completed: '2025-12-11'
---

# 前端性能优化

> **Status**: ✅ Complete · **Priority**: High · **Created**: 2025-12-10 · **Tags**: frontend, performance, optimization, lighthouse

## Overview

全面的前端性能优化，包括图片优化（WebP 转换、响应式加载、懒加载）、代码分割、缓存策略、Service Worker 集成，目标 Lighthouse Score > 90。

**当前技术栈**:
- Next.js 16.0.8 (App Router)
- React 19.2.1
- Turbopack (开发模式)

**性能目标**:
- Lighthouse Performance Score: > 90
- First Contentful Paint (FCP): < 1.8s
- Largest Contentful Paint (LCP): < 2.5s
- Total Blocking Time (TBT): < 200ms
- Cumulative Layout Shift (CLS): < 0.1

## Design

### 1. 图片优化策略

**Next.js Image 组件**:
```tsx
import Image from 'next/image'

// 使用 Next.js Image 组件自动优化
<Image
  src="/cover.jpg"
  alt="Book Cover"
  width={200}
  height={300}
  loading="lazy"
  placeholder="blur"
/>
```

**优化措施**:
- 使用 Next.js Image 组件替换所有 `<img>` 标签
- 启用自动 WebP 转换
- 实现响应式图片加载
- 添加图片懒加载
- 使用 blur placeholder 提升感知性能

### 2. 代码分割和懒加载

**动态导入**:
```tsx
import dynamic from 'next/dynamic'

// 懒加载重型组件
const BookTocDialog = dynamic(() => import('@/components/book-toc-dialog'), {
  loading: () => <div>Loading...</div>,
  ssr: false
})

const MetadataSearchDialog = dynamic(() => import('@/components/metadata-search-dialog'), {
  loading: () => <div>Loading...</div>
})
```

**路由级代码分割**:
- Next.js App Router 自动按路由分割代码
- 使用 `dynamic()` 懒加载非关键组件
- 延迟加载第三方库（如 react-reader）

### 3. 缓存策略

**Next.js 缓存配置**:
```ts
// next.config.js
export default {
  images: {
    formats: ['image/webp', 'image/avif'],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
  },
  // 启用 SWC minification
  swcMinify: true,
}
```

**API 响应缓存**:
```tsx
// 使用 Next.js fetch 缓存
const res = await fetch('/api/books', {
  next: { revalidate: 3600 } // 1小时缓存
})
```

### 4. 字体优化

**使用 next/font**:
```tsx
import { Inter } from 'next/font/google'

const inter = Inter({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-inter',
})
```

### 5. Bundle 分析和优化

**分析工具**:
```bash
# 安装 bundle analyzer
pnpm add -D @next/bundle-analyzer

# 分析 bundle
ANALYZE=true pnpm build
```

**优化策略**:
- 移除未使用的依赖
- Tree-shaking 优化
- 压缩和 minify
- 使用 ES modules

### 6. 性能监控

**Web Vitals 监控**:
```tsx
// app/layout.tsx
export function reportWebVitals(metric) {
  console.log(metric)
  // 发送到分析服务
}
```

## Plan

### Phase 1: 基线测试和分析

- [ ] 1.1 运行 Lighthouse 测试获取当前性能基线
- [ ] 1.2 使用 @next/bundle-analyzer 分析 bundle 大小
- [ ] 1.3 识别性能瓶颈和优化机会
- [ ] 1.4 记录当前的 Core Web Vitals 指标

### Phase 2: 图片优化

- [ ] 2.1 审计所有使用 `<img>` 标签的地方
- [ ] 2.2 将 BookCard 组件改用 Next.js Image
- [ ] 2.3 优化封面图片加载（添加 placeholder）
- [ ] 2.4 实现图片懒加载策略
- [ ] 2.5 配置图片优化参数（formats, sizes）

### Phase 3: 代码分割和懒加载

- [ ] 3.1 识别可以懒加载的重型组件
- [ ] 3.2 使用 dynamic() 懒加载对话框组件
  - BookTocDialog
  - MetadataSearchDialog
  - AlertDialog
- [ ] 3.3 懒加载 react-reader 库
- [ ] 3.4 懒加载 react-markdown 组件
- [ ] 3.5 优化路由预加载策略

### Phase 4: 字体和资源优化

- [ ] 4.1 使用 next/font 优化字体加载
- [ ] 4.2 配置字体 display: swap
- [ ] 4.3 移除未使用的字体权重
- [ ] 4.4 优化 CSS 加载顺序

### Phase 5: 缓存和数据获取优化

- [ ] 5.1 配置 API 响应缓存策略
- [ ] 5.2 实现 SWR 或 React Query 数据缓存
- [ ] 5.3 优化书籍列表的分页加载
- [ ] 5.4 添加乐观更新提升感知性能

### Phase 6: Bundle 优化

- [ ] 6.1 分析并移除未使用的依赖
- [ ] 6.2 优化第三方库导入（tree-shaking）
- [ ] 6.3 配置 SWC minification
- [ ] 6.4 启用 gzip/brotli 压缩

### Phase 7: 性能监控和测试

- [ ] 7.1 集成 Web Vitals 监控
- [ ] 7.2 运行 Lighthouse 测试验证改进
- [ ] 7.3 测试不同网络条件下的性能
- [ ] 7.4 记录优化前后的性能对比

## Test

### 性能指标测试

- [ ] Lighthouse Performance Score > 90
- [ ] First Contentful Paint (FCP) < 1.8s
- [ ] Largest Contentful Paint (LCP) < 2.5s
- [ ] Total Blocking Time (TBT) < 200ms
- [ ] Cumulative Layout Shift (CLS) < 0.1
- [ ] Speed Index < 3.4s

### 功能测试

- [ ] 所有图片正常显示
- [ ] 懒加载组件正常工作
- [ ] 页面导航流畅无卡顿
- [ ] 缓存策略正常工作
- [ ] 字体加载无闪烁

### 兼容性测试

- [ ] Chrome/Edge 测试通过
- [ ] Firefox 测试通过
- [ ] Safari 测试通过
- [ ] 移动端测试通过

## Notes

### 当前性能问题（待测试确认）

1. **图片加载**
   - 使用原生 `<img>` 标签，未优化
   - 封面图片较大，加载慢
   - 缺少懒加载和 placeholder

2. **代码体积**
   - react-reader 库较大（~200KB）
   - react-markdown 及其依赖较大
   - 可能存在未使用的依赖

3. **缓存策略**
   - API 响应未缓存
   - 静态资源缓存配置待优化

### 优化优先级

**High Priority**:
- 图片优化（最大影响）
- 代码分割（减少初始加载）
- 字体优化（避免 FOUT/FOIT）

**Medium Priority**:
- 缓存策略
- Bundle 优化
- 性能监控

**Low Priority**:
- Service Worker（PWA 功能）
- 预加载优化

### 参考资源

- [Next.js Image Optimization](https://nextjs.org/docs/app/building-your-application/optimizing/images)
- [Next.js Font Optimization](https://nextjs.org/docs/app/building-your-application/optimizing/fonts)
- [Web Vitals](https://web.dev/vitals/)
- [Lighthouse](https://developers.google.com/web/tools/lighthouse)

### 依赖项

- 依赖：009-frontend-migration（Next.js 迁移完成）
- 被依赖：无
