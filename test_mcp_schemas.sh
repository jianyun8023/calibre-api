#!/bin/bash

# 测试 MCP 参数说明功能的脚本

echo "=== Calibre API MCP 参数说明测试 ==="
echo

# 检查服务器是否正在运行
if ! curl -s http://localhost:8080/ping > /dev/null; then
    echo "❌ 服务器未运行，请先启动服务器："
    echo "   ./calibre-api"
    echo
    exit 1
fi

echo "✅ 服务器正在运行"
echo

# 测试 MCP 端点
echo "🔍 测试 MCP 端点..."
if curl -s http://localhost:8080/mcp > /dev/null; then
    echo "✅ MCP 端点可访问"
else
    echo "❌ MCP 端点不可访问"
fi
echo

# 测试搜索 API（带参数说明）
echo "🔍 测试搜索 API..."
SEARCH_RESPONSE=$(curl -s "http://localhost:8080/api/search?q=test&limit=5")
if echo "$SEARCH_RESPONSE" | grep -q "data"; then
    echo "✅ 搜索 API 正常工作"
else
    echo "❌ 搜索 API 返回错误"
    echo "响应: $SEARCH_RESPONSE"
fi
echo

# 测试元数据搜索 API
echo "🔍 测试元数据搜索 API..."
METADATA_RESPONSE=$(curl -s "http://localhost:8080/api/metadata/search?query=test&limit=3")
if echo "$METADATA_RESPONSE" | grep -q "data\|message"; then
    echo "✅ 元数据搜索 API 正常工作"
else
    echo "❌ 元数据搜索 API 返回错误"
    echo "响应: $METADATA_RESPONSE"
fi
echo

# 测试出版社列表 API
echo "🔍 测试出版社列表 API..."
PUBLISHER_RESPONSE=$(curl -s "http://localhost:8080/api/publisher?limit=10")
if echo "$PUBLISHER_RESPONSE" | grep -q "data\|message"; then
    echo "✅ 出版社列表 API 正常工作"
else
    echo "❌ 出版社列表 API 返回错误"
    echo "响应: $PUBLISHER_RESPONSE"
fi
echo

echo "=== 测试完成 ==="
echo
echo "📝 要验证 MCP 参数说明，请："
echo "1. 在 Cursor 中打开设置"
echo "2. 找到 MCP 配置部分"
echo "3. 添加新的 MCP 服务器："
echo "   - 名称：Calibre API"
echo "   - URL：http://localhost:8080/mcp"
echo "4. 连接后，您应该能看到所有 API 工具都包含详细的参数说明"
echo 