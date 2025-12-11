package ratelimiter

import (
	"sync"
	"time"
)

//TokenBucket => token bucket for single client
type TokenBucket struct{
	mu sync.Mutex
	tokens float64
	capacity int32
	refillRate float64
	lastRefill time.Time
}

// Create New Token Bucket
func NewTokenBucket(capacity int32, refillRate float64) *TokenBucket{
	return &TokenBucket{
		tokens: float64(capacity),
		capacity: capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

//Refill function based on time elapsed
func (tb *TokenBucket) refill(){
	now:= time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	//Calculate tokens to add
	tokensToAdd := elapsed* tb.refillRate

	//Adding tokens with limit
	tb.tokens = min(tb.tokens + tokensToAdd, float64(tb.capacity))
	tb.lastRefill=now
}

// Check when Requst can be made
func (tb *TokenBucket) Allow(tokensRequested int32) (bool, int32, int64){
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens
	tb.refill()

	// Check current tokens count
	if tb.tokens >= float64(tokensRequested){
		tb.tokens -= float64(tokensRequested)
		return true, int32(tb.tokens),0
	}

	//Not Enough Tokens
	tokensNeeded := float64(tokensRequested) - tb.tokens
	retryAfterSeconds:= tokensNeeded / tb.refillRate
	retryAfterMS := int64(retryAfterSeconds*1000)

	return false, int32(tb.tokens), retryAfterMS
}

//Bucket Status
func (tb *TokenBucket) GetStatus() (int32,int32, float64){
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	return int32(tb.tokens),tb.capacity, tb.refillRate
}

func min(a,b float64) float64{
	if a<b {
		return a
	}
	return b
}