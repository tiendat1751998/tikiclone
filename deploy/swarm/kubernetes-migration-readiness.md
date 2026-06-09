# Kubernetes Migration Readiness Report

## Overview

This document assesses the Tikiclone application's readiness for Kubernetes migration based on Swarm validation results.

## Container Compliance

### ✅ PASS - All Services Containerized
- All 9 services have Dockerfiles
- Multi-stage builds used
- Non-root user configured
- Health check endpoints implemented

### ⚠️ WARN - Health Check Paths
Some services use `/health`, others use `/ready` - standardize for Kubernetes:
- identity-auth: `/health` → `/actuator/health/liveness` & `/actuator/health/readiness`
- gateway-service: `/health` → `/healthz`
- product-service: `/health` → `/ready`

## Service Discovery

### ✅ Swarm DNS Compatibility
Services communicate via DNS names (e.g., `mysql-primary`) which translates directly to Kubernetes service names.

### ⚠️ WARN - Environment Variables
Current Swarm uses compose environment variables. Kubernetes requires:
- ConfigMap for non-sensitive configs
- Secret for sensitive values (JWT secrets, DB passwords)

## Resource Requirements

### Memory Mapping

| Service | Swarm Limit | K8s Request | K8s Limit | Notes |
|---------|-------------|-------------|-----------|-------|
| gateway-service | 512Mi | 256Mi | 512Mi | Increase to 1Gi for production |
| product-service | 512Mi | 256Mi | 512Mi | Similar to Swarm |
| search-service | 256Mi | 128Mi | 256Mi | Can use smaller limits |
| cart-service | 512Mi | 256Mi | 512Mi | Session memory overhead |
| identity-auth | 256Mi | 128Mi | 512Mi | JWT processing overhead |
| category-service | 256Mi | 128Mi | 256Mi | Read-heavy, lower limits OK |
| checkout-service | 512Mi | 256Mi | 512Mi | Transactional, keep limits |
| order-service | 512Mi | 256Mi | 512Mi | Persistent connections |
| notification-service | 256Mi | 128Mi | 256Mi | Asynchronous, lower limits |

## Stateful Services Assessment

### MySQL
- **Type**: Single instance
- **K8s Migration**: Use StatefulSet with persistent volumes
- **Current**: No clustering, no replication
- **Recommendation**: Deploy MySQL operator or use managed service

### Redis
- **Type**: Single master
- **K8s Migration**: Redis Sentinel or Cluster
- **Recommendation**: Use Redis operator for HA

### MongoDB
- **Type**: Single instance
- **K8s Migration**: MongoDB ReplicaSet via Helm chart
- **Recommendation**: Use MongoDB community operator

### Kafka
- **Type**: Single broker
- **K8s Migration**: Use Strimzi Kafka operator
- **Existing**: `deploy/platform/messaging/kafka/strimzi/kafka-cluster.yaml`

## Network Policies

### ✅ Required for K8s
- frontend (ingress)
- backend (service-to-service)
- monitoring (metrics scraping)

### Missing: Network Policy Definitions
Create policies in `deploy/k8s/base/network-policies/`

## Secrets Management

### Current
- Environment variables in compose file
- Hardcoded development secrets

### Required for K8s
- `JWT_ACCESS_SECRET` → Kubernetes Secret
- `JWT_REFRESH_SECRET` → Kubernetes Secret  
- `ENCRYPTION_KEY` → Kubernetes Secret
- Use External Secrets Operator for production

## Horizontal Pod Autoscaling

### ✅ HPA Ready
HPA configurations exist in:
- `deploy/k8s/identity-auth/hpa.yaml` (but not for all services)

### Required
Create HPA for each service:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gateway-service
spec:
  minReplicas: 5
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

## Persistent Volumes

| Volume | Size | StorageClass | Backup |
|--------|------|--------------|--------|
| mysql_data | 10Gi | standard | Required |
| redis_data | 5Gi | standard | Required |
| mongo_data | 10Gi | standard | Required |
| grafana_data | 1Gi | standard | Optional |
| prometheus_data | 20Gi | standard | Required |

## Readiness Checklist

| Item | Status | Notes |
|------|--------|-------|
| ✅ Container images build successfully | PASS | All services have valid Dockerfiles |
| ✅ Health check endpoints | WARN | Need standardization |
| ✅ Resource limits defined | PASS | In docker-stack.yml |
| ✅ Config externalized | WARN | Hardcoded in compose, need secrets |
| ⚠️ Liveness/readiness probes | WARN | Need proper K8s probe configs |
| ⚠️ Network policies | MISSING | Need to create |
| ⚠️ Service mesh support | PARTIAL | Istio/Linkerd integration possible |
| ✅ Rolling update strategy | PASS | Configured in both compose and K8s |
| ✅ Pod anti-affinity | PASS | Spread across nodes |
| ❓ External dependencies | - | Kafka, Redis, MySQL need operators |

## Migration Steps

1. **Phase 1**: Deploy to K8s with same replica counts
2. **Phase 2**: Enable HPA and validate autoscaling
3. **Phase 3**: Migrate stateful services to operators
4. **Phase 4**: Implement network policies
5. **Phase 5**: Add service mesh (optional)
6. **Phase 6**: Production tuning

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Database migration | HIGH | Use managed DB or operators |
| Cache migration | MEDIUM | Redis cluster with sentinel |
| Service discovery | LOW | Kubernetes DNS handles this |
| Network policies | MEDIUM | Implement before production |
| Secrets rotation | MEDIUM | Use External Secrets Operator |

## Recommendations

1. **Standardize health checks** across all services before K8s migration
2. **Create Helm charts** for all services (partially exists)
3. **Implement init containers** for database migrations
4. **Add preStop hooks** for graceful shutdown
5. **Configure resource requests** based on Swarm test data
6. **Set up monitoring stack** in K8s (existing Prometheus/Grafana configs)