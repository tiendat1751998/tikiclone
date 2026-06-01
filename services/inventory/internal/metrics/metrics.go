package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ReservationLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "tiki_inventory_reservation_duration_seconds", Help: "Reservation latency",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})

	StockDeductionLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "tiki_inventory_stock_deduction_duration_seconds", Help: "Stock deduction latency",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})

	ReservationFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_inventory_reservation_failures_total", Help: "Reservation failures",
	}, []string{"reason"})

	IdempotentRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_inventory_idempotent_requests_total", Help: "Idempotent request deduplication count",
	})

	OversellPreventionCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_inventory_oversell_prevention_total", Help: "Oversell prevention count",
	})

	FlashSaleThrottlingCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_inventory_flash_sale_throttling_total", Help: "Flash sale throttling count",
	})

	ReconciliationFailures = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_inventory_reconciliation_failures_total", Help: "Reconciliation failures",
	})

	KafkaPublishLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "tiki_inventory_kafka_publish_duration_seconds", Help: "Kafka publish latency",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
	}, []string{"event_type"})

	KafkaPublishErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_inventory_kafka_publish_errors_total", Help: "Kafka publish errors",
	}, []string{"event_type"})
)
