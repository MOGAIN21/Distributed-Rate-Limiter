.PHONY:	proto	clean	build	run-server	run-client	test

# Generating the protobuf code
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/ratelimiter.proto
	@echo "Protobuf code generated."

# Cleaning up generated files
clean:
	rm -f proto/*.pb.go
	@echo "Cleaned up generated files."

# Building the server and client binaries
build:
	go build -o bin/server cmd/server/main.go
	go build -o bin/client cmd/client/main.go
	@echo "Built server and client."

# Running the server
run-server:
	go run cmd/server/main.go

# Running the client
run-client:
	go run cmd/client/main.go

# Running tests
test:
	go test -v	./...

#Installing the necessary dependencies
deps:
	go mod download
	go mod tidy
	@echo "Dependencies installed."

#Format the code
fmt:
	go fmt ./...
	@echo "Code formatted."