#!/bin/bash

echo "☸️  Deploying to Kubernetes..."
echo ""

# Check if cluster is running
if ! kubectl cluster-info &>/dev/null; then
    echo " Kubernetes cluster not found!"
    echo "Start with: minikube start"
    exit 1
fi

echo "Cluster found"
echo ""

# Build and load image
echo "📦 Building Docker image..."
docker-compose build ratelimiter

echo "📤 Loading image into cluster..."
if command -v minikube &>/dev/null; then
    eval $(minikube docker-env)
    docker-compose build ratelimiter
elif command -v kind &>/dev/null; then
    kind load docker-image distributed-rate-limiter-ratelimiter:latest
fi

echo ""
echo "🚀 Deploying to Kubernetes..."

# Apply manifests
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/redis-deployment.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/ratelimiter-deployment.yaml
kubectl apply -f k8s/hpa.yaml

echo ""
echo "⏳ Waiting for pods..."
kubectl wait --for=condition=ready pod -l app=ratelimiter -n ratelimiter --timeout=120s

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Status:"
kubectl get all -n ratelimiter

echo ""
echo "🔗 Access the service:"
echo "   kubectl port-forward -n ratelimiter svc/ratelimiter-service 50051:50051"
echo ""
echo "📈 Watch with k9s:"
echo "   k9s -n ratelimiter"
