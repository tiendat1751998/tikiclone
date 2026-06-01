package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_http_request_duration_seconds",
		Help:    "HTTP request latency distributions",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"service", "method", "path", "status"})

	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"service", "method", "path", "status"})

	HTTPRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tiki_http_requests_in_flight",
		Help: "Current number of in-flight HTTP requests",
	})

	GRPCRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_grpc_request_duration_seconds",
		Help:    "gRPC request latency distributions",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"service", "method", "status"})

	GRPCRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_grpc_requests_total",
		Help: "Total number of gRPC requests",
	}, []string{"service", "method", "status"})

	DatabaseQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_db_query_duration_seconds",
		Help:    "Database query latency distributions",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "operation", "table"})

	DatabaseQueriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_db_queries_total",
		Help: "Total number of database queries",
	}, []string{"service", "operation", "table", "status"})

	CacheHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_cache_hits_total",
		Help: "Total number of cache hits",
	}, []string{"service", "cache_type"})

	CacheMissesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_cache_misses_total",
		Help: "Total number of cache misses",
	}, []string{"service", "cache_type"})

	KafkaMessagesProduced = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_kafka_messages_produced_total",
		Help: "Total number of Kafka messages produced",
	}, []string{"service", "topic"})

	KafkaMessagesConsumed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_kafka_messages_consumed_total",
		Help: "Total number of Kafka messages consumed",
	}, []string{"service", "topic", "status"})

	BusinessErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_business_errors_total",
		Help: "Total number of business logic errors",
	}, []string{"service", "error_code"})
)

func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func ObserveHTTPMetrics(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		HTTPRequestsInFlight.Inc()
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath()

		HTTPRequestsTotal.WithLabelValues(service, method, path, status).Inc()
		HTTPRequestDuration.WithLabelValues(service, method, path, status).Observe(duration.Seconds())
		HTTPRequestsInFlight.Dec()
	}
}
