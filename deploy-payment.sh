#! /bin/bash
# Deploy payment service to Docker Swarm worker nodes
# Usage: ./deploy-payment.sh

set -e

# Build the image
docker build -t tikiclone-payment:latest -f services/payment/Dockerfile .

# Redeploy payment service from the main docker-stack.yml
docker stack deploy -c docker-stack.yml tiki

echo "Payment service deploying. Check with: docker service logs tiki_payment"