# 三角机构TRPG单人引擎 - 快速启动指南

## 🚀 快速开始

### 前置要求

- Go 1.21+
- Docker 和 Docker Compose（推荐）
- 或者 PostgreSQL 15+ 和 Redis 7+（本地运行）

### 方式1: 使用Docker Compose（推荐）

这是最简单的方式，会自动启动所有依赖服务。

```bash
# 1. 启动所有服务
docker compose up -d

# 2. 查看日志
docker compose logs -f backend

# 3. 测试API
curl http://localhost:8080/health
curl http://localhost:8080/api/version

# 4. 停止服务
docker compose down
```

### 方式2: 测试服务器（无需数据库）

如果你只想快速测试核心功能，可以使用测试服务器：

```bash
# 1. 编译测试服务器
go build -o trpg-testserver ./cmd/testserver

# 2. 运行
./trpg-testserver

# 3. 测试API
curl http://localhost:8080/health
curl http://localhost:8080/api/version

# 4. 测试骰子系统
curl -X POST http://localhost:8080/api/dice/roll \
  -H "Content-Type: application/json" \
  -d '{"count": 6}'
```

**注意**: 测试服务器不支持数据持久化，仅用于测试核心功能。

### 方式3: 本地运行（需要手动启动依赖）

如果你想在本地开发环境运行：

```bash
# 1. 启动PostgreSQL
# 方式A: 使用Docker
docker run -d \
  --name trpg-postgres \
  -e POSTGRES_USER=trpg \
  -e POSTGRES_PASSWORD=trpg_password \
  -e POSTGRES_DB=trpg_solo_engine \
  -p 5432:5432 \
  postgres:15-alpine

# 方式B: 使用本地PostgreSQL
# 创建数据库: CREATE DATABASE trpg_solo_engine;

# 2. 启动Redis
# 方式A: 使用Docker
docker run -d \
  --name trpg-redis \
  -p 6379:6379 \
  redis:7-alpine

# 方式B: 使用本地Redis
# redis-server

# 3. 配置环境变量（可选）
cp .env.example .env
# 编辑 .env 文件，填入你的配置

# 4. 运行服务器
go run cmd/server/main.go

# 或者编译后运行
go build -o trpg-engine ./cmd/server
./trpg-engine
```

## 📋 验证安装

运行测试脚本验证项目状态：

```bash
./scripts/test-project.sh
```

这会检查：
- ✅ 项目编译
- ✅ 配置文件
- ✅ 单元测试
- ✅ 依赖管理
- ✅ 数据文件

## 🧪 测试API

### 健康检查

```bash
curl http://localhost:8080/health
```

响应：
```json
{
  "status": "ok",
  "service": "trpg-solo-engine",
  "database": "ok",
  "redis": "ok"
}
```

### 版本信息

```bash
curl http://localhost:8080/api/version
```

响应：
```json
{
  "version": "0.1.0",
  "name": "TRPG Solo Engine"
}
```

### 骰子掷骰

```bash
curl -X POST http://localhost:8080/api/dice/roll \
  -H "Content-Type: application/json" \
  -d '{
    "count": 6,
    "quality": "focus"
  }'
```

响应：
```json
{
  "success": true,
  "data": {
    "dice": [3, 1, 4, 3, 2, 1],
    "threes": 2,
    "success": true,
    "chaos": 0,
    "overload": 0,
    "triple_ascension": false
  }
}
```

### 创建角色

```bash
curl -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试特工",
    "pronouns": "他/him",
    "anomaly_type": "whisper",
    "reality_type": "caretaker",
    "career_type": "pr"
  }'
```

### 查询剧本

```bash
curl http://localhost:8080/api/scenarios
```

## 🔧 配置

### 主配置文件

编辑 `configs/config.yaml` 来调整配置：

```yaml
server:
  port: "8080"
  mode: "debug"  # debug, release, test

log:
  level: "info"  # debug, info, warn, error

database:
  host: "localhost"
  port: 5432
  user: "trpg"
  password: "trpg_password"
  dbname: "trpg_solo_engine"

redis:
  host: "localhost"
  port: 6379

ai:
  provider: "openai"
  api_key: ""  # 从环境变量读取
  model: "gpt-4"
```

详细配置说明见 [configs/CONFIG_GUIDE.md](configs/CONFIG_GUIDE.md)

### 环境变量

创建 `.env` 文件（从 `.env.example` 复制）：

```bash
cp .env.example .env
```

必须设置的环境变量（生产环境）：

```bash
# 数据库密码
DATABASE_PASSWORD=your_secure_password

# AI API密钥
AI_API_KEY=your_openai_api_key

# JWT密钥
JWT_SECRET=your_jwt_secret_key
```

## 📊 运行测试

### 运行所有测试

```bash
go test ./... -v
```

### 运行单元测试

```bash
go test ./internal/domain/... -v
go test ./internal/service/... -v
go test ./internal/handler/... -v
```

### 运行属性测试

```bash
go test ./internal/domain/... -run Property -v
go test ./internal/service/... -run Property -v
```

### 测试覆盖率

```bash
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 🐛 故障排查

### 数据库连接失败

```
Error: failed to connect to database
```

**解决方案：**
1. 确认PostgreSQL正在运行：`docker ps` 或 `pg_isready`
2. 检查配置：`configs/config.yaml` 中的数据库设置
3. 测试连接：`psql -h localhost -U trpg -d trpg_solo_engine`

### Redis连接失败

```
Error: failed to connect to redis
```

**解决方案：**
1. 确认Redis正在运行：`docker ps` 或 `redis-cli ping`
2. 检查配置：`configs/config.yaml` 中的Redis设置
3. 测试连接：`redis-cli -h localhost -p 6379 ping`

### 端口已被占用

```
Error: bind: address already in use
```

**解决方案：**
1. 更改端口：编辑 `configs/config.yaml` 中的 `server.port`
2. 或者停止占用端口的进程：`lsof -i :8080`

### 配置文件未找到

```
Error: Config File "config" Not Found
```

**解决方案：**
1. 确保在项目根目录运行
2. 确保 `configs/config.yaml` 存在
3. 或者使用环境变量配置

## 📚 下一步

- 阅读 [API文档](docs/API.md)（待创建）
- 查看 [配置指南](configs/CONFIG_GUIDE.md)
- 了解 [ARC系统](.kiro/specs/trpg-solo-engine/arc-system.md)
- 阅读 [设计文档](.kiro/specs/trpg-solo-engine/design.md)

## 🆘 获取帮助

- 查看 [需求文档](.kiro/specs/trpg-solo-engine/requirements.md)
- 查看 [任务列表](.kiro/specs/trpg-solo-engine/tasks.md)
- 运行验证脚本：`./scripts/validate-config.sh`
- 运行测试脚本：`./scripts/test-project.sh`

## 📝 开发工作流

```bash
# 1. 拉取最新代码
git pull

# 2. 安装依赖
go mod download

# 3. 运行测试
go test ./...

# 4. 启动开发服务器
go run cmd/testserver/main.go

# 5. 进行开发...

# 6. 运行测试
go test ./...

# 7. 提交代码
git add .
git commit -m "feat: 添加新功能"
git push
```

## 🎯 常用命令

```bash
# 编译
go build -o trpg-engine ./cmd/server

# 运行
./trpg-engine

# 测试
go test ./...

# 格式化代码
go fmt ./...

# 检查代码
go vet ./...

# 更新依赖
go mod tidy

# 查看依赖
go mod graph

# Docker相关
docker compose up -d          # 启动
docker compose down           # 停止
docker compose logs -f        # 查看日志
docker compose ps             # 查看状态
docker compose restart        # 重启
```

## ✨ 特性

- ✅ 完整的6d4骰子系统
- ✅ ARC角色创建（9种异常×9种现实×9种职能）
- ✅ 资质保证和过载机制
- ✅ 混沌池管理
- ✅ 请求机构系统
- ✅ 异常能力系统
- ✅ 伤害和人寿保险
- ✅ 绩效追踪（嘉奖/申诫）
- ✅ 游戏会话管理
- ✅ 剧本系统
- ✅ 场景和NPC管理
- ✅ 线索追踪
- ✅ 存档系统
- ✅ RESTful API
- ✅ 速率限制
- ✅ 日志系统
- ✅ 健康检查

## 📄 许可证

[添加许可证信息]
