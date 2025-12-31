package main

import (
	"fmt"
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "github.com/MKR-24/distributed-rate-limiter/proto"
)

func main(){
	// Parse command-line flags
	serverAddr := flag.String("server", "localhost:50051", "The Server address")
	clientID := flag.String("client", "test-client", "Client_ID")
	requests := flag.Int("requests", 10, "Number of requests to send")
	interval := flag.Duration("interval", 100*time.Millisecond, "Interval between requests")
	flag.Parse()

	// Set up a connection to the server
	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewRateLimiterClient(conn)

	fmt.Printf("Client %s sending %d requests to %s with interval %v\n", *clientID, *requests, *serverAddr, *interval)

	//Get Status
	statusResp, err := client.GetStatus(context.Background(), &pb.GetStatusRequest{ClientId: *clientID})
	if err != nil {
		log.Fatalf("Failed to get status: %v", err)
	}else{
		fmt.Printf("Initial Status: %d/%d tokens (%.2f tokens/sec)\n\n",
			statusResp.RemainingTokens, statusResp.Capacity, statusResp.RefillRate)
	}

	//Make Requests
	allowed:= 0
	denied:= 0

	for i:=1 ; i<=*requests; i++{
		resp, err := client.CheckLimit(context.Background(),&pb.CheckLimitRequest{
			ClientId: *clientID,
			TokensRequested: 1,
		})
		if err != nil {
			log.Printf("Request %d: Error checking limit: %v", i, err)
			continue
		}
		if resp.Allowed {
			allowed++
			fmt.Printf("Request %d: Allowed | Remaining Tokens: %d tokens | %s\n", i, resp.RemainingTokens, resp.Message)
		}else{
			denied++
			fmt.Printf(" Request %d: DENIED   | Retry after: %d ms | %s\n",
				i, resp.RetryAfterMs, resp.Message)
		}
		time.Sleep(*interval)
	}
	//Final Status
	fmt.Printf("\n" + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Summary: %d allowed, %d denied\n", allowed, denied)
	
	statusResp, err = client.GetStatus(context.Background(), &pb.GetStatusRequest{
		ClientId: *clientID,
	})
	if err == nil {
		fmt.Printf("Final Status: %d/%d tokens remaining\n",
			statusResp.RemainingTokens, statusResp.Capacity)
	}
}