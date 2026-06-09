# Replica Placement Strategy

## Cluster Node Labeling

Label nodes for optimal placement:

```bash
# Manager Node (Control Plane)
docker node update --label-add manager=true $(docker node ls --filter role=manager -q)
docker node update --label-add tier=control $(docker node ls --filter role=manager -q)

# Worker Nodes
docker node update --label-add worker=true worker-1
docker node update --label-add worker=true worker-2
docker node update --label-add worker=true worker-3

# Specialized labels for stateful services
docker node update --label-add db=primary worker-1
docker node update --label-add cache=primary worker-2
docker node update --label-add messaging=primary worker-3
```

## Placement Constraints

| Service | Replicas | Constraint | Spread Preference |
|---------|----------|------------|-------------------|
| gateway-service | 5 | `node.labels.worker != null` | `node.id` (spread across all workers) |
| product-service | 5 | `node.labels.worker != null` | `node.id` |
| search-service | 3 | `node.labels.worker != null` | `node.id` |
| cart-service | 5 | `node.labels.worker != null` | `node.id` |
| identity-auth | 3 | `node.labels.worker != null` | `node.id` |
| category-service | 3 | `node.labels.worker != null` | `node.id` |
| checkout-service | 3 | `node.labels.worker != null` | `node.id` |
| order-service | 3 | `node.labels.worker != null` | `node.id` |
| notification-service | 3 | `node.labels.worker != null` | `node.id` |

## Stateful Service Placement

| Service | Replicas | Node Constraint |
|---------|----------|-----------------|
| mysql-primary | 1 | `node.labels.db == primary` |
| redis-master | 1 | `node.labels.cache == primary` |
| mongodb | 1 | `node.labels.db == primary` |
| kafka | 1 | `node.labels.messaging == primary` |
| zookeeper | 1 | `node.labels.messaging == primary` |
| prometheus | 1 | `node.role == manager` |
| grafana | 1 | `node.role == manager` |
| otel-collector | 1 | `node.role == manager` |

## Memory Distribution Across Workers

Each worker has 8GB RAM. Total allocated memory per worker:

### Per Worker Memory Allocation (8GB total)

| Service Category | Replicas per Worker | Memory/Replica | Total |
|------------------|---------------------|----------------|-------|
| Gateway (5 total) | ~2 replicas | 512Mi | ~1GB |
| Product (5 total) | ~2 replicas | 512Mi | ~1GB |
| Cart (5 total) | ~2 replicas | 512Mi | ~1GB |
| Auth (3 total) | ~1 replica | 256Mi | ~256Mi |
| Search (3 total) | ~1 replica | 256Mi | ~256Mi |
| Category (3 total) | ~1 replica | 256Mi | ~256Mi |
| Checkout (3 total) | ~1 replica | 512Mi | ~256Mi |
| Order (3 total) | ~1 replica | 512Mi | ~256Mi |
| Notification (3 total) | ~1 service | 256Mi | ~256Mi |
| Node Exporter / cAdvisor | global | 256Mi | ~256Mi |
| **TOTAL PER WORKER** | | | **~5GB** |

This leaves ~3GB headroom per worker for:
- System overhead
- Peak traffic spikes
- Docker daemon and networking

## Deployment Order

1. Label manager and worker nodes
2. Deploy stateful services (infra layer)
3. Wait for health checks to pass
4. Deploy stateless services in order:
   - auth services first (identity-auth)
   - then business services
5. Deploy monitoring stack
6. Deploy Traefik load balancer

## Rolling Update Strategy

- **Parallelism**: 2 (for 5-replica services) or 1 (for 3-replica services)
- **Delay**: 10-20s between batches
- **Failure Action**: rollback
- **Order**: stop-first (terminate old before starting new)

## Resource Reservations

Critical for preventing OOM:

```yaml
# Example resource allocation
deploy:
  resources:
    limits:
      memory: 512M  # Hard limit
      cpus: '0.5'
    reservations:
      memory: 256M  # Guaranteed minimum
      cpus: '0.1'
```

## Health Check Protocol

All services use HTTP health checks with:
- `initial_delay_seconds`: 20-30s (service startup time)
- `period_seconds`: 10s
- `timeout_seconds`: 5s
- `failure_threshold`: 5 retries

## Anti-Affinity Rules

Services configured with `spread: node.id` to ensure replicas
are distributed across different nodes, maximizing fault tolerance.