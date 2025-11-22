#!/bin/bash

# 修复数据库问题
# 此脚本会进入PostgreSQL容器并创建缺失的数据库

echo "=========================================="
echo "修复数据库配置"
echo "=========================================="
echo ""

# 检查Docker Compose是否运行
if ! docker-compose ps | grep -q "postgres"; then
    echo "❌ PostgreSQL容器未运行"
    echo "请先启动服务: docker-compose up -d"
    exit 1
fi

echo "✓ PostgreSQL容器正在运行"
echo ""

# 创建数据库
echo "创建数据库..."
docker-compose exec -T postgres psql -U trpg -d postgres <<EOF
-- 创建trpg数据库（如果应用尝试连接这个名字）
CREATE DATABASE trpg;

-- 创建trpg_solo_engine数据库（配置文件中的名字）
CREATE DATABASE trpg_solo_engine;

-- 列出所有数据库
\l
EOF

echo ""
echo "✓ 数据库创建完成"
echo ""

# 重启后端服务以重新连接
echo "重启后端服务..."
docker-compose restart backend

echo ""
echo "=========================================="
echo "✓ 修复完成"
echo "=========================================="
echo ""
echo "查看日志:"
echo "  docker-compose logs -f backend"
echo ""
echo "测试连接:"
echo "  curl http://localhost:8080/health"
echo ""
