#!/bin/bash

# Claude Proxy 测试脚本

echo "=== Claude Proxy 测试脚本 ==="
echo ""

# 设置颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# 代理地址
PROXY_URL="http://localhost:27015"

# 测试健康检查
echo "1. 测试健康检查端点..."
HEALTH_RESPONSE=$(curl -s "${PROXY_URL}/health")
if [[ $? -eq 0 ]]; then
    echo -e "${GREEN}✓ 健康检查成功${NC}: $HEALTH_RESPONSE"
else
    echo -e "${RED}✗ 健康检查失败${NC}"
    exit 1
fi
echo ""

# 测试状态端点
echo "2. 测试状态端点..."
STATUS_RESPONSE=$(curl -s "${PROXY_URL}/status" | jq .)
if [[ $? -eq 0 ]]; then
    echo -e "${GREEN}✓ 状态检查成功${NC}"
    echo "$STATUS_RESPONSE"
else
    echo -e "${RED}✗ 状态检查失败${NC}"
fi
echo ""

# 测试普通请求
echo "3. 测试普通 Claude API 请求..."
NORMAL_REQUEST='{
  "model": "claude-haiku-4-5-20251001",
  "messages": [
    {
      "role": "user",
      "content": [
        {
          "type": "text",
          "text": "Hello, this is a test"
        }
      ]
    }
  ],
  "metadata": {
    "user_id": "user_test123_account__session_test456"
  },
  "max_tokens": 100,
  "stream": false
}'

echo "发送请求到代理..."
RESPONSE=$(curl -s -X POST "${PROXY_URL}/api/v1/messages?beta=true" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test_key" \
  -H "anthropic-version: 2023-06-01" \
  -d "$NORMAL_REQUEST")

if [[ $? -eq 0 ]]; then
    echo -e "${GREEN}✓ 普通请求成功${NC}"
    echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
else
    echo -e "${RED}✗ 普通请求失败${NC}"
fi
echo ""

# 测试流式请求
echo "4. 测试流式 Claude API 请求..."
STREAM_REQUEST='{
  "model": "claude-sonnet-4-5-20250929",
  "messages": [
    {
      "role": "user",
      "content": [
        {
          "type": "text",
          "text": "Tell me a short story"
        }
      ]
    }
  ],
  "metadata": {
    "user_id": "user_test123_account__session_test456"
  },
  "max_tokens": 200,
  "stream": true
}'

echo "发送流式请求到代理..."
echo -e "${YELLOW}流式响应：${NC}"
curl -s -X POST "${PROXY_URL}/api/v1/messages?beta=true" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test_key" \
  -H "anthropic-version: 2023-06-01" \
  -H "Accept: text/event-stream" \
  -d "$STREAM_REQUEST" \
  --no-buffer

echo ""
echo ""
echo "=== 测试完成 ==="
