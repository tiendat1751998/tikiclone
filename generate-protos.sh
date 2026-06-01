#!/bin/sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "${SCRIPT_DIR}"

echo "Starting protobuf compilation inside a golang Docker container..."

docker run --rm \
  -v "$(pwd)":/workspace \
  -w /workspace \
  golang:1.26.3-bookworm \
  sh -c "
    apt-get update && apt-get install -y protobuf-compiler libprotobuf-dev git wget && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    
    echo 'Cloning protoc-gen-validate for imports...' && \
    git clone --depth 1 https://github.com/bufbuild/protoc-gen-validate.git /tmp/validate && \
    
    echo 'Generating central catalog proto...' && \
    mkdir -p proto/catalog/v1 && \
    protoc --proto_path=proto/tiki \
           --proto_path=/tmp/validate \
           --proto_path=/usr/include \
           --go_out=proto \
           --go_opt=module=github.com/tiki-clone/tiki/proto \
           --go-grpc_out=proto \
           --go-grpc_opt=module=github.com/tiki-clone/tiki/proto \
           proto/tiki/catalog/v1/catalog.proto && \
           
    echo 'Generating services/order proto...' && \
    protoc --proto_path=services/order \
           --proto_path=/usr/include \
           --go_out=services/order \
           --go_opt=paths=source_relative \
           --go-grpc_out=services/order \
           --go-grpc_opt=paths=source_relative \
           services/order/proto/order/v1/order.proto && \
           
    echo 'Generating services/payment proto...' && \
    protoc --proto_path=services/payment \
           --proto_path=/usr/include \
           --go_out=services/payment \
           --go_opt=paths=source_relative \
           --go-grpc_out=services/payment \
           --go-grpc_opt=paths=source_relative \
           services/payment/proto/payment/v1/payment.proto && \
           
    echo 'Generating services/inventory proto...' && \
    protoc --proto_path=services/inventory \
           --proto_path=/usr/include \
           --go_out=services/inventory \
           --go_opt=paths=source_relative \
           --go-grpc_out=services/inventory \
           --go-grpc_opt=paths=source_relative \
           services/inventory/proto/inventory/v1/inventory.proto && \
           
    echo 'Generating services/product-catalog proto...' && \
    protoc --proto_path=services/product-catalog \
           --proto_path=/usr/include \
           --go_out=services/product-catalog \
           --go_opt=paths=source_relative \
           --go-grpc_out=services/product-catalog \
           --go-grpc_opt=paths=source_relative \
           services/product-catalog/proto/productcatalog/v1/catalog.proto && \
           
    echo 'Generating services/shipment proto...' && \
    protoc --proto_path=services/shipment \
           --proto_path=/usr/include \
           --go_out=services/shipment \
           --go_opt=paths=source_relative \
           --go-grpc_out=services/shipment \
           --go-grpc_opt=paths=source_relative \
           services/shipment/proto/shipment/v1/shipment.proto && \
    
    echo 'Verifying generated catalog-product files:' && \
    ls -la proto/catalog/v1/
  "

echo "All protobuf compilation completed successfully!"
