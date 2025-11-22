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

## API端点

- `GET /health` - 健康检查
- `GET /api/version` - 版本信息

更多API文档请访问 `/swagger` (开发中)

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
