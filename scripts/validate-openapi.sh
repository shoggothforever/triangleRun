#!/bin/bash

# 验证OpenAPI规范文件
# 使用在线工具或本地工具验证OpenAPI规范的正确性

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OPENAPI_FILE="$PROJECT_ROOT/api/openapi.yaml"

echo "=========================================="
echo "OpenAPI 规范验证"
echo "=========================================="
echo ""

# 检查文件是否存在
if [ ! -f "$OPENAPI_FILE" ]; then
    echo "❌ 错误: OpenAPI文件不存在: $OPENAPI_FILE"
    exit 1
fi

echo "✓ OpenAPI文件存在: $OPENAPI_FILE"
echo ""

# 检查文件大小
FILE_SIZE=$(wc -c < "$OPENAPI_FILE")
echo "✓ 文件大小: $FILE_SIZE 字节"
echo ""

# 基本YAML语法检查
echo "检查YAML语法..."
if command -v python3 &> /dev/null; then
    if python3 -c "import yaml" 2>/dev/null; then
        python3 -c "import yaml; yaml.safe_load(open('$OPENAPI_FILE'))" 2>&1
        if [ $? -eq 0 ]; then
            echo "✓ YAML语法正确"
        else
            echo "❌ YAML语法错误"
            exit 1
        fi
    else
        echo "⚠ 警告: 未安装PyYAML模块，跳过YAML语法检查"
        echo "   安装方法: pip3 install pyyaml"
    fi
else
    echo "⚠ 警告: 未安装python3，跳过YAML语法检查"
fi
echo ""

# 检查必需的OpenAPI字段
echo "检查OpenAPI必需字段..."
REQUIRED_FIELDS=("openapi" "info" "paths")
for field in "${REQUIRED_FIELDS[@]}"; do
    if grep -q "^$field:" "$OPENAPI_FILE"; then
        echo "✓ 找到字段: $field"
    else
        echo "❌ 缺少必需字段: $field"
        exit 1
    fi
done
echo ""

# 统计API端点数量
echo "统计API端点..."
ENDPOINT_COUNT=$(grep -c "^  /api/" "$OPENAPI_FILE" || true)
echo "✓ 发现 $ENDPOINT_COUNT 个API端点"
echo ""

# 检查是否有示例
echo "检查示例..."
EXAMPLE_COUNT=$(grep -c "example:" "$OPENAPI_FILE" || true)
echo "✓ 发现 $EXAMPLE_COUNT 个示例"
echo ""

# 如果安装了swagger-cli，使用它进行验证
if command -v swagger-cli &> /dev/null; then
    echo "使用swagger-cli进行完整验证..."
    swagger-cli validate "$OPENAPI_FILE"
    echo "✓ swagger-cli验证通过"
elif command -v npx &> /dev/null; then
    echo "使用npx swagger-cli进行完整验证..."
    npx @apidevtools/swagger-cli validate "$OPENAPI_FILE"
    echo "✓ swagger-cli验证通过"
else
    echo "⚠ 警告: 未安装swagger-cli，跳过完整验证"
    echo "   安装方法: npm install -g @apidevtools/swagger-cli"
fi
echo ""

echo "=========================================="
echo "✓ OpenAPI规范验证完成"
echo "=========================================="
echo ""
echo "下一步:"
echo "1. 启动服务器: ./trpg-engine 或 go run cmd/server/main.go"
echo "2. 访问文档: http://localhost:8080/api/docs"
echo "3. 测试API: 使用Postman导入 api/postman-collection.json"
echo ""
