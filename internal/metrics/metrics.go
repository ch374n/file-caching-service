package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Cache metrics
	CacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	CacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	CacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_operation_duration_seconds",
			Help:    "Cache operation duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation"},
	)

	// R2 metrics
	R2RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "r2_requests_total",
			Help: "Total number of R2 requests",
		},
		[]string{"operation", "status"},
	)

	R2RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "r2_request_duration_seconds",
			Help:    "R2 request duration in seconds",
			Buckets: []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"operation"},
	)
)
