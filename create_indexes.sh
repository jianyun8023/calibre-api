#!/bin/bash

QDRANT_URL="http://192.168.2.236:6333"
COLLECTION_NAME="books"

echo "开始为 collection '$COLLECTION_NAME' 创建索引..."
echo "Qdrant 地址: $QDRANT_URL"
echo ""

# 定义索引配置
declare -a indexes=(
  "book_id:integer:支持按ID排序"
  "last_modified:datetime:支持按时间排序"
  "publisher:keyword:按出版社过滤"
  "authors:keyword:按作者过滤"
  "tags:keyword:按标签过滤"
  "isbn:keyword:ISBN精确查询"
)

# 创建索引
for index in "${indexes[@]}"; do
  IFS=':' read -r field_name field_type description <<< "$index"
  echo "[$field_name] 创建 $field_type 索引 - $description"
  
  response=$(curl -s -X PUT "$QDRANT_URL/collections/$COLLECTION_NAME/index" \
    -H 'Content-Type: application/json' \
    -d "{
      \"field_name\": \"$field_name\",
      \"field_schema\": \"$field_type\"
    }")
  
  # 检查响应
  if echo "$response" | grep -q '"status":"ok"'; then
    echo "  ✅ 成功"
  else
    echo "  ⚠️  响应: $response"
  fi
  echo ""
done

echo "验证 collection 状态..."
curl -s "$QDRANT_URL/collections/$COLLECTION_NAME" | python3 -m json.tool | grep -A 20 '"payload_schema"'

echo ""
echo "索引创建完成！"

