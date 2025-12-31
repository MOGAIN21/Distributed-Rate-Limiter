package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/MKR-24/distributed-rate-limiter/proto"
	"github.com/MKR-24/distributed-rate-limiter/internal/ratelimiter"
)

// server is used to implement the RateLimiter grpc Service
type server struct {
	pb.UnimplementedRateLimiterServer
	manager *ratelimiter.Manager
}

func (s *server) CheckLimit(ctx context.Context, req *pb.CheckLimitRequest) (*pb.CheckLimitResponse, error) {
	// Validate request
	if req.ClientId == "" {
		return nil, status.Error(codes.InvalidArgument, "ClientId is required")
	}

	tokensRequested := req.TokensRequested
	if tokensRequested <= 0 {
		tokensRequested = 1 // Default to 1 token if not specified
	}

	// Check rate limit
	allowed, remaining, retryAfterMs := s.manager.CheckLimit(req.ClientId, tokensRequested)

	response := &pb.CheckLimitResponse{
		Allowed:      allowed,
		RemainingTokens: remaining,
		RetryAfterMs: retryAfterMs,
	}

	if allowed{
		response.Message = fmt.Sprintf("Request allowed. %d tokens remaining.", remaining)
	} else {
		response.Message = fmt.Sprintf("Rate limit exceeded. Try again in %d ms.", retryAfterMs)
	}
	return response, nil
}

func (s *server) GetStatus(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
	if req.ClientId == "" {
		return nil, status.Error(codes.InvalidArgument, "ClientId is required")
	}
	remaining, capacity , refillRate := s.manager.GetStatus(req.ClientId)

	response := &pb.GetStatusResponse{
		RemainingTokens: remaining,
		Capacity:        capacity,
		RefillRate:      refillRate,
		NextRefillMs:    0,
	}

	log.Printf("Status for ClientID %s: Remaining=%d, Capacity=%d, RefillRate=%.2f", req.ClientId, remaining, capacity, refillRate)
	return response, nil
}

func main() {
	// Load configuration
	config, err := ratelimiter.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	//Create Rate Limiter Manager
	rlConfig := ratelimiter.NewConfig(
		config.RateLimiter.DefaultCapacity,
		config.RateLimiter.DefaultRefillRate,
	)
	manager := ratelimiter.NewManager(rlConfig)

	// Set up gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterRateLimiterServer(grpcServer, &server{manager: manager,})
	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Listen on specified host and port
	address := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", address, err)
	}

	go func(){
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
		grpcServer.GracefulStop()
	}()

	log.Printf("Starting gRPC server on %s", address)
	log.Printf("Rate Limiter Config - Capacity: %d, Refill Rate: %.2f tokens/sec", rlConfig.Capacity, rlConfig.RefillRate)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}

}