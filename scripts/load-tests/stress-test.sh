#!/bin/bash

echo "Stress Test"
echo "Duration: 30 seconds"
echo "Total: 15,000 requests/sec"
echo ""

ghz --insecure \
    --proto proto/ratelimiter.proto \
    --call ratelimiter.RateLimiter/CheckLimit \
    -d '{"client_id": "load-test-{{.RequestNumber}}", "tokens_requested": 1}' \
    -c 300 \
    -n 450000 \
    --rps 15000 \
    localhost:50051

echo ""
echo "Stress Test Completed"

chmod +x scripts/load-tests/stress-test.sh