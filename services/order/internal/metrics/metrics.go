package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OrdersCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_orders_created_total",
		Help: "Total number of orders created",
	}, []string{"currency"})

	OrdersCancelledTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_orders_cancelled_total",
		Help: "Total number of orders cancelled",
	}, []string{"reason"})

	OrderCreationLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "tiki_order_creation_duration_seconds",
		Help:    "Order creation latency",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})

	OrderTransitionLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_order_transition_duration_seconds",
		Help:    "Order lifecycle transition latency",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	}, []string{"from_status", "to_status"})

	OrderCancellationLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "tiki_order_cancellation_duration_seconds",
		Help:    "Order cancellation latency",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})

	ReconciliationLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_order_reconciliation_duration_seconds",
		Help:    "Order reconciliation latency",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 12),
	}, []string{"reconciliation_type"})

	ReconciliationFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_order_reconciliation_failures_total",
		Help: "Total number of reconciliation failures",
	}, []string{"reconciliation_type"})

	TransitionRetries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_order_transition_retries_total",
		Help: "Total number of order transition retries",
	}, []string{"from_status", "to_status"})

	IdempotencyHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_order_idempotency_hits_total",
		Help: "Total number of idempotent request hits",
	})

	CacheHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_order_cache_hits_total",
		Help: "Total cache hits",
	}, []string{"operation"})

	CacheMisses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_order_cache_misses_total",
		Help: "Total cache misses",
	}, []string{"operation"})

	KafkaPublishLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_order_kafka_publish_duration_seconds",
		Help:    "Kafka event publish latency",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
	}, []string{"event_type"})

	KafkaPublishErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_order_kafka_publish_errors_total",
		Help: "Total Kafka publish errors",
	}, []string{"event_type"})

	ActiveOrdersByStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tiki_orders_active_by_status",
		Help: "Current number of active orders by status",
	}, []string{"status"})
)
