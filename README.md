# Trae Agent Go

一个基于Go语言实现的智能代理系统，支持多种LLM提供商，具备重试机制、缓存系统和完整的监控能力。

## 🌟 主要特性

### 1. **多LLM提供商支持**
- **OpenAI**: 使用官方Go包，完整的API集成
- **豆包**: 使用OpenAI兼容的API格式
- 支持工具调用和完整的API功能

### 2. **智能重试机制**
- 指数退避算法，避免API过载
- 智能错误分类，只重试可恢复的错误
- 可配置的重试策略和回调
- 支持上下文超时和取消

### 3. **高性能缓存系统**
- 内存缓存，支持TTL过期
- LRU清理策略，防止内存泄漏
- 智能缓存键生成，确保缓存命中率
- 实时统计和监控

### 4. **结构化日志系统**
- 多级别日志支持（DEBUG, INFO, WARN, ERROR, FATAL）
- 结构化字段支持
- 环境变量配置
- 全局日志记录器

### 5. **性能监控系统**
- Prometheus兼容的指标格式
- 计数器、仪表、直方图指标类型
- API调用统计和延迟监控
- 缓存命中率统计

### 6. **容器化部署**
- 多阶段Docker构建
- Docker Compose完整服务栈
- 健康检查和自动重启
- 生产就绪的配置

## 🚀 快速开始

### 前置要求
- Go 1.21+
- Docker & Docker Compose (可选)

### 安装依赖
```bash
go mod download
go mod tidy
```

### 构建
```bash
make build
```

### 运行
```bash
# 显示帮助
./build/trage-cli --help

# 显示配置
./build/trage-cli show-config --config-file trae_config.yaml

# 执行任务
./build/trage-cli run "Hello World" --config-file trae_config.yaml

# 交互模式
./build/trage-cli interactive --config-file trae_config.yaml
```

## 🐳 Docker部署

### 快速部署
```bash
# 启动完整服务栈
make deploy

# 查看服务状态
make deploy-status

# 查看日志
make deploy-logs

# 停止服务
make deploy-stop

# 清理资源
make deploy-cleanup
```

### 手动部署
```bash
# 构建镜像
docker build -t trage-agent:latest .

# 运行容器
docker run -d --name trage-agent -p 8080:8080 trage-agent:latest

# 查看日志
docker logs -f trage-agent
```

## ⚙️ 配置

### 基本配置
```yaml
agents:
  trae_agent:
    enable_lakeview: true
    model: gpt4_model
    max_steps: 200
    tools:
      - bash
      - edit_file
      - sequential_thinking
      - task_done

model_providers:
  openai:
    api_key: "your_openai_api_key"
    provider: "openai"
    base_url: "https://api.openai.com"
    api_version: "v1"
  
  doubao:
    api_key: "your_doubao_api_key"
    provider: "doubao"
    base_url: "https://api.doubao.com"
    api_version: "v1"

models:
  gpt4_model:
    model_provider: openai
    model: "gpt-4"
    max_tokens: 4096
    temperature: 0.5
    supports_tool_calling: true
```

### 环境变量
```bash
export LOG_LEVEL=DEBUG
export OPENAI_API_KEY="your_api_key"
export DOUBAO_API_KEY="your_api_key"
```

## 🔧 开发

### 运行测试
```bash
# 运行所有测试
make test

# 运行特定包测试
go test ./pkg/llm -v
go test ./pkg/utils -v

# 生成测试覆盖率报告
make test-coverage
```

### 代码质量
```bash
# 格式化代码
make format

# 代码检查
make lint

# 运行go vet
make vet
```

### 完整构建流程
```bash
# 完整构建流程
make all

# 生产构建
make prod
```

## 📊 监控

### 指标端点
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

### 关键指标
- `api_calls_total`: API调用总数
- `api_latency_seconds`: API调用延迟
- `cache_hit_rate`: 缓存命中率
- `retry_attempts_total`: 重试次数
- `errors_total`: 错误总数

## 🏗️ 架构

```
trage-agent-go/
├── cmd/trage-cli/          # 命令行入口
├── pkg/
│   ├── agent/              # 代理系统实现
│   ├── config/             # 配置管理
│   ├── llm/                # LLM客户端
│   │   ├── openai_client.go    # OpenAI客户端
│   │   ├── doubao_client.go    # 豆包客户端
│   │   ├── retry_wrapper.go    # 重试包装器
│   │   └── cache.go            # 缓存系统
│   ├── tools/              # 工具系统
│   └── utils/              # 工具函数
│       ├── logger.go            # 日志系统
│       └── metrics.go           # 监控系统
├── scripts/                 # 部署脚本
├── Dockerfile              # Docker构建文件
├── docker-compose.yml      # Docker Compose配置
├── Makefile                # 构建和部署脚本
└── README.md               # 项目文档
```

## 🔄 使用示例

### 基本使用
```go
import (
    "trage-agent-go/pkg/llm"
    "trage-agent-go/pkg/utils"
)

// 创建日志记录器
logger := utils.NewLogger(utils.LogLevelInfo)

// 创建性能监控器
monitor := utils.NewPerformanceMonitor(logger)

// 创建OpenAI客户端
openaiClient := llm.NewOpenAIClient(apiKey, baseURL, apiVersion)

// 添加重试机制
retryableClient := llm.NewRetryableLLMClient(openaiClient, retryConfig)

// 添加缓存
cachedClient := llm.NewCachedLLMClient(retryableClient, cache)

// 使用客户端
messages := []llm.LLMMessage{{Role: "user", Content: "Hello"}}
response, err := cachedClient.Chat(messages, tools, config)
```

### 日志记录
```go
// 使用全局日志记录器
utils.Info("Application started", utils.F("version", "1.0.0"))

// 创建带字段的日志记录器
logger := utils.NewLogger(utils.LogLevelDebug)
logger.WithFields(utils.F("user_id", "123")).Info("User logged in")
```

### 性能监控
```go
// 记录API调用
monitor.RecordAPICall("openai", duration, success)

// 记录缓存命中
monitor.RecordCacheHit(true)

// 导出指标
metrics := monitor.ExportMetrics()
```

## 🚧 待实现功能

- [ ] 持久化缓存（Redis支持）
- [ ] 更多LLM提供商
- [ ] 流式响应支持
- [ ] 高级负载均衡
- [ ] 限流和熔断机制

## 🤝 贡献

欢迎提交Issue和Pull Request！

### 开发流程
1. Fork项目
2. 创建功能分支
3. 提交更改
4. 运行测试
5. 提交Pull Request

## 📄 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [OpenAI Go](https://github.com/sashabaranov/go-openai) - OpenAI官方Go包
- [Cobra](https://github.com/spf13/cobra) - 强大的CLI框架
- [Prometheus](https://prometheus.io/) - 监控系统
- [Grafana](https://grafana.com/) - 可视化平台
