#!/bin/bash

echo "Medium-Load Test: 2000 requests/sec"
echo "Duration: 20 seconds"
echo "Total: 40,000 requests"
echo ""

ghz --insecure \
    --proto proto/ratelimiter.proto \
    --call ratelimiter.RateLimiter/CheckLimit \
    -d '{"client_id": "load-test-{{.RequestNumber}}", "tokens_requested": 1}' \
    -c 50 \
    --rps 2000 \
    --duration 60s \
    localhost:50051

echo ""
echo "Medium-Load Test Completed"

chmod +x scripts/load-tests/medium-load.sh