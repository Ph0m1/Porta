# Porta Gateway - 安全性和监控指南

本文档详细介绍了 Porta Gateway 的安全性和监控功能。

## 🛡️ 安全功能

### 1. 认证和授权

#### JWT 认证
```yaml
security:
  auth:
    jwt_secret: "your-super-secret-jwt-key"
    jwt_expiration: "24h"
    required_roles:
      "/admin/*": ["admin"]
      "/api/users": ["user", "admin"]
```

#### API Key 认证
```yaml
security:
  auth:
    api_keys:
      "api-key-123": "client-1"
      "api-key-456": "client-2"
```

使用方式：
```bash
# 在请求头中添加 API Key
curl -H "X-API-Key: api-key-123" http://localhost:8080/api/users

# 或在查询参数中
curl "http://localhost:8080/api/users?api_key=api-key-123"
```

#### Basic 认证
```yaml
security:
  auth:
    basic_auth:
      "admin": "admin123"
      "user": "user123"
```

#### 请求签名认证
```yaml
security:
  signature_auth:
    enabled: true
    secrets:
      "client-1": "secret-for-client-1"
```

### 2. 限流 (Rate Limiting)

#### 全局限流
```yaml
security:
  rate_limit:
    requests_per_second: 100
    burst_size: 200
    window_size: "1m"
```

#### 端点级限流
```yaml
security:
  rate_limit:
    endpoints:
      "/api/auth/login":
        requests_per_second: 5
        burst_size: 10
```

#### 限流策略
- **Token Bucket**: 允许突发流量，适合大多数场景
- **Sliding Window**: 精确控制，适合严格限流场景

### 3. CORS 配置

```yaml
security:
  cors:
    allowed_origins:
      - "https://yourdomain.com"
      - "https://*.yourdomain.com"
    allowed_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
    allowed_headers:
      - "Content-Type"
      - "Authorization"
      - "X-API-Key"
    allow_credentials: true
```

### 4. 安全头

```yaml
security:
  security_headers:
    content_type_nosniff: true
    frame_deny: true
    browser_xss_filter: true
    content_security_policy: "default-src 'self'"
    referrer_policy: "strict-origin-when-cross-origin"
    hsts_max_age: 31536000
```

### 5. IP 白名单

```yaml
security:
  ip_whitelist:
    enabled: true
    allowed_ips:
      - "127.0.0.1"
      - "10.0.0.0/8"
      - "192.168.0.0/16"
```

## 📊 监控功能

### 1. Prometheus 指标

#### 请求指标
- `porta_requests_total`: 总请求数
- `porta_request_duration_seconds`: 请求延迟
- `porta_requests_in_flight`: 正在处理的请求数
- `porta_request_size_bytes`: 请求大小
- `porta_response_size_bytes`: 响应大小

#### 后端指标
- `porta_backend_requests_total`: 后端请求总数
- `porta_backend_request_duration_seconds`: 后端请求延迟
- `porta_backend_errors_total`: 后端错误数

#### 系统指标
- `porta_goroutines_count`: Goroutine 数量
- `porta_memory_usage_bytes`: 内存使用量
- `porta_cpu_usage_percent`: CPU 使用率

#### 安全指标
- `porta_rate_limit_hits_total`: 限流命中数
- `porta_rate_limit_blocks_total`: 限流阻止数

### 2. 健康检查

#### 端点
- `/__health`: 详细健康状态
- `/__ready`: 就绪检查
- `/__live`: 存活检查

#### 健康检查项
- 内存使用量检查
- Goroutine 数量检查
- 后端连通性检查

#### 响应示例
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "uptime": "2h30m15s",
  "checks": {
    "memory": {
      "status": "healthy",
      "message": "Memory usage: 256 MB"
    },
    "backend_api_users_0": {
      "status": "healthy",
      "message": "Backend http://backend:8080 is healthy"
    }
  },
  "system_info": {
    "goroutines": 45,
    "memory_alloc_bytes": 268435456,
    "cpu_count": 4,
    "go_version": "go1.24"
  }
}
```

### 3. 告警规则

#### 高错误率告警
```yaml
- alert: HighErrorRate
  expr: rate(porta_requests_total{status_code=~"5.."}[5m]) / rate(porta_requests_total[5m]) > 0.1
  for: 2m
  labels:
    severity: critical
```

#### 高响应时间告警
```yaml
- alert: HighResponseTime
  expr: histogram_quantile(0.95, rate(porta_request_duration_seconds_bucket[5m])) > 1
  for: 5m
  labels:
    severity: warning
```

## 🚀 部署和使用

### 1. 基本部署

```bash
# 构建镜像
make docker-build

# 启动完整环境（包括监控）
make docker-run

# 查看服务状态
docker-compose ps
```

### 2. 监控访问

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Gateway Metrics**: http://localhost:8080/metrics
- **Health Check**: http://localhost:8080/__health

### 3. 安全配置示例

```bash
# 启动带安全配置的网关
./porta -c config.yaml -s security.yaml

# 测试 API Key 认证
curl -H "X-API-Key: api-key-123" http://localhost:8080/api/users

# 测试 JWT 认证
curl -H "Authorization: Bearer your-jwt-token" http://localhost:8080/api/users

# 测试限流
for i in {1..150}; do curl http://localhost:8080/api/test; done
```

## 📈 性能调优

### 1. 限流配置

```yaml
# 高流量场景
rate_limit:
  requests_per_second: 1000
  burst_size: 2000
  window_size: "1m"

# 低延迟场景
rate_limit:
  requests_per_second: 100
  burst_size: 150
  window_size: "10s"
```

### 2. 监控优化

```yaml
# 减少监控开销
monitoring:
  metrics:
    enabled: true
    sample_rate: 0.1  # 只采样 10% 的请求
  
  health:
    check_interval: "60s"  # 降低检查频率
```

### 3. 安全性平衡

```yaml
# 生产环境推荐配置
security:
  auth:
    jwt_expiration: "1h"  # 较短的过期时间
  
  rate_limit:
    requests_per_second: 500
    burst_size: 1000
  
  security_headers:
    content_security_policy: "default-src 'self'; script-src 'self'"
```

## 🔧 故障排除

### 1. 常见问题

#### 认证失败
```bash
# 检查 JWT 密钥配置
curl -H "Authorization: Bearer invalid-token" http://localhost:8080/api/test
# 响应: 401 Unauthorized: invalid JWT token
```

#### 限流触发
```bash
# 检查限流状态
curl -I http://localhost:8080/api/test
# 响应头: X-RateLimit-Remaining: 0
```

#### 后端不可达
```bash
# 检查健康状态
curl http://localhost:8080/__health
# 查看 backend 检查状态
```

### 2. 日志分析

```bash
# 查看网关日志
docker-compose logs -f porta-gateway

# 查看 Prometheus 日志
docker-compose logs prometheus

# 查看特定错误
docker-compose logs porta-gateway | grep ERROR
```

### 3. 监控调试

```bash
# 检查指标端点
curl http://localhost:8080/metrics

# 查看特定指标
curl http://localhost:8080/metrics | grep porta_requests_total

# 检查 Prometheus 目标
curl http://localhost:9090/api/v1/targets
```

## 📚 最佳实践

### 1. 安全配置

- 使用强密码和密钥
- 定期轮换 API 密钥
- 启用 HTTPS
- 配置适当的 CORS 策略
- 使用最小权限原则

### 2. 监控配置

- 设置合适的告警阈值
- 监控关键业务指标
- 定期检查健康状态
- 保留足够的历史数据

### 3. 性能优化

- 根据业务需求调整限流参数
- 优化后端连接池
- 使用缓存减少后端压力
- 定期清理过期数据

## 🔗 相关链接

- [Prometheus 文档](https://prometheus.io/docs/)
- [Grafana 文档](https://grafana.com/docs/)
- [JWT 标准](https://jwt.io/)
- [CORS 规范](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) 