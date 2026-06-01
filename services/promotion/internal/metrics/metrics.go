package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ValidationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_promotion_validations_total",
		Help: "Total voucher validations",
	}, []string{"result"})

	RedeemTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_promotion_redemptions_total",
		Help: "Total voucher redemptions",
	})

	EvaluationsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_promotion_evaluations_total",
		Help: "Total promotion evaluations",
	})

	IdempotentRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_promotion_idempotent_requests_total",
		Help: "Total idempotent requests served from cache",
	})

	ValidationLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "tiki_promotion_validation_duration_seconds",
		Help: "Voucher validation latency",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5},
	})

	AbuseDetections = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_promotion_abuse_detections_total",
		Help: "Total abuse detection triggers",
	}, []string{"type"})
)
