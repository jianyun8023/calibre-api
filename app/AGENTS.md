# Calibre Pages - 前端开发指南

> Vue.js 3 + TypeScript 前端应用，书籍管理和智能阅读界面

## 🎯 核心原则

### 技术选型
- **框架**: Vue 3 Composition API + TypeScript
- **UI**: Element Plus
- **构建**: Vite

### 设计目标
1. **类型安全**: TypeScript 覆盖率 > 80%
2. **性能优先**: 首屏 < 2s, 列表滚动 60fps
3. **用户体验**: 加载状态完善，错误提示友好，移动端适配

## 🏗️ 分层架构

```
API 层 → 封装 HTTP 请求，统一错误处理
  ↓
Types → TypeScript 类型定义（与后端对齐）
  ↓
Components → 可复用 UI 组件（单一职责）
  ↓
Views → 页面组件（组合业务逻辑）
```

### 关键约束
- API 调用必须封装在 `api/` 层，不在组件中直接 fetch
- 组件类型使用 `interface`（可扩展）而非 `type`
- 所有 API 响应必须定义 TypeScript 类型
- 样式使用 `scoped`，避免全局污染

## 📐 核心规范

### 1. 组件设计

**原则**: 单一职责、类型明确、状态最小化

```typescript
// ✅ Props/Emits 使用 TypeScript 接口
interface Props {
  bookId: number
  readonly?: boolean
}
const props = withDefaults(defineProps<Props>(), {
  readonly: false
})

interface Emits {
  (e: 'update', book: Book): void
}
const emit = defineEmits<Emits>()

// ✅ 状态选择：基本类型用 ref，相关状态用 reactive
const loading = ref(false)
const filters = reactive({ keyword: '', category: 'all' })
```

### 2. API 封装

**原则**: 统一入口、类型安全、错误处理

```typescript
// api/books.ts
export const booksApi = {
  getBook: (id: number) => apiRequest<Book>(`/api/books/${id}`),
  search: (opts: SearchOptions) => apiRequest<Book[]>('/api/search', {
    method: 'POST', body: JSON.stringify(opts)
  })
}

// apiUtils.ts - 统一错误处理
async function apiRequest<T>(url: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(url, { headers: { 'Content-Type': 'application/json' }, ...opts })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}
```

**权衡**: 
- ✅ 集中管理 API，易于维护和模拟测试
- ⚠️ 泛型类型需要手动指定，运行时无法验证

### 3. 类型定义

**原则**: 与后端对齐、使用 interface（可扩展）

```typescript
// types/book.ts - 对应后端 Go 结构
export interface Book {
  id: number
  title: string
  authors: string[]
  publisher?: string
  isbn?: string
  last_modified: string
}

export interface SearchOptions {
  query: string
  strategy?: 'keyword' | 'semantic' | 'hybrid'
  limit?: number
}
```

**成功标准**: 类型定义与后端 API 响应完全一致，避免运行时类型错误

### 4. 样式管理

**原则**: 组件隔离、主题统一、响应式优先

```vue
<style scoped>
.book-card { padding: 16px; }
</style>
```

```scss
// CSS 变量主题
:root {
  --primary-color: #409eff;
  --bg-color: #ffffff;
}

[data-theme='dark'] {
  --bg-color: #1a1a1a;
}
```

```css
/* 移动端断点 */
@media (max-width: 768px) {
  .book-grid { grid-template-columns: repeat(2, 1fr); }
}
```

### 5. 性能优化

**策略**: 懒加载、虚拟滚动、防抖节流

```typescript
// 路由懒加载（减少首屏体积）
const routes = [
  { path: '/books', component: () => import('@/views/BookList.vue') }
]

// 搜索防抖（减少 API 调用）
import { useDebounceFn } from '@vueuse/core'
const search = useDebounceFn(async (q: string) => {
  await booksApi.search({ query: q })
}, 300)
```

```vue
<!-- 大列表虚拟滚动 -->
<el-virtual-scroll :items="books" :item-size="120">
  <template #default="{ item }"><BookCard :book="item" /></template>
</el-virtual-scroll>

<!-- 图片懒加载 -->
<el-image :src="book.cover_url" lazy />
```

**权衡**: 虚拟滚动增加复杂度，仅用于 > 100 项的列表

## 🔧 关键场景

### 数据加载
```typescript
// 标准模式：loading + error 状态
const loading = ref(false)
const books = ref<Book[]>([])

async function load() {
  loading.value = true
  try {
    books.value = await booksApi.getBooks({ limit: 20 })
  } catch (error) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}
```

### 分页
```typescript
const pagination = reactive({ page: 1, limit: 20, total: 0 })

async function loadPage() {
  const { data, total } = await booksApi.getBooks(pagination)
  books.value = data
  pagination.total = total
}
```

### 搜索缓存（可选优化）
```typescript
const cache = new Map<string, Book[]>()

async function search(query: string) {
  if (cache.has(query)) return cache.get(query)!
  const results = await booksApi.search({ query })
  cache.set(query, results)
  return results
}
```

**权衡**: 缓存提升体验，但需要考虑失效策略（如时间过期、容量限制）

## 🚀 开发流程

```bash
# 本地开发
cd app/calibre-pages
pnpm install
pnpm dev  # localhost:5173

# 构建生产
pnpm build  # 输出到 dist/
```

```typescript
// vite.config.js - API 代理
export default {
  server: {
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: true }
    }
  }
}
```

## ⚠️ 关键约束

### 必须遵守
1. **类型安全**: 所有 API 调用指定返回类型
2. **错误处理**: 所有异步操作 try-catch + 用户提示
3. **样式隔离**: 组件样式使用 `scoped`
4. **性能**: 避免 computed 中调用 API，大列表用虚拟滚动

### 安全检查
1. 用户输入转义（防 XSS）
2. Markdown 渲染配置安全选项
3. 敏感操作二次确认

## 🎯 成功标准

**代码质量**
- TypeScript 类型覆盖率 > 80%
- 组件可复用性高（单一职责）
- API 封装完善，错误处理统一

**性能指标**
- 首屏加载 < 2s
- 列表滚动 60fps
- 图片懒加载 + 占位

**用户体验**
- 移动端适配（响应式断点）
- 加载/错误状态完善
- 操作响应及时（防抖 300ms）

---

**版本**: 1.0.0 | **更新**: 2024-12-08 | **维护**: jianyun8023

