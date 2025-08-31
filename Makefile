# Trae Agent Makefile
.PHONY: help build test clean docker-build docker-run docker-stop docker-logs docker-clean deploy

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
BINARY_NAME := trage-cli
BUILD_DIR := build
DOCKER_IMAGE := trage-agent
DOCKER_TAG := latest

help: ## 显示帮助信息
	@echo "Trae Agent 构建和部署工具"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## 构建Go二进制文件
	@echo "构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/trage-cli
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

test: ## 运行测试
	@echo "运行测试..."
	@go test -v ./...
	@echo "测试完成"

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率报告..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告生成完成: coverage.html"

clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "清理完成"

docker-build: ## 构建Docker镜像
	@echo "构建Docker镜像: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker镜像构建完成"

docker-run: ## 运行Docker容器
	@echo "运行Docker容器..."
	@docker run -d --name trage-agent-container -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "容器启动完成，访问 http://localhost:8080"

docker-stop: ## 停止Docker容器
	@echo "停止Docker容器..."
	@docker stop trage-agent-container || true
	@docker rm trage-agent-container || true
	@echo "容器已停止并删除"

docker-logs: ## 查看Docker容器日志
	@docker logs -f trage-agent-container

docker-clean: ## 清理Docker资源
	@echo "清理Docker资源..."
	@docker stop trage-agent-container || true
	@docker rm trage-agent-container || true
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true
	@echo "Docker资源清理完成"

deploy: ## 使用docker-compose部署完整服务栈
	@echo "部署完整服务栈..."
	@./scripts/deploy.sh start

deploy-stop: ## 停止docker-compose服务
	@echo "停止服务..."
	@./scripts/deploy.sh stop

deploy-restart: ## 重启docker-compose服务
	@echo "重启服务..."
	@./scripts/deploy.sh restart

deploy-logs: ## 查看服务日志
	@./scripts/deploy.sh logs

deploy-status: ## 查看服务状态
	@./scripts/deploy.sh status

deploy-cleanup: ## 清理所有部署资源
	@echo "清理部署资源..."
	@./scripts/deploy.sh cleanup

install-deps: ## 安装Go依赖
	@echo "安装Go依赖..."
	@go mod download
	@go mod tidy
	@echo "依赖安装完成"

lint: ## 运行代码检查
	@echo "运行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

format: ## 格式化代码
	@echo "格式化代码..."
	@go fmt ./...
	@echo "代码格式化完成"

vet: ## 运行go vet
	@echo "运行go vet..."
	@go vet ./...
	@echo "go vet完成"

# 开发相关命令
dev: build ## 开发模式：构建并运行
	@echo "启动开发模式..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --help

# 生产构建
prod-build: ## 生产环境构建
	@echo "生产环境构建..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/trage-cli
	@echo "生产构建完成"

# 快速测试构建
quick-test: ## 快速测试构建
	@echo "快速测试构建..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/trage-cli
	@echo "快速测试构建完成"

# 显示版本信息
version: ## 显示版本信息
	@echo "Trae Agent Go Version"
	@echo "Go Version: $(shell go version)"
	@echo "Git Commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Build Time: $(shell date)"

# 检查依赖更新
deps-check: ## 检查依赖更新
	@echo "检查依赖更新..."
	@go list -u -m all
	@echo "依赖检查完成"

# 更新依赖
deps-update: ## 更新依赖
	@echo "更新依赖..."
	@go get -u ./...
	@go mod tidy
	@echo "依赖更新完成"

# 生成文档
docs: ## 生成文档
	@echo "生成文档..."
	@mkdir -p docs
	@echo "文档生成完成（待实现）"

# 性能测试
bench: ## 运行性能测试
	@echo "运行性能测试..."
	@go test -bench=. ./...
	@echo "性能测试完成"

# 安全扫描
security-scan: ## 安全扫描
	@echo "安全扫描..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found, skipping..."; \
	fi

# 完整构建流程
all: clean install-deps lint test build ## 完整构建流程：清理、安装依赖、检查、测试、构建
	@echo "完整构建流程完成"

# 生产部署流程
prod: prod-build docker-build ## 生产部署流程：生产构建、Docker构建
	@echo "生产部署流程完成"
