#!/bin/bash
set -euo pipefail

echo "=== Tiki Clone Platform Bootstrap ==="

NAMESPACES=(
  "tiki"
  "tiki-infra"
  "tiki-observability"
  "tiki-gitops"
  "tiki-ingress"
)

echo "Creating namespaces..."
for ns in "${NAMESPACES[@]}"; do
  kubectl create namespace "$ns" --dry-run=client -o yaml | kubectl apply -f -
done

echo "Installing cert-manager..."
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml
kubectl -n cert-manager wait --for=condition=Available deployment --all --timeout=300s

echo "Installing ArgoCD..."
kubectl create namespace tiki-gitops --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -n tiki-gitops -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl -n tiki-gitops wait --for=condition=Available deployment --all --timeout=300s

echo "Installing Prometheus Stack..."
kubectl apply -f deploy/platform/monitoring/prometheus/prometheus.yaml
kubectl apply -f deploy/platform/monitoring/grafana/

echo "Installing Loki + Tempo..."
kubectl apply -f deploy/platform/monitoring/loki/
kubectl apply -f deploy/platform/monitoring/tempo/
kubectl apply -f deploy/platform/monitoring/otel-collector/

echo "Installing databases..."
kubectl apply -f deploy/platform/database/postgres/
kubectl apply -f deploy/platform/database/mongodb/
kubectl apply -f deploy/platform/database/elasticsearch/

echo "Installing infrastructure..."
kubectl apply -f deploy/platform/cache/redis/
kubectl apply -f deploy/platform/messaging/kafka/
kubectl apply -f deploy/platform/storage/minio/

echo "Installing Kubernetes configs..."
kubectl apply -f deploy/platform/kubernetes/namespaces/
kubectl apply -f deploy/platform/kubernetes/rbac/
kubectl apply -f deploy/platform/kubernetes/resource-management/
kubectl apply -f deploy/platform/kubernetes/network-policies/
kubectl apply -f deploy/platform/kubernetes/pod-security/
kubectl apply -f deploy/platform/kubernetes/storage/

echo "Installing Ingress..."
kubectl apply -f deploy/platform/ingress/nginx/
kubectl apply -f deploy/platform/ingress/certificates/

echo "Installing Istio..."
istioctl install -f deploy/platform/istio/control-plane/istio-operator.yaml -y
kubectl apply -f deploy/platform/istio/gateways/
kubectl apply -f deploy/platform/istio/mtls/

echo "Installing Secrets management..."
kubectl apply -f deploy/platform/secrets/external-secrets/
kubectl apply -f deploy/platform/secrets/vault/

echo "Installing ArgoCD applications..."
kubectl apply -f deploy/platform/gitops/argocd/projects/
kubectl apply -f deploy/platform/gitops/argocd/applicationsets/

echo "=== Bootstrap complete ==="
echo "ArgoCD URL: https://argocd.tiki-clone.com"
echo "Grafana URL: https://grafana.tiki-clone.com"
