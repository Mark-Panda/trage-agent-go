# Trae Agent 交互式对话使用指南

## 概述

Trae Agent 提供了强大的交互式对话功能，允许用户通过命令行与AI代理进行多轮对话，执行各种软件工程任务。

## 快速开始

### 1. 设置环境变量

在使用交互式对话之前，需要设置相应的API密钥：

```bash
# OpenAI API密钥
export OPENAI_API_KEY="your_openai_api_key_here"

# 豆包API密钥（如果使用豆包）
export DOUBAO_API_KEY="your_doubao_api_key_here"

# 其他提供商...
```

### 2. 启动交互式对话

```bash
# 使用默认配置文件
./trage-cli interactive

# 指定配置文件
./trage-cli interactive --config-file ./trae_config.yaml

# 指定工作目录
./trage-cli interactive --working-dir /path/to/project
```

### 3. 基本命令

交互式模式支持以下命令：

- `help` - 显示帮助信息
- `status` - 显示代理状态和配置
- `clear` - 清屏
- `exit` 或 `quit` - 退出会话
- 输入任何任务描述来执行任务

## 配置说明

### 模型配置

在 `trae_config.yaml` 中配置不同的模型：

```yaml
models:
  trae_agent_model:
    model_provider: openai
    model: gpt-4o
    max_tokens: 4096
    temperature: 0.5
    supports_tool_calling: true
```

### 工具配置

代理可以使用以下工具：

- `bash` - 执行bash命令
- `edit_file` - 编辑文件
- `sequential_thinking` - 结构化思考
- `task_done` - 标记任务完成

## 使用示例

### 示例1：文件操作

```
trae-agent> 创建一个新的README文件
🚀 执行任务: 创建一个新的README文件
✅ 任务执行成功！
```

### 示例2：代码分析

```
trae-agent> 分析这个Go项目的结构
🚀 执行任务: 分析这个Go项目的结构
✅ 任务执行成功！
```

### 示例3：问题诊断

```
trae-agent> 为什么我的程序编译失败？
🚀 执行任务: 为什么我的程序编译失败？
✅ 任务执行成功！
```

## 故障排除

### 常见问题

1. **API密钥错误**
   - 检查环境变量是否正确设置
   - 验证API密钥是否有效

2. **模型不存在**
   - 检查模型名称是否正确
   - 确认模型提供商配置

3. **工具执行失败**
   - 检查工作目录权限
   - 验证工具配置

### 调试模式

使用 `--console-type rich` 启用详细输出：

```bash
./trage-cli interactive --console-type rich
```

## 高级功能

### 轨迹记录

启用轨迹记录来保存所有操作：

```bash
./trage-cli interactive --trajectory-file ./trajectory.json
```

### 自定义工具

可以注册自定义工具来扩展代理能力。

### MCP服务器

支持Model Context Protocol服务器集成。

## 最佳实践

1. **明确任务描述**：提供清晰、具体的任务描述
2. **分步执行**：复杂任务可以分解为多个步骤
3. **验证结果**：检查工具执行结果
4. **使用help命令**：遇到问题时查看帮助信息

## 技术支持

如果遇到问题，请：

1. 检查配置文件语法
2. 验证环境变量设置
3. 查看错误日志
4. 提交issue到项目仓库
