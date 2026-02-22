#!/bin/bash

echo "Heavy-Load Test: 3,000 requests/sec"
echo "Duration: 20 seconds"
echo "Total: 60,000 requests"
echo ""

ghz --insecure \
    --proto proto/ratelimiter.proto \
    --call ratelimiter.RateLimiter/CheckLimit \
    -d '{"client_id": "load-test-{{.RequestNumber}}", "tokens_requested": 1}' \
    -c 75 \
    -n 60000 \
    --rps 3000 \
    localhost:50051

echo ""
echo "Heavy-Load Test Completed"

chmod +x scripts/load-tests/heavy-load.sh