# 多阶段构建Dockerfile
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的系统依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o trage-cli ./cmd/trage-cli

# 第二阶段：运行时镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S trae && \
    adduser -u 1001 -S trae -G trae

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/trage-cli .

# 复制配置文件
COPY --from=builder /app/trae_config.yaml .

# 创建必要的目录
RUN mkdir -p /app/logs /app/cache && \
    chown -R trae:trae /app

# 切换到非root用户
USER trae

# 暴露端口（如果需要HTTP服务）
EXPOSE 8080

# 设置环境变量
ENV LOG_LEVEL=INFO
ENV CONFIG_FILE=/app/trae_config.yaml

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ./trage-cli --help || exit 1

# 设置入口点
ENTRYPOINT ["./trage-cli"]

# 默认命令
CMD ["--help"]
