# TRPG Solo Engine API 文档

本目录包含TRPG Solo Engine的完整API文档。

## 文件说明

- `openapi.yaml` - OpenAPI 3.0规范文件，定义了所有API端点、请求/响应格式和数据模型
- `swagger-ui.html` - Swagger UI界面，用于交互式浏览和测试API

## 查看文档

### 方法1: 使用服务器内置路由

启动服务器后，访问以下URL：

```
http://localhost:8080/api/docs
```

这将显示交互式的Swagger UI界面。

### 方法2: 使用在线Swagger Editor

1. 访问 https://editor.swagger.io/
2. 将 `openapi.yaml` 的内容粘贴到编辑器中
3. 在右侧查看渲染后的文档

### 方法3: 使用本地工具

使用任何支持OpenAPI 3.0的工具，例如：

- Postman: 导入 `openapi.yaml` 文件
- Insomnia: 导入 `openapi.yaml` 文件
- VS Code: 安装 OpenAPI (Swagger) Editor 扩展

## API概览

### 端点分类

1. **健康检查** (`/health`, `/api/version`)
   - 检查服务状态
   - 获取版本信息

2. **角色管理** (`/api/agents`)
   - 创建、查询、更新、删除角色
   - 管理ARC系统（异常、现实、职能）

3. **游戏会话** (`/api/sessions`)
   - 创建和管理游戏会话
   - 执行游戏行动
   - 转换游戏阶段

4. **骰子系统** (`/api/dice`)
   - 基础掷骰
   - 异常能力掷骰
   - 请求机构掷骰

5. **剧本管理** (`/api/scenarios`)
   - 列出可用剧本
   - 获取剧本详情
   - 获取场景信息

6. **存档管理** (`/api/saves`)
   - 保存游戏状态
   - 加载存档
   - 管理存档列表

## 响应格式

所有API响应遵循统一格式：

### 成功响应
```json
{
  "success": true,
  "data": {
    // 响应数据
  }
}
```

### 错误响应
```json
{
  "success": false,
  "error": "错误信息",
  "details": {
    // 可选的详细错误信息
  }
}
```

## HTTP状态码

- `200 OK` - 请求成功
- `201 Created` - 资源创建成功
- `400 Bad Request` - 请求参数无效
- `404 Not Found` - 资源不存在
- `409 Conflict` - 状态冲突（如阶段转换不合法）
- `422 Unprocessable Entity` - 数据损坏或格式错误
- `500 Internal Server Error` - 服务器内部错误

## 示例请求

### 创建角色

```bash
curl -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "张三",
    "pronouns": "他/him",
    "anomaly_type": "whisper",
    "reality_type": "caretaker",
    "career_type": "pr",
    "relationships": [
      {
        "name": "李四",
        "description": "童年好友",
        "connection": 4,
        "played_by": "GM"
      },
      {
        "name": "王五",
        "description": "前同事",
        "connection": 4,
        "played_by": "GM"
      },
      {
        "name": "赵六",
        "description": "邻居",
        "connection": 4,
        "played_by": "GM"
      }
    ]
  }'
```

### 创建游戏会话

```bash
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "123e4567-e89b-12d3-a456-426614174000",
    "scenario_id": "eternal-spring"
  }'
```

### 执行掷骰

```bash
curl -X POST http://localhost:8080/api/dice/roll \
  -H "Content-Type: application/json" \
  -d '{
    "count": 6
  }'
```

## 开发说明

### 更新文档

修改 `openapi.yaml` 后，文档会自动更新。无需重启服务器。

### 验证规范

使用在线工具验证OpenAPI规范：

```bash
# 使用swagger-cli（需要先安装）
npm install -g @apidevtools/swagger-cli
swagger-cli validate openapi.yaml
```

### 生成客户端代码

可以使用OpenAPI Generator生成各种语言的客户端代码：

```bash
# 安装openapi-generator-cli
npm install -g @openapitools/openapi-generator-cli

# 生成TypeScript客户端
openapi-generator-cli generate \
  -i openapi.yaml \
  -g typescript-axios \
  -o ./client/typescript

# 生成Python客户端
openapi-generator-cli generate \
  -i openapi.yaml \
  -g python \
  -o ./client/python
```

## 相关资源

- [OpenAPI 3.0 规范](https://swagger.io/specification/)
- [Swagger UI 文档](https://swagger.io/tools/swagger-ui/)
- [《三角机构》规则书](../docs/triangle-agency-rules.pdf)

## 许可证

MIT License
