# API测试指南

本指南帮助你测试TRPG Solo Engine的API功能。

## 前置准备

### 启动服务器

```bash
# 方法1: 使用编译后的二进制
./trpg-engine

# 方法2: 使用go run
go run cmd/server/main.go

# 方法3: 使用Docker
docker-compose up
```

确认服务器启动成功：
```bash
curl http://localhost:8080/health
```

预期响应：
```json
{
  "status": "ok",
  "service": "trpg-solo-engine",
  "database": "ok",
  "redis": "ok"
}
```

## 测试方法

### 方法1: 使用Swagger UI（推荐）

1. 打开浏览器访问：`http://localhost:8080/api/docs`
2. 选择要测试的端点
3. 点击 "Try it out"
4. 填写参数
5. 点击 "Execute"
6. 查看响应

**优点**：
- 可视化界面
- 自动填充示例
- 实时查看响应
- 无需额外工具

### 方法2: 使用Postman

1. 打开Postman
2. 导入 `api/postman-collection.json`
3. 设置环境变量：
   - `baseUrl`: `http://localhost:8080`
4. 按顺序执行请求
5. 保存返回的ID到对应变量

**优点**：
- 专业的API测试工具
- 可保存测试历史
- 支持自动化测试
- 团队协作功能

### 方法3: 使用curl

直接在命令行测试，适合脚本化和自动化。

**优点**：
- 快速简单
- 易于脚本化
- 无需额外工具
- 适合CI/CD

## 完整测试流程

### 1. 健康检查

```bash
# 检查服务状态
curl http://localhost:8080/health

# 获取版本信息
curl http://localhost:8080/api/version
```

### 2. 创建角色

```bash
curl -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试角色",
    "pronouns": "他/him",
    "anomaly_type": "whisper",
    "reality_type": "caretaker",
    "career_type": "pr",
    "relationships": [
      {
        "name": "关系1",
        "description": "童年好友",
        "connection": 4,
        "played_by": "GM"
      },
      {
        "name": "关系2",
        "description": "前同事",
        "connection": 4,
        "played_by": "GM"
      },
      {
        "name": "关系3",
        "description": "邻居",
        "connection": 4,
        "played_by": "GM"
      }
    ]
  }'
```

**保存返回的角色ID**，例如：`123e4567-e89b-12d3-a456-426614174000`

### 3. 查询角色

```bash
# 列出所有角色
curl http://localhost:8080/api/agents

# 获取特定角色
curl http://localhost:8080/api/agents/123e4567-e89b-12d3-a456-426614174000
```

### 4. 查看可用剧本

```bash
# 列出所有剧本
curl http://localhost:8080/api/scenarios

# 获取剧本详情
curl http://localhost:8080/api/scenarios/eternal-spring
```

### 5. 创建游戏会话

```bash
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "123e4567-e89b-12d3-a456-426614174000",
    "scenario_id": "eternal-spring"
  }'
```

**保存返回的会话ID**

### 6. 执行掷骰

```bash
# 基础掷骰
curl -X POST http://localhost:8080/api/dice/roll \
  -H "Content-Type: application/json" \
  -d '{"count": 6}'

# 异常能力掷骰
curl -X POST http://localhost:8080/api/dice/ability \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "123e4567-e89b-12d3-a456-426614174000",
    "ability_id": "whisper_read_thoughts",
    "qa_spend": 2
  }'

# 请求机构掷骰
curl -X POST http://localhost:8080/api/dice/request \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "123e4567-e89b-12d3-a456-426614174000",
    "quality": "subtlety",
    "effect": "我在进入房间前已经准备好了开锁工具",
    "causal_chain": "我一直有随身携带工具的习惯",
    "qa_spend": 1
  }'
```

### 7. 执行游戏行动

```bash
# 移动到场景
curl -X POST http://localhost:8080/api/sessions/YOUR_SESSION_ID/actions \
  -H "Content-Type: application/json" \
  -d '{
    "action_type": "move_to_scene",
    "target": "scene_01"
  }'

# 收集线索
curl -X POST http://localhost:8080/api/sessions/YOUR_SESSION_ID/actions \
  -H "Content-Type: application/json" \
  -d '{
    "action_type": "collect_clue",
    "target": "clue_fountain"
  }'

# 添加混沌
curl -X POST http://localhost:8080/api/sessions/YOUR_SESSION_ID/actions \
  -H "Content-Type: application/json" \
  -d '{
    "action_type": "add_chaos",
    "parameters": {
      "amount": 3
    }
  }'
```

### 8. 转换游戏阶段

```bash
curl -X POST http://localhost:8080/api/sessions/YOUR_SESSION_ID/phase \
  -H "Content-Type: application/json" \
  -d '{
    "phase": "investigation"
  }'
```

### 9. 保存游戏

```bash
curl -X POST http://localhost:8080/api/saves \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "YOUR_SESSION_ID",
    "name": "调查阶段-喷泉前"
  }'
```

**保存返回的存档ID**

### 10. 管理存档

```bash
# 列出所有存档
curl http://localhost:8080/api/saves

# 获取存档详情
curl http://localhost:8080/api/saves/YOUR_SAVE_ID

# 加载存档
curl -X POST http://localhost:8080/api/saves/YOUR_SAVE_ID/load

# 删除存档
curl -X DELETE http://localhost:8080/api/saves/YOUR_SAVE_ID
```

## 测试脚本

创建一个完整的测试脚本 `test-api.sh`：

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

echo "1. 健康检查..."
curl -s $BASE_URL/health | jq

echo -e "\n2. 创建角色..."
AGENT_RESPONSE=$(curl -s -X POST $BASE_URL/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试角色",
    "pronouns": "他/him",
    "anomaly_type": "whisper",
    "reality_type": "caretaker",
    "career_type": "pr",
    "relationships": [
      {"name": "关系1", "description": "描述1", "connection": 4, "played_by": "GM"},
      {"name": "关系2", "description": "描述2", "connection": 4, "played_by": "GM"},
      {"name": "关系3", "description": "描述3", "connection": 4, "played_by": "GM"}
    ]
  }')

AGENT_ID=$(echo $AGENT_RESPONSE | jq -r '.data.id')
echo "角色ID: $AGENT_ID"

echo -e "\n3. 创建会话..."
SESSION_RESPONSE=$(curl -s -X POST $BASE_URL/api/sessions \
  -H "Content-Type: application/json" \
  -d "{
    \"agent_id\": \"$AGENT_ID\",
    \"scenario_id\": \"eternal-spring\"
  }")

SESSION_ID=$(echo $SESSION_RESPONSE | jq -r '.data.id')
echo "会话ID: $SESSION_ID"

echo -e "\n4. 执行掷骰..."
curl -s -X POST $BASE_URL/api/dice/roll \
  -H "Content-Type: application/json" \
  -d '{"count": 6}' | jq

echo -e "\n5. 保存游戏..."
SAVE_RESPONSE=$(curl -s -X POST $BASE_URL/api/saves \
  -H "Content-Type: application/json" \
  -d "{
    \"session_id\": \"$SESSION_ID\",
    \"name\": \"测试存档\"
  }")

SAVE_ID=$(echo $SAVE_RESPONSE | jq -r '.data.id')
echo "存档ID: $SAVE_ID"

echo -e "\n测试完成！"
echo "角色ID: $AGENT_ID"
echo "会话ID: $SESSION_ID"
echo "存档ID: $SAVE_ID"
```

使用方法：
```bash
chmod +x test-api.sh
./test-api.sh
```

## 错误处理测试

### 测试无效输入

```bash
# 缺少必需字段
curl -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试"
  }'

# 预期: 400 Bad Request
```

### 测试不存在的资源

```bash
# 不存在的角色ID
curl http://localhost:8080/api/agents/00000000-0000-0000-0000-000000000000

# 预期: 404 Not Found
```

### 测试无效的阶段转换

```bash
# 尝试转换到无效阶段
curl -X POST http://localhost:8080/api/sessions/YOUR_SESSION_ID/phase \
  -H "Content-Type: application/json" \
  -d '{
    "phase": "invalid_phase"
  }'

# 预期: 400 Bad Request
```

## 性能测试

使用 `ab` (Apache Bench) 进行简单的性能测试：

```bash
# 测试健康检查端点
ab -n 1000 -c 10 http://localhost:8080/health

# 测试掷骰端点
ab -n 100 -c 5 -p dice-payload.json -T application/json \
  http://localhost:8080/api/dice/roll
```

## 常见问题

### Q: 如何查看详细的请求/响应？

A: 使用 `curl -v` 查看详细信息：
```bash
curl -v http://localhost:8080/api/agents
```

### Q: 如何格式化JSON响应？

A: 使用 `jq` 工具：
```bash
curl http://localhost:8080/api/agents | jq
```

### Q: 如何保存响应到文件？

A: 使用重定向：
```bash
curl http://localhost:8080/api/agents > agents.json
```

### Q: 如何测试并发请求？

A: 使用 `&` 并行执行：
```bash
for i in {1..10}; do
  curl http://localhost:8080/api/dice/roll &
done
wait
```

## 自动化测试

### 使用Newman（Postman CLI）

```bash
# 安装Newman
npm install -g newman

# 运行Postman集合
newman run api/postman-collection.json \
  --environment postman-env.json \
  --reporters cli,json
```

### 集成到CI/CD

在 `.github/workflows/api-test.yml` 中：

```yaml
name: API Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Start services
        run: docker-compose up -d
      - name: Wait for services
        run: sleep 10
      - name: Run API tests
        run: ./test-api.sh
      - name: Stop services
        run: docker-compose down
```

## 下一步

- 阅读 [API文档](README.md)
- 查看 [快速开始指南](QUICKSTART.md)
- 探索 [OpenAPI规范](openapi.yaml)
- 了解 [变更日志](CHANGELOG.md)

## 获取帮助

如果遇到问题：

1. 检查服务器日志
2. 验证请求格式
3. 查看API文档
4. 提交Issue

---

祝测试愉快！🎲
