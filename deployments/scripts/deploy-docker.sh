#!/bin/bash

# 三角机构TRPG单人引擎 - Docker部署脚本
# 使用方法: ./deploy-docker.sh [dev|prod]

set -e

ENVIRONMENT=${1:-dev}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=========================================="
echo "TRPG Solo Engine - Docker Deployment"
echo "Environment: $ENVIRONMENT"
echo "=========================================="

cd "$PROJECT_ROOT"

# 检查Docker和Docker Compose
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "Error: Docker Compose is not installed"
    exit 1
fi

# 加载环境变量
if [ -f .env ]; then
    echo "Loading environment variables from .env"
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "Warning: .env file not found, using defaults"
fi

# 设置构建参数
export VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
export BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
export GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "Version: $VERSION"
echo "Build Time: $BUILD_TIME"
echo "Git Commit: $GIT_COMMIT"

# 构建和部署
if [ "$ENVIRONMENT" = "prod" ]; then
    echo "Deploying production environment..."
    docker-compose -f docker-compose.yml -f deployments/docker-compose.prod.yml build --no-cache
    docker-compose -f docker-compose.yml -f deployments/docker-compose.prod.yml up -d
    
    # 启动监控（如果需要）
    read -p "Start monitoring stack? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose --profile monitoring up -d
    fi
else
    echo "Deploying development environment..."
    docker-compose up -d --build
fi

# 等待服务启动
echo "Waiting for services to be ready..."
sleep 10

# 健康检查
echo "Checking service health..."
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:${SERVER_PORT:-8080}/health > /dev/null 2>&1; then
        echo "✓ Backend service is healthy"
        break
    fi
    
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo "Waiting for backend... ($RETRY_COUNT/$MAX_RETRIES)"
    sleep 2
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "✗ Backend service failed to start"
    docker-compose logs backend
    exit 1
fi

# 显示服务状态
echo ""
echo "=========================================="
echo "Deployment completed successfully!"
echo "=========================================="
docker-compose ps

echo ""
echo "Service URLs:"
echo "  Backend API: http://localhost:${SERVER_PORT:-8080}"
echo "  Health Check: http://localhost:${SERVER_PORT:-8080}/health"
echo "  Swagger UI: http://localhost:${SERVER_PORT:-8080}/swagger/index.html"
echo "  PostgreSQL: localhost:${DATABASE_PORT:-5432}"
echo "  Redis: localhost:${REDIS_PORT:-6379}"

if docker-compose ps | grep -q prometheus; then
    echo "  Prometheus: http://localhost:${PROMETHEUS_PORT:-9090}"
fi

if docker-compose ps | grep -q grafana; then
    echo "  Grafana: http://localhost:${GRAFANA_PORT:-3001}"
fi

echo ""
echo "Useful commands:"
echo "  View logs: docker-compose logs -f [service]"
echo "  Stop services: docker-compose down"
echo "  Restart service: docker-compose restart [service]"
echo "  Execute command: docker-compose exec [service] [command]"
echo ""
