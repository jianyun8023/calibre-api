#!/bin/bash

# 启动服务
echo "Starting calibre-api..."
./calibre-api > server.log 2>&1 &
PID=$!
echo "Server PID: $PID"

# 等待服务启动
sleep 5

# 基础 URL
BASE_URL="http://localhost:8080/api"

echo "----------------------------------------"
echo "Testing Chat API"
echo "----------------------------------------"

# 1. 创建对话
echo "1. Creating conversation..."
CREATE_RES=$(curl -s -X POST "$BASE_URL/chat/conversations" \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Chat"}')
echo "Response: $CREATE_RES"

CONV_ID=$(echo $CREATE_RES | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Conversation ID: $CONV_ID"

if [ -z "$CONV_ID" ]; then
    echo "Failed to create conversation"
    kill $PID
    exit 1
fi

# 2. 列出对话
echo -e "\n2. Listing conversations..."
curl -s "$BASE_URL/chat/conversations" | head -c 200
echo "..."

# 3. 发送消息 (模拟)
# 注意：由于 LLM 调用可能较慢且消耗 Token，这里我们只测试接口连通性
# 如果配置正确，应该能收到流式响应
echo -e "\n3. Sending message (Hello)..."
curl -N -X POST "$BASE_URL/chat/conversations/$CONV_ID/messages" \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello, this is a test message."}' > response.txt 2>&1 &
CURL_PID=$!

# 等待几秒钟让它生成一些输出
sleep 5
kill $CURL_PID

echo "Response preview:"
head -n 10 response.txt

# 4. 获取消息历史
echo -e "\n4. Getting message history..."
curl -s "$BASE_URL/chat/conversations/$CONV_ID/messages" | head -c 200
echo "..."

# 5. 删除对话
echo -e "\n5. Deleting conversation..."
curl -X DELETE "$BASE_URL/chat/conversations/$CONV_ID"

# 停止服务
echo -e "\nStopping server..."
kill $PID
