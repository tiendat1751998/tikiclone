#!/bin/bash
# ============================================================
# TikiClone Production Deploy Script
# Run on k8s master node (10.0.0.200)
# ============================================================
set -e

echo "=== TikiClone Production Deploy ==="
cd /home/datdt/tikiclone

# Pull latest code
echo "[1/5] Pulling latest code..."
git pull origin dev-datdt

# Build images
echo "[2/5] Building images..."

# Build web (Next.js)
echo "  → Building tikiclone-web..."
docker build -t tikiclone-web:latest -f apps/web/Dockerfile apps/web/

# Build gateway (Go)
echo "  → Building tikiclone-gateway..."
cd services/gateway
export PATH=$PATH:/usr/local/go/bin
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /tmp/gateway .
cd ../..

# Build all other Go services
for svc in identity-auth product catalog-product cart checkout order inventory payment delivery notification; do
    if [ -f "services/$svc/Dockerfile" ]; then
        echo "  → Building tikiclone-$svc..."
        docker build -t "tikiclone-$svc:latest" -f "services/$svc/Dockerfile" "services/$svc/" 2>/dev/null || echo "    ⚠️  Build failed for $svc, using existing image"
    fi
done

# Deploy stack
echo "[3/5] Deploying stack..."
docker stack deploy -c docker-stack.yml tiki

# Wait for services
echo "[4/5] Waiting for services to start..."
sleep 10

# Verify
echo "[5/5] Verifying services..."
echo ""
echo "=== Service Status ==="
docker service ls --format "table {{.Name}}\t{{.Image}}\t{{.Replicas}}\t{{.Ports}}"
echo ""

echo "=== Traefik Logs (last 20 lines) ==="
docker service logs tiki_traefik --tail 20 2>/dev/null || echo "Traefik not ready yet"
echo ""

echo "=== Gateway Logs (last 20 lines) ==="
docker service logs tiki_gateway --tail 20 2>/dev/null || echo "Gateway not ready yet"
echo ""

echo "=== Web Logs (last 20 lines) ==="
docker service logs tiki_web --tail 20 2>/dev/null || echo "Web not ready yet"
echo ""

echo "=== Deploy Complete ==="
echo "Access frontend at: http://$(hostname -I | awk '{print $1}')/"
echo "Traefik dashboard:  http://$(hostname -I | awk '{print $1}'):8080/dashboard/"
