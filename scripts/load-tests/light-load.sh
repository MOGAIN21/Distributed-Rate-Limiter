#!/bin/bash

echo "Light-Load Test: 500 requests/sec"
echo "Duration: 20 seconds"
echo "Total: 10,000 requests"
echo ""

ghz --insecure \
    --proto proto/ratelimiter.proto \
    --call ratelimiter.RateLimiter/CheckLimit \
    -d '{"client_id": "load-test-{{.RequestNumber}}", "tokens_requested": 1}' \
    -c 25 \
    -n 10000 \
    --rps 500 \
    localhost:50051

echo ""
echo "Light-Load Test Completed"

chmod +x scripts/load-tests/light-load.sh