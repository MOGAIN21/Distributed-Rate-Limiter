#!/bin/bash

echo "═══════════════════════════════════════════════════════════"
echo "  Distributed Rate Limiter - Performance Benchmark Suite"
echo "═══════════════════════════════════════════════════════════"
echo ""

# Serrvices are running check
echo "Checking if services are running..."
if ! docker ps | grep -q "ratelimiter"; then
    echo "Error: Rate Limiter service is not running. Please start the services before running the benchmarks: docker-compose up."
    exit 1
fi
echo "Services are running. Starting benchmarks..."
echo ""

#Create results directory
mkdir -p results
TIMESTAMP=$(date +"%Y%m%d-%H%M%S")
RESULT_FILE="results/benchmark-results-$TIMESTAMP.txt"

#Header
echo "Performance Benchmark Results - $(date)" > $RESULT_FILE
echo "==========================================" >> $RESULT_FILE
echo "" >> $RESULT_FILE

#Workload-1: Light Load
echo "Running Light-Load Test: 1000 requests/sec for 30 seconds..."
./scripts/load-tests/light-load.sh | tee -a $RESULT_FILE
echo "" >> $RESULT_FILE
sleep 5

#Workload-2: Medium Load
echo "Running Medium-Load Test: 5000 requests/sec for 30 seconds..."
./scripts/load-tests/medium-load.sh | tee -a $RESULT_FILE
echo "" >> $RESULT_FILE
sleep 5 

#Workload-3: Heavy Load
echo "Running Heavy-Load Test: 10,000 requests/sec for 30 seconds..."
./scripts/load-tests/heavy-load.sh | tee -a $RESULT_FILE
echo "" >> $RESULT_FILE
sleep 5

#Workload-4: Stress Test
echo "Running Stress Test: 15,000 requests/sec for 30 seconds..."
./scripts/load-tests/stress-test.sh | tee -a $RESULT_FILE
echo "" >> $RESULT_FILE

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "All benchmarks completed. Results saved to $RESULT_FILE"
echo "═══════════════════════════════════════════════════════════"

chmod +x scripts/run.sh