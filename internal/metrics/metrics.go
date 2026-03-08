package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	//RequestCount counts the total number of requests received by the rate limiter
	RequestTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "rate_limiter_request_total",
		Help: "Total number of requests received by the rate limiter",
	},
	[]string{"allowed"},
	)

	//RateLimit Hits
	RateLimitHitsTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "rate_limiter_hits_total",
		Help: "Total number of requests that were rate limited",
	},
	)
	//Active clients gauge
	ActiveClients = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "rate_limiter_active_clients",
		Help: "Current number of active clients being tracked by the rate limiter",
	},
	)

	//Token bucket metrics
	TokenBucketSize = promauto.NewHistogram(
	prometheus.HistogramOpts{
		Name: "rate_limiter_token_bucket_size",
		Help: "Distribution of the token bucket sizes for each client",
		Buckets: []float64{0,10,25,50,75,100},
	},
	)

	//Request duration histogram
	RequestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "rate_limiter_request_duration_seconds",
		Help: "Duration of request processing in seconds",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method"},
	)
	// Redis operations
	RedisOperationsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "rate_limiter_redis_operations_total",
		Help: "Total number of Redis operations performed by the rate limiter",
	},
	[]string{"operation", "status"},
	)
)