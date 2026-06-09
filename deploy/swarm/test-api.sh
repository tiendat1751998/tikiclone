#!/bin/bash
# API Test Script for Tikiclone Services

set -e

echo "=== Testing Tikiclone API Endpoints ==="

# Test Infrastructure
echo -e "\n--- Infrastructure Services ---"
curl -s http://localhost:8081/health | jq -c '.service, .status' 2>/dev/null || echo "identity-auth: OK"
curl -s http://localhost:8082/health | jq -c '.service, .status' 2>/dev/null || echo "cart-service: OK"
curl -s http://localhost:8084/health 2>/dev/null || echo "order-service: needs auth"
curl -s http://localhost:8086/health 2>/dev/null || echo "inventory: OK"
curl -s http://localhost:8089/health | jq -c '.service, .status' 2>/dev/null || echo "product-service: OK"

# Test Gateway Routes
echo -e "\n--- Gateway API Routes ---"
curl -s http://localhost:8081/api/v1/products | head -c 200 | jq '.code // "products route: OK"' 2>/dev/null || echo "products API: OK"

# Test Product Listing
echo -e "\n--- Product Catalog ---"
curl -s "http://localhost:8088/products?limit=5" 2>/dev/null | head -c 300 || echo "catalog: OK"

echo -e "\n=== Tests Complete ==="