package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ProductsCreated = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_catalog_products_created_total", Help: "Total products created"})
	ProductsUpdated = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_catalog_products_updated_total", Help: "Total products updated"})
	SKUsCreated     = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_catalog_skus_created_total", Help: "Total SKUs created"})
	IdempotentRequests = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_catalog_idempotent_requests_total", Help: "Idempotent requests"})
	ProductsDeleted    = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_catalog_products_deleted_total", Help: "Total products deleted"})

	CacheHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_catalog_cache_hits_total", Help: "Cache hits by layer",
	}, []string{"service", "cache_layer"})

	CacheMissesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_catalog_cache_misses_total", Help: "Cache misses by layer",
	}, []string{"service", "cache_layer"})

	KafkaPublishLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "tiki_catalog_kafka_publish_duration_seconds", Help: "Kafka publish latency",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
	}, []string{"event_type"})

	KafkaPublishErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_catalog_kafka_publish_errors_total", Help: "Kafka publish errors",
	}, []string{"event_type"})
)
