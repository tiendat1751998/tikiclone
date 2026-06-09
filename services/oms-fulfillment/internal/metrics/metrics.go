package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	FulfillmentCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oms_fulfillment_created_total",
		Help: "Total number of fulfillments created",
	})

	FulfillmentStatusTransitions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "oms_fulfillment_status_transitions_total",
		Help: "Total number of fulfillment status transitions",
	}, []string{"from", "to"})

	ReturnRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oms_return_requests_total",
		Help: "Total number of return requests",
	})

	FulfillmentLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "oms_fulfillment_request_duration_seconds",
		Help:    "Request latency for fulfillment operations",
		Buckets: prometheus.DefBuckets,
	})

	KafkaPublishErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oms_kafka_publish_errors_total",
		Help: "Total number of kafka publish errors",
	})
)
