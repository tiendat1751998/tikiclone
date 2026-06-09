#!/bin/bash
# deploy-all.sh — Deploy full tikiclone stack to Swarm
# Run this on the swarm manager node.
set -euo pipefail

STACK_NAME="${1:-tiki}"
SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$SCRIPT_DIR"

echo "=== Creating Docker configs ==="
for cfg in \
  prometheus_config:configs/prometheus.yml \
  prometheus_alerts:configs/prometheus-alerts.yml \
  grafana_datasources:deploy/platform/monitoring/grafana/datasources/datasources.yaml \
  otel_collector_config:configs/otel-collector.yaml \
  promtail_config:configs/promtail/docker-config.yaml \
  traefik_dynamic:deploy/platform/ingress/traefik/dynamic.yml
do
  name="${cfg%%:*}"
  file="${cfg#*:}"
  if docker config inspect "$name" >/dev/null 2>&1; then
    docker config rm "$name"
  fi
  docker config create "$name" "$file"
done

echo "=== Creating Docker secrets ==="
for sec in \
  tiki_cert:certs/cert.pem \
  tiki_key:certs/key.pem
do
  name="${sec%%:*}"
  file="${sec#*:}"
  if docker secret inspect "$name" >/dev/null 2>&1; then
    docker secret rm "$name"
  fi
  docker secret create "$name" "$file"
done

echo "=== Deploying stack: $STACK_NAME ==="
docker stack deploy -c docker-stack.yml "$STACK_NAME"

echo "=== Waiting for services to stabilize ==="
sleep 10
docker service ls --filter name="${STACK_NAME}_"

echo "=== Done ==="
echo "Check status: docker service ls"
echo "Check logs:   docker service logs ${STACK_NAME}_traefik"
echo "Test HTTPS:   curl -sk https://localhost/api/v1/products?limit=1"
