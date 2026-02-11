# Docker Deployment

## Quick Start
```bash
# Build and start all services
docker-compose up

# Or run in background
docker-compose up -d
```

## Services

- **Rate Limiter Server**: `localhost:50051` (gRPC)
- **Redis**: `localhost:6379`

## Commands
```bash
# Build images
docker-compose build

# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Clean everything (including data)
docker-compose down -v
```

## Environment Variables

Override in `docker-compose.yml`:
```yaml
environment:
  - REDIS_HOST=redis
  - REDIS_PORT=6379
  - REDIS_ENABLED=true
```

## Accessing Redis
```bash
# Connect to Redis CLI
docker exec -it ratelimiter-redis redis-cli

# View rate limit data
docker exec -it ratelimiter-redis redis-cli KEYS "ratelimit:*"
```

## Testing
```bash
# Build client locally
make build

# Test against Docker server
./bin/client -server localhost:50051 -requests 10
```

## Production Notes

- Data persists in Docker volume `redis-data`
- Health checks ensure Redis is ready before starting server
- Automatic restart on failure
