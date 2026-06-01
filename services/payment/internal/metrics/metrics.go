package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	PaymentsAuthorizedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_payments_authorized_total",
		Help: "Total number of authorized payments",
	}, []string{"psp_provider", "payment_method"})

	PaymentsCapturedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_payments_captured_total",
		Help: "Total number of captured payments",
	}, []string{"psp_provider"})

	PaymentsFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_payments_failed_total",
		Help: "Total number of failed payments",
	}, []string{"reason"})

	PaymentAuthorizationLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_payment_authorization_duration_seconds",
		Help:    "Payment authorization latency",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 12),
	}, []string{"psp_provider"})

	PaymentCaptureLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_payment_capture_duration_seconds",
		Help:    "Payment capture latency",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 12),
	}, []string{"psp_provider"})

	WebhookLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_payment_webhook_duration_seconds",
		Help:    "Webhook processing latency",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
	}, []string{"psp_provider", "event_type"})

	ReconciliationLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "tiki_payment_reconciliation_duration_seconds",
		Help:    "Payment reconciliation latency",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 12),
	})

	DuplicatePreventionCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_payment_duplicate_prevention_total",
		Help: "Total number of duplicate payment attempts prevented",
	})

	PSPFailureRate = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_payment_psp_failures_total",
		Help: "Total number of PSP failures",
	}, []string{"psp_provider"})

	ReplayAttackCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_payment_replay_attacks_total",
		Help: "Total number of replay attacks detected",
	})

	RefundsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_payment_refunds_total",
		Help: "Total number of refunds processed",
	}, []string{"status"})

	FraudDetectedCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_payment_fraud_detected_total",
		Help: "Total number of fraud detection triggers",
	})

	ActivePayments = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tiki_payments_active_by_status",
		Help: "Current number of active payments by status",
	}, []string{"status"})
)

var (
	KafkaPublishLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_payment_kafka_publish_duration_seconds",
		Help:    "Kafka publish latency",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
	}, []string{"event_type"})

	KafkaPublishErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_payment_kafka_publish_errors_total",
		Help: "Kafka publish errors",
	}, []string{"event_type"})

	WebhookProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_payment_webhook_processed_total",
		Help: "Total webhooks processed",
	}, []string{"psp_provider", "event_type"})
)
