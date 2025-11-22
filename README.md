# TRPG Solo Engine - 三角机构单人引擎

基于《三角机构》TRPG规则的单人自助体验应用后端系统。

## 功能特性

- 完整的6d4骰子判定系统
- ARC角色系统（异常、现实、职能）
- 资质保证和过载机制
- 混沌池和请求机构
- 剧本模组系统
- AI总经理叙事
- 完整的存档系统

## 快速开始

### 前置要求

- Go 1.21+
- Docker & Docker Compose (可选)
- PostgreSQL 15+ (如果不使用Docker)
- Redis 7+ (如果不使用Docker)

### 本地开发

1. 克隆仓库
```bash
git clone <repository-url>
cd trpg-solo-engine
```

2. 安装依赖
```bash
make deps
```

3. 启动开发数据库
```bash
make dev-db
```

4. 测试数据库连接
```bash
go run ./cmd/dbtest/main.go
```

5. 配置环境
```bash
cp .env.example .env
# 编辑 .env 设置AI服务等配置
```

6. 运行应用
```bash
make run
```

### 使用Docker

1. 启动所有服务
```bash
make docker-up
```

2. 查看日志
```bash
make docker-logs
```

3. 停止服务
```bash
make docker-down
```

## API文档

完整的API文档已通过OpenAPI 3.0规范提供。

### 访问文档

启动服务器后，访问：

```
http://localhost:8080/api/docs
```

你将看到交互式的Swagger UI界面，可以：
- 浏览所有API端点
- 查看请求/响应示例
- 直接在浏览器中测试API

### 主要端点

- **健康检查**: `GET /health`, `GET /api/version`
- **角色管理**: `POST /api/agents`, `GET /api/agents`, `GET /api/agents/{id}`, `PUT /api/agents/{id}`, `DELETE /api/agents/{id}`
- **游戏会话**: `POST /api/sessions`, `GET /api/sessions/{id}`, `POST /api/sessions/{id}/actions`, `POST /api/sessions/{id}/phase`
- **骰子系统**: `POST /api/dice/roll`, `POST /api/dice/ability`, `POST /api/dice/request`
- **剧本管理**: `GET /api/scenarios`, `GET /api/scenarios/{id}`, `GET /api/scenarios/{id}/scenes/{sceneId}`
- **存档管理**: `POST /api/saves`, `GET /api/saves`, `GET /api/saves/{id}`, `POST /api/saves/{id}/load`, `DELETE /api/saves/{id}`

### 文档资源

- [API快速开始](api/QUICKSTART.md) - 快速上手指南
- [API详细文档](api/README.md) - 完整API说明
- [OpenAPI规范](api/openapi.yaml) - 原始规范文件
- [Postman集合](api/postman-collection.json) - 导入到Postman测试
- [API变更日志](api/CHANGELOG.md) - 版本变更记录

## 项目结构

```
.
├── cmd/
│   └── server/          # 主应用入口
├── internal/
│   ├── domain/          # 领域模型
│   ├── service/         # 业务逻辑
│   ├── repository/      # 数据访问
│   └── handler/         # HTTP处理器
├── pkg/                 # 公共包
├── configs/             # 配置文件
└── docs/                # 文档
```

## 开发

```bash
# 构建
make build

# 运行测试
make test

# 格式化代码
make fmt

# 代码检查
make lint
```

## 许可证

MIT
