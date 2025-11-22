#!/bin/bash

# 三角机构TRPG单人引擎 - 配置验证脚本

set -e

echo "🔍 验证配置文件..."

# 检查配置文件是否存在
echo "📁 检查配置文件..."
if [ ! -f "configs/config.yaml" ]; then
    echo "❌ 错误: configs/config.yaml 不存在"
    exit 1
fi
echo "✅ config.yaml 存在"

if [ ! -f ".env.example" ]; then
    echo "❌ 错误: .env.example 不存在"
    exit 1
fi
echo "✅ .env.example 存在"

# 检查ARC配置文件
echo ""
echo "📁 检查ARC配置文件..."
for file in anomalies.json realities.json careers.json; do
    if [ ! -f "configs/$file" ]; then
        echo "❌ 错误: configs/$file 不存在"
        exit 1
    fi
    echo "✅ $file 存在"
done

# 验证JSON格式
echo ""
echo "🔍 验证JSON格式..."
for file in anomalies.json realities.json careers.json; do
    if command -v jq &> /dev/null; then
        if jq empty "configs/$file" 2>/dev/null; then
            echo "✅ $file JSON格式正确"
        else
            echo "❌ 错误: $file JSON格式错误"
            exit 1
        fi
    else
        echo "⚠️  警告: jq未安装，跳过JSON验证"
        break
    fi
done

# 验证YAML格式（如果安装了yamllint）
echo ""
echo "🔍 验证YAML格式..."
if command -v yamllint &> /dev/null; then
    if yamllint configs/config.yaml; then
        echo "✅ config.yaml YAML格式正确"
    else
        echo "❌ 错误: config.yaml YAML格式错误"
        exit 1
    fi
else
    echo "⚠️  警告: yamllint未安装，跳过YAML验证"
fi

# 检查必需的环境变量文档
echo ""
echo "📋 检查环境变量文档..."
required_vars=(
    "DATABASE_PASSWORD"
    "AI_API_KEY"
    "JWT_SECRET"
)

for var in "${required_vars[@]}"; do
    if grep -q "$var" .env.example; then
        echo "✅ $var 已在 .env.example 中文档化"
    else
        echo "❌ 错误: $var 未在 .env.example 中找到"
        exit 1
    fi
done

# 检查配置文档
echo ""
echo "📚 检查配置文档..."
if [ -f "configs/CONFIG_GUIDE.md" ]; then
    echo "✅ CONFIG_GUIDE.md 存在"
else
    echo "⚠️  警告: CONFIG_GUIDE.md 不存在"
fi

if [ -f "configs/README.md" ]; then
    echo "✅ README.md 存在"
else
    echo "⚠️  警告: README.md 不存在"
fi

echo ""
echo "✅ 所有配置文件验证通过！"
echo ""
echo "📝 下一步："
echo "1. 复制 .env.example 为 .env"
echo "2. 填写 .env 中的敏感信息"
echo "3. 根据需要调整 configs/config.yaml"
echo "4. 运行应用: go run cmd/server/main.go"
