#!/bin/bash

echo "🚀 Porta Gateway 完整功能测试"
echo "=================================="

# 测试基础功能
echo ""
echo "1. 测试基础代理功能:"
echo "   请求 1: $(curl -s http://localhost:8080/test | jq -r .service)"
echo "   请求 2: $(curl -s http://localhost:8080/test | jq -r .service)"
echo "   请求 3: $(curl -s http://localhost:8080/test | jq -r .service)"

# 测试负载均衡
echo ""
echo "2. 测试负载均衡 (10 个请求):"
backend1_count=0
backend2_count=0
for i in {1..10}; do
    service=$(curl -s http://localhost:8080/test | jq -r .service)
    if [ "$service" = "backend-1" ]; then
        ((backend1_count++))
    elif [ "$service" = "backend-2" ]; then
        ((backend2_count++))
    fi
    echo "   请求 $i: $service"
done
echo "   统计: backend-1: $backend1_count, backend-2: $backend2_count"

# 测试监控端点
echo ""
echo "3. 测试监控端点:"
echo "   调试端点: $(curl -s http://localhost:8080/__debug/endpoints | jq -r .message)"

# 测试 Prometheus
echo ""
echo "4. 测试 Prometheus:"
if curl -s http://localhost:9090/api/v1/status/config > /dev/null 2>&1; then
    echo "   ✅ Prometheus 运行正常"
else
    echo "   ❌ Prometheus 无法访问"
fi

# 测试 Grafana
echo ""
echo "5. 测试 Grafana:"
if curl -s http://localhost:3000/api/health > /dev/null 2>&1; then
    echo "   ✅ Grafana 运行正常"
    version=$(curl -s http://localhost:3000/api/health | jq -r .version)
    echo "   版本: $version"
else
    echo "   ❌ Grafana 无法访问"
fi

# 测试 Redis
echo ""
echo "6. 测试 Redis:"
if docker exec gateway-redis-1 redis-cli ping > /dev/null 2>&1; then
    echo "   ✅ Redis 运行正常"
else
    echo "   ❌ Redis 无法访问"
fi

# 测试后端服务
echo ""
echo "7. 测试后端服务:"
echo "   Backend 1: $(curl -s http://localhost:8081 | jq -r .service)"
echo "   Backend 2: $(curl -s http://localhost:8082 | jq -r .service)"

# 测试性能
echo ""
echo "8. 测试性能 (100 个并发请求):"
start_time=$(date +%s.%N)
for i in {1..100}; do
    curl -s http://localhost:8080/test > /dev/null &
done
wait
end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc)
echo "   100 个请求耗时: ${duration} 秒"

echo ""
echo "🎉 测试完成！"
echo ""
echo "访问地址:"
echo "   Gateway: http://localhost:8080"
echo "   Prometheus: http://localhost:9090"
echo "   Grafana: http://localhost:3000 (admin/admin)"
echo "   Backend 1: http://localhost:8081"
echo "   Backend 2: http://localhost:8082"
echo "   Redis: localhost:6380" 