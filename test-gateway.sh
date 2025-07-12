#!/bin/bash

echo "=== Porta Gateway Test Script ==="
echo ""

# Test backend services
echo "1. Testing Backend Services:"
echo "   Backend 1 (port 8081):"
curl -s http://localhost:8081 | head -1
echo "   Backend 2 (port 8082):"
curl -s http://localhost:8082 | head -1
echo ""

# Test gateway (if running on host)
echo "2. Testing Gateway (if running on host port 8080):"
if curl -s --connect-timeout 2 http://localhost:8080/test > /dev/null 2>&1; then
    echo "   Gateway is responding on port 8080"
    echo "   Response:"
    curl -s http://localhost:8080/test
else
    echo "   Gateway not responding on port 8080"
fi
echo ""

# Show Docker services status
echo "3. Docker Services Status:"
docker-compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
echo ""

echo "=== Test Complete ==="
echo ""
echo "To start the gateway manually:"
echo "  ./build/porta -c ./examples/etc/config.yaml -p 8080"
echo ""
echo "To access services:"
echo "  Gateway:    http://localhost:8080/test"
echo "  Backend 1:  http://localhost:8081"
echo "  Backend 2:  http://localhost:8082"
echo "  Grafana:    http://localhost:3000"
echo "  Redis:      localhost:6380" 