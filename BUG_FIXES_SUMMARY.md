# Bug 修复总结

## 修复的 Bug

### Bug 1: Detail.vue - 重复调用 `fetchBook()`
**位置**: `app/calibre-pages/src/views/Detail.vue` 第 277-293 行

**问题描述**:
组件同时使用了 `watch: { '$route' }` 和 `beforeRouteUpdate` 导航守卫，当路由参数（如书籍 ID）变化时，两者都会调用 `fetchBook()`，导致：
- 每次路由参数变化时发送 2 次 API 请求
- 浪费服务器资源和网络带宽
- 可能导致竞态条件

**修复方案**:
删除 `watch: { '$route' }` 监听器，仅保留 `beforeRouteUpdate` 导航守卫。这是 Vue Router 推荐的处理同一组件内路由参数变化的标准方式。

**修复后代码**:
```typescript
// 路由守卫：在同一组件内路由参数变化时调用
beforeRouteUpdate(to, from, next) {
  // 当路由参数变化时，重新获取书籍数据
  if (to.params.id && to.params.id !== from.params.id) {
    this.fetchBook(to.params.id as string)
  }
  next()
}
```

**影响**:
- ✅ 消除重复的 API 调用
- ✅ 提升性能
- ✅ 避免潜在的竞态条件

---

### Bug 2: Chat.vue - 消息 ID 不唯一
**位置**: `app/calibre-pages/src/views/Chat.vue` 第 273-292 行

**问题描述**:
消息 ID 使用简单的时间戳生成：
- 用户消息 ID: `Date.now().toString()`
- AI 消息 ID: `(Date.now() + 1).toString()`

这导致以下问题：
1. 如果用户在同一毫秒内快速发送多条消息，会产生重复 ID
2. `Date.now() + 1` 无法保证唯一性，下一条用户消息可能与当前 AI 消息 ID 冲突
3. 在 Vue 的 `v-for` 中使用重复的 key 会导致渲染错误和状态混乱

**修复方案**:
使用 `timestamp + random` 的组合生成唯一 ID：
```typescript
const generateUniqueId = () => {
  return `${Date.now()}-${Math.random().toString(36).substring(2, 9)}`
}
```

这种方案：
- 时间戳保证时间序列性
- 随机字符串（7 位 base36）保证同一毫秒内的唯一性
- 简单高效，无需外部依赖

**修复后代码**:
```typescript
const sendMessage = async () => {
  if (!inputMessage.value.trim() || sending.value || !currentConversation.value) return
  
  // 生成唯一的消息 ID (使用 timestamp + random)
  const generateUniqueId = () => {
    return `${Date.now()}-${Math.random().toString(36).substring(2, 9)}`
  }
  
  const userMessage: Message = {
    id: generateUniqueId(),
    conversation_id: currentConversation.value.id,
    role: 'user',
    content: inputMessage.value,
    created_at: new Date().toISOString()
  }
  messages.value.push(userMessage)
  
  // ... 其他代码 ...
  
  const aiMessage: Message = {
    id: generateUniqueId(),
    conversation_id: currentConversation.value.id,
    role: 'assistant',
    content: '',
    created_at: new Date().toISOString(),
    renderedContent: ''
  }
  messages.value.push(aiMessage)
  // ...
}
```

**影响**:
- ✅ 保证消息 ID 唯一性
- ✅ 避免 Vue 渲染错误
- ✅ 支持快速连续发送消息
- ✅ 简单高效，无性能影响

---

### Bug 3: Books.vue - `prevPage()` 未重置 `nextCursor`
**位置**: `app/calibre-pages/src/views/Books.vue` 第 106-114 行

**问题描述**:
在基于游标的分页系统中：
1. 用户向前翻页时，`nextCursor` 被设置为下一页的游标
2. 当用户点击"上一页"回退时，`prevPage()` 函数从栈中恢复 `cursor`
3. **但是 `nextCursor` 没有被重置**，仍然保留前进导航时的旧值
4. 当用户再次点击"下一页"时，会使用错误的 `nextCursor`，跳过页面或显示错误数据

**示例场景**:
```
1. 用户在第1页，nextCursor = "cursor_for_page2"
2. 用户点击"下一页" → 到达第2页，nextCursor = "cursor_for_page3"
3. 用户点击"上一页" → 回到第1页，但 nextCursor 仍然是 "cursor_for_page3" (错误!)
4. 用户再次点击"下一页" → 应该去第2页，但实际会去第3页 (因为使用了错误的游标)
```

**修复方案**:
在 `prevPage()` 函数中重置 `nextCursor` 为空字符串，强制 `fetchBooks()` 从服务器获取正确的下一页游标。

**修复后代码**:
```typescript
const prevPage = () => {
  if (prevCursors.value.length > 0) {
    // Pop the last cursor from stack
    cursor.value = prevCursors.value.pop() || ''
    currentPage.value--
    // Reset nextCursor to ensure correct forward navigation
    nextCursor.value = ''
    window.scrollTo({ top: 0, behavior: 'smooth' })
    fetchBooks()
  }
}
```

**工作流程**:
1. 回退时重置 `nextCursor` 为空
2. `fetchBooks()` 执行并从服务器获取当前页数据
3. 服务器返回正确的 `next_cursor` 值
4. 后续的"下一页"操作使用正确的游标

**影响**:
- ✅ 修复游标分页的前进/后退逻辑
- ✅ 确保数据一致性
- ✅ 避免跳页或重复页面
- ✅ 符合游标分页的最佳实践

---

## 测试建议

### Bug 1 测试
1. 打开浏览器开发者工具的网络面板
2. 访问某本书的详情页 (例如 `/detail/1`)
3. 点击另一本书的链接 (例如 `/detail/2`)
4. **验证**: 应该只看到 **1 次** API 请求 `/api/book/2`

### Bug 2 测试
1. 打开聊天页面
2. 快速连续发送多条消息（尽可能快地按回车）
3. 打开 Vue DevTools 查看 `messages` 数组
4. **验证**: 所有消息的 `id` 应该都是唯一的，格式如 `1732800000000-abc1234`

### Bug 3 测试
1. 访问书籍列表页 `/books`
2. 点击"下一页"到第 2 页
3. 点击"上一页"回到第 1 页
4. 再次点击"下一页"
5. **验证**: 应该显示第 2 页的数据（与步骤 2 相同），而不是第 3 页

---

## 代码质量改进

所有修复都遵循了项目的代码规范：
- ✅ 添加了清晰的注释说明修复原因
- ✅ 保持了代码风格的一致性
- ✅ 没有引入新的 linter 错误
- ✅ 符合 Vue 3 和 TypeScript 最佳实践

## 相关文件
- `app/calibre-pages/src/views/Detail.vue` (修复 Bug 1)
- `app/calibre-pages/src/views/Chat.vue` (修复 Bug 2)
- `app/calibre-pages/src/views/Books.vue` (修复 Bug 3)

