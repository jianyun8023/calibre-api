# BookGrid Component

统一的图书网格布局组件，支持加载状态、空状态和响应式布局。

## 使用示例

### 基础用法

```tsx
import { BookGrid } from '@/components/book-grid'

<BookGrid books={books} loading={loading} />
```

### 自定义列数

```tsx
<BookGrid 
  books={books}
  columns={{
    base: 2,  // 移动端 2 列
    md: 3,    // 平板 3 列
    lg: 4,    // 桌面 4 列
    xl: 5     // 大屏 5 列
  }}
/>
```

### 显示更多信息

```tsx
<BookGrid 
  books={books}
  moreInfo={true}  // 显示出版社、出版日期等
/>
```

### 使用代理图片

```tsx
<BookGrid 
  books={books}
  proxyImage={true}  // 使用后端代理加载图片
/>
```

### 自定义空状态消息

```tsx
<BookGrid 
  books={books}
  emptyMessage="没有找到相关书籍"
/>
```

### 自定义骨架屏数量

```tsx
<BookGrid 
  books={books}
  loading={loading}
  skeletonCount={15}  // 显示 15 个骨架屏
/>
```

## Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `books` | `Book[]` | **required** | 书籍数组 |
| `loading` | `boolean` | `false` | 是否显示加载状态 |
| `emptyMessage` | `string` | `"No books found."` | 空状态消息 |
| `moreInfo` | `boolean` | `false` | 是否显示更多信息 |
| `proxyImage` | `boolean` | `false` | 是否使用代理图片 |
| `skeletonCount` | `number` | `10` | 骨架屏数量 |
| `columns` | `object` | `{base:2, sm:2, md:3, lg:4, xl:5}` | 响应式列数配置 |
| `className` | `string` | `undefined` | 自定义 CSS 类名 |

## 状态

### 加载状态
当 `loading={true}` 时，显示骨架屏：
```tsx
<BookGrid books={[]} loading={true} skeletonCount={10} />
```

### 空状态
当 `loading={false}` 且 `books.length === 0` 时，显示空状态消息：
```tsx
<BookGrid books={[]} loading={false} emptyMessage="暂无书籍" />
```

### 正常状态
当 `loading={false}` 且 `books.length > 0` 时，显示书籍网格：
```tsx
<BookGrid books={bookList} loading={false} />
```

## 性能优化

- 使用 `React.memo` 避免不必要的重渲染
- 使用 `useMemo` 缓存 grid class 计算
- 使用 `book.id` 作为 key 优化列表渲染

## 替换示例

### 替换前
```tsx
<div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6">
  {loading ? (
    Array.from({ length: 10 }).map((_, i) => (
      <Skeleton key={i} className="h-48 w-full rounded-xl" />
    ))
  ) : books.length === 0 ? (
    <div className="text-center py-20 text-muted-foreground">
      No books found.
    </div>
  ) : (
    books.map((book) => <BookCard key={book.id} book={book} />)
  )}
</div>
```

### 替换后
```tsx
<BookGrid books={books} loading={loading} />
```

代码减少 ~80%！
