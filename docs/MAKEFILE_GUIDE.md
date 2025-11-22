# Makefile 使用指南

本项目使用 Makefile 来简化常用的开发和部署操作。

## 查看所有命令

```bash
make help
```

输出示例：
```
可用命令:
  build           构建应用
  run             运行应用
  test            运行测试
  clean           清理构建文件
  docker-build    构建Docker镜像
  docker-up       启动Docker容器
  docker-down     停止Docker容器
  docker-logs     查看Docker日志
  fmt             格式化代码
  lint            代码检查
  deps            下载依赖
  dev-db          启动开发数据库（仅PostgreSQL和Redis）
  dev-db-down     停止开发数据库
```

## 常用工作流

### 1. 快速开始（Docker）

```bash
# 启动所有服务
make docker-up

# 查看日志
make docker-logs

# 停止服务
make docker-down
```

### 2. 本地开发

```bash
# 启动数据库
make dev-db

# 运行应用
make run

# 在另一个终端运行测试
make test

# 停止数据库
make dev-db-down
```

### 3. 代码质量检查

```bash
# 格式化代码
make fmt

# 代码检查（需要安装 golangci-lint）
make lint

# 运行测试
make test
```

### 4. 构建和部署

```bash
# 构建本地二进制
make build

# 构建Docker镜像
make docker-build

# 启动Docker服务
make docker-up
```

## 命令详解

### 开发命令

| 命令 | 说明 | 等价命令 |
|------|------|----------|
| `make build` | 构建应用 | `go build -o trpg-engine ./cmd/server` |
| `make run` | 运行应用 | `go run ./cmd/server/main.go` |
| `make test` | 运行测试 | `go test -v ./...` |
| `make fmt` | 格式化代码 | `go fmt ./...` |
| `make lint` | 代码检查 | `golangci-lint run` |
| `make deps` | 下载依赖 | `go mod download && go mod tidy` |
| `make clean` | 清理构建文件 | `rm -f trpg-engine && go clean` |

### Docker命令

| 命令 | 说明 | 等价命令 |
|------|------|----------|
| `make docker-build` | 构建Docker镜像 | `docker-compose build` |
| `make docker-up` | 启动所有服务 | `docker-compose up -d` |
| `make docker-down` | 停止所有服务 | `docker-compose down` |
| `make docker-logs` | 查看后端日志 | `docker-compose logs -f backend` |
| `make dev-db` | 启动数据库 | `docker-compose up -d postgres redis` |
| `make dev-db-down` | 停止数据库 | `docker-compose down postgres redis` |

## 典型场景

### 场景1: 第一次启动项目

```bash
# 1. 克隆项目
git clone <repository-url>
cd trpg-solo-engine

# 2. 下载依赖
make deps

# 3. 启动所有服务
make docker-up

# 4. 查看日志确认启动
make docker-logs
```

### 场景2: 日常开发

```bash
# 早上开始工作
make dev-db          # 启动数据库
make run             # 运行应用

# 修改代码...

# 运行测试
make test

# 格式化代码
make fmt

# 晚上结束工作
make dev-db-down     # 停止数据库
```

### 场景3: 提交代码前

```bash
# 格式化代码
make fmt

# 运行代码检查
make lint

# 运行所有测试
make test

# 构建确认无错误
make build

# 清理构建文件
make clean
```

### 场景4: 部署测试

```bash
# 构建Docker镜像
make docker-build

# 启动服务
make docker-up

# 查看日志
make docker-logs

# 测试API
curl http://localhost:8080/health

# 停止服务
make docker-down
```

## 自定义 Makefile

如果你需要添加自己的命令，编辑 `Makefile`：

```makefile
.PHONY: my-command

my-command: ## 我的自定义命令
	@echo "执行自定义命令"
	# 你的命令...
```

然后运行：
```bash
make my-command
```

## 故障排除

### 问题1: make: command not found

**解决方案**：

macOS:
```bash
xcode-select --install
```

Linux (Ubuntu/Debian):
```bash
sudo apt-get install build-essential
```

Linux (CentOS/RHEL):
```bash
sudo yum groupinstall "Development Tools"
```

### 问题2: golangci-lint: command not found

`make lint` 需要安装 golangci-lint：

```bash
# macOS
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# 或使用 go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 问题3: 权限错误

如果遇到权限问题：

```bash
# 给 Makefile 执行权限
chmod +x Makefile

# 或使用 sudo（不推荐）
sudo make <command>
```

## 最佳实践

1. **使用 `make help`** - 忘记命令时随时查看
2. **本地开发用 `make dev-db` + `make run`** - 更快的迭代速度
3. **生产部署用 `make docker-up`** - 完整的容器化环境
4. **提交前运行 `make fmt` 和 `make test`** - 保证代码质量
5. **定期运行 `make deps`** - 保持依赖最新

## 相关文档

- [快速开始指南](../QUICKSTART.md)
- [项目README](../README.md)
- [Docker Compose配置](../docker-compose.yml)

## 贡献

如果你有好的 Makefile 命令建议，欢迎提交 PR！

---

**提示**：Makefile 让开发更简单！🚀
