# 草稿合并模式使用指南

## 功能说明

UPDATE 草稿现在支持**字段级合并**：
- ✅ 新提交的字段**覆盖**旧字段
- ✅ 未提交的字段**保留**旧值
- ✅ 支持分多次提交逐步补全元数据

## 合并行为示例

### 示例 1：逐步补全信息

```bash
# 第一次提交：清空垃圾标签
curl -X POST "http://localhost:3000/api/drafts/update" \
  -H "Content-Type: application/json" \
  -d '{
    "updates": [{
      "id": "274576",
      "data": {"tags": []}
    }]
  }'
# 草稿数据：{"tags": []}

# 第二次提交：补充出版社
curl -X POST "http://localhost:3000/api/drafts/update" \
  -H "Content-Type: application/json" \
  -d '{
    "updates": [{
      "id": "274576",
      "data": {"publisher": "清华大学出版社"}
    }]
  }'
# 草稿数据：{"tags": [], "publisher": "清华大学出版社"}  ← tags 被保留

# 第三次提交：补充 ISBN 和作者
curl -X POST "http://localhost:3000/api/drafts/update" \
  -H "Content-Type: application/json" \
  -d '{
    "updates": [{
      "id": "274576",
      "data": {
        "isbn": "9787302395423",
        "authors": ["李永华"]
      }
    }]
  }'
# 最终草稿：{
#   "tags": [],
#   "publisher": "清华大学出版社",
#   "isbn": "9787302395423",
#   "authors": ["李永华"]
# }
```

### 示例 2：覆盖已有字段

```bash
# 第一次提交
curl -X POST "http://localhost:3000/api/drafts/update" \
  -H "Content-Type: application/json" \
  -d '{
    "updates": [{
      "id": "274579",
      "data": {"publisher": "旧出版社"}
    }]
  }'
# 草稿数据：{"publisher": "旧出版社"}

# 第二次提交：纠正出版社
curl -X POST "http://localhost:3000/api/drafts/update" \
  -H "Content-Type: application/json" \
  -d '{
    "updates": [{
      "id": "274579",
      "data": {"publisher": "机械工业出版社"}
    }]
  }'
# 草稿数据：{"publisher": "机械工业出版社"}  ← 新值覆盖旧值
```

## 使用场景

### ✅ 推荐场景

1. **分步补全元数据**
   - 第一步：清理垃圾标签
   - 第二步：补充出版社
   - 第三步：添加 ISBN

2. **纠正错误提交**
   - 发现之前提交的出版社有误
   - 重新提交正确的出版社
   - 其他字段（tags、authors）保持不变

3. **AI 辅助元数据补全**
   - AI 第一次提交部分字段
   - AI 获取更多信息后补充其他字段
   - 无需重复提交已有字段

### ❌ 不适用场景

如果您想**完全替换**草稿内容（而非合并），请：
1. 先取消旧草稿：`POST /api/drafts/cancel {"ids": ["274576"]}`
2. 再提交新草稿

## 重启服务

合并模式需要重启后端才能生效。

```bash
# 在终端 1 中按 Ctrl+C 停止服务
# 然后重新运行
go run ./main.go
```

## 验证合并行为

重启后，您可以测试合并：

```bash
# 1. 提交第一个草稿（只有 tags）
curl -X POST "http://localhost:3000/api/drafts/update" \
  -H "Content-Type: application/json" \
  -d '{
    "updates": [{
      "id": "274576",
      "data": {"tags": []}
    }]
  }'

# 2. 查看草稿
curl "http://localhost:3000/api/drafts?limit=10" | jq '.data[] | select(.book_id == "274576")'
# 应该看到：{"tags": []}

# 3. 提交第二个草稿（添加 publisher 和 isbn）
curl -X POST "http://localhost:3000/api/drafts/update" \
  -H "Content-Type: application/json" \
  -d '{
    "updates": [{
      "id": "274576",
      "data": {
        "publisher": "清华大学出版社",
        "isbn": "9787302395423",
        "authors": ["李永华"]
      }
    }]
  }'

# 4. 再次查看草稿
curl "http://localhost:3000/api/drafts?limit=10" | jq '.data[] | select(.book_id == "274576")'
# 应该看到所有字段都在：{"tags": [], "publisher": "清华大学出版社", ...}
```

## 技术细节

### 合并规则

对于每个字段：
- `nil` (未提交) → 保留旧值
- 非 `nil` (已提交) → 使用新值覆盖

### 特殊字段处理

| 字段 | 合并逻辑 |
|------|----------|
| **Tags** | `nil`=保留旧值，`[]`=清空，`[...]`=设置新标签 |
| **Authors** | `nil`=保留旧值，`[]`=被拒绝（不允许清空），`[...]`=更新作者 |
| **String 字段** | `nil`=保留旧值，`""`=被拒绝（不允许清空），非空=更新 |
| **Rating** | `0`=保留旧值，非0=更新 |
| **PubDate** | `nil`/零值=保留旧值，有效日期=更新 |

### 数据库存储

草稿数据以 JSON 格式存储在 `book_drafts.data` 字段：
```sql
SELECT id, book_id, data FROM book_drafts WHERE book_id = '274576';
-- data: {"tags":[],"publisher":"清华大学出版社","isbn":"9787302395423","authors":["李永华"]}
```

## 与旧行为的对比

| 场景 | 旧行为（完全覆盖） | 新行为（字段合并） |
|------|-------------------|-------------------|
| 第一次提交 `{"tags": []}` | 草稿：`{"tags": []}` | 草稿：`{"tags": []}` |
| 第二次提交 `{"publisher": "ABC"}` | 草稿：`{"publisher": "ABC"}` ❌ tags 丢失 | 草稿：`{"tags": [], "publisher": "ABC"}` ✅ tags 保留 |
| 第三次提交 `{"tags": ["新"]}` | 草稿：`{"tags": ["新"]}` ❌ publisher 丢失 | 草稿：`{"tags": ["新"], "publisher": "ABC"}` ✅ publisher 保留，tags 被覆盖 |
