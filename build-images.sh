#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status
set -e

# Setup script variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKIP_JAVA=false
SKIP_GO=false

# Parse arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --skip-java) SKIP_JAVA=true ;;
        --skip-go) SKIP_GO=true ;;
        *) echo "Unknown parameter: $1"; exit 1 ;;
    esac
    shift
done

echo -e "\033[0;36m=========================================\033[0m"
echo -e "\033[0;36mStarting Docker Image Build Process\033[0m"
echo -e "\033[0;36m=========================================\033[0m"

# 1. Build Java Image
if [ "$SKIP_JAVA" = false ]; then
    JAVA_DIR="${SCRIPT_DIR}/services/identity-auth"
    DOCKERFILE_PATH="${JAVA_DIR}/Dockerfile"
    if [ -f "$DOCKERFILE_PATH" ]; then
        echo -e "\n\033[0;33m[Java] Building docker image for identity-auth...\033[0m"
        IMAGE_NAME="ghcr.io/tiki-clone/identity-auth:latest"
        docker build -t "$IMAGE_NAME" -f "$DOCKERFILE_PATH" "$JAVA_DIR"
        echo -e "\033[0;32m[Java] Success: Built ${IMAGE_NAME}\033[0m"
    fi
fi

# 2. Build Go Images
if [ "$SKIP_GO" = false ]; then
    SHARED_PACKAGE_PATH="${SCRIPT_DIR}/packages/go-shared"
    GO_MODULES="services/auth services/cart services/catalog-product services/checkout services/gateway services/inventory services/order services/payment services/product services/product-catalog services/promotion services/shipment platforms/advertising platforms/aiml platforms/analytics platforms/api-gateway platforms/billing platforms/developer platforms/fraud platforms/fraud-risk platforms/global-infra platforms/live-commerce platforms/live-scale platforms/logistics-delivery platforms/notification platforms/notification-campaign platforms/oms-fulfillment platforms/payment-ledger platforms/rec-vector platforms/recommendation platforms/search platforms/search-indexing platforms/service-mesh platforms/sre platforms/user-behavior"

    for MODULE in $GO_MODULES; do
        FULL_PATH="${SCRIPT_DIR}/${MODULE}"
        if [ -d "$FULL_PATH" ]; then
            MODULE_NAME=$(basename "${MODULE}")
            
            DOCKERFILE_PATH=""
            if [ -f "${FULL_PATH}/Dockerfile" ]; then
                DOCKERFILE_PATH="${FULL_PATH}/Dockerfile"
            elif [ -f "${FULL_PATH}/deployments/Dockerfile" ]; then
                DOCKERFILE_PATH="${FULL_PATH}/deployments/Dockerfile"
            fi
            
            if [ -n "$DOCKERFILE_PATH" ]; then
                echo -e "\n\033[0;33m[Go] Building docker image for ${MODULE}...\033[0m"
                IMAGE_NAME="ghcr.io/tiki-clone/${MODULE_NAME}:latest"
                
                echo "Creating temporary build context for ${MODULE_NAME}..."
                
                TEMP_CONTEXT_DIR="${SCRIPT_DIR}/temp_build_${MODULE_NAME}"
                rm -rf "$TEMP_CONTEXT_DIR"
                mkdir -p "$TEMP_CONTEXT_DIR"
                
                # Copy module and shared pkg into temp context
                cp -r "${FULL_PATH}/"* "$TEMP_CONTEXT_DIR/"
                mkdir -p "${TEMP_CONTEXT_DIR}/packages/go-shared"
                cp -r "${SHARED_PACKAGE_PATH}/"* "${TEMP_CONTEXT_DIR}/packages/go-shared/"
                
                # Copy central proto directory if needed
                HAS_PROTO=0
                if grep -q "github.com/tiki-clone/tiki/proto" "${FULL_PATH}/go.mod" 2>/dev/null || [ "${MODULE_NAME}" = "catalog-product" ]; then
                    echo "-> Copying central proto module into build context..."
                    mkdir -p "${TEMP_CONTEXT_DIR}/proto"
                    cp -r "${SCRIPT_DIR}/proto/"* "${TEMP_CONTEXT_DIR}/proto/"
                    HAS_PROTO=1
                    echo "-> Central proto folder context copy contents:"
                    ls -la "${TEMP_CONTEXT_DIR}/proto/"
                    if [ -d "${TEMP_CONTEXT_DIR}/proto/catalog" ]; then
                        ls -la "${TEMP_CONTEXT_DIR}/proto/catalog/v1/"
                    fi
                fi
                
                # Apply replacements to go.mod inside the temp context
                if [ -f "${TEMP_CONTEXT_DIR}/go.mod" ]; then
                    sed 's|replace github.com/tiki-clone/tiki/proto => ../../proto|replace github.com/tiki-clone/tiki/proto => /proto|g' "${TEMP_CONTEXT_DIR}/go.mod" > "${TEMP_CONTEXT_DIR}/go.mod.tmp"
                    mv "${TEMP_CONTEXT_DIR}/go.mod.tmp" "${TEMP_CONTEXT_DIR}/go.mod"
                    sed 's|replace github.com/tiki-clone/tiki/packages/go-shared => ../../packages/go-shared|replace github.com/tiki-clone/tiki/packages/go-shared => ./packages/go-shared|g' "${TEMP_CONTEXT_DIR}/go.mod" > "${TEMP_CONTEXT_DIR}/go.mod.tmp"
                    mv "${TEMP_CONTEXT_DIR}/go.mod.tmp" "${TEMP_CONTEXT_DIR}/go.mod"
                fi
                
                # Create adjusted temp Dockerfile
                TEMP_DOCKERFILE="${TEMP_CONTEXT_DIR}/Dockerfile.build"
                sed -e 's/; /\n/g' \
                    -e 's|COPY \.\./\.\./packages/go-shared /app/packages/go-shared|COPY packages/go-shared /packages/go-shared|g' \
                    -e 's|RUN go mod download|# RUN go mod download|g' \
                    "$DOCKERFILE_PATH" > "${TEMP_DOCKERFILE}.tmp"
                
                # Safely inject the protoc builder step replacing COPY . .
                awk '/COPY \. \./ {
                    print "COPY . ."
                    print "RUN apt-get update && apt-get install -y protobuf-compiler libprotobuf-dev git wget && \\"
                    print "    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \\"
                    print "    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \\"
                    print "    mkdir -p /tmp/validate/validate && \\"
                    print "    wget -qO /tmp/validate/validate/validate.proto https://raw.githubusercontent.com/bufbuild/protoc-gen-validate/main/validate/validate.proto && \\"
                    print "    if [ -d \"/proto/tiki/catalog/v1\" ]; then \\"
                    print "        echo \"Compiling catalog.proto...\" && \\"
                    print "        mkdir -p /proto/catalog/v1 && \\"
                    print "        protoc --proto_path=/proto/tiki --proto_path=/tmp/validate --proto_path=/usr/include --go_out=/proto --go_opt=module=github.com/tiki-clone/tiki/proto --go-grpc_out=/proto --go-grpc_opt=module=github.com/tiki-clone/tiki/proto /proto/tiki/catalog/v1/catalog.proto || exit 1; \\"
                    print "    fi && \\"
                    print "    find . -name \"*.proto\" -type f -not -path \"./proto/tiki/*\" | while read f; do \\"
                    print "        echo \"Compiling $f...\" && \\"
                    print "        protoc --proto_path=. --proto_path=/tmp/validate --proto_path=/usr/include --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative \"$f\" || exit 1; \\"
                    print "    done"
                    print "RUN go mod tidy"
                    next
                }1' "${TEMP_DOCKERFILE}.tmp" > "$TEMP_DOCKERFILE"
                rm -f "${TEMP_DOCKERFILE}.tmp"
                
                # Check if packages/go-shared COPY is now present, if not, inject it
                if ! grep -q "packages/go-shared" "$TEMP_DOCKERFILE"; then
                    sed 's|COPY go.mod go.sum ./|COPY go.mod go.sum ./\nCOPY packages/go-shared /packages/go-shared|g' "$TEMP_DOCKERFILE" > "${TEMP_DOCKERFILE}.tmp"
                    mv "${TEMP_DOCKERFILE}.tmp" "$TEMP_DOCKERFILE"
                fi
                
                # If proto directory was copied, inject COPY proto /proto into the Dockerfile
                if [ "$HAS_PROTO" -eq 1 ]; then
                    sed 's|COPY packages/go-shared /packages/go-shared|COPY packages/go-shared /packages/go-shared\nCOPY proto /proto|g' "$TEMP_DOCKERFILE" > "${TEMP_DOCKERFILE}.tmp"
                    mv "${TEMP_DOCKERFILE}.tmp" "$TEMP_DOCKERFILE"
                fi
                
                docker build -t "$IMAGE_NAME" -f "$TEMP_DOCKERFILE" "$TEMP_CONTEXT_DIR"
                
                # Clean up
                rm -rf "$TEMP_CONTEXT_DIR"
                echo -e "\033[0;32m[Go] Success: Built ${IMAGE_NAME}\033[0m"
            fi
        fi
    done
fi

echo -e "\n\033[0;36m=========================================\033[0m"
echo -e "\033[0;36mAll Docker image builds completed!\033[0m"
echo -e "\033[0;36m=========================================\033[0m"
