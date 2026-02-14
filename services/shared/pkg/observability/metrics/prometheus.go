package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Стандартные метрики HTTP
var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method", "path"},
	)

	HTTPRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
		[]string{"service"},
	)
)

// Database метрики
var (
	DBQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"service", "database", "operation", "status"},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.0001, .0005, .001, .005, .01, .05, .1, .5, 1, 5},
		},
		[]string{"service", "database", "operation"},
	)
)

// Cache метрики
var (
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"service", "cache_name"},
	)

	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"service", "cache_name"},
	)
)

// RecordHTTPMetrics записывает метрики HTTP запроса
func RecordHTTPMetrics(serviceName, method, path, status string, duration time.Duration) {
	HTTPRequestsTotal.WithLabelValues(serviceName, method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(serviceName, method, path).Observe(duration.Seconds())
}

// RecordDBMetrics записывает метрики запроса к БД
func RecordDBMetrics(serviceName, database, operation, status string, duration time.Duration) {
	DBQueriesTotal.WithLabelValues(serviceName, database, operation, status).Inc()
	DBQueryDuration.WithLabelValues(serviceName, database, operation).Observe(duration.Seconds())
}
