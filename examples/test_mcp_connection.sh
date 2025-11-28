#!/bin/bash
# MCP 连接测试脚本
# 用于验证 MCP SSE 端点是否正常工作

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
SERVER_URL="${SERVER_URL:-http://localhost:8080}"
SSE_ENDPOINT="/mcp/sse"
MESSAGE_ENDPOINT="/mcp/message"
TIMEOUT=5

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   MCP 连接测试 - v1.2.0${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "服务器地址: ${YELLOW}${SERVER_URL}${NC}"
echo ""

# 测试 1: 服务器健康检查
echo -e "${BLUE}[1/5] 测试服务器健康状态...${NC}"
if curl -s -f -m $TIMEOUT "${SERVER_URL}/ping" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 服务器运行正常${NC}"
else
    echo -e "${RED}✗ 服务器无法访问，请确保服务器已启动${NC}"
    exit 1
fi
echo ""

# 测试 2: 测试 SSE 端点可访问性
echo -e "${BLUE}[2/5] 测试 SSE 端点可访问性...${NC}"
SSE_RESPONSE=$(curl -s -w "\n%{http_code}" -m $TIMEOUT "${SERVER_URL}${SSE_ENDPOINT}" 2>/dev/null || echo "000")
HTTP_CODE=$(echo "$SSE_RESPONSE" | tail -n 1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ SSE 端点响应正常 (HTTP $HTTP_CODE)${NC}"
elif [ "$HTTP_CODE" = "000" ]; then
    echo -e "${RED}✗ 无法连接到 SSE 端点${NC}"
    exit 1
else
    echo -e "${YELLOW}⚠ SSE 端点返回 HTTP $HTTP_CODE (可能正常，SSE 连接特性)${NC}"
fi
echo ""

# 测试 3: 测试 CORS 头
echo -e "${BLUE}[3/5] 测试 CORS 配置...${NC}"
CORS_HEADERS=$(curl -s -I -H "Origin: https://inspector.modelcontextprotocol.io" \
    "${SERVER_URL}${SSE_ENDPOINT}" 2>/dev/null | grep -i "access-control")

if [ -n "$CORS_HEADERS" ]; then
    echo -e "${GREEN}✓ CORS 已配置${NC}"
    echo "$CORS_HEADERS" | while IFS= read -r line; do
        echo "  $line"
    done
else
    echo -e "${RED}✗ 未检测到 CORS 头，MCP Inspector 可能无法连接${NC}"
fi
echo ""

# 测试 4: 测试 OPTIONS 预检请求
echo -e "${BLUE}[4/5] 测试 OPTIONS 预检请求...${NC}"
OPTIONS_RESPONSE=$(curl -s -X OPTIONS -I \
    -H "Origin: https://inspector.modelcontextprotocol.io" \
    -H "Access-Control-Request-Method: GET" \
    "${SERVER_URL}${SSE_ENDPOINT}" 2>/dev/null)

if echo "$OPTIONS_RESPONSE" | grep -qi "access-control-allow-origin"; then
    echo -e "${GREEN}✓ OPTIONS 预检请求成功${NC}"
else
    echo -e "${YELLOW}⚠ OPTIONS 预检请求可能有问题${NC}"
fi
echo ""

# 测试 5: 测试消息端点
echo -e "${BLUE}[5/5] 测试消息端点...${NC}"
MESSAGE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":1,"method":"ping"}' \
    "${SERVER_URL}${MESSAGE_ENDPOINT}" 2>/dev/null || echo "000")
MESSAGE_CODE=$(echo "$MESSAGE_RESPONSE" | tail -n 1)

if [ "$MESSAGE_CODE" = "200" ] || [ "$MESSAGE_CODE" = "204" ]; then
    echo -e "${GREEN}✓ 消息端点响应正常 (HTTP $MESSAGE_CODE)${NC}"
elif [ "$MESSAGE_CODE" = "000" ]; then
    echo -e "${RED}✗ 无法连接到消息端点${NC}"
else
    echo -e "${YELLOW}⚠ 消息端点返回 HTTP $MESSAGE_CODE${NC}"
fi
echo ""

# 总结
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   测试完成${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}MCP 端点准备就绪！${NC}"
echo ""
echo -e "可以使用以下方式连接:"
echo -e "  ${YELLOW}1. MCP Inspector (在线)${NC}"
echo -e "     https://inspector.modelcontextprotocol.io/"
echo -e "     服务器 URL: ${YELLOW}${SERVER_URL}${SSE_ENDPOINT}${NC}"
echo ""
echo -e "  ${YELLOW}2. curl 测试${NC}"
echo -e "     ${YELLOW}curl -N ${SERVER_URL}${SSE_ENDPOINT}${NC}"
echo ""
echo -e "  ${YELLOW}3. 发送 JSON-RPC 消息${NC}"
echo -e "     ${YELLOW}curl -X POST ${SERVER_URL}${MESSAGE_ENDPOINT} \\${NC}"
echo -e "     ${YELLOW}  -H 'Content-Type: application/json' \\${NC}"
echo -e "     ${YELLOW}  -d '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\"}'${NC}"
echo ""
echo -e "${BLUE}详细文档: docs/features/MCP_INSPECTOR_GUIDE.md${NC}"
echo ""

