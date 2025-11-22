# TRPG Solo Engine - 快速开始指南

欢迎使用三角机构TRPG单人引擎！本指南将帮助你在5分钟内启动并运行系统。

## 前置要求

- **Go 1.21+** - [安装指南](https://golang.org/doc/install)
- **Docker & Docker Compose** - [安装指南](https://docs.docker.com/get-docker/)
- **Git** - 用于克隆仓库
- **Make** - 通常系统自带（macOS/Linux）

## Makefile 命令速查

项目提供了 Makefile 来简化常用操作：

```bash
make help        # 显示所有可用命令
make docker-up   # 启动所有服务（最快开始）
make dev-db      # 仅启动数据库（用于本地开发）
make run         # 本地运行应用
make test        # 运行测试
make build       # 构建二进制文件
make fmt         # 格式化代码
make clean       # 清理构建文件
```

💡 **提示**：本文档中的大部分命令都可以用 Makefile 简化！

## 快速启动（推荐）

### 方法1: 使用 Docker Compose（最简单）

```bash
# 1. 克隆仓库
git clone <repository-url>
cd trpg-solo-engine

# 2. 启动所有服务（后端、PostgreSQL、Redis）
make docker-up

# 3. 查看日志确认启动成功
make docker-logs

# 4. 等待服务就绪（约10秒）
# 看到 "starting server" 日志后按 Ctrl+C 退出日志查看
```

**验证服务运行**：
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

### 方法2: 本地开发模式

如果你需要修改代码并实时看到效果：

```bash
# 1. 启动数据库和Redis（使用Docker）
make dev-db

# 2. 等待服务就绪（约5秒）
sleep 5

# 3. 在本地运行应用
make run
```

或使用便捷脚本：
```bash
./scripts/start-local.sh
```

### 查看所有可用命令

```bash
make help
```

## 访问服务

服务启动后，你可以访问：

| 服务 | URL | 说明 |
|------|-----|------|
| **API文档** | http://localhost:8080/api/docs | 交互式Swagger UI文档 |
| **健康检查** | http://localhost:8080/health | 服务健康状态 |
| **API版本** | http://localhost:8080/api/version | API版本信息 |

## 快速测试

### 1. 创建角色

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
      {"name": "关系1", "description": "童年好友", "connection": 4, "played_by": "GM"},
      {"name": "关系2", "description": "前同事", "connection": 4, "played_by": "GM"},
      {"name": "关系3", "description": "邻居", "connection": 4, "played_by": "GM"}
    ]
  }'
```

保存返回的 `id` 字段（角色ID）。

### 2. 查看可用剧本

```bash
curl http://localhost:8080/api/scenarios
```

### 3. 创建游戏会话

```bash
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "你的角色ID",
    "scenario_id": "eternal-spring"
  }'
```

### 4. 执行掷骰

```bash
curl -X POST http://localhost:8080/api/dice/roll \
  -H "Content-Type: application/json" \
  -d '{"count": 6}'
```

## 使用 Postman 测试

1. 打开 Postman
2. 导入 `api/postman-collection.json`
3. 设置环境变量 `baseUrl` 为 `http://localhost:8080`
4. 开始测试所有API端点

## 常用命令

### 使用 Makefile（推荐）

```bash
# 查看所有可用命令
make help

# 构建应用
make build

# 运行应用（本地）
make run

# 运行测试
make test

# 代码格式化
make fmt

# 下载依赖
make deps

# Docker相关
make docker-build    # 构建Docker镜像
make docker-up       # 启动所有服务
make docker-down     # 停止所有服务
make docker-logs     # 查看后端日志

# 开发数据库
make dev-db          # 启动PostgreSQL和Redis
make dev-db-down     # 停止数据库服务

# 清理
make clean           # 清理构建文件
```

### Docker Compose 原始命令

如果你更喜欢直接使用 docker-compose：

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f backend

# 重启服务
docker-compose restart

# 停止服务
docker-compose stop

# 停止并删除容器
docker-compose down

# 重新构建并启动
docker-compose up -d --build
```

### 数据库管理

```bash
# 进入PostgreSQL容器
docker-compose exec postgres psql -U trpg -d trpg_solo_engine

# 查看数据库列表
docker-compose exec postgres psql -U trpg -c "\l"

# 备份数据库
docker-compose exec postgres pg_dump -U trpg trpg_solo_engine > backup.sql

# 恢复数据库
docker-compose exec -T postgres psql -U trpg trpg_solo_engine < backup.sql
```

### 开发命令

使用 Makefile：
```bash
# 运行所有测试
make test

# 代码格式化
make fmt

# 代码检查（需要安装 golangci-lint）
make lint

# 构建二进制文件
make build

# 下载并整理依赖
make deps
```

直接使用 Go 命令：
```bash
# 运行测试
go test ./...

# 运行特定包的测试
go test ./internal/service/...

# 运行属性测试
go test -v ./internal/domain/... -run Property

# 代码格式化
go fmt ./...

# 代码检查
go vet ./...

# 构建二进制文件
go build -o trpg-engine ./cmd/server
```

## 故障排除

### 问题1: 端口8080已被占用

**错误信息**：`bind: address already in use`

**解决方案**：
```bash
# 查找占用端口的进程
lsof -i :8080

# 停止进程
kill <PID>

# 或者修改端口
export SERVER_PORT=8081
docker-compose up -d
```

### 问题2: 数据库连接失败

**错误信息**：`database "trpg" does not exist`

**解决方案**：
```bash
# 运行修复脚本
./scripts/fix-database.sh

# 或手动创建数据库
docker-compose exec postgres psql -U trpg -d postgres -c "CREATE DATABASE trpg_solo_engine"
```

### 问题3: Docker容器无法启动

**解决方案**：
```bash
# 使用 Makefile
make docker-down
make docker-up
make docker-logs

# 或使用 docker-compose
docker-compose down -v
docker-compose up -d
docker-compose logs
```

### 问题4: API文档404

如果使用Docker运行，API文档路由可能不可用。使用本地模式：

```bash
# 停止Docker中的后端
docker-compose stop backend

# 启动数据库
make dev-db

# 本地运行
make run

# 或使用脚本
./scripts/start-local.sh
```

然后访问 http://localhost:8080/api/docs

## 配置说明

### 环境变量

主要环境变量（在 `docker-compose.yml` 中配置）：

```yaml
# 服务器配置
SERVER_PORT=8080
SERVER_MODE=release

# 数据库配置（注意：使用 DATABASE_DBNAME 而不是 DATABASE_NAME）
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_USER=trpg
DATABASE_PASSWORD=trpg_password
DATABASE_DBNAME=trpg_solo_engine
DATABASE_SSLMODE=disable

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379

# 日志配置
LOG_LEVEL=info
```

### 配置文件

主配置文件位于 `configs/config.yaml`，包含详细的配置选项。

环境变量优先级高于配置文件。

## 下一步

- 🛠️ 学习 [Makefile 使用指南](docs/MAKEFILE_GUIDE.md) - 掌握所有开发命令
- 📖 阅读 [API文档](api/README.md)
- 🧪 查看 [测试指南](api/TESTING_GUIDE.md)
- 🎮 了解 [游戏规则](.kiro/specs/trpg-solo-engine/design.md)
- 📜 探索 [剧本系统](scenarios/README.md)
- 🔧 查看 [配置指南](configs/CONFIG_GUIDE.md)

## 获取帮助

遇到问题？

1. 查看 [故障排除文档](docs/TROUBLESHOOTING.md)
2. 检查 [项目状态](PROJECT_STATUS.md)
3. 查看 [API变更日志](api/CHANGELOG.md)
4. 提交 Issue 到项目仓库

## 开发工作流

### 典型的开发流程

使用 Makefile 简化流程：

```bash
# 1. 启动依赖服务
make dev-db

# 2. 本地运行应用（便于调试）
make run

# 3. 修改代码...

# 4. 运行测试
make test

# 5. 提交前检查
make fmt
make lint  # 需要安装 golangci-lint

# 6. 构建Docker镜像测试
make docker-build
make docker-up
make docker-logs
```

或使用原始命令：

```bash
# 1. 启动依赖服务
docker-compose up -d postgres redis

# 2. 本地运行应用
go run cmd/server/main.go

# 3. 修改代码...

# 4. 运行测试
go test ./...

# 5. 提交前检查
go fmt ./...
go vet ./...

# 6. 构建Docker镜像测试
docker-compose build backend
docker-compose up -d
```

### 热重载开发

推荐使用 [air](https://github.com/cosmtrek/air) 实现热重载：

```bash
# 安装air
go install github.com/cosmtrek/air@latest

# 启动热重载
air
```

## 性能优化建议

### 开发环境

- 使用本地运行模式（`./scripts/start-local.sh`）获得更快的启动速度
- 启用调试日志：`export LOG_LEVEL=debug`
- 使用 `air` 实现热重载

### 生产环境

- 使用 Docker Compose 部署
- 设置 `SERVER_MODE=release`
- 配置适当的连接池大小
- 启用Redis缓存
- 配置速率限制

## 许可证

MIT License

---

**祝你使用愉快！** 🎲

如有问题，欢迎提交 Issue 或查看文档。
