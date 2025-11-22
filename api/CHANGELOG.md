# API 变更日志

本文档记录TRPG Solo Engine API的所有重要变更。

## [0.1.0] - 2024-11-22

### 新增

#### 健康检查端点
- `GET /health` - 检查服务、数据库和Redis状态
- `GET /api/version` - 获取API版本信息

#### 角色管理端点
- `POST /api/agents` - 创建新角色
- `GET /api/agents` - 列出所有角色
- `GET /api/agents/{id}` - 获取角色详情
- `PUT /api/agents/{id}` - 更新角色信息
- `DELETE /api/agents/{id}` - 删除角色

支持完整的ARC系统：
- 9种异常体类型（whisper, catalog, drain, timepiece, growth, firearm, dream, manifold, absence）
- 9种现实类型（caretaker, overbooked, hunted, star, struggling, newborn, romantic, pillar, outsider）
- 9种职能类型（pr, rd, barista, ceo, intern, gravedigger, reception, hotline, clown）

#### 游戏会话端点
- `POST /api/sessions` - 创建游戏会话
- `GET /api/sessions/{id}` - 获取会话详情
- `POST /api/sessions/{id}/actions` - 执行游戏行动
- `POST /api/sessions/{id}/phase` - 转换游戏阶段

支持的行动类型：
- `move_to_scene` - 移动到场景
- `collect_clue` - 收集线索
- `unlock_location` - 解锁地点
- `add_chaos` - 添加混沌
- `update_npc_state` - 更新NPC状态

支持的游戏阶段：
- `morning` - 晨会阶段
- `investigation` - 调查阶段
- `encounter` - 遭遇阶段
- `aftermath` - 余波阶段

#### 骰子系统端点
- `POST /api/dice/roll` - 基础6d4掷骰
- `POST /api/dice/ability` - 异常能力掷骰
- `POST /api/dice/request` - 请求机构掷骰

特性：
- 统计"3"的数量判定成功/失败
- 支持资质保证（QA）调整
- 自动计算混沌生成
- 三重升华检测

#### 剧本管理端点
- `GET /api/scenarios` - 列出所有可用剧本
- `GET /api/scenarios/{id}` - 获取剧本详情
- `GET /api/scenarios/{id}/scenes/{sceneId}` - 获取场景详情

#### 存档管理端点
- `POST /api/saves` - 保存游戏状态
- `GET /api/saves` - 列出所有存档（支持按会话ID过滤）
- `GET /api/saves/{id}` - 获取存档详情
- `POST /api/saves/{id}/load` - 加载存档
- `DELETE /api/saves/{id}` - 删除存档

#### 文档
- OpenAPI 3.0规范文件
- Swagger UI交互式文档
- Postman集合
- 快速开始指南
- API使用说明

### 响应格式

所有API响应遵循统一格式：

成功响应：
```json
{
  "success": true,
  "data": { ... }
}
```

错误响应：
```json
{
  "success": false,
  "error": "错误信息",
  "details": { ... }
}
```

### HTTP状态码

- `200 OK` - 请求成功
- `201 Created` - 资源创建成功
- `400 Bad Request` - 请求参数无效
- `404 Not Found` - 资源不存在
- `409 Conflict` - 状态冲突
- `422 Unprocessable Entity` - 数据损坏
- `500 Internal Server Error` - 服务器错误

### 已知限制

- 当前版本未实现认证机制
- 不支持多用户并发游戏
- AI服务集成尚未完成
- 部分混沌效应需要手动处理

### 技术细节

- API版本：0.1.0
- OpenAPI规范：3.0.3
- 默认端口：8080
- 内容类型：application/json
- 字符编码：UTF-8

## 未来计划

### [0.2.0] - 计划中

- [ ] JWT认证支持
- [ ] WebSocket实时通信
- [ ] AI总经理集成
- [ ] 批量操作API
- [ ] 高级查询和过滤
- [ ] 速率限制配置
- [ ] API密钥管理

### [0.3.0] - 计划中

- [ ] 多人游戏支持
- [ ] 实时协作功能
- [ ] 游戏回放系统
- [ ] 统计和分析API
- [ ] 自定义剧本上传
- [ ] 社区功能API

## 迁移指南

### 从无到0.1.0

这是首个版本，无需迁移。

## 贡献

如果你发现API问题或有改进建议，请：

1. 查看现有的Issue
2. 创建新的Issue描述问题
3. 提交Pull Request

## 许可证

MIT License

---

**注意**: 本API仍在积极开发中，可能会有破坏性变更。建议在生产环境使用前等待1.0.0稳定版本。
