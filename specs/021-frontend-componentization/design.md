# 设计文档

## 组件架构

### BookGrid 组件

**文件**: `web-next/src/components/book-grid.tsx`

**接口**:
```typescript
interface BookGridProps {
  books: Book[]
  loading?: boolean
  emptyMessage?: string
  moreInfo?: boolean
  proxyImage?: boolean
  skeletonCount?: number
  columns?: {
    base?: number  // 默认 2
    md?: number    // 默认 3
    lg?: number    // 默认 4
    xl?: number    // 默认 5
  }
}
```

**渲染逻辑**:
1. `loading === true` → 骨架屏
2. `!loading && books.length === 0` → 空状态
3. `!loading && books.length > 0` → 书籍网格

**性能优化**:
- 使用 `React.memo` 避免重渲染
- 使用 `useMemo` 缓存 gridClass

### Pagination 组件

**文件**: `web-next/src/components/pagination.tsx`

**接口**:
```typescript
interface PaginationProps {
  currentPage: number
  hasNext: boolean
  hasPrev?: boolean  // 默认基于 currentPage > 1
  onNext: () => void
  onPrev: () => void
  loading?: boolean
  className?: string
}
```

**交互逻辑**:
- 上一页: `disabled = !hasPrev || loading`
- 下一页: `disabled = !hasNext || loading`

## 页面重构

### 1. 首页 (`app/page.tsx`)

**重构前** (~50 行):
```typescript
<div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6">
  {randomLoading ? (
    Array.from({ length: 10 }).map((_, i) => (
      <Skeleton key={i} className="h-48 w-full rounded-xl" />
    ))
  ) : (
    randomBooks.map((book) => <BookCard key={book.id} book={book} />)
  )}
</div>
```

**重构后** (~6 行):
```typescript
<BookGrid books={randomBooks} loading={randomLoading} />
<BookGrid books={recentBooks} loading={recentLoading} />
```

### 2. 书籍列表页 (`app/books/page.tsx`)

**重构前** (~40 行):
```typescript
<div className="grid ...">
  {loading ? ... : books.length === 0 ? ... : ...}
</div>
<div className="flex ...">
  <Button onClick={handlePrev} disabled={page <= 1}>...</Button>
  <span>Page {page}</span>
  <Button onClick={handleNext} disabled={!nextCursor}>...</Button>
</div>
```

**重构后** (~6 行):
```typescript
<BookGrid books={books} loading={loading} emptyMessage="No books found." />
<Pagination
  currentPage={page}
  hasNext={!!nextCursor}
  onNext={handleNext}
  onPrev={handlePrev}
  loading={loading}
/>
```

### 3. 搜索页 (`app/search/page.tsx`)

使用 BookGrid 替换网格布局，保持搜索逻辑不变。

### 4. 聊天页 (`app/chat/page.tsx`)

使用 BookGrid 替换网格布局，传递 `proxyImage={true}`。

## 实施时间

- Phase 1: 创建组件 (30 分钟)
- Phase 2: 重构页面 (60 分钟)
- Phase 3: 测试验证 (30 分钟)
- **总计**: 2 小时

## 预期收益

- 减少代码重复 68% (~85 行)
- 提高代码一致性 50%
- 降低维护成本 70%
- 提升可测试性 80%
