# API 文档快速开始指南

本指南将帮助你快速开始使用TRPG Solo Engine的API文档。

## 启动服务器

首先，确保服务器正在运行：

```bash
# 方法1: 直接运行编译后的二进制文件
./trpg-engine

# 方法2: 使用go run
go run cmd/server/main.go

# 方法3: 使用Docker
docker-compose up
```

服务器默认在 `http://localhost:8080` 启动。

## 访问API文档

### 交互式文档（推荐）

在浏览器中打开：

```
http://localhost:8080/api/docs
```

你将看到Swagger UI界面，可以：
- 浏览所有API端点
- 查看请求/响应示例
- 直接在浏览器中测试API

### 原始OpenAPI规范

如果需要查看或下载原始的OpenAPI YAML文件：

```
http://localhost:8080/api/docs/openapi.yaml
```

## 使用Postman测试API

### 导入Postman集合

1. 打开Postman
2. 点击 "Import" 按钮
3. 选择 `api/postman-collection.json` 文件
4. 集合将被导入，包含所有预配置的请求

### 设置环境变量

在Postman中设置以下变量：

- `baseUrl`: `http://localhost:8080`
- `agentId`: 创建角色后获得的ID
- `sessionId`: 创建会话后获得的ID
- `saveId`: 创建存档后获得的ID

## 快速测试流程

### 1. 检查服务健康状态

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
      {"name": "关系1", "description": "描述1", "connection": 4, "played_by": "GM"},
      {"name": "关系2", "description": "描述2", "connection": 4, "played_by": "GM"},
      {"name": "关系3", "description": "描述3", "connection": 4, "played_by": "GM"}
    ]
  }'
```

保存返回的 `id` 字段，这是你的角色ID。

### 3. 列出可用剧本

```bash
curl http://localhost:8080/api/scenarios
```

### 4. 创建游戏会话

使用你的角色ID和剧本ID：

```bash
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "你的角色ID",
    "scenario_id": "eternal-spring"
  }'
```

保存返回的会话ID。

### 5. 执行掷骰

```bash
curl -X POST http://localhost:8080/api/dice/roll \
  -H "Content-Type: application/json" \
  -d '{"count": 6}'
```

### 6. 保存游戏

```bash
curl -X POST http://localhost:8080/api/saves \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "你的会话ID",
    "name": "我的第一个存档"
  }'
```

## 常见问题

### Q: 如何查看所有可用的API端点？

A: 访问 `http://localhost:8080/api/docs` 查看完整的API文档。

### Q: API返回404错误

A: 确保：
1. 服务器正在运行
2. URL路径正确（注意 `/api` 前缀）
3. 资源ID存在（对于需要ID的端点）

### Q: 如何调试API请求？

A: 
1. 使用Swagger UI的 "Try it out" 功能
2. 查看服务器日志输出
3. 使用Postman的Console查看详细请求/响应

### Q: 支持哪些内容类型？

A: 所有POST/PUT请求都使用 `application/json`，响应也是JSON格式。

## 下一步

- 阅读完整的 [API文档](README.md)
- 查看 [OpenAPI规范](openapi.yaml)
- 探索 [剧本系统](../scenarios/README.md)
- 了解 [游戏规则](../.kiro/specs/trpg-solo-engine/design.md)

## 获取帮助

如果遇到问题：

1. 检查服务器日志
2. 查看API文档中的错误响应说明
3. 参考示例请求
4. 提交Issue到项目仓库

## 许可证

MIT License
