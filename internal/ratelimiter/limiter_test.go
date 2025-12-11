package ratelimiter
import (
	"testing"
	"time"
)

func TestTokenBucket_Allow( t *testing.T) {
	bucket:= NewTokenBucket(10,5.0)
	for i:=0 ; i<10;i++{
		allowed, remaining, _ := bucket.Allow(1)
		if !allowed{
			t.Errorf("Request %d should be allowed", i)
		}
		if int(remaining) != 9-i {
			t.Errorf("Expected %d tokens, got %d", 9-i, remaining)
		}
	}

	// 11th request should be denied
	allowed, _, retryAfter := bucket.Allow(1)
	if allowed {
		t.Error("Request should be rate limited")
	}
	if retryAfter == 0 {
		t.Error("Should have retry-after time")
	}
	
	// Wait for refill (200ms = 1 token at 5 tokens/sec)
	time.Sleep(200 * time.Millisecond)
	
	// Should allow 1 request now
	allowed, _, _ = bucket.Allow(1)
	if !allowed {
		t.Error("Request should be allowed after refill")
	}

}

func TestManager_MultipleClients(t *testing.T) {
	config := NewConfig(5, 10.0) // 5 tokens, 10/sec
	manager := NewManager(config)
	
	// Client A makes 5 requests
	for i := 0; i < 5; i++ {
		allowed, _, _ := manager.CheckLimit("client-a", 1)
		if !allowed {
			t.Errorf("Client A request %d should be allowed", i)
		}
	}
	
	// Client A's 6th request should fail
	allowed, _, _ := manager.CheckLimit("client-a", 1)
	if allowed {
		t.Error("Client A should be rate limited")
	}
	
	// Client B should still have full capacity
	allowed, remaining, _ := manager.CheckLimit("client-b", 1)
	if !allowed {
		t.Error("Client B should be allowed")
	}
	if remaining != 4 {
		t.Errorf("Client B should have 4 tokens, got %d", remaining)
	}
	
	// Check tracking of 2 clients
	if manager.GetClientCount() != 2 {
		t.Errorf("Expected 2 clients, got %d", manager.GetClientCount())
	}
}