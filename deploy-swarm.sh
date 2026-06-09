#!/bin/bash
# ============================================================
# TIKICLONE SWARM DEPLOYMENT SCRIPT
# 5 Replica Performance Validation Mode
# ============================================================
# Topology: 1 Manager + 3 Workers
# ============================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "========================================"
echo " TIKICLONE SWARM DEPLOYMENT"
echo "========================================"

# 1. Verify cluster
echo "==> Checking swarm cluster status..."
docker node ls --format "table {{.Hostname}}\t{{.Status}}\t{{.Availability}}\t{{.Manager Status}}" 2>/dev/null || {
    echo "ERROR: Docker Swarm not initialized. Run: docker swarm init"
    exit 1
}

NODE_COUNT=$(docker node ls --format "{{.Hostname}}" | wc -l)
WORKER_COUNT=$(docker node ls --filter "role=worker" --format "{{.Hostname}}" | wc -l)
echo "   Nodes: $NODE_COUNT total ($WORKER_COUNT workers)"

# 2. Verify images exist
echo "==> Checking required images..."
REQUIRED_IMAGES=(
    "tikiclone-gateway:optimized"
    "tikiclone-identity-auth:latest"
    "tikiclone-product:latest"
    "tikiclone-catalog-product:latest"
    "tikiclone-cart:latest"
    "tikiclone-checkout:latest"
    "tikiclone-order:latest"
    "tikiclone-auth:latest"
    "mysql:8.0"
    "redis:7-alpine"
    "traefik:v3.0"
    "mongo:7"
    "confluentinc/cp-kafka:7.5.0"
    "confluentinc/cp-zookeeper:7.5.0"
    "prom/prometheus:v3.2.1"
    "grafana/grafana:11.5.2"
    "otel/opentelemetry-collector-contrib:0.91.0"
)

for img in "${REQUIRED_IMAGES[@]}"; do
    if docker images --format "{{.Repository}}:{{.Tag}}" | grep -q "^${img}$"; then
        echo "   [OK] $img"
    else
        echo "   [MISSING] $img - will pull now"
        docker pull "$img" 2>&1 | tail -1
    fi
done

# 3. Clean any previous stack
echo "==> Cleaning previous stack (if any)..."
docker stack rm tiki 2>/dev/null || true
sleep 10
echo "   Cleaning networks..."
docker network ls --filter "name=tiki_" --format "{{.ID}}" | xargs -r docker network rm 2>/dev/null || true

# 4. Deploy stack
echo "==> Deploying stack: docker-stack.yml"
docker stack deploy -c docker-stack.yml tiki
echo "   Stack deployed. Waiting for services to initialize..."

# 5. Wait for services
SERVICES=$(docker stack services tiki --format "{{.Name}}" | wc -l)
echo "   Total services: $SERVICES"

for i in {1..12}; do
    RUNNING=$(docker stack services tiki --format "{{.Name}}\t{{.Replicas}}" | grep -c "[0-9]/[0-9]" || true)
    echo "   Poll $i/12: checking service status..."
    sleep 10
done

# 6. Show final status
echo ""
echo "========================================"
echo " SERVICE STATUS"
echo "========================================"
docker service ls --format "table {{.Name}}\t{{.Mode}}\t{{.Replicas}}\t{{.Image}}"

echo ""
echo "========================================"
echo " NODE STATUS"
echo "========================================"
docker node ls

echo ""
echo "========================================"
echo " NETWORKS"
echo "========================================"
docker network ls --filter "name=tiki_"

echo ""
echo "===== DEPLOYMENT COMPLETE ====="
echo "Traefik Dashboard: http://localhost:8080"
echo "Grafana:          http://localhost:3001 (admin/admin)"
echo ""
echo "Next steps:"
echo "  1. Import dump data: ./scripts/migrate-dump.sh"
echo "  2. Run health check: ./scripts/health-check.sh"
echo "  3. Run stress test:  python3 stress_test.py"
