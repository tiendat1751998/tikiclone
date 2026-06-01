package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	CheckoutsInitiated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_checkout_initiated_total",
		Help: "Total checkouts initiated",
	})

	CheckoutsCompleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_checkout_completed_total",
		Help: "Total checkouts completed",
	})

	CheckoutsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_checkout_failed_total",
		Help: "Total checkouts failed",
	})

	CheckoutRetries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_checkout_retries_total",
		Help: "Total checkout retries",
	})

	IdempotentRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_checkout_idempotent_requests_total",
		Help: "Total idempotent requests served from cache",
	})

	CheckoutLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_checkout_step_duration_seconds",
		Help: "Checkout step latency",
		Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"step"})
)
