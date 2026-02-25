# Kubernetes Deployment Guide

## Overview

The distributed rate limiter is designed for production Kubernetes deployment with high availability, auto-scaling, and distributed state management. The deployment architecture supports horizontal scaling, automatic failover, and zero-downtime updates.

## Architecture
```
┌─────────────────────────────────────────────────────┐
│              Kubernetes Cluster                     │
│                                                     │
│  ┌───────────────────────────────────────────────┐  │
│  │         Namespace: ratelimiter                │  │
│  │                                               │  │
│  │  ┌─────────────────────────────────────────┐  │  │
│  │  │      LoadBalancer Service               │  │  │
│  │  │   (External Access: gRPC + Metrics)     │  │  │
│  │  └─────────────┬───────────────────────────┘  │  │
│  │                │                              │  │
│  │    ┌───────────┼───────────────┐              │  │
│  │    │           │               │              │  │
│  │  ┌─▼──┐     ┌─▼──┐         ┌─▼──┐             │  │
│  │  │Pod1│     │Pod2│   ...   │PodN│             │  │
│  │  │RL  │     │RL  │         │RL  │             │  │
│  │  └─┬──┘     └─┬──┘         └─┬──┘             │  │
│  │    │          │              │                │  │
│  │    └──────────┼──────────────┘                │  │
│  │               │                               │  │
│  │         ┌─────▼──────┐                        │  │
│  │         │   Redis    │                        │  │
│  │         │  Service   │                        │  │ 
│  │         └────────────┘                        │  │
│  │                                               │  │
│  │  HPA: Auto-scales 2-10 pods (70% CPU)         │  │
│  └──────────────────────────────────────────────-┘  │
└─────────────────────────────────────────────────────┘
```

## Components

### 1. Rate Limiter Deployment
- **Replicas**: 3 (configurable via HPA: 2-10)
- **Image**: distributed-rate-limiter-ratelimiter:latest
- **Resources**:
  - Requests: 100m CPU, 128Mi memory
  - Limits: 500m CPU, 256Mi memory
- **Probes**:
  - Liveness: HTTP GET /health (port 8080)
  - Readiness: HTTP GET /health (port 8080)

### 2. Redis Deployment
- **Replicas**: 1 (single instance, can be upgraded to StatefulSet for persistence)
- **Image**: redis:7-alpine
- **Resources**:
  - Requests: 100m CPU, 128Mi memory
  - Limits: 200m CPU, 256Mi memory

### 3. Horizontal Pod Autoscaler
- **Minimum Replicas**: 2
- **Maximum Replicas**: 10
- **Target CPU Utilization**: 70%
- **Scale-Up Policy**: Aggressive (100% increase or 2 pods per 30s)
- **Scale-Down Policy**: Conservative (50% decrease per 60s with 60s stabilization)

### 4. Services
- **ratelimiter-service** (LoadBalancer):
  - Port 50051: gRPC API
  - Port 8080: Metrics and health
- **redis-service** (ClusterIP):
  - Port 6379: Redis connection

## Deployment Instructions

### Prerequisites
```bash
# Kubernetes cluster with kubectl access
kubectl version --client

# For local testing with minikube
minikube start --cpus=4 --memory=4096
minikube addons enable metrics-server

# For local testing with kind
kind create cluster --name ratelimiter
```

### Step 1: Build and Load Container Image

**For minikube:**
```bash
eval $(minikube docker-env)
docker build -t distributed-rate-limiter-ratelimiter:latest -f Dockerfile .
docker images | grep ratelimiter
```

**For kind:**
```bash
docker build -t distributed-rate-limiter-ratelimiter:latest -f Dockerfile .
kind load docker-image distributed-rate-limiter-ratelimiter:latest --name ratelimiter
```

**For cloud providers (GKE, EKS, AKS):**
```bash
# Tag and push to container registry
docker tag distributed-rate-limiter-ratelimiter:latest gcr.io/PROJECT_ID/ratelimiter:v1.0.0
docker push gcr.io/PROJECT_ID/ratelimiter:v1.0.0

# Update kubernetes/ratelimiter-deployment.yaml to use the registry image
```

### Step 2: Deploy to Kubernetes
```bash
# Create namespace
kubectl apply -f kubernetes/namespace.yaml

# Deploy Redis
kubectl apply -f kubernetes/redis-deployment.yaml

# Wait for Redis to be ready
kubectl wait --for=condition=ready pod -l app=redis -n ratelimiter --timeout=60s

# Deploy ConfigMap
kubectl apply -f kubernetes/configmap.yaml

# Deploy Rate Limiter
kubectl apply -f kubernetes/ratelimiter-deployment.yaml

# Wait for Rate Limiter pods
kubectl wait --for=condition=ready pod -l app=ratelimiter -n ratelimiter --timeout=120s

# Deploy HPA
kubectl apply -f kubernetes/hpa.yaml
```

### Step 3: Verify Deployment
```bash
# Check all resources
kubectl get all -n ratelimiter

# Check HPA status
kubectl get hpa -n ratelimiter

# View pod logs
kubectl logs -n ratelimiter -l app=ratelimiter --tail=50

# Describe pods for detailed status
kubectl describe pod -n ratelimiter -l app=ratelimiter
```

### Step 4: Access the Service

**Local development (port forwarding):**
```bash
kubectl port-forward -n ratelimiter svc/ratelimiter-service 50051:50051
kubectl port-forward -n ratelimiter svc/ratelimiter-service 8080:8080
```

**Minikube (LoadBalancer):**
```bash
minikube service ratelimiter-service -n ratelimiter --url
```

**Cloud providers:**
```bash
# Get external IP
kubectl get svc ratelimiter-service -n ratelimiter
# Use EXTERNAL-IP for connections
```

## Testing Auto-Scaling

### Monitor HPA
```bash
# Terminal 1: Watch HPA in real-time
kubectl get hpa -n ratelimiter -w

# Terminal 2: Watch pod scaling
kubectl get pods -n ratelimiter -w
```

### Generate Load
```bash
# Using ghz for gRPC load testing
ghz --insecure \
  --call ratelimiter.RateLimiter/CheckLimit \
  -d '{"client_id":"loadtest-{{.RequestNumber}}", "tokens_requested":1}' \
  -c 200 \
  -n 100000 \
  --rps 5000 \
  localhost:50051

# Or using the test client
for i in {1..20}; do
  ./bin/client -client "k8s-test-$i" -requests 200 -interval 10ms &
done
```

Observe pods scaling from 3 to higher numbers as CPU utilization increases.

## Monitoring and Observability

### Prometheus Integration

The rate limiter exposes metrics on port 8080:
```bash
# View metrics endpoint
kubectl port-forward -n ratelimiter svc/ratelimiter-service 8080:8080
curl http://localhost:8080/metrics
```

Key metrics:
- `rate_limiter_request_total` - Total requests by client and outcome
- `rate_limiter_hits_total` - Total rate limit denials
- `rate_limiter_active_clients` - Number of tracked clients
- `rate_limiter_token_bucket_size` - Current token levels
- `rate_limiter_request_duration_seconds` - Request latency histogram

### Grafana Dashboards

Import dashboards from `monitoring/grafana/dashboards/` to visualize:
- Request throughput
- Rate limit hit rate
- Token bucket levels
- P50/P95/P99 latency
- Pod resource utilization

## Production Considerations

### High Availability

1. **Multi-Zone Deployment**: Spread pods across availability zones
2. **Pod Disruption Budgets**: Ensure minimum available pods during updates
3. **Anti-Affinity Rules**: Prevent multiple pods on same node
```yaml
# Add to deployment spec
spec:
  template:
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - ratelimiter
              topologyKey: kubernetes.io/hostname
```

### Redis High Availability

For production, upgrade to Redis Cluster or Redis Sentinel:
```yaml
# Use StatefulSet for Redis with persistent volumes
# Consider managed Redis (ElastiCache, Cloud Memorystore, Azure Cache)
```

### Security

1. **Network Policies**: Restrict pod-to-pod communication
2. **Pod Security Standards**: Enforce restricted PSS
3. **Secrets Management**: Use Kubernetes Secrets or external secret managers
4. **RBAC**: Implement least-privilege service accounts
5. **mTLS**: Enable mutual TLS for gRPC connections

### Resource Optimization

Based on load testing results:
- 500 req/s: 1 pod sufficient
- 2,000 req/s: 2-3 pods recommended
- 3,000+ req/s: 3-5 pods with potential Redis scaling

Adjust HPA thresholds based on your traffic patterns.

## Troubleshooting

### Pods Not Starting
```bash
# Check pod status
kubectl get pods -n ratelimiter

# Describe pod for events
kubectl describe pod <pod-name> -n ratelimiter

# Common issues:
# - ImagePullBackOff: Image not available or incorrect imagePullPolicy
# - CrashLoopBackOff: Application error, check logs
# - Pending: Insufficient resources or scheduling constraints
```

### Image Pull Errors
```bash
# Verify image exists in cluster (minikube/kind)
eval $(minikube docker-env)
docker images | grep ratelimiter

# For cloud: verify registry permissions
kubectl describe pod <pod-name> -n ratelimiter | grep -A 5 "Events:"
```

### Health Check Failures
```bash
# Check if health endpoint is responding
kubectl port-forward -n ratelimiter <pod-name> 8080:8080
curl http://localhost:8080/health

# View application logs
kubectl logs -n ratelimiter <pod-name> --tail=100
```

### Redis Connection Issues
```bash
# Verify Redis is running
kubectl get pods -n ratelimiter -l app=redis

# Test Redis connectivity from rate limiter pod
kubectl exec -it -n ratelimiter <ratelimiter-pod> -- sh
# Inside pod:
# apk add redis
# redis-cli -h redis-service ping
```

### HPA Not Scaling
```bash
# Check metrics-server is running
kubectl get deployment metrics-server -n kube-system

# Verify HPA can read metrics
kubectl get hpa -n ratelimiter
kubectl describe hpa ratelimiter-hpa -n ratelimiter

# Check pod resource utilization
kubectl top pods -n ratelimiter
```

## Updating the Deployment

### Rolling Update
```bash
# Update image version
kubectl set image deployment/ratelimiter \
  ratelimiter=distributed-rate-limiter-ratelimiter:v2.0.0 \
  -n ratelimiter

# Monitor rollout
kubectl rollout status deployment/ratelimiter -n ratelimiter

# Rollback if needed
kubectl rollout undo deployment/ratelimiter -n ratelimiter
```

### Configuration Updates
```bash
# Edit ConfigMap
kubectl edit configmap ratelimiter-config -n ratelimiter

# Restart pods to pick up new config
kubectl rollout restart deployment/ratelimiter -n ratelimiter
```

## Cleanup
```bash
# Delete all resources in namespace
kubectl delete namespace ratelimiter

# Or delete individual resources
kubectl delete -f kubernetes/
```

## Manifest Files

All Kubernetes manifests are located in the `k8s/` directory:

- `namespace.yaml` - Creates isolated namespace
- `configmap.yaml` - Application configuration
- `redis-deployment.yaml` - Redis deployment and service
- `ratelimiter-deployment.yaml` - Rate limiter deployment and LoadBalancer service
- `hpa.yaml` - Horizontal Pod Autoscaler configuration

## Performance in Kubernetes

Based on load testing, the Kubernetes deployment achieves:
- 3,000+ requests/second sustained throughput
- Sub-2ms average latency
- Linear scaling with additional pods
- Zero-downtime rolling updates
- Automatic recovery from pod failures


---

**Status**: Project Completed
