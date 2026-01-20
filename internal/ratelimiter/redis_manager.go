package ratelimiter

import (
	"fmt"
	"time"
	"log"
	"sync"
	"github.com/MKR-24/distributed-rate-limiter/internal/storage"
)

//Manager manages rate limiters with Redis Backend
type RedisManager struct {
	storage      *storage.RedisStorage
	config 	     *Config
	localCache   sync.Map //Reduce Redis calls by caching token buckets locally
	ttl		     time.Duration
}

func NewRedisManager(config *Config, redisAddr string, redisPassword string, redisDB int) (*RedisManager, error) {
	redisStorage, err := storage.NewRedisStorage(redisAddr, redisPassword, redisDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis storage: %w", err)
	}

	log.Printf("Connected to Redis at %s", redisAddr)

	return &RedisManager{
		storage:    redisStorage,
		config:     config,
		ttl:        10 * time.Minute, //Cache TTL
	}, nil
}

func (rm *RedisManager) CheckLimit(clientID string, tokensRequested int32) (bool, int32, int64) {
	if tokensRequested <= 0 {
		tokensRequested = 1
	}
	tokens, lastRefill, err := rm.storage.GetTokenBucket(clientID, rm.config.Capacity, rm.config.RefillRate)
	if err != nil {
		log.Printf("Error retrieving token bucket for client %s: %v", clientID, err)
		return true, rm.config.Capacity, 0
	}

	now := time.Now().UnixMilli()
	elapsed := float64(now - lastRefill)/1000.0
	tokensToAdd := elapsed * rm.config.RefillRate
	tokens=minFloat(tokens + tokensToAdd, float64(rm.config.Capacity))

	if tokens >= float64(tokensRequested) {
		tokens -= float64(tokensRequested)
		err = rm.storage.UpdateTokenBucket(clientID, tokens, now, rm.ttl); if err!= nil {
			log.Printf("Error updating token bucket for client %s: %v", clientID, err)
		}
		return true, int32(tokens), 0
	}

	tokensNeeded := float64(tokensRequested) - tokens
	retryAfterSeconds := tokensNeeded / rm.config.RefillRate
	retryAfterMs := int64(retryAfterSeconds * 1000)
	return false, int32(tokens), retryAfterMs
}

func (rm *RedisManager) GetStatus(clientID string) (int32, int32, float64){
	tokens, lastRefill, err := rm.storage.GetTokenBucket(clientID, rm.config.Capacity, rm.config.RefillRate)
	if err != nil {
		log.Printf("Error retrieving token bucket status for client %s: %v", clientID, err)
		return rm.config.Capacity, rm.config.Capacity, rm.config.RefillRate
	}
	now := time.Now().UnixMilli()
	elapsed := float64(now - lastRefill)/1000.0
	tokensToAdd := elapsed * rm.config.RefillRate
	tokens=minFloat(tokens + tokensToAdd, float64(rm.config.Capacity))

	return int32(tokens), rm.config.Capacity, rm.config.RefillRate
}

func (rm *RedisManager) GetClientCount() int {
	count, err := rm.storage.GetClientCount()
	if err != nil {
		log.Printf("Error retrieving client count: %v", err)
		return 0
	}
	return count
}

func (rm *RedisManager) Close() error{
	return rm.storage.Close()
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}