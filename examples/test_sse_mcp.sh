#!/bin/bash

# Calibre MCP SSE 测试脚本
# 用于测试 SSE MCP 服务的各项功能

echo "🚀 开始测试 Calibre MCP SSE 服务"

BASE_URL="http://localhost:8080"
API_BASE="$BASE_URL/api/mcp"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试函数
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "\n${BLUE}📡 测试: $description${NC}"
    echo -e "${YELLOW}请求: $method $endpoint${NC}"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$endpoint")
    fi
    
    # 分离响应体和状态码
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✅ 成功 (HTTP $http_code)${NC}"
        echo -e "${YELLOW}响应:${NC}"
        echo "$response_body" | jq . 2>/dev/null || echo "$response_body"
    else
        echo -e "${RED}❌ 失败 (HTTP $http_code)${NC}"
        echo -e "${RED}响应:${NC}"
        echo "$response_body"
    fi
}

# 检查服务器是否运行
echo -e "${BLUE}🔍 检查服务器状态${NC}"
if curl -s "$BASE_URL/api/search?q=test&limit=1" > /dev/null; then
    echo -e "${GREEN}✅ Calibre API 服务器正在运行${NC}"
else
    echo -e "${RED}❌ Calibre API 服务器未运行或不可访问${NC}"
    echo -e "${YELLOW}请先启动服务器: ./calibre-api${NC}"
    exit 1
fi

# 1. 测试初始化
test_endpoint "POST" "$API_BASE/initialize" '{
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
    }
}' "MCP 初始化"

# 2. 测试获取工具列表
test_endpoint "GET" "$API_BASE/tools/list" "" "获取工具列表"

# 3. 测试搜索书籍工具
test_endpoint "POST" "$API_BASE/tools/call" '{
    "name": "search_books",
    "arguments": {
        "query": "Python",
        "limit": 3
    }
}' "搜索书籍工具"

# 4. 测试获取最近书籍工具
test_endpoint "POST" "$API_BASE/tools/call" '{
    "name": "get_recent_books",
    "arguments": {
        "limit": 5
    }
}' "获取最近书籍工具"

# 5. 测试获取书籍详情工具（需要有效的书籍ID）
test_endpoint "POST" "$API_BASE/tools/call" '{
    "name": "get_book",
    "arguments": {
        "id": "1"
    }
}' "获取书籍详情工具"

# 6. 测试搜索元数据工具
test_endpoint "POST" "$API_BASE/tools/call" '{
    "name": "search_metadata",
    "arguments": {
        "query": "Python编程"
    }
}' "搜索元数据工具"

# 7. 测试无效工具调用
test_endpoint "POST" "$API_BASE/tools/call" '{
    "name": "invalid_tool",
    "arguments": {}
}' "无效工具调用测试"

# 8. 测试 SSE 连接（后台测试）
echo -e "\n${BLUE}📡 测试 SSE 连接${NC}"
echo -e "${YELLOW}启动 SSE 连接测试（10秒）...${NC}"

# 使用 timeout 命令限制 curl 运行时间
timeout 10s curl -s -N "$API_BASE/connect" | head -20 &
SSE_PID=$!

# 等待几秒钟
sleep 3

# 检查 SSE 进程是否还在运行
if kill -0 $SSE_PID 2>/dev/null; then
    echo -e "${GREEN}✅ SSE 连接正常建立${NC}"
    kill $SSE_PID 2>/dev/null
else
    echo -e "${RED}❌ SSE 连接失败${NC}"
fi

echo -e "\n${GREEN}🎉 测试完成！${NC}"
echo -e "\n${BLUE}📋 使用说明：${NC}"
echo -e "1. 在浏览器中打开 examples/mcp_sse_client.html 进行交互式测试"
echo -e "2. 使用 JavaScript 客户端库进行编程集成"
echo -e "3. 查看 docs/MCP_SSE_README.md 获取详细文档"

echo -e "\n${YELLOW}🔗 相关链接：${NC}"
echo -e "- SSE 连接: $API_BASE/connect"
echo -e "- 初始化端点: $API_BASE/initialize"
echo -e "- 工具列表: $API_BASE/tools/list"
echo -e "- 工具调用: $API_BASE/tools/call"