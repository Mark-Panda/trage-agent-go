#!/bin/bash

# Trae Agent 交互式对话演示脚本

echo "🚀 Trae Agent 交互式对话演示"
echo "================================"
echo ""

# 检查可执行文件是否存在
if [ ! -f "../trage-cli" ]; then
    echo "❌ 错误: trage-cli 可执行文件不存在"
    echo "请先运行: go build -o trage-cli ./cmd/trage-cli/"
    exit 1
fi

# 检查配置文件是否存在
if [ ! -f "../trae_config.yaml" ]; then
    echo "❌ 错误: 配置文件 trae_config.yaml 不存在"
    exit 1
fi

# 检查环境变量
if [ -z "$OPENAI_API_KEY" ]; then
    echo "⚠️  警告: OPENAI_API_KEY 环境变量未设置"
    echo "请设置: export OPENAI_API_KEY='your_api_key_here'"
    echo ""
    echo "演示将使用默认配置，但可能无法正常工作"
    echo ""
fi

echo "✅ 环境检查完成"
echo ""

echo "📋 可用命令:"
echo "• help - 显示帮助信息"
echo "• status - 显示代理状态"
echo "• clear - 清屏"
echo "• exit/quit - 退出会话"
echo "• 输入任何任务描述来执行任务"
echo ""

echo "🎯 演示任务建议:"
echo "• '你会什么' - 了解代理能力"
echo "• '分析这个项目' - 项目结构分析"
echo "• '创建一个测试文件' - 文件操作演示"
echo "• '执行 ls -la' - 命令执行演示"
echo ""

echo "🚀 启动交互式对话..."
echo ""

# 启动交互式对话
cd .. && ./trage-cli interactive --config-file ./trae_config.yaml
