package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ShipmentsCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_shipments_created_total", Help: "Total shipments created",
	}, []string{"carrier"})

	ShipmentTransitionLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "tiki_shipment_transition_duration_seconds", Help: "Shipment transition latency",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	}, []string{"from_status", "to_status"})

	TrackingSyncLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "tiki_shipment_tracking_sync_duration_seconds", Help: "Tracking sync latency",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 12),
	})

	WebhookReplayCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_shipment_webhook_replay_total", Help: "Webhook replay attacks detected",
	})

	CarrierAPILatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "tiki_shipment_carrier_api_duration_seconds", Help: "Carrier API latency",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 12),
	}, []string{"carrier"})

	ActiveShipments = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tiki_shipments_active_by_status", Help: "Active shipments by status",
	}, []string{"status"})

	KafkaPublishLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "tiki_shipment_kafka_publish_duration_seconds", Help: "Kafka publish latency",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
	}, []string{"event_type"})

	KafkaPublishErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_shipment_kafka_publish_errors_total", Help: "Kafka publish errors",
	}, []string{"event_type"})
)
