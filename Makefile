.PHONY:	proto	clean	build	run-server	run-client	test	deps	fmt

# Generating the protobuf code
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/ratelimiter.proto
	@echo "Protobuf code generated."

# Cleaning up generated files
clean:
	rm -f proto/*.pb.go
	rm -rf bin/
	@echo "Cleaned up generated files."

# Building the server and client binaries
build:
	mkdir -p bin
	go build -o bin/server cmd/server/main.go
	go build -o bin/client cmd/client/main.go
	@echo "Built server and client in bin."

# Running the server
run-server:
	go run cmd/server/main.go

# Running the client
run-client:
	go run cmd/client/main.go	-requests 10

# Run client with custom number of requests
client-test:
	go run cmd/client/main.go -client	user-123 -requests 150	-interval 50ms

# Running tests
test:
	go test -v	./...

#Run test with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out	-o coverage.html
	@echo "Test coverage report generated in coverage.html."

#Installing the necessary dependencies
deps:
	go mod download
	go mod tidy
	@echo "Dependencies installed."

#Format the code
fmt:
	go fmt ./...
	@echo "Code formatted."

#Run linter
lint:
	golangci-lint run

#Help command
help:
	@echo "Available commands:"
	@echo "  make proto          - Generate protobuf code"
	@echo "  make clean          - Clean up generated files"
	@echo "  make build          - Build server and client binaries"
	@echo "  make run-server     - Run the server"
	@echo "  make run-client     - Run the client"
	@echo "  make client-test    - Run client with custom number of requests"
	@echo "  make test           - Run tests"

#Docker commands
.PHONY:	docker-build	docker-up	docker-down	docker-logs	docker-clean

docker-build:
	docker-compose build
	@echo "Docker images built."

docker-up:
	docker-compose up -d
	@echo "Docker containers started."

docker-down:
	docker-compose down
	@echo "Docker containers stopped."

docker-logs:
	docker-compose logs -f

docker-clean:
	docker-compose down -v
	docker system prune -f
	@echo "Docker containers and volumes cleaned."

docker-restart:
	docker-compose restart
	@echo "Docker containers restarted."

#Monitoring commands
.PHONY:	monnitoring-up	monitoring-down	grafana	prometheus	metrics

monitoring-up:
	docker-compose  up -d	prometheus grafana
	@echo "Monitoring stack started."
	@echo "Access Grafana at http://localhost:3000 (default credentials: admin/admin)."
	@echo "Access Prometheus at http://localhost:9090."

monitoring-down:
	docker-compose stop prometheus grafana

grafana:
	@echo "Opening Grafana dashboard..."
	@(open http://localhost:3000 || xdg-open http://localhost:3000 || echo "Open http://localhost:3000 in your browser")	2>/dev/null

prometheus:
	@echo "Opening Prometheus..."
	@(open http://localhost:9090 || xdg-open http://localhost:9090 || echo "Open http://localhost:9090 in your browser")	2>/dev/null

metrics:
	@echo "Fetching metrics..."
	@curl -s http://localhost:8080/metrics | grep	rate_limiter	| grep -v	"^#"

#Load testing commands
.PHONY:	load-test-light	load-test-medium	load-test-heavy	load-test-stress	load-test-all

load-test-light:
	@./scripts/load-tests/light-load.sh

load-test-medium:
	@./scripts/load-tests/medium-load.sh

load-test-heavy:
	@./scripts/load-tests/heavy-load.sh

load-test-stress:
	@./scripts/load-tests/stress-test.sh

load-test-all:
	@./scripts/run.sh

#Help
.PHONY:	help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build           - Build server and client binaries"
	@echo "  make run-server      - Run the server locally"
	@echo "  make run-client      - Run test client"
	@echo "  make proto           - Generate protobuf code"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build    - Build Docker images"
	@echo "  make docker-up       - Start all services"
	@echo "  make docker-down     - Stop all services"
	@echo "  make docker-logs     - View logs"
	@echo "  make docker-clean    - Clean Docker resources"
	@echo ""
	@echo "Monitoring:"
	@echo "  make grafana         - Open Grafana dashboard"
	@echo "  make prometheus      - Open Prometheus"
	@echo "  make metrics         - View current metrics"
	@echo ""
	@echo "Load Testing:"
	@echo "  make load-test-light - Run light load test"
	@echo "  make load-test-medium- Run medium load test"
	@echo "  make load-test-heavy - Run heavy load test"
	@echo "  make load-test-all   - Run all tests"
	@echo ""
	@echo "Other:"
	@echo "  make test            - Run unit tests"
	@echo "  make fmt             - Format code"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make help            - Show this help"