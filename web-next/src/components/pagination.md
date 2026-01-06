# Pagination Component

统一的分页组件，支持上一页/下一页导航和禁用状态。

## 使用示例

### 基础用法

```tsx
import { Pagination } from '@/components/pagination'

<Pagination
  currentPage={page}
  totalPages={totalPages}
  hasNext={page < totalPages}
  hasPrev={page > 1}
  onNext={handleNext}
  onPrev={handlePrev}
/>
```

### 带加载状态

```tsx
<Pagination
  currentPage={page}
  totalPages={totalPages}
  hasNext={!!nextCursor}
  onNext={handleNext}
  onPrev={handlePrev}
  loading={isLoading}  // 禁用所有按钮
/>
```

### 自定义样式

```tsx
<Pagination
  currentPage={page}
  totalPages={totalPages}
  hasNext={page < totalPages}
  hasPrev={page > 1}
  onNext={handleNext}
  onPrev={handlePrev}
  className="mt-8"
/>
```

## Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `currentPage` | `number` | **required** | 当前页码 |
| `totalPages` | `number` | **required** | 总页数 |
| `hasNext` | `boolean` | **required** | 是否有下一页 |
| `hasPrev` | `boolean` | `currentPage > 1` | 是否有上一页 |
| `onNext` | `() => void` | **required** | 下一页回调 |
| `onPrev` | `() => void` | **required** | 上一页回调 |
| `loading` | `boolean` | `false` | 是否加载中 |
| `className` | `string` | `undefined` | 自定义 CSS 类名 |

## 状态

### 禁用状态
- 上一页按钮: 当 `!hasPrev || loading` 时禁用
- 下一页按钮: 当 `!hasNext || loading` 时禁用

### 加载状态
当 `loading={true}` 时，所有按钮都被禁用

## 可访问性

- 按钮包含 `aria-label` 属性
- 禁用状态清晰可见
- 支持键盘导航

## 使用场景

### 1. 前端分页 (Publisher 页面)
```tsx
const ITEMS_PER_PAGE = 50
const [currentPage, setCurrentPage] = useState(1)

const totalPages = Math.ceil(items.length / ITEMS_PER_PAGE)
const paginatedItems = items.slice(
  (currentPage - 1) * ITEMS_PER_PAGE,
  currentPage * ITEMS_PER_PAGE
)

<Pagination
  currentPage={currentPage}
  totalPages={totalPages}
  hasNext={currentPage < totalPages}
  hasPrev={currentPage > 1}
  onNext={() => setCurrentPage(prev => prev + 1)}
  onPrev={() => setCurrentPage(prev => prev - 1)}
/>
```

### 2. 后端分页 (Books 页面)
```tsx
const [page, setPage] = useState(1)
const [nextCursor, setNextCursor] = useState<string | null>(null)

<Pagination
  currentPage={page}
  totalPages={999}  // 未知总页数
  hasNext={!!nextCursor}
  onNext={() => setPage(prev => prev + 1)}
  onPrev={() => setPage(prev => prev - 1)}
  loading={loading}
/>
```

## 替换示例

### 替换前
```tsx
<div className="flex items-center justify-center gap-4 py-8">
  <Button onClick={handlePrev} disabled={page <= 1}>
    <ArrowLeft className="h-4 w-4 mr-2" />
    Previous
  </Button>
  <span>Page {page}</span>
  <Button onClick={handleNext} disabled={!nextCursor}>
    Next
    <ArrowRight className="h-4 w-4 ml-2" />
  </Button>
</div>
```

### 替换后
```tsx
<Pagination
  currentPage={page}
  totalPages={totalPages}
  hasNext={!!nextCursor}
  onNext={handleNext}
  onPrev={handlePrev}
/>
```

代码减少 ~70%！
