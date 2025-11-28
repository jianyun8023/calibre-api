# Qdrant Collection 创建指南

## 创建 Collection 并预设索引

在使用 calibre-api 之前，需要先在 Qdrant 中创建 collection 并配置必要的索引。

### 1. 创建 Collection（推荐配置）

使用以下命令创建优化的 collection，**包含预设的 Payload 索引**：

```bash
curl -X PUT http://localhost:6333/collections/books \
  -H 'Content-Type: application/json' \
  -d '{
    "vectors": {
      "size": 4096,
      "distance": "Cosine",
      "hnsw_config": {
        "m": 32,
        "ef_construct": 200,
        "full_scan_threshold": 10000,
        "max_indexing_threads": 4,
        "on_disk": false
      }
    },
    "optimizers_config": {
      "indexing_threshold": 50000,
      "max_segment_size": 200000,
      "memmap_threshold": 50000,
      "flush_interval_sec": 10
    },
    "on_disk_payload": true,
    "replication_factor": 1,
    "write_consistency_factor": 1
  }'
```

### 2. 创建必需的 Payload 索引

创建完 collection 后，立即创建必需的 payload 索引：

#### 2.1 为 book_id 创建索引（integer 类型）

```bash
curl -X PUT http://localhost:6333/collections/books/index \
  -H 'Content-Type: application/json' \
  -d '{
    "field_name": "book_id",
    "field_schema": "integer"
  }'
```

**用途**：支持按 book_id 排序和过滤，实现书籍列表的倒序显示。

#### 2.2 为 last_modified 创建索引（datetime 类型）

```bash
curl -X PUT http://localhost:6333/collections/books/index \
  -H 'Content-Type: application/json' \
  -d '{
    "field_name": "last_modified",
    "field_schema": "datetime"
  }'
```

**用途**：支持按最后修改时间排序，实现"最近更新"功能。

#### 2.3 为 publisher 创建索引（keyword 类型）

```bash
curl -X PUT http://localhost:6333/collections/books/index \
  -H 'Content-Type: application/json' \
  -d '{
    "field_name": "publisher",
    "field_schema": "keyword"
  }'
```

**用途**：支持按出版社精确过滤。

#### 2.4 为 authors 创建索引（keyword 类型，数组）

```bash
curl -X PUT http://localhost:6333/collections/books/index \
  -H 'Content-Type: application/json' \
  -d '{
    "field_name": "authors",
    "field_schema": "keyword"
  }'
```

**用途**：支持按作者精确过滤。

#### 2.5 为 tags 创建索引（keyword 类型，数组）

```bash
curl -X PUT http://localhost:6333/collections/books/index \
  -H 'Content-Type: application/json' \
  -d '{
    "field_name": "tags",
    "field_schema": "keyword"
  }'
```

**用途**：支持按标签精确过滤。

#### 2.6 为 isbn 创建索引（keyword 类型）

```bash
curl -X PUT http://localhost:6333/collections/books/index \
  -H 'Content-Type: application/json' \
  -d '{
    "field_name": "isbn",
    "field_schema": "keyword"
  }'
```

**用途**：支持按 ISBN 精确匹配查询。

### 3. 验证索引创建

查看 collection 信息，确认索引已创建：

```bash
curl http://localhost:6333/collections/books
```

响应中应该包含索引信息：

```json
{
  "result": {
    "status": "green",
    "indexes": {
      "book_id": {
        "data_type": "integer",
        "points_count": 0
      },
      "last_modified": {
        "data_type": "datetime",
        "points_count": 0
      },
      "publisher": {
        "data_type": "keyword",
        "points_count": 0
      },
      "authors": {
        "data_type": "keyword",
        "points_count": 0
      },
      "tags": {
        "data_type": "keyword",
        "points_count": 0
      },
      "isbn": {
        "data_type": "keyword",
        "points_count": 0
      }
    }
  }
}
```

### 4. 一键创建脚本

将以上命令整合为一个脚本（`setup_qdrant.sh`）：

```bash
#!/bin/bash

QDRANT_URL="${QDRANT_URL:-http://localhost:6333}"
COLLECTION_NAME="${COLLECTION_NAME:-books}"

echo "Creating collection: $COLLECTION_NAME"

# 创建 collection
curl -X PUT "$QDRANT_URL/collections/$COLLECTION_NAME" \
  -H 'Content-Type: application/json' \
  -d '{
    "vectors": {
      "size": 4096,
      "distance": "Cosine",
      "hnsw_config": {
        "m": 32,
        "ef_construct": 200,
        "on_disk": false
      }
    },
    "on_disk_payload": true
  }'

echo -e "\n\nCreating indexes..."

# 创建索引
declare -a indexes=(
  "book_id:integer"
  "last_modified:datetime"
  "publisher:keyword"
  "authors:keyword"
  "tags:keyword"
  "isbn:keyword"
)

for index in "${indexes[@]}"; do
  IFS=':' read -r field_name field_type <<< "$index"
  echo "Creating index for $field_name ($field_type)"
  
  curl -X PUT "$QDRANT_URL/collections/$COLLECTION_NAME/index" \
    -H 'Content-Type: application/json' \
    -d "{
      \"field_name\": \"$field_name\",
      \"field_schema\": \"$field_type\"
    }"
  echo ""
done

echo -e "\n\nVerifying collection..."
curl "$QDRANT_URL/collections/$COLLECTION_NAME"
```

使用方法：

```bash
chmod +x setup_qdrant.sh
./setup_qdrant.sh
```

或指定自定义 Qdrant 地址：

```bash
QDRANT_URL=http://192.168.2.236:6333 ./setup_qdrant.sh
```

## 索引说明

### 为什么需要这些索引？

根据 [Qdrant 文档](https://qdrant.tech/documentation/concepts/indexing/#payload-index)：

1. **Range 索引**（integer, datetime）：
   - 支持 `order_by` 排序
   - 支持 Range 过滤条件（`lt`, `gt`, `gte`, `lte`）
   - 必须预先创建才能使用排序功能

2. **Keyword 索引**：
   - 支持 Match 精确匹配过滤
   - 加速按字段值查询的性能

### 索引类型对应表

| 字段 | 类型 | 用途 | 是否必需 |
|------|------|------|----------|
| book_id | integer | 按ID倒序排列书籍 | ✅ 必需 |
| last_modified | datetime | 按时间排序 | ✅ 推荐 |
| publisher | keyword | 按出版社过滤 | ✅ 推荐 |
| authors | keyword | 按作者过滤 | ✅ 推荐 |
| tags | keyword | 按标签过滤 | ✅ 推荐 |
| isbn | keyword | ISBN精确查询 | ⭕ 可选 |

## 索引维护

### 查看现有索引

查看 collection 的所有索引：

```bash
curl http://localhost:6333/collections/books
```

响应中查看 `payload_schema` 部分：

```json
{
  "result": {
    "payload_schema": {
      "book_id": {
        "data_type": "integer"
      },
      "last_modified": {
        "data_type": "datetime"
      },
      "publisher": {
        "data_type": "keyword"
      }
    }
  }
}
```

### 删除索引

如果需要删除某个索引：

```bash
curl -X DELETE http://localhost:6333/collections/books/index/book_id
```

**注意**：删除索引会影响查询性能，但不会删除数据本身。

### 重建索引

Qdrant 的索引是自动维护的，通常不需要手动重建。但如果遇到以下情况：

1. **添加新数据后索引未更新**：
   - Qdrant 会自动为新数据建立索引
   - 索引构建是异步的，可能需要等待几秒

2. **索引损坏或不一致**：
   - 删除索引后重新创建即可

```bash
# 删除旧索引
curl -X DELETE http://localhost:6333/collections/books/index/book_id

# 重新创建索引
curl -X PUT http://localhost:6333/collections/books/index \
  -H 'Content-Type: application/json' \
  -d '{
    "field_name": "book_id",
    "field_schema": "integer"
  }'
```

### 批量更新索引脚本

如果需要批量重建所有索引：

```bash
#!/bin/bash

QDRANT_URL="${QDRANT_URL:-http://localhost:6333}"
COLLECTION_NAME="${COLLECTION_NAME:-books}"

echo "Rebuilding indexes for collection: $COLLECTION_NAME"

# 定义所有索引
declare -A indexes=(
  ["book_id"]="integer"
  ["last_modified"]="datetime"
  ["publisher"]="keyword"
  ["authors"]="keyword"
  ["tags"]="keyword"
  ["isbn"]="keyword"
)

for field_name in "${!indexes[@]}"; do
  field_type="${indexes[$field_name]}"
  
  echo "Deleting index: $field_name"
  curl -X DELETE "$QDRANT_URL/collections/$COLLECTION_NAME/index/$field_name"
  
  echo ""
  echo "Recreating index: $field_name ($field_type)"
  curl -X PUT "$QDRANT_URL/collections/$COLLECTION_NAME/index" \
    -H 'Content-Type: application/json' \
    -d "{
      \"field_name\": \"$field_name\",
      \"field_schema\": \"$field_type\"
    }"
  
  echo -e "\n---"
done

echo "Index rebuild complete"
```

### 添加新索引

如果将来需要为其他字段添加索引（例如 `rating`）：

```bash
curl -X PUT http://localhost:6333/collections/books/index \
  -H 'Content-Type: application/json' \
  -d '{
    "field_name": "rating",
    "field_schema": "float"
  }'
```

**支持的索引类型**：
- `keyword` - 用于精确匹配（Match）
- `integer` - 用于整数范围和排序（Range, order_by）
- `float` - 用于浮点数范围和排序
- `bool` - 用于布尔值匹配
- `geo` - 用于地理位置查询
- `datetime` - 用于日期时间范围和排序（Range, order_by）
- `text` - 用于全文搜索
- `uuid` - 用于 UUID 类型的精确匹配

### 索引更新策略

Qdrant 的索引更新策略：

1. **自动更新**：
   - 插入新数据时，索引自动更新
   - 更新现有数据时，索引自动重建
   - 删除数据时，索引自动清理

2. **异步构建**：
   - 大批量数据导入时，索引构建是异步的
   - 可以通过 collection 状态查看构建进度

3. **查看索引构建状态**：
```bash
curl http://localhost:6333/collections/books
```

查看响应中的 `optimizer_status` 字段：
- `ok` - 索引已完成
- `optimizing` - 正在构建索引

### 性能优化建议

1. **大批量导入数据时**：
   - 先导入数据，后创建索引（索引构建更快）
   - 或者在导入前就创建好索引（推荐）

2. **选择性创建索引**：
   - 只为经常查询的字段创建索引
   - 过多索引会增加写入延迟和存储开销

3. **监控索引性能**：
   ```bash
   # 查看 collection 统计信息
   curl http://localhost:6333/collections/books
   ```
   
   关注：
   - `points_count` - 总点数
   - `indexed_vectors_count` - 已索引向量数
   - `segments_count` - 段数量

## 注意事项

1. **索引创建是幂等的**：多次调用创建索引API不会报错
2. **已有数据会自动索引**：为已有 collection 添加索引时，Qdrant 会自动为现有数据建立索引
3. **索引占用空间**：每个索引会占用额外的存储空间，建议只为常用查询字段创建索引
4. **性能影响**：索引会轻微增加写入延迟，但大幅提升查询性能
5. **索引是持久化的**：索引数据存储在磁盘上，重启不会丢失
6. **索引自动维护**：通常不需要手动重建或更新索引

## 自动索引初始化

calibre-api 会在启动时自动检查并创建必需的索引。如果索引已存在，会跳过创建。

相关日志：
```
Qdrant payload indexes ensured successfully
```

如果看到此日志，说明索引已就绪。

## 故障排查

### 问题：order_by 报错 "No range index"

**错误信息**：
```
Wrong input: No range index for `order_by` key: `book_id`
```

**原因**：缺少 Range 类型的索引（integer 或 datetime）

**解决方法**：
```bash
curl -X PUT http://localhost:6333/collections/books/index \
  -H 'Content-Type: application/json' \
  -d '{
    "field_name": "book_id",
    "field_schema": "integer"
  }'
```

### 问题：过滤查询很慢

**原因**：过滤字段没有索引

**解决方法**：为常用过滤字段创建 keyword 索引

```bash
curl -X PUT http://localhost:6333/collections/books/index \
  -H 'Content-Type: application/json' \
  -d '{
    "field_name": "publisher",
    "field_schema": "keyword"
  }'
```

### 问题：索引构建卡住

**检查状态**：
```bash
curl http://localhost:6333/collections/books
```

**如果 `optimizer_status` 长时间显示 `optimizing`**：
1. 检查磁盘空间是否充足
2. 检查 Qdrant 日志
3. 如果必要，可以删除索引重建

## 参考文档

- [Qdrant Payload Index 官方文档](https://qdrant.tech/documentation/concepts/indexing/#payload-index)
- [Qdrant Filtering 文档](https://qdrant.tech/documentation/concepts/filtering/)
- [Qdrant Search 文档](https://qdrant.tech/documentation/concepts/search/)

