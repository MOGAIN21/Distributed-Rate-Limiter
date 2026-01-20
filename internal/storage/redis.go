package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

//Redis Storage implements distributed token bucket storage
type RedisStorage struct {
	client *redis.Client
	ctx   context.Context
}

type TokenBucketState struct {
	Tokens float64
	LastRefill int64
}

//NewRedisStorage creates a new RedisStorage instance
func NewRedisStorage(addr string, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		Password: password,
		DB: db,
		DialTimeout: 5 * time.Second,
		ReadTimeout: 3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize: 10,
		MinIdleConns: 5,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{
		client: client,
		ctx: ctx,
	}, nil
}

func (r *RedisStorage) GetTokenBucket(clientID string, capacity int32, refillRate float64) (float64, int64, error) {
	key := fmt.Sprintf("ratelimit:%s", clientID)

	result, err := r.client.HGetAll(r.ctx, key).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get token bucket from Redis: %w", err)
	}

	// If no data exists, initialize with full capacity
	if len(result) == 0 {
		return float64(capacity), time.Now().UnixMilli(), nil
	}

	// Parse tokens
	tokens, err := strconv.ParseFloat(result["tokens"], 64)
	if err != nil {
		return float64(capacity), time.Now().UnixMilli(), nil
	}

	// Parse last_refill
	lastRefill, err := strconv.ParseInt(result["last_refill"], 10, 64)
	if err != nil {
		return tokens, time.Now().UnixMilli(), nil
	}

	return tokens, lastRefill, nil
}

// UpdateTokenBucket updates the token bucket in Redis
func (r *RedisStorage) UpdateTokenBucket(clientID string, tokens float64, lastRefill int64, ttl time.Duration) error {
	key := fmt.Sprintf("ratelimit:%s", clientID)

	// Use a pipeline for atomicity
	pipe:=r.client.Pipeline()
	pipe.HSet(r.ctx,key, "tokens",tokens)
	pipe.HSet(r.ctx,key, "last_refill",lastRefill)
	pipe.Expire(r.ctx,key, ttl)

	_, err := pipe.Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to update token bucket in Redis: %w", err)
	}

	return nil
}

//DeleteTokenBucket deletes the token bucket from Redis
func (r *RedisStorage) DeleteTokenBucket(clientID string) error{
	key := fmt.Sprintf("ratelimit:%s", clientID)
	return r.client.Del(r.ctx,key).Err()
}

//Close Redis connection
func (r *RedisStorage) Close() error {
	return r.client.Close()
}

//Ping Redis server
func (r *RedisStorage) Ping() error {
	return r.client.Ping(r.ctx).Err()
}

//GetClientCount returns number of clients stored in Redis
func (r *RedisStorage) GetClientCount() (int, error){
	keys, err := r.client.Keys(r.ctx, "ratelimit:*").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get client count from Redis: %w", err)
	}
	return len(keys), nil
}