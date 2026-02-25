# Kubernetes Deployment Guide

## Architecture
```
┌─────────────────────────────────────────────┐
│           Kubernetes Cluster                │
│  ┌──────────────────────────────────────┐   │
│  │  Namespace: ratelimiter              │   │
│  │                                      │   │
│  │  ┌────────────┐  ┌────────────┐      │   │
│  │  │  Pod 1     │  │  Pod 2     │      │   │
│  │  │ Rate       │  │ Rate       │      │   │
│  │  │ Limiter    │  │ Limiter    │ ...  │   │
│  │  └────────────┘  └────────────┘      │   │
│  │         │              │             │   │
│  │         └──────┬───────┘             │   │
│  │                │                     │   │
│  │         ┌──────▼───────┐             │   │
│  │         │   Redis      │             │   │
│  │         │   Service    │             │   │
│  │         └──────────────┘             │   │
│  └──────────────────────────────────────┘   │
│                                             │
│  HPA: Auto-scales 2-10 pods based on CPU    │
└─────────────────────────────────────────────┘
```

## Deployment Commands

\`\`\`bash
# 1. Load image
eval $(minikube docker-env)
docker-compose build
minikube image load distributed-rate-limiter-ratelimiter:latest
# 2. Deploy
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/redis-deployment.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/ratelimiter-deployment.yaml
kubectl apply -f k8s/hpa.yaml

# 3. Verify
kubectl get all -n ratelimiter
\`\`\`

## Features

- 3 replicas for high availability
- Auto-scaling (2-10 pods based on 70% CPU)
- Health checks (liveness & readiness probes)
- Resource limits (prevent resource exhaustion)
- LoadBalancer service for external access
- Redis for distributed state

## Manifests Created

- \`namespace.yaml\` - Isolated namespace
- \`configmap.yaml\` - Configuration
- \`redis-deployment.yaml\` - Redis StatefulSet
- \`ratelimiter-deployment.yaml\` - Rate limiter pods
- \`hpa.yaml\` - Horizontal Pod Autoscaler
