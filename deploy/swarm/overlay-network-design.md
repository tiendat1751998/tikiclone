# Docker Swarm Overlay Network Design

## Network Topology

```
┌─────────────────────────────────────────────────────────────────┐
│                     Docker Swarm Network                          │
└─────────────────────────────────────────────────────────────────┘

                    ┌──────────────────┐
                    │   Manager Node   │
                    │   (Control)      │
                    └────────┬─────────┘
                             │
           ┌─────────────────┼─────────────────┐
           │                 │                 │
           ▼                 ▼                 ▼
    ┌──────────┐      ┌──────────┐      ┌──────────┐
    │ Worker-1 │      │ Worker-2 │      │ Worker-3 │
    │  8GB RAM │      │  8GB RAM │      │  8GB RAM │
    └──────────┘      └──────────┘      └──────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Network Segmentation                           │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  frontend (overlay) - Public facing traffic                      │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ Traefik Load Balancer, Gateway, Web UI                    │  │
│  │ Attachable: Yes                                           │  │
│  │ Encrypted: Yes (default)                                  │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  backend (overlay) - Internal service communication                │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ MySQL, Redis, MongoDB, Kafka, Services communication       │  │
│  │ Attachable: Yes                                           │  │
│  │ Encrypted: Yes (default)                                  │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  monitoring (overlay) - Observability stack                      │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ Prometheus, Grafana, cAdvisor, Node Exporter, Loki, Promtail│  │
│  │ Attachable: Yes                                           │  │
│  │ Encrypted: Yes (default)                                  │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Network Configuration

### frontend Network
- **Driver**: overlay
- **Attachable**: true (for external containers if needed)
- **Scope**: Swarm-scope
- **Purpose**: External HTTP/HTTPS traffic, load balancer ingress
- **Services**: traefik, gateway-service, web

### backend Network
- **Driver**: overlay  
- **Attachable**: true
- **Scope**: Swarm-scope
- **Purpose**: Internal service-to-service communication, database access
- **Services**: mysql-primary, redis-master, mongodb, kafka, zookeeper, all microservices

### monitoring Network
- **Driver**: overlay
- **Attachable**: true
- **Scope**: Swarm-scope
- **Purpose**: Metrics collection, log aggregation
- **Services**: prometheus, grafana, node-exporter, cadvisor, loki, promtail

## Service Discovery

Services communicate via DNS names:
- `gateway-service` → resolves to all 5 gateway replicas
- `mysql-primary` → single MySQL instance
- `redis-master` → single Redis instance
- `identity-auth` → all 3 auth replicas

## Traffic Flow

```
Internet → [Traefik:80,443] → gateway-service (5 replicas) → Backend Services
                                            ├─→ identity-auth (3 replicas)
                                            ├─→ product-service (5 replicas)
                                            ├─→ cart-service (5 replicas)
                                            ├─→ search-service (3 replicas)
                                            ├─→ category-service (3 replicas)
                                            ├─→ checkout-service (3 replicas)
                                            ├─→ order-service (3 replicas)
                                            └─→ notification-service (3 replicas)

Monitoring: Prometheus scrapes metrics from all services via monitoring network
Logging: Promtail collects logs → Loki → Grafana
```

## Security Considerations

- All overlay networks encrypted by default
- No direct external access to backend network
- Secrets managed via Docker secrets
- Traefik handles TLS termination