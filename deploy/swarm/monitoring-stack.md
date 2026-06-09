# Monitoring Stack

## Docker Stack Configuration

The monitoring stack runs as part of the main docker-stack.yml with the following services:

### Prometheus Configuration

```yaml
prometheus:
   image: prom/prometheus:v3.12.0
  volumes:
    - prometheus_data:/prometheus
    - ./configs/prometheus.yml:/etc/prometheus/prometheus.yml:ro
  networks:
    - monitoring
    - backend
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--storage.tsdb.path=/prometheus'
    - '--storage.tsdb.retention.time=7d'
    - '--web.enable-lifecycle'
    - '--web.enable-admin-api'
    - '--web.console.libraries=/usr/share/prometheus/console_libraries'
  deploy:
    mode: replicated
    replicas: 1
    placement:
      constraints: [node.role == manager]
    resources:
      limits:
        memory: 1G
        cpus: '1'
      reservations:
        memory: 512M
        cpus: '0.5'
```

### Grafana Configuration

```yaml
grafana:
  image: grafana/grafana:11.5.2
  volumes:
    - grafana_data:/var/lib/grafana
    - ./deploy/platform/monitoring/grafana/datasources:/etc/grafana/provisioning/datasources:ro
    - ./deploy/platform/monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards:ro
  networks:
    - monitoring
    - frontend
  environment:
    - GF_SECURITY_ADMIN_USER=admin
    - GF_SECURITY_ADMIN_PASSWORD=admin
    - GF_USERS_ALLOW_SIGN_UP=false
  deploy:
    mode: replicated
    replicas: 1
    placement:
      constraints: [node.role == manager]
    resources:
      limits:
        memory: 512M
        cpus: '0.5'
```

### Node Exporter (Global)

```yaml
node-exporter:
  image: prom/node-exporter:v1.8.0
  networks:
    - monitoring
  deploy:
    mode: global
    resources:
      limits:
        memory: 128M
        cpus: '0.2'
```

### cAdvisor (Global)

```yaml
cadvisor:
  image: gcr.io/cadvisor/cadvisor:v0.47.0
  volumes:
    - /:/rootfs:ro
    - /var/run:/var/run:ro
    - /sys:/sys:ro
    - /var/lib/docker:/var/lib/docker:ro
  networks:
    - monitoring
  deploy:
    mode: global
    resources:
      limits:
        memory: 256M
        cpus: '0.5'
```

### Loki & Promtail

```yaml
loki:
   image: grafana/loki:3.6.11
  networks:
    - monitoring
    - backend
  command: -config.file=/etc/loki/local-config.yaml
  deploy:
    mode: replicated
    replicas: 1
    resources:
      limits:
        memory: 512M
        cpus: '0.5'

promtail:
   image: grafana/promtail:3.6.11
  volumes:
    - /var/log:/var/log
    - /var/lib/docker/containers:/var/lib/docker/containers:ro
  networks:
    - monitoring
  deploy:
    mode: global
    resources:
      limits:
        memory: 128M
        cpus: '0.2'
```

## Monitored Metrics

### System Metrics
- **CPU Usage**: Per container and per node
- **RAM Usage**: Per container and per node (target < 75%)
- **Disk I/O**: Read/write operations and throughput
- **Network I/O**: Bytes in/out, connection counts

### Application Metrics
- **API Latency**: 95th and 99th percentile response times
- **Error Rate**: HTTP 5xx errors, service exceptions
- **Request Throughput**: requests/second per service
- **Replica Status**: Running, pending, failed replica counts

### Database Metrics
- **MySQL**: Connection count, query latency, buffer pool usage
- **Redis**: Memory usage, hit rate, connected clients
- **MongoDB**: Operation counts, connection pool, memory usage

### Kafka Metrics
- **Broker**: Message throughput, partition counts
- **Topics**: Message rates, lag metrics

## Grafana Dashboards

Pre-configured dashboards for:
1. **Swarm Overview** - Cluster health, replica distribution
2. **Service Metrics** - Per-service resource usage
3. **Business Metrics** - Orders, carts, inventory levels
4. **Infrastructure** - MySQL, Redis, MongoDB, Kafka metrics
5. **Alerting** - Active alerts and thresholds

## Alert Rules

```yaml
# Critical Alerts
- Alert: HighMemoryUsage
  expr: container_memory_usage_bytes / container_spec_memory_limit > 0.85
  for: 2m
  severity: critical

- Alert: HighCPUUsage  
  expr: rate(container_cpu_usage_seconds_total[5m]) > 0.8
  for: 2m
  severity: warning

- Alert: ServiceDown
  expr: absent(container_last_seen)
  for: 1m
  severity: critical

- Alert: ReplicaFailure
  expr: count by (service) (swarm_task_state{state!="running"}) > 0
  for: 1m
  severity: warning
```

## Access URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | http://manager:3001 | admin/admin |
| Prometheus | http://manager:9090 | - |
| Traefik Dashboard | http://manager:8080/dashboard | - |

## Retention Policies

- **Prometheus**: 7 days retention
- **Loki**: 7 days retention
- **Grafana**: Indefinite (persistent volume)
- **cAdvisor**: No persistent storage (real-time only)