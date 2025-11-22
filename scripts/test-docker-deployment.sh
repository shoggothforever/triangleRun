#!/bin/bash

# 三角机构TRPG单人引擎 - Docker部署测试脚本

set -e

echo "🐳 测试Docker Compose部署..."
echo ""

# 检查服务状态
echo "📊 检查服务状态..."
docker compose ps

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 测试API端点"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 1. 健康检查
echo "1️⃣  测试健康检查..."
HEALTH=$(curl -s http://localhost:8080/health)
echo "$HEALTH" | jq .
if echo "$HEALTH" | jq -e '.status == "ok"' > /dev/null; then
    echo "✅ 健康检查通过"
else
    echo "❌ 健康检查失败"
    exit 1
fi

echo ""

# 2. 版本信息
echo "2️⃣  测试版本信息..."
VERSION=$(curl -s http://localhost:8080/api/version)
echo "$VERSION" | jq .
if echo "$VERSION" | jq -e '.version' > /dev/null; then
    echo "✅ 版本信息正常"
else
    echo "❌ 版本信息失败"
    exit 1
fi

echo ""

# 3. 骰子掷骰
echo "3️⃣  测试骰子系统..."
DICE=$(curl -s -X POST http://localhost:8080/api/dice/roll \
  -H "Content-Type: application/json" \
  -d '{"count": 6}')
echo "$DICE" | jq .
if echo "$DICE" | jq -e '.success == true' > /dev/null; then
    echo "✅ 骰子系统正常"
else
    echo "❌ 骰子系统失败"
    exit 1
fi

echo ""

# 4. 创建角色
echo "4️⃣  测试角色创建..."
AGENT=$(curl -s -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Docker测试特工",
    "pronouns": "他/him",
    "anomaly_type": "低语",
    "reality_type": "看护者",
    "career_type": "公关"
  }')
echo "$AGENT" | jq '.data | {id, name, anomaly: .anomaly.type, reality: .reality.type, career: .career.type}'
if echo "$AGENT" | jq -e '.success == true' > /dev/null; then
    AGENT_ID=$(echo "$AGENT" | jq -r '.data.id')
    echo "✅ 角色创建成功 (ID: $AGENT_ID)"
else
    echo "❌ 角色创建失败"
    exit 1
fi

echo ""

# 5. 查询角色
echo "5️⃣  测试角色查询..."
AGENT_GET=$(curl -s http://localhost:8080/api/agents/$AGENT_ID)
echo "$AGENT_GET" | jq '.data | {id, name, qa}'
if echo "$AGENT_GET" | jq -e '.success == true' > /dev/null; then
    echo "✅ 角色查询成功"
else
    echo "❌ 角色查询失败"
    exit 1
fi

echo ""

# 6. 查询剧本
echo "6️⃣  测试剧本查询..."
SCENARIOS=$(curl -s http://localhost:8080/api/scenarios)
echo "$SCENARIOS" | jq '.data[] | {id, name}'
if echo "$SCENARIOS" | jq -e '.success == true' > /dev/null; then
    echo "✅ 剧本查询成功"
else
    echo "❌ 剧本查询失败"
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 测试总结"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "✅ 所有API测试通过！"
echo ""
echo "🎯 服务信息："
echo "  - 后端API: http://localhost:8080"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6379"
echo ""
echo "📝 查看日志："
echo "  docker compose logs -f backend"
echo ""
echo "🛑 停止服务："
echo "  docker compose down"
echo ""
