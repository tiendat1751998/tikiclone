# Kubernetes Cluster Deployment & Autoscaling Rules

All production deployments in the Tiki Clone cluster must adhere to these policies.

## 1. Resource Limits & Probes
- **Requests vs Limits**: CPU request/limit ratio must be `1:2`, Memory request/limit ratio must be `1:1` (to prevent Out-Of-Memory termination during dynamic Java heap expansions).
- **Probes Configurations**:
  - `readinessProbe`: Tells K8s when a Pod is ready to receive traffic.
  - `livenessProbe`: Tells K8s when a Pod has crashed and needs to be restarted.
  - Minimum `initialDelaySeconds` must match service startup times (e.g., 30s for Java Spring Boot, 5s for Go).

## 2. Auto-scaling & High Availability
- **HPA Target metrics**: Horizontal Pod Autoscalers (HPA) must scale up instances when target CPU utilization exceeds **70%** or target Memory exceeds **80%**.
- **Pod Disruption Budgets (PDB)**: Critical services (Payment, Order) must configure PDB to ensure at least **50%** of pods are available during cluster upgrades.
