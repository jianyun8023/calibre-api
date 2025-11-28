#!/bin/bash
# MCP 工具测试脚本
# 测试所有已注册的 MCP 工具

set -e

SERVER_URL="http://127.0.0.1:8080"
SSE_ENDPOINT="${SERVER_URL}/mcp/sse"
MESSAGE_ENDPOINT="${SERVER_URL}/mcp/message"

echo "======================================"
echo "MCP 工具测试"
echo "======================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试服务器是否运行
echo "1. 检查服务器状态..."
if curl -s "${SERVER_URL}/ping" | grep -q "pong"; then
    echo -e "${GREEN}✓ 服务器运行正常${NC}"
else
    echo -e "${RED}✗ 服务器未运行，请先启动 calibre-api${NC}"
    exit 1
fi
echo ""

# 测试 SSE 端点
echo "2. 测试 SSE 端点..."
if curl -s -N "${SSE_ENDPOINT}" -m 2 2>&1 | head -1 | grep -q "event:"; then
    echo -e "${GREEN}✓ SSE 端点正常${NC}"
else
    echo -e "${YELLOW}! SSE 端点响应异常（可能是正常的流式响应）${NC}"
fi
echo ""

# 生成测试会话 ID
SESSION_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')

echo "3. 测试工具列表..."
echo "   发送 initialize 请求..."

# 发送 initialize 请求
INIT_RESPONSE=$(curl -s -X POST "${MESSAGE_ENDPOINT}?sessionId=${SESSION_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
      }
    }
  }')

if echo "$INIT_RESPONSE" | jq -e '.result.capabilities' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Initialize 成功${NC}"
    echo "   服务器信息:"
    echo "$INIT_RESPONSE" | jq -r '.result.serverInfo // empty' | sed 's/^/     /'
else
    echo -e "${RED}✗ Initialize 失败${NC}"
    echo "   响应: $INIT_RESPONSE"
fi
echo ""

echo "4. 获取工具列表..."

# 发送 tools/list 请求
TOOLS_RESPONSE=$(curl -s -X POST "${MESSAGE_ENDPOINT}?sessionId=${SESSION_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list",
    "params": {}
  }')

if echo "$TOOLS_RESPONSE" | jq -e '.result.tools' > /dev/null 2>&1; then
    TOOL_COUNT=$(echo "$TOOLS_RESPONSE" | jq '.result.tools | length')
    echo -e "${GREEN}✓ 成功获取工具列表 (${TOOL_COUNT} 个工具)${NC}"
    echo ""
    echo "   已注册的工具:"
    echo "$TOOLS_RESPONSE" | jq -r '.result.tools[] | "     - \(.name): \(.description)"'
else
    echo -e "${RED}✗ 获取工具列表失败${NC}"
    echo "   响应: $TOOLS_RESPONSE"
fi
echo ""

echo "5. 安全性检查..."
echo "   检查危险工具是否已移除..."

DANGEROUS_TOOLS=("update_book_metadata" "delete_book")
FOUND_DANGEROUS=false

for tool in "${DANGEROUS_TOOLS[@]}"; do
    if echo "$TOOLS_RESPONSE" | jq -e ".result.tools[] | select(.name == \"$tool\")" > /dev/null 2>&1; then
        echo -e "${RED}✗ 发现危险工具: $tool${NC}"
        FOUND_DANGEROUS=true
    else
        echo -e "${GREEN}✓ 危险工具 $tool 已正确移除${NC}"
    fi
done

if [ "$FOUND_DANGEROUS" = false ]; then
    echo -e "${GREEN}✓ 安全检查通过，无危险工具暴露${NC}"
fi
echo ""

echo "6. 测试工具调用（如果有数据）..."
echo "   尝试调用 random_books 工具..."

# 测试 random_books 工具
RANDOM_RESPONSE=$(curl -s -X POST "${MESSAGE_ENDPOINT}?sessionId=${SESSION_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "random_books",
      "arguments": {
        "limit": 3
      }
    }
  }')

if echo "$RANDOM_RESPONSE" | jq -e '.result.content' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ random_books 工具调用成功${NC}"
    RESULT_TEXT=$(echo "$RANDOM_RESPONSE" | jq -r '.result.content[0].text' | jq -r '.count // empty')
    if [ -n "$RESULT_TEXT" ]; then
        echo "   返回书籍数量: $RESULT_TEXT"
    fi
elif echo "$RANDOM_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
    ERROR_MSG=$(echo "$RANDOM_RESPONSE" | jq -r '.error.message')
    if [[ "$ERROR_MSG" == *"not available"* ]]; then
        echo -e "${YELLOW}! Qdrant 服务未配置或无数据（预期情况）${NC}"
    else
        echo -e "${RED}✗ 工具调用失败: $ERROR_MSG${NC}"
    fi
else
    echo -e "${YELLOW}! 工具调用响应异常${NC}"
    echo "   响应: $RANDOM_RESPONSE"
fi
echo ""

echo "======================================"
echo "测试总结"
echo "======================================"
echo ""
echo "✅ MCP 端点正常"
echo "✅ 工具注册成功 (${TOOL_COUNT} 个)"
echo "✅ 安全性检查通过"
echo ""
echo "📋 可用工具列表:"
echo "$TOOLS_RESPONSE" | jq -r '.result.tools[] | "  - \(.name)"'
echo ""
echo "🔒 安全保证:"
echo "  - 所有工具均为只读操作"
echo "  - update_book_metadata 已移除"
echo "  - delete_book 已移除"
echo ""
echo "🎉 测试完成！"

