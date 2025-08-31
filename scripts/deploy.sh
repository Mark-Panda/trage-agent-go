#!/bin/bash

# Trae Agent 部署脚本
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    log_success "Docker and Docker Compose are available"
}

# 检查配置文件
check_config() {
    if [ ! -f "trae_config.yaml" ]; then
        log_warning "trae_config.yaml not found. Creating from example..."
        if [ -f "trae_config.yaml.example" ]; then
            cp trae_config.yaml.example trae_config.yaml
            log_info "Please update trae_config.yaml with your API keys"
        else
            log_error "No configuration file found. Please create trae_config.yaml"
            exit 1
        fi
    fi
}

# 创建必要的目录
create_directories() {
    log_info "Creating necessary directories..."
    mkdir -p logs cache monitoring/grafana/dashboards monitoring/grafana/datasources
    log_success "Directories created"
}

# 创建监控配置
create_monitoring_config() {
    log_info "Creating monitoring configuration..."
    
    # Prometheus配置
    cat > monitoring/prometheus.yml << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  # - "first_rules.yml"
  # - "second_rules.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'trage-agent'
    static_configs:
      - targets: ['trage-agent:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
EOF

    # Grafana数据源配置
    cat > monitoring/grafana/datasources/prometheus.yml << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
EOF

    log_success "Monitoring configuration created"
}

# 构建镜像
build_image() {
    log_info "Building Trae Agent Docker image..."
    docker build -t trage-agent:latest .
    log_success "Image built successfully"
}

# 启动服务
start_services() {
    log_info "Starting services..."
    docker-compose up -d
    
    # 等待服务启动
    log_info "Waiting for services to start..."
    sleep 10
    
    # 检查服务状态
    docker-compose ps
    log_success "Services started successfully"
}

# 显示访问信息
show_access_info() {
    log_info "Access Information:"
    echo "  - Trae Agent CLI: docker exec -it trage-agent ./trage-cli --help"
    echo "  - Prometheus: http://localhost:9090"
    echo "  - Grafana: http://localhost:3000 (admin/admin)"
    echo "  - Redis: localhost:6379"
    echo ""
    log_info "To view logs: docker-compose logs -f trage-agent"
}

# 停止服务
stop_services() {
    log_info "Stopping services..."
    docker-compose down
    log_success "Services stopped"
}

# 清理资源
cleanup() {
    log_info "Cleaning up resources..."
    docker-compose down -v --remove-orphans
    docker rmi trage-agent:latest 2>/dev/null || true
    log_success "Cleanup completed"
}

# 主函数
main() {
    case "${1:-start}" in
        "start")
            log_info "Starting Trae Agent deployment..."
            check_docker
            check_config
            create_directories
            create_monitoring_config
            build_image
            start_services
            show_access_info
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            stop_services
            start_services
            show_access_info
            ;;
        "cleanup")
            cleanup
            ;;
        "logs")
            docker-compose logs -f trage-agent
            ;;
        "status")
            docker-compose ps
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [command]"
            echo ""
            echo "Commands:"
            echo "  start     Start all services (default)"
            echo "  stop      Stop all services"
            echo "  restart   Restart all services"
            echo "  cleanup   Stop and remove all containers and volumes"
            echo "  logs      Show Trae Agent logs"
            echo "  status    Show service status"
            echo "  help      Show this help message"
            ;;
        *)
            log_error "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
