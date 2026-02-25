# Distributed Rate Limiter Service

A production-ready, horizontally scalable distributed rate limiter built with Go, gRPC, and Redis. Implements the token bucket algorithm for fair and efficient API rate limiting with distributed state coordination.

## Overview

This rate limiter service provides high-performance, distributed rate limiting for protecting APIs and services from abuse. It achieves 3,000+ requests/second with sub-2ms latency while maintaining accurate rate limits across multiple server instances using Redis for state coordination.

## Key Features

- **Token Bucket Algorithm**: Fair rate limiting with burst handling
- **Distributed Architecture**: Redis-backed state sharing across multiple instances
- **High Performance**: 3,000+ req/s with p99 latency under 20ms
- **gRPC API**: Efficient binary protocol for low-latency communication
- **Auto-Scaling**: Kubernetes HPA for automatic horizontal scaling
- **Production Monitoring**: Prometheus metrics and Grafana dashboards
- **Docker Ready**: Complete containerization with Docker Compose
- **Graceful Degradation**: Automatic fallback to in-memory mode if Redis unavailable

## Architecture
```
┌─────────────┐
│   Clients   │
└──────┬──────┘
       │ gRPC
       ▼
┌──────────────────────────────────────┐
│     Rate Limiter Instances (N)       │
│  ┌────────┐ ┌────────┐  ┌────────┐   │
│  │Server 1│ │Server 2│  │Server N│   │
│  └────┬───┘ └────┬───┘  └────┬───┘   │
│       │          │           │       │
└───────┼──────────┼───────────┼──────-┘
        │          │           │
        └──────────┼───────────┘
                   │
            ┌──────▼───────┐
            │  Redis       │
            │  (Shared     │
            │   State)     │
            └──────────────┘
```

## Performance Benchmarks

Based on comprehensive load testing:

| Metric | Light (500/s) | Medium (2K/s) | Heavy (3K/s) |
|--------|---------------|---------------|--------------|
| Throughput | 500 req/s | 2,000 req/s | 3,000 req/s |
| Avg Latency | 0.98 ms | 1.83 ms | 1.57 ms |
| p95 Latency | 1.39 ms | 5.13 ms | 3.18 ms |
| p99 Latency | 1.77 ms | 18.71 ms | 12.00 ms |
| Success Rate | 100% | 100% | 100% |

See [PERFORMANCE.md](PERFORMANCE.md) for detailed benchmark results.

## Quick Start

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Redis (for distributed mode)
- Protocol Buffers compiler

### Installation
```bash
# Clone repository
git clone https://github.com/MKR-24/distributed-rate-limiter.git
cd distributed-rate-limiter

# Install dependencies
go mod download

# Generate protobuf code
make proto

# Build binaries
make build
```

### Running with Docker Compose
```bash
# Start all services (Redis, Rate Limiter, Prometheus, Grafana)
docker-compose up -d

# View logs
docker-compose logs -f ratelimiter

# Test the service
./bin/client -requests 10

# Access Grafana dashboards
open http://localhost:3000  # admin/admin
```

### Running Locally
```bash
# Start Redis
redis-server

# Run server
./bin/server

# In another terminal, run client
./bin/client -client my-app -requests 100
```

## Project Structure
```
distributed-rate-limiter/
├── cmd/
│   ├── server/              # gRPC server entrypoint
│   └── client/              # Test client application
├── internal/
│   ├── ratelimiter/         # Core rate limiting logic
│   │   ├── limiter.go       # Token bucket implementation
│   │   ├── manager.go       # Multi-client coordination
│   │   ├── redis_manager.go # Redis-backed manager
│   │   └── config.go        # Configuration structures
│   ├── storage/
│   │   └── redis.go         # Redis storage layer
│   └── metrics/
│       └── metrics.go       # Prometheus metrics
├── proto/
│   └── ratelimiter.proto    # gRPC service definition
├── kubernetes/              # Kubernetes manifests
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── hpa.yaml
│   └── configmap.yaml
├── monitoring/
│   ├── prometheus.yml       # Prometheus configuration
│   └── grafana/             # Grafana dashboards
├── scripts/
│   ├── load-tests/          # Load testing scripts
│   └── k8s-deploy.sh        # Kubernetes deployment script
├── config/
│   └── config.yaml          # Application configuration
├── Dockerfile               # Container image definition
├── docker-compose.yml       # Multi-service orchestration
├── Makefile                 # Build automation
├── PERFORMANCE.md           # Detailed benchmark results
└── KUBERNETES.md            # Kubernetes deployment guide
```

## API Documentation

### gRPC Service Definition
```protobuf
service RateLimiter {
  rpc CheckLimit(CheckLimitRequest) returns (CheckLimitResponse);
  rpc GetStatus(GetStatusRequest) returns (GetStatusResponse);
}
```

### CheckLimit

Check if a request should be allowed based on rate limits.

**Request:**
```json
{
  "client_id": "user-123",
  "tokens_requested": 1
}
```

**Response:**
```json
{
  "allowed": true,
  "remaining_tokens": 95,
  "retry_after_ms": 0,
  "message": "Request allowed. 95 tokens remaining."
}
```

### GetStatus

Retrieve current rate limit status for a client.

**Request:**
```json
{
  "client_id": "user-123"
}
```

**Response:**
```json
{
  "remaining_tokens": 95,
  "capacity": 100,
  "refill_rate": 1.67,
  "next_refill_ms": 0
}
```

## Configuration

Edit `config/config.yaml`:
```yaml
server:
  port: 50051
  host: "0.0.0.0"

ratelimiter:
  default_capacity: 100      # Maximum tokens in bucket
  default_refill_rate: 1.67  # Tokens per second (~100/min)

redis:
  enabled: true              # Use distributed mode
  host: "localhost"
  port: 6379
```

## Development

### Running Tests
```bash
# Unit tests
make test

# With coverage
make test-coverage

# Load testing
make load-test-light
make load-test-medium
make load-test-heavy
```

### Building
```bash
# Build binaries
make build

# Build Docker image
make docker-build

# Generate protobuf code
make proto

# Format code
make fmt
```

## Deployment

### Docker
```bash
# Build and start all services
make docker-up

# View metrics at http://localhost:8080/metrics
# Access Grafana at http://localhost:3000
```

### Kubernetes
```bash
# Deploy to Kubernetes cluster
make k8s-deploy

# Check status
kubectl get all -n ratelimiter

# View logs
kubectl logs -n ratelimiter -l app=ratelimiter

# Access with port forwarding
kubectl port-forward -n ratelimiter svc/ratelimiter-service 50051:50051
```

See [KUBERNETES.md](KUBERNETES.md) for comprehensive deployment guide.

## Monitoring

### Prometheus Metrics

Available at `http://localhost:8080/metrics`:

- `rate_limiter_request_total` - Total requests by client and outcome
- `rate_limiter_hits_total` - Total rate limit denials
- `rate_limiter_active_clients` - Number of active clients
- `rate_limiter_token_bucket_size` - Current token levels per client
- `rate_limiter_request_duration_seconds` - Request latency histogram

### Grafana Dashboards

Pre-configured dashboards showing:
- Request throughput (requests/second)
- Rate limit hit rate
- Token bucket levels over time
- Request latency (p50, p95, p99)
- Active client count

## Load Testing
```bash
# Light load: 500 req/s
./scripts/load-tests/light-load.sh

# Medium load: 2,000 req/s
./scripts/load-tests/medium-load.sh

# Heavy load: 3,000 req/s
./scripts/load-tests/heavy-load.sh

# Run complete benchmark suite
./scripts/run-all-tests.sh
```

Results are saved to `results/benchmark_<timestamp>.txt`

## How It Works

### Token Bucket Algorithm

Each client gets a bucket with a maximum capacity of tokens. Tokens are consumed with each request and refill at a constant rate.
```
Initial: [**********] 100 tokens
Request: [*********]  99 tokens (1 consumed)
Wait:    [**********] 100 tokens (refilled)
Burst:   [*****]      50 tokens (50 consumed rapidly)
Refill:  [*******]    70 tokens (20 added over time)
```

When tokens reach zero, requests are denied until tokens refill.

### Distributed Coordination

Multiple rate limiter instances share state via Redis:

1. Request arrives at any instance
2. Instance fetches current token count from Redis
3. If tokens available, decrements count and allows request
4. If no tokens, request is denied
5. Tokens refill automatically based on elapsed time

This ensures consistent rate limiting across all instances.

## Production Considerations

### High Availability

- Deploy multiple rate limiter instances
- Use Redis Cluster for distributed state
- Configure health checks and readiness probes
- Implement circuit breakers for Redis failures

### Security

- Implement mutual TLS for gRPC connections
- Use network policies in Kubernetes
- Secure Redis with authentication
- Rate limit the rate limiter itself

### Scaling

- Horizontal: Add more rate limiter instances
- Vertical: Increase Redis resources
- Auto-scaling: Use Kubernetes HPA based on CPU/memory
- Expected: Linear scaling to 15,000+ req/s with proper infrastructure

## Troubleshooting

**Pods not starting in Kubernetes:**
```bash
kubectl describe pod -n ratelimiter <pod-name>
kubectl logs -n ratelimiter <pod-name>
```

**Redis connection failures:**
- Verify Redis is running: `redis-cli ping`
- Check network connectivity
- Review firewall rules
- Service automatically falls back to in-memory mode

**High latency:**
- Check Redis performance
- Monitor CPU/memory usage
- Review network latency
- Consider adding more instances

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request


## Acknowledgments

Built using:
- Go programming language
- gRPC and Protocol Buffers
- Redis for distributed state
- Prometheus and Grafana for monitoring
- Docker and Kubernetes for containerization

## Contact

**Author**: MKR-24  
**Repository**: https://github.com/MKR-24/distributed-rate-limiter  
**Issues**: https://github.com/MKR-24/distributed-rate-limiter/issues

---

Built as part of a systems engineering learning project demonstrating distributed systems, high-performance computing, and production engineering principles.
