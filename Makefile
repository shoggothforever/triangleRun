.PHONY: help build run test clean docker-build docker-up docker-down

help: ## 显示帮助信息
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 构建应用
	go build -o trpg-engine ./cmd/server

run: ## 运行应用
	go run ./cmd/server/main.go

test: ## 运行测试
	go test -v ./...

clean: ## 清理构建文件
	rm -f trpg-engine
	go clean

docker-build: ## 构建Docker镜像
	docker-compose build

docker-up: ## 启动Docker容器
	docker-compose up -d

docker-down: ## 停止Docker容器
	docker-compose down

docker-logs: ## 查看Docker日志
	docker-compose logs -f backend

fmt: ## 格式化代码
	go fmt ./...

lint: ## 代码检查
	golangci-lint run

deps: ## 下载依赖
	go mod download
	go mod tidy

dev-db: ## 启动开发数据库（仅PostgreSQL和Redis）
	docker-compose up -d postgres redis

dev-db-down: ## 停止开发数据库
	docker-compose down postgres redis
