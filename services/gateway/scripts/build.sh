#!/bin/bash
set -euo pipefail

APP_NAME="tiki-gateway"
VERSION="${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
OUTPUT_DIR="./bin"

mkdir -p "${OUTPUT_DIR}"

echo "Building ${APP_NAME}:${VERSION}"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-w -s -X main.version=${VERSION}" \
  -o "${OUTPUT_DIR}/${APP_NAME}" \
  ./cmd/server

echo "Build complete: ${OUTPUT_DIR}/${APP_NAME}"
ls -lh "${OUTPUT_DIR}/${APP_NAME}"
