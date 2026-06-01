package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	CartsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_cart_carts_created_total",
		Help: "Total number of carts created",
	})

	CartsMerged = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_cart_carts_merged_total",
		Help: "Total number of cart merges",
	})

	ItemsAdded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_cart_items_added_total",
		Help: "Total number of items added to carts",
	})

	ItemsUpdated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_cart_items_updated_total",
		Help: "Total number of cart items updated",
	})

	ItemsRemoved = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_cart_items_removed_total",
		Help: "Total number of items removed from carts",
	})

	CheckoutPreviewsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_cart_checkout_previews_created_total",
		Help: "Total number of checkout previews generated",
	})

	IdempotentRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_cart_idempotent_requests_total",
		Help: "Total number of idempotent requests served from cache",
	})

	CartOperationLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_cart_operation_duration_seconds",
		Help: "Cart operation latency",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	}, []string{"operation"})

	MergeConflicts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_cart_merge_conflicts_total",
		Help: "Total number of cart merge conflicts",
	})

	CacheHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_cart_cache_hits_total",
		Help: "Total number of cache hits",
	}, []string{"cache_type", "provider"})

	CacheMissesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_cart_cache_misses_total",
		Help: "Total number of cache misses",
	}, []string{"cache_type", "provider"})
)
