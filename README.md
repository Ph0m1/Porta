# 🚀 Porta Gateway

一个功能完整的 Go 语言 API 网关，具备监控、负载均衡、缓存和安全性功能。

## ✨ 特性

- 🔄 **智能负载均衡** - 支持轮询和随机负载均衡算法
- 📊 **完整监控系统** - 集成 Prometheus 和 Grafana
- 🚀 **高性能代理** - 支持并发请求和缓存
- 🐳 **容器化部署** - Docker 和 Docker Compose 支持
- 🔒 **安全中间件** - 请求验证和错误处理
- 📝 **结构化日志** - 详细的请求和错误日志
- ⚡ **缓存支持** - Redis 集成和内存缓存
- 🎯 **灵活配置** - 支持 YAML、JSON、TOML 等格式

## 🏗️ 架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client        │    │   Porta Gateway │    │   Backend       │
│                 │───▶│                 │───▶│   Services      │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   Monitoring    │
                       │                 │
                       │ • Prometheus    │
                       │ • Grafana       │
                       │ • Redis         │
                       └─────────────────┘
```

## 🚀 快速开始

### 前置要求

- Go 1.21+
- Docker & Docker Compose
- Make (可选)

### 本地开发

```bash
# 克隆项目
git clone <repository-url>
cd gateway

# 安装依赖
make deps

# 构建项目
make build

# 启动后端服务
docker-compose up -d backend-service-1 backend-service-2

# 运行网关 (使用本地配置)
./build/porta -c ./examples/etc/config.yaml -p 9090
```

### Docker 部署

```bash
# 启动完整环境
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f porta-gateway
```

## 📊 监控和可视化

### Prometheus 指标

网关自动暴露以下指标：

- `porta_requests_total` - 总请求数
- `porta_request_duration_seconds` - 请求耗时
- `porta_memory_usage_bytes` - 内存使用
- `porta_goroutines_count` - Goroutine 数量
- `porta_rate_limit_blocks_total` - 限流阻止数
- `porta_circuit_breaker_state` - 熔断器状态

### Grafana 仪表板

访问 http://localhost:3000 (admin/admin) 查看预配置的仪表板：

- 请求量和响应时间
- 错误率和成功率
- 后端服务健康状态
- 系统资源使用情况

## 🔧 配置

### 基础配置

```yaml
version: 1
name: "Porta Gateway"
port: 8080
timeout: 10
cache_ttl: 3600

host:
  - "http://backend-service-1:80"
  - "http://backend-service-2:80"

endpoints:
  - endpoint: "/test"
    method: "GET"
    concurrent_calls: 1
    timeout: 1000
    cache_ttl: 3600
    backend:
      - host:
          - "http://backend-service-1:80"
          - "http://backend-service-2:80"
        url_pattern: "/"
        encoding: "json"
```

### 环境变量

| 变量 | 描述 | 默认值 |
|------|------|--------|
| `GO111MODULE` | Go 模块模式 | `on` |
| `http_proxy` | HTTP 代理 | 无 |
| `https_proxy` | HTTPS 代理 | 无 |

## 🛠️ 开发

### 项目结构

```
gateway/
├── config/          # 配置管理
├── encoding/        # 编码解码器
├── examples/        # 示例和配置
├── logging/         # 日志系统
├── proxy/           # 代理核心
├── router/          # 路由框架
├── sd/              # 服务发现
├── docker/          # Docker 配置
├── Dockerfile       # 容器构建
├── docker-compose.yml # 服务编排
├── Makefile         # 构建脚本
└── README.md        # 项目文档
```

### 构建命令

```bash
# 下载依赖
make deps

# 构建项目
make build

# 运行测试
make test

# 清理构建
make clean

# Docker 构建
make docker-build

# 启动服务
make docker-run

# 停止服务
make docker-stop
```

### 添加新功能

1. **新中间件**: 在 `proxy/` 目录下创建新的中间件
2. **新编码器**: 在 `encoding/` 目录下实现新的编码器
3. **新路由**: 在 `router/` 目录下扩展路由功能

## 📡 API 端点

### 网关端点

- `GET /test` - 测试端点，返回后端服务响应
- `GET /__debug/endpoints` - 调试端点列表

### 监控端点

- `GET /__health` - 健康检查
- `GET /__live` - 存活检查
- `GET /__ready` - 就绪检查
- `GET /metrics` - Prometheus 指标

## 🔍 故障排除

### 常见问题

1. **"Bad Host" 错误**
   - 检查配置文件中的后端地址
   - 确认 Docker 网络配置
   - 清除代理环境变量

2. **代理连接失败**
   - 检查后端服务状态
   - 验证网络连通性
   - 查看容器日志

3. **Prometheus 配置错误**
   - 检查 YAML 语法
   - 避免重复字段定义
   - 重启 Prometheus 容器

### 调试命令

```bash
# 查看网关日志
docker-compose logs porta-gateway

# 检查网络连通性
docker exec gateway-porta-gateway-1 wget -qO- http://backend-service-1:80

# 测试负载均衡
for i in {1..10}; do curl -s http://localhost:8080/test | jq .service; done

# 运行完整测试
./test_complete.sh
```

## 📈 性能测试

### 基准测试结果

- **并发请求**: 100 个请求耗时 ~0.1 秒
- **响应时间**: 平均 2-3ms
- **吞吐量**: 支持高并发负载
- **内存使用**: 优化的内存管理

### 负载测试

```bash
# 使用 Apache Bench
ab -n 1000 -c 100 http://localhost:8080/test

# 使用 wrk
wrk -t12 -c400 -d30s http://localhost:8080/test
```

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Prometheus](https://prometheus.io/) - 监控系统
- [Grafana](https://grafana.com/) - 数据可视化

---

**Porta Gateway** - 让 API 网关更简单、更强大！ 🚀

