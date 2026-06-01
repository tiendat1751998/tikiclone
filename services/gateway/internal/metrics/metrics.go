package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	GatewayRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_gateway_requests_total",
		Help: "Total number of requests processed by the gateway",
	}, []string{"service", "method", "path", "status", "auth"})

	GatewayRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_gateway_request_duration_seconds",
		Help:    "Request latency distributions for the gateway",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"service", "method", "path", "status"})

	GatewayUpstreamDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_gateway_upstream_duration_seconds",
		Help:    "Upstream service response latency",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"upstream_service", "method"})

	GatewayRateLimitHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_gateway_rate_limit_hits_total",
		Help: "Total number of rate limit hits",
	}, []string{"type", "key"})

	GatewayAuthErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_gateway_auth_errors_total",
		Help: "Total number of authentication errors",
	}, []string{"reason"})

	GatewayActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tiki_gateway_active_connections",
		Help: "Current number of active connections",
	})

	GatewayRequestBytesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_gateway_request_bytes_total",
		Help: "Total bytes received by the gateway",
	}, []string{"service", "method"})

	GatewayResponseBytesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_gateway_response_bytes_total",
		Help: "Total bytes sent by the gateway",
	}, []string{"service", "status"})

	GatewayUpstreamErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_gateway_upstream_errors_total",
		Help: "Total number of upstream service errors",
	}, []string{"service", "error_type"})

	GatewayTokenBlacklistCheck = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_gateway_token_blacklist_checks_total",
		Help: "Total number of token blacklist checks",
	}, []string{"result"})
)

func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
