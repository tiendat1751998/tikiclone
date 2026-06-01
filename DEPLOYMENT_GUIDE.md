# Tiki Clone — Deployment Guide

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Kubernetes Cluster                        │
│                                                                  │
│  ┌──────────────┐    ┌──────────────────────────────────────┐   │
│  │   Ingress     │───▶│         API Gateway (tiki-gateway) │   │
│  │   Controller  │    │  - JWT Validation                    │   │
│  └──────────────┘    │  - Rate Limiting                     │   │
│                       │  - Request Routing                   │   │
│                       └──────────────────────────────────────┘   │
│                                      │                           │
│         ┌────────────────────────────┼────────────────────┐     │
│         │                            │                    │     │
│  ┌──────▼──────┐  ┌──────▼──────┐  ┌─▼──────────────┐    │     │
│  │  Inventory   │  │  Payment    │  │  Order          │    │     │
│  │  Service     │  │  Service    │  │  Service        │    │     │
│  └──────┬──────┘  └──────┬──────┘  └─┬──────────────┘    │     │
│         │                │            │                    │     │
│  ┌──────▼──────┐  ┌──────▼──────┐  ┌─▼──────────────┐    │     │
│  │  MySQL       │  │  Redis      │  │  Kafka          │    │     │
│  │  (Primary)   │  │  (Cluster)  │  │  (Cluster)      │    │     │
│  └─────────────┘  └─────────────┘  └────────────────┘    │     │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Observability Stack                     │   │
│  │  Prometheus │ Grafana │ Jaeger │ OpenTelemetry Collector  │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Quick Start

### Local Development

```bash
# Start all infrastructure + services
make deploy-local-build

# Check status
docker-compose ps

# View logs
make logs

# Run tests
make test

# Stop everything
make stop-local-clean
```

### Kubernetes Deployment

```bash
# Deploy to staging
make deploy-staging

# Deploy to production
make deploy-prod

# Check status
make status

# Rollback a service
make rollback-inventory
```

## Services

| Service | HTTP Port | gRPC Port | Replicas | Description |
|---------|-----------|-----------|----------|-------------|
| gateway | 8080 | 9090 | 3 | API Gateway, JWT validation, rate limiting |
| inventory | 8086 | 9096 | 3 | Stock management, anti-oversell |
| payment | 8083 | 9093 | 3 | Payment processing, PSP integration |
| order | 8084 | 9094 | 3 | Order lifecycle management |
| checkout | 8085 | 9095 | 3 | Checkout orchestration (Saga) |
| cart | 8082 | 9092 | 3 | Shopping cart management |
| product | 8087 | 9097 | 3 | Product catalog |
| promotion | 8089 | 9099 | 3 | Vouchers, campaigns, pricing rules |
| shipment | 8090 | 9091 | 3 | Shipping & delivery tracking |
| auth | 8081 | 9090 | 3 | Authentication & RBAC |

## Deployment Files Structure

Each service has a `deployments/` directory containing:

```
services/inventory/deployments/
├── deployment.yaml      # Kubernetes Deployment
├── service.yaml         # Kubernetes Service
├── configmap.yaml       # Environment configuration
├── secrets.yaml         # Secrets (passwords, keys)
├── hpa.yaml             # Horizontal Pod Autoscaler
├── pdb.yaml             # Pod Disruption Budget
├── network-policy.yaml  # Network isolation rules
└── service-monitor.yaml # Prometheus monitoring
```

## Security Features

- **JWT Authentication**: All services require valid JWT tokens
- **Network Policies**: Services can only communicate with authorized peers
- **Pod Security**: All containers run as non-root with read-only filesystems
- **Resource Limits**: CPU/Memory limits prevent resource exhaustion
- **Health Checks**: Liveness, readiness, and startup probes
- **Graceful Shutdown**: 60-second termination grace period

## Monitoring

- **Metrics**: Prometheus metrics at `/metrics` endpoint
- **Tracing**: OpenTelemetry with Jaeger backend
- **Logging**: Structured JSON logs with correlation IDs
- **Alerts**: ServiceMonitor for Prometheus alerting

## CI/CD Pipeline

1. **Lint**: `go vet` for all Go services
2. **Test**: Unit tests with MySQL/Redis test containers
3. **Security Scan**: Trivy vulnerability scanner
4. **Build**: Docker image build and push
5. **Deploy Staging**: Automatic deployment to staging
6. **Deploy Production**: Manual approval required
