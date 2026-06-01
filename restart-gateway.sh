#!/bin/bash
set -e
cd /home/datdt/tikiclone

# Restart gateway with all env vars from docker-compose
docker run -d --name tikiclone-gateway-1 --network tikiclone_default \
  -p 8080:8080 \
  -e APP_ENV=development \
  -e GATEWAY_HTTP_PORT=8080 \
  -e GATEWAY_GRPC_PORT=9090 \
  -e REDIS_ADDR=redis-master:6379 \
  -e KAFKA_BROKERS=kafka:9092 \
  -e JWT_ACCESS_SECRET=dev-access-secret-key-for-local-development-only \
  -e ENCRYPTION_KEY=b8477fd5d6f4c7b22878f8d3b873778e28fcac8e34a0be45cdf014bbb31571d7 \
  -e OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318 \
  -e UPSTREAM_AUTH_SERVICE=auth:8087 \
  -e UPSTREAM_CATALOG_SERVICE=catalog-product:8088 \
  -e UPSTREAM_CART_SERVICE=cart:8082 \
  -e UPSTREAM_ORDER_SERVICE=order:8084 \
  -e UPSTREAM_INVENTORY_SERVICE=inventory:8086 \
  -e UPSTREAM_PAYMENT_SERVICE=payment:8083 \
  -e UPSTREAM_PRODUCT_SERVICE=product:8089 \
  -e UPSTREAM_PRODUCT_CATALOG_SERVICE=product-catalog:8090 \
  -e UPSTREAM_PROMOTION_SERVICE=promotion:8091 \
  -e UPSTREAM_SHIPMENT_SERVICE=shipment:8092 \
  tikiclone-gateway 2>&1 || true

# If image doesn't exist, build it
if [ $? -ne 0 ]; then
  docker build -t tikiclone-gateway -f services/gateway/Dockerfile .
  docker run -d --name tikiclone-gateway-1 --network tikiclone_default \
    -p 8080:8080 \
    -e APP_ENV=development \
    -e GATEWAY_HTTP_PORT=8080 \
    -e GATEWAY_GRPC_PORT=9090 \
    -e REDIS_ADDR=redis-master:6379 \
    -e KAFKA_BROKERS=kafka:9092 \
    -e JWT_ACCESS_SECRET=dev-access-secret-key-for-local-development-only \
    -e ENCRYPTION_KEY=b8477fd5d6f4c7b22878f8d3b873778e28fcac8e34a0be45cdf014bbb31571d7 \
    -e OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318 \
    -e UPSTREAM_AUTH_SERVICE=auth:8087 \
    -e UPSTREAM_CATALOG_SERVICE=catalog-product:8088 \
    -e UPSTREAM_CART_SERVICE=cart:8082 \
    -e UPSTREAM_ORDER_SERVICE=order:8084 \
    -e UPSTREAM_INVENTORY_SERVICE=inventory:8086 \
    -e UPSTREAM_PAYMENT_SERVICE=payment:8083 \
    -e UPSTREAM_PRODUCT_SERVICE=product:8089 \
    -e UPSTREAM_PRODUCT_CATALOG_SERVICE=product-catalog:8090 \
    -e UPSTREAM_PROMOTION_SERVICE=promotion:8091 \
    -e UPSTREAM_SHIPMENT_SERVICE=shipment:8092 \
    tikiclone-gateway
fi

echo "Gateway started. Waiting 10s..."
sleep 10
