package ratelimiter

import "time"

//Config holds rate limiter configuration
type Config struct{
	Capacity int32 //Maximum tokens in bucket
	RefillRate float64  //Tokens added per second
	RefillTime time.Duration 
}

func DefaultConfig() *Config{
	return &Config{
		Capacity: 100,  //100 token capacity
		RefillRate: 1.67, // ~100 requests per minute
		RefillTime: time.Millisecond * 100,//Refill every 100ms
	}
}

//Custom Configuartion
func NewConfig(capacity int32, refillRate float64) *Config{
	return &Config{
		Capacity:capacity,
		RefillRate: refillRate,
		RefillTime: time.Millisecond*100,
	}
}