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

//RateLimiterService defines the interface for both in-memory and Redis backed rate limiters
type RateLimiterService interface {
	CheckLimit(clientID string, tokensRequested int32) (bool, int32, int64)
	GetStatus(clientID string) (int32, int32, float64)
	GetClientCount() int
}

// server is used to implement the RateLimiter grpc Service
type server struct {
	pb.UnimplementedRateLimiterServer
	limiter RateLimiterService
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
	allowed, remaining, retryAfterMs := s.limiter.CheckLimit(req.ClientId, tokensRequested)

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
	log.Printf("[CheckLimit] client=%s, tokens=%d, allowed=%v, remaining=%d",
		req.ClientId, tokensRequested, allowed, remaining)

	return response, nil
}

func (s *server) GetStatus(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
	if req.ClientId == "" {
		return nil, status.Error(codes.InvalidArgument, "ClientId is required")
	}
	remaining, capacity , refillRate := s.limiter.GetStatus(req.ClientId)

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
	var limiter RateLimiterService
	var redisManager *ratelimiter.RedisManager

	if config.Redis.Enabled {
		redisAddr := fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port)
		redisManager, err = ratelimiter.NewRedisManager(rlConfig, redisAddr,"", 0)
		if err != nil {
			log.Printf("Failed to connect to Redis: %v", err)
			log.Println("Falling back to in-memory rate limiter")
			limiter = ratelimiter.NewManager(rlConfig)
		} else{
			log.Printf("Using Redis-backed rate limiter")
			limiter = redisManager
			defer redisManager.Close()
		}
	} else {
		log.Printf("Using in-memory rate limiter")
		limiter = ratelimiter.NewManager(rlConfig)
	}

	// Set up gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterRateLimiterServer(grpcServer, &server{limiter: limiter})
	// Enable reflection for debugging with grpcurl
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