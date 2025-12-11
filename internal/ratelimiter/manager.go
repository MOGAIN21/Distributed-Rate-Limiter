package ratelimiter

import (
	"sync"
)

//Manager manages rate limiters for multiple clients
type Manager struct{
	mu sync.RWMutex
	buckets map[string]*TokenBucket //clientID -> TokenBucket
	config *Config
}

// New Manager creates new rate limiter manager
func NewManager(config *Config) *Manager{
	if config == nil{
		config = DefaultConfig()
	}

	return &Manager{
		buckets: make(map[string]*TokenBucket),
		config:config,
	}
}

//Creating and getting Buckets
func (m *Manager) getOrCreateBucket(clientID string) *TokenBucket{
	m.mu.RLock()
	bucket, exists := m.buckets[clientID]
	m.mu.RUnlock()

	if exists{
		return bucket
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	
	//Just for confirmation double checking 
	if bucket, exists := m.buckets[clientID]; exists {
		return bucket
	}

	bucket = NewTokenBucket(m.config.Capacity, m.config.RefillRate)
	m.buckets[clientID] = bucket
	return bucket
}

//Check if client can make request
func (m *Manager) CheckLimit(clientID string, tokensRequested int32) (bool ,int32, int64){
	if tokensRequested <=0{
		tokensRequested = 1
	}

	bucket:= m.getOrCreateBucket(clientID)
	return bucket.Allow(tokensRequested)
}

// Current status of client bucket
func (m *Manager) GetStatus(clientID string) (int32,int32,float64){
	bucket:= m.getOrCreateBucket(clientID)
	return bucket.GetStatus()
}

//Number of clients
func (m *Manager) GetClientCount() int{
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.buckets)
}