#!/bin/bash

# 本地启动脚本 - 使用Docker数据库，本地运行应用

echo "=========================================="
echo "启动TRPG Solo Engine（本地模式）"
echo "=========================================="
echo ""

# 检查Docker服务是否运行
echo "1. 检查Docker服务..."
if ! docker-compose ps | grep -q "postgres.*Up"; then
    echo "   启动PostgreSQL和Redis..."
    docker-compose up -d postgres redis
    echo "   等待服务就绪..."
    sleep 5
else
    echo "   ✓ Docker服务已运行"
fi

echo ""
echo "2. 检查数据库..."
# 确保数据库存在
docker-compose exec -T postgres psql -U trpg -d postgres -c "SELECT 1 FROM pg_database WHERE datname='trpg_solo_engine'" | grep -q 1
if [ $? -ne 0 ]; then
    echo "   创建数据库..."
    docker-compose exec -T postgres psql -U trpg -d postgres -c "CREATE DATABASE trpg_solo_engine"
fi
echo "   ✓ 数据库就绪"

echo ""
echo "3. 停止Docker中的后端（如果在运行）..."
docker-compose stop backend 2>/dev/null
echo "   ✓ 已停止"

echo ""
echo "4. 设置环境变量..."
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_USER=trpg
export DATABASE_PASSWORD=trpg_password
export DATABASE_NAME=trpg_solo_engine
export DATABASE_SSLMODE=disable
export REDIS_HOST=localhost
export REDIS_PORT=6379
export SERVER_PORT=8080
echo "   ✓ 环境变量已设置"

echo ""
echo "=========================================="
echo "✓ 准备完成"
echo "=========================================="
echo ""
echo "启动应用服务器..."
echo ""
echo "访问:"
echo "  - API文档: http://localhost:8080/api/docs"
echo "  - 健康检查: http://localhost:8080/health"
echo "  - API版本: http://localhost:8080/api/version"
echo ""
echo "按 Ctrl+C 停止服务器"
echo ""
echo "=========================================="
echo ""

# 启动应用
go run cmd/server/main.go
