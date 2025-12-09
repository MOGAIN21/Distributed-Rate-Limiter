# Distributed Rate Limiter Service

A production-ready, horizontally scalable token-bucket rate limiter built with Go, gRPC, and Redis.

## Features

- 🚀 Token bucket algorithm for rate limiting
- ⚡ gRPC API for high-performance communication  
- 📦 Redis for distributed state management
- 🔄 Horizontally scalable architecture
- 🐳 Docker containerization
- ☸️ Kubernetes ready

## Tech Stack

- **Language**: Go
- **Protocol**: gRPC
- **Storage**: Redis
- **Containerization**: Docker
- **Orchestration**: Kubernetes

## Project Structure
```
distributed-rate-limiter/
├── cmd/
│   ├── server/          # gRPC server
│   └── client/          # Test client
├── internal/
│   ├── ratelimiter/     # Core rate limiting logic
│   └── storage/         # Redis integration
├── proto/               # Protocol buffer definitions
├── config/              # Configuration files
└── scripts/             # Build and deployment scripts
```

## Development Roadmap

- [x] Project structure
- [ ] Protocol buffer definitions
- [ ] Token bucket implementation
- [ ] Redis integration
- [ ] gRPC server
- [ ] Docker setup
- [ ] Unit tests
- [ ] Load testing
- [ ] Kubernetes deployment

## Author

Built by MKR-24 as part of SWE learning journey
