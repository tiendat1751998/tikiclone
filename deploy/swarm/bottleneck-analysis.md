# Bottleneck Analysis

## Overview

Analysis of potential bottlenecks in the Tikiclone Swarm environment based on service architecture and resource constraints.

## Memory Bottleneck Analysis

### Total Memory Allocation (24GB across 3 workers)

| Layer | Services | Memory/Replica | Replicas | Total |
|-------|----------|---------------|----------|-------|
| Gateway | gateway-service | 512Mi | 5 | 2.5GB |
| Product | product-service | 512Mi | 5 | 2.5GB |
| Cart | cart-service | 512Mi | 5 | 2.5GB |
| Auth | identity-auth | 256Mi | 3 | 768Mi |
| Search | search-service | 256Mi | 3 | 768Mi |
| Category | category-service | 256Mi | 3 | 768Mi |
| Checkout | checkout-service | 512Mi | 3 | 1.5GB |
| Order | order-service | 512Mi | 3 | 1.5GB |
| Notification | notification-service | 256Mi | 3 | 768Mi |
| Infrastructure | mysql, redis, mongodb, kafka | varies | 1 each | ~8GB |
| Monitoring | prometheus, grafana, node-exporter | varies | 1 each | ~2GB |

### Per-Worker Distribution (8GB each)

Each worker handles ~5GB of service allocations, leaving ~3GB headroom.

**Critical paths potentially hitting CPU limits first**:
- Gateway (handles all incoming traffic)
- Product (most frequently accessed)
- Cart (high write volume)

## CPU Bottleneck Analysis

### Expected CPU Saturation Points

| Service | Replicas | Target CPU < 70% | Bottleneck TPS |
|---------|----------|------------------|---------------|
| gateway-service | 5 | ~2 core-equivalents | ~500-1000 TPS |
| product-service | 5 | ~2.5 core-equivalents | ~2000-5000 TPS |
| cart-service | 5 | ~2.5 core-equivalents | ~1000-2000 TPS |
| order-service | 3 | ~1.5 core-equivalents | ~500-1000 TPS |

## Database Bottleneck Analysis

### MySQL Primary (Single Instance)

**Current Configuration**:
- max-connections: 2000
- innodb-buffer-pool-size: 1G (in docker-stack.yml)

**Bottleneck Risk**: HIGH
- All services share single MySQL instance
- No read replicas for scaling reads
- Connection pool exhaustion under 1000+ users

**Mitigation**:
- Connection pooling in applications
- Query optimization
- Consider Aurora/MySQL operator for production

### Redis Master (Single Instance)

**Bottleneck Risk**: MEDIUM
- All session/cart data in single instance
- No clustering for sharding
- Network latency to workers

### MongoDB (Single Instance)

**Bottleneck Risk**: LOW
- Used primarily for catalog-product
- Read-heavy workload
- Can add read replicas

## Network Bottleneck Analysis

### Overlay Network Overhead

- All service communication via encrypted overlay networks
- Additional latency for cross-node calls
- Bandwidth: limited by worker node network (typically 1Gbps)

**Critical Paths**:
- gateway → all backend services
- cart → inventory → order (chained calls)

## Identified Bottlenecks

### 1. Database Layer (Priority: HIGH)

**Impact**: System-wide slowdown under load

**Evidence**:
- Single MySQL instance for all services
- Authentication, cart, order, product all query MySQL
- Connection pool saturation likely at 500+ concurrent users

**Recommendations**:
- Add read replicas for product queries
- Implement connection pooling (HikariCP for Java, sqlx for Go)
- Add query timeouts and circuit breakers

### 2. Gateway Layer (Priority: HIGH)

**Impact**: Request queuing, increased latency

**Evidence**:
- All 5000+ RPS through gateway
- HTTP/2 + WebSocket support configured
- Single entry point (Traefik) before gateway

**Recommendations**:
- Ensure 5 replicas are utilized (load balance check)
- Enable connection pooling downstream
- Add caching at gateway level

### 3. Cart Service (Priority: MEDIUM)

**Impact**: Race conditions, cart data inconsistency

**Evidence**:
- High write frequency (add/update/remove)
- Redis atomic operations required
- Session affinity may be needed

**Recommendations**:
- Implement Lua scripts for cart operations
- Add Redis cluster for sharding
- Verify atomicity under load

## Scalability Limits

Based on current configuration:

| Users | Expected Status | Bottleneck Risk |
|-------|-----------------|-----------------|
| 50 | ✅ Optimal | None |
| 100 | ✅ Optimal | None |
| 250 | ✅ Good | Low |
| 500 | ⚠️ Degraded | Medium |
| 1000 | ⚠️ Bottlenecked | High (MySQL) |

## Resource Optimization Opportunities

### 1. Decrease Memory Limits
```yaml
# Can reduce these for test environment
identity-auth: 256Mi → 192Mi
search-service: 256Mi → 128Mi
notification-service: 256Mi → 192Mi
```

### 2. Add Read Replicas
```yaml
# MySQL read replica for product queries
mysql-read-1:
  image: mysql:8.0
  command: --read-only --innodb-buffer-pool-size=1G
```

### 3. Optimize Connection Pools
```yaml
# Current: no explicit pool config
# Add to each service:
environment:
  MYSQL_MAX_OPEN_CONNS: "100"
  MYSQL_MAX_IDLE_CONNS: "25"
  REDIS_POOL_SIZE: "50"
```

## Monitoring Thresholds

Configure alerts for:

```yaml
alerts:
  - name: HighCPU
    expr: rate(container_cpu_usage_seconds_total[1m]) > 0.8
    severity: warning

  - name: HighMemory
    expr: container_memory_usage_bytes / container_spec_memory_limit > 0.85
    severity: warning

  - name: DatabaseSlow
    expr: mysql_global_status_slow_queries > 100
    severity: critical

  - name: GatewayLatency
    expr: histogram_quantile(0.95, gateway_request_duration_seconds) > 0.5
    severity: warning
```

## Conclusion

The Swarm environment is optimized for test validation (up to ~500 concurrent users). For production Kubernetes rollout, prioritize:

1. **Database scaling** - MySQL read replicas, connection pooling
2. **Gateway tuning** - Ensure load distribution across 5 replicas
3. **Cache optimization** - Redis cluster consideration