package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/gateway/internal/discovery"
	"github.com/tikiclone/tiki/services/gateway/internal/middleware"
	"github.com/tikiclone/tiki/services/gateway/internal/resilience"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	proxyRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_gateway_proxy_requests_total",
		Help: "Total proxy requests",
	}, []string{"service", "method", "status"})

	proxyRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_gateway_proxy_duration_seconds",
		Help:    "Proxy request duration",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"service", "method"})

	proxyRetriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_gateway_proxy_retries_total",
		Help: "Total proxy retries",
	}, []string{"service"})

	proxyCircuitBreakerState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tiki_gateway_circuit_breaker_state",
		Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
	}, []string{"service"})

	proxyErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_gateway_proxy_errors_total",
		Help: "Total proxy errors by type",
	}, []string{"service", "error_type"})
)

type Proxy struct {
	discovery       *discovery.ServiceDiscovery
	transport       *http.Transport
	circuitBreakers map[string]*resilience.CircuitBreaker
	retryConfig     map[string]resilience.RetryConfig
	cbMu            sync.RWMutex
	timeouts        map[string]time.Duration
	cachedProxies   sync.Map
}

type ProxyTarget struct {
	ServiceName string
	PathPrefix  string
	StripPrefix string
	Timeout     time.Duration
}

type ProxyOption struct {
	ServiceName    string
	CircuitBreaker resilience.CircuitBreakerOptions
	RetryConfig    resilience.RetryConfig
	Timeout        time.Duration
}

func NewProxy(
	svcDiscovery *discovery.ServiceDiscovery,
	maxIdleConns int,
	idleConnTimeout time.Duration,
) *Proxy {
	transport := &http.Transport{
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   maxIdleConns,
		IdleConnTimeout:       idleConnTimeout,
		DisableCompression:    false,
		MaxConnsPerHost:        maxIdleConns * 2,
		DisableKeepAlives:     false,
		ForceAttemptHTTP2:     false,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}
	transport.MaxResponseHeaderBytes = 1 << 20
	return &Proxy{
		discovery:       svcDiscovery,
		circuitBreakers: make(map[string]*resilience.CircuitBreaker),
		retryConfig:     make(map[string]resilience.RetryConfig),
		timeouts:        make(map[string]time.Duration),
		transport:       transport,
	}
}

func (p *Proxy) Configure(opts []ProxyOption) {
	for _, opt := range opts {
		cb := resilience.NewCircuitBreaker(opt.ServiceName, opt.CircuitBreaker)
		p.cbMu.Lock()
		p.circuitBreakers[opt.ServiceName] = cb
		p.retryConfig[opt.ServiceName] = opt.RetryConfig
		if opt.Timeout > 0 {
			p.timeouts[opt.ServiceName] = opt.Timeout
		}
		p.cbMu.Unlock()
		proxyCircuitBreakerState.WithLabelValues(opt.ServiceName).Set(float64(cb.State()))
	}
}

func (p *Proxy) getCircuitBreaker(service string) *resilience.CircuitBreaker {
	p.cbMu.RLock()
	defer p.cbMu.RUnlock()
	return p.circuitBreakers[service]
}

func (p *Proxy) getRetryConfig(service string) resilience.RetryConfig {
	p.cbMu.RLock()
	defer p.cbMu.RUnlock()
	if cfg, ok := p.retryConfig[service]; ok {
		return cfg
	}
	return resilience.DefaultRetryConfig()
}

func (p *Proxy) getTimeout(service string) time.Duration {
	p.cbMu.RLock()
	defer p.cbMu.RUnlock()
	if t, ok := p.timeouts[service]; ok {
		return t
	}
	return 10 * time.Second
}

func (p *Proxy) getOrCreateProxy(serviceName string) *httputil.ReverseProxy {
	if cached, ok := p.cachedProxies.Load(serviceName); ok {
		return cached.(*httputil.ReverseProxy)
	}
	proxy := &httputil.ReverseProxy{
		Transport: p.transport,
		Director:  func(req *http.Request) {},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			observability.GetLogger().Error("proxy error",
				zap.String("service", serviceName),
				zap.Error(err),
			)
			proxyErrorsTotal.WithLabelValues(serviceName, "PROXY_ERROR").Inc()
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(gin.H{
				"error_code": "BAD_GATEWAY",
				"message":    "upstream connection error",
			})
		},
	}
	p.cachedProxies.Store(serviceName, proxy)
	return proxy
}

func (p *Proxy) ReverseProxy(target *ProxyTarget) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := otel.Tracer("tiki-gateway").Start(c.Request.Context(),
			fmt.Sprintf("reverse_proxy.%s", target.ServiceName),
			trace.WithSpanKind(trace.SpanKindClient),
		)
		defer span.End()

		span.SetAttributes(
			attribute.String("upstream.service", target.ServiceName),
			attribute.String("upstream.path", c.Request.URL.Path),
		)

		instance, err := p.discovery.GetInstance(target.ServiceName)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "no healthy upstream instances")
			observability.GetLogger().Error("no healthy upstream instances",
				zap.String("service", target.ServiceName),
				zap.Error(err),
			)
			proxyErrorsTotal.WithLabelValues(target.ServiceName, "NO_HEALTHY_INSTANCE").Inc()
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error_code": "SERVICE_UNAVAILABLE",
				"message":    fmt.Sprintf("service %s is currently unavailable", target.ServiceName),
			})
			return
		}

		targetURL, _ := url.Parse(fmt.Sprintf("http://%s:%d", instance.Address, instance.Port))
		baseDirector := p.buildDirector(targetURL, target)

		req := c.Request
		ctx := c.Request.Context()
		if timeout := p.getTimeout(target.ServiceName); timeout > 0 {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			req = req.WithContext(ctx)
		}

		proxy := p.getOrCreateProxy(target.ServiceName)
		proxy.Director = func(r *http.Request) {
			baseDirector(r)
			p.injectGatewayHeaders(c, r)
		}

		recorder := &responseRecorder{ResponseWriter: c.Writer, statusCode: http.StatusOK}

		retryCfg := p.getRetryConfig(target.ServiceName)
		cb := p.getCircuitBreaker(target.ServiceName)

		start := time.Now()
		statusCode := http.StatusOK

		if cb != nil {
			err = cb.Execute(func() error {
				return p.doProxyRequest(c, recorder, req, proxy)
			})
		} else if retryCfg.MaxAttempts > 1 {
			err = resilience.DoWithRetry(c.Request.Context(), retryCfg, target.ServiceName, func(ctx context.Context) error {
				return p.doProxyRequest(c, recorder, req, proxy)
			})
		} else {
			err = p.doProxyRequest(c, recorder, req, proxy)
		}

		duration := time.Since(start)

		if err != nil {
			statusCode = http.StatusBadGateway
			if strings.Contains(err.Error(), "circuit breaker") {
				statusCode = http.StatusServiceUnavailable
			}
			proxyErrorsTotal.WithLabelValues(target.ServiceName, errorType(err)).Inc()
		} else {
			statusCode = c.Writer.Status()
		}

		proxyRequestsTotal.WithLabelValues(
			target.ServiceName,
			c.Request.Method,
			fmt.Sprintf("%d", statusCode),
		).Inc()
		proxyRequestDuration.WithLabelValues(target.ServiceName, c.Request.Method).Observe(duration.Seconds())

		if cb != nil {
			proxyCircuitBreakerState.WithLabelValues(target.ServiceName).Set(float64(cb.State()))
		}
	}
}

func (p *Proxy) buildDirector(upstreamURL *url.URL, target *ProxyTarget) func(req *http.Request) {
	return func(req *http.Request) {
		req.URL.Scheme = upstreamURL.Scheme
		req.URL.Host = upstreamURL.Host
		req.Host = upstreamURL.Host

		if target.StripPrefix != "" {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, target.StripPrefix)
			if !strings.HasPrefix(req.URL.Path, "/") {
				req.URL.Path = "/" + req.URL.Path
			}
		}

		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "Tiki-Clone-Gateway")
		}
	}
}

func (p *Proxy) doProxyRequest(c *gin.Context, recorder *responseRecorder, req *http.Request, proxy *httputil.ReverseProxy) error {
	proxy.ServeHTTP(recorder, req)

	select {
	case <-req.Context().Done():
		if req.Context().Err() == context.DeadlineExceeded {
			return fmt.Errorf("upstream timeout")
		}
		return req.Context().Err()
	default:
	}

	if recorder.statusCode >= 500 {
		return fmt.Errorf("upstream returned %d", recorder.statusCode)
	}

	return nil
}

func (p *Proxy) injectGatewayHeaders(c *gin.Context, req *http.Request) {
	if correlationID, exists := c.Get(string(middleware.CorrelationIDKey)); exists {
		req.Header.Set("X-Correlation-ID", fmt.Sprintf("%v", correlationID))
	}
	if requestID, exists := c.Get(string(middleware.RequestIDKey)); exists {
		req.Header.Set("X-Request-ID", fmt.Sprintf("%v", requestID))
	}
	if userID, exists := c.Get(string(middleware.UserIDKey)); exists {
		req.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))
	}
	if roles, exists := c.Get(string(middleware.UserRolesKey)); exists {
		if roleList, ok := roles.([]string); ok {
			req.Header.Set("X-User-Roles", strings.Join(roleList, ","))
		}
	}
	if deviceInfo, exists := c.Get(string(middleware.DeviceInfoKey)); exists {
		if info, ok := deviceInfo.(map[string]string); ok {
			for k, v := range info {
				req.Header.Set(fmt.Sprintf("X-Device-%s", k), v)
			}
		}
	}

	req.Header.Set("X-Forwarded-For", c.ClientIP())
	req.Header.Set("X-Forwarded-Proto", c.Request.URL.Scheme)
	req.Header.Set("X-Forwarded-Host", c.Request.Host)

	traceID := trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()
	if traceID != "" {
		req.Header.Set("X-Trace-ID", traceID)
	}
}

func (p *Proxy) HealthCheck(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		instance, err := p.discovery.GetInstance(serviceName)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"service": serviceName,
				"healthy": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"service": serviceName,
			"healthy": true,
			"address": fmt.Sprintf("%s:%d", instance.Address, instance.Port),
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.ResponseWriter.Write(b)
}

func errorType(err error) string {
	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "circuit breaker"):
		return "CIRCUIT_BREAKER_OPEN"
	case strings.Contains(errStr, "timeout"):
		return "TIMEOUT"
	case strings.Contains(errStr, "connection refused"):
		return "CONNECTION_REFUSED"
	case strings.Contains(errStr, "connection reset"):
		return "CONNECTION_RESET"
	case strings.Contains(errStr, "no such host"):
		return "DNS_FAILURE"
	default:
		return "UPSTREAM_ERROR"
	}
}