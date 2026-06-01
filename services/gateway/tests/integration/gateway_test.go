package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
	"github.com/tikiclone/tiki/services/gateway/internal/auth"
	"github.com/tikiclone/tiki/services/gateway/internal/config"
	"github.com/tikiclone/tiki/services/gateway/internal/discovery"
	"github.com/tikiclone/tiki/services/gateway/internal/ratelimit"
	"github.com/tikiclone/tiki/services/gateway/internal/routing"
	"github.com/tikiclone/tiki/services/gateway/internal/transport"
)

func newTestRequest(method, path string) *http.Request {
	req, _ := http.NewRequest(method, path, nil)
	req.Header.Set("User-Agent", "test-agent/1.0")
	req.Header.Set("Host", "localhost")
	return req
}

func setupTestRouter() *gin.Engine {
	cfg := &config.Config{
		AppName:  "test-gateway",
		AppEnv:   "test",
		LogLevel: "error",
		HTTPPort: 8080,
		RateLimit: config.RateLimitConfig{
			Enabled: false,
		},
		Auth: config.AuthConfig{
			EnableRBAC: false,
		},
		Server: config.ServerConfig{
			MaxBodySize: 10485760,
		},
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
		OpenTelemetry: config.OTELConfig{
			ServiceName: "test-gateway",
		},
		Upstreams: config.UpstreamConfig{
			DefaultTimeout:  5 * time.Second,
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second,
		},
	}

	jwtValidator := auth.NewJWTValidator(cfg.Auth, nil)
	authMiddleware := auth.NewAuthMiddleware(jwtValidator)

	rateLimiter := ratelimit.NewRateLimiter(nil, cfg.RateLimit)

	svcDiscovery := discovery.NewServiceDiscovery()
	svcDiscovery.RegisterStatic("test-service", []*discovery.ServiceInstance{
		{ID: "test-1", Name: "test-service", Address: "localhost", Port: 9999, Weight: 1},
	})

	proxy := transport.NewProxy(
		svcDiscovery,
		cfg.Upstreams.MaxIdleConns,
		cfg.Upstreams.IdleConnTimeout,
	)

	grpcProxy := transport.NewGRPCProxy(svcDiscovery)

	healthChecker := health.NewChecker("test-gateway", "1.0.0")

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	router := routing.NewRouter(cfg, proxy, grpcProxy, rateLimiter, authMiddleware, svcDiscovery, healthChecker, nil)
	router.Setup(engine)

	return engine
}

func TestGateway_HealthEndpoint(t *testing.T) {
	engine := setupTestRouter()

	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodGet, "/health")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response["status"] != "alive" {
		t.Errorf("expected alive status, got %v", response["status"])
	}
}

func TestGateway_ReadinessEndpoint(t *testing.T) {
	engine := setupTestRouter()

	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodGet, "/ready")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGateway_MetricsEndpoint(t *testing.T) {
	engine := setupTestRouter()

	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodGet, "/metrics")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGateway_NotFound(t *testing.T) {
	engine := setupTestRouter()

	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodGet, "/nonexistent")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGateway_CORSHeaders(t *testing.T) {
	engine := setupTestRouter()

	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodOptions, "/api/v1/products")
	req.Header.Set("Origin", "http://localhost:3000")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("expected CORS origin to be echoed, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestGateway_CORSRejectsInvalidOrigin(t *testing.T) {
	engine := setupTestRouter()

	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodOptions, "/api/v1/products")
	req.Header.Set("Origin", "http://evil.com")
	engine.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "http://evil.com" {
		t.Error("CORS should not allow invalid origins")
	}
}

func TestGateway_SecurityHeaders(t *testing.T) {
	engine := setupTestRouter()

	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodGet, "/health")
	engine.ServeHTTP(w, req)

	headers := []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Strict-Transport-Security",
		"Referrer-Policy",
	}

	for _, header := range headers {
		if w.Header().Get(header) == "" {
			t.Errorf("missing security header: %s", header)
		}
	}
}

func TestGateway_CorrelationID(t *testing.T) {
	engine := setupTestRouter()

	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodGet, "/health")
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	engine.ServeHTTP(w, req)

	if w.Header().Get("X-Correlation-ID") != "test-correlation-id" {
		t.Errorf("expected correlation id to be preserved")
	}
}

func TestGateway_RequestIDGenerated(t *testing.T) {
	engine := setupTestRouter()

	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodGet, "/health")
	engine.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("expected request ID to be generated")
	}
}

func TestGateway_UpstreamsEndpoint(t *testing.T) {
	engine := setupTestRouter()

	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodGet, "/upstreams")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if response["services"] == nil {
		t.Error("expected services list")
	}
}

func TestGateway_MissingUserAgent(t *testing.T) {
	engine := setupTestRouter()

	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodGet, "/api/v1/products")
	req.Header.Del("User-Agent")
	engine.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Log("user-agent not enforced for this route")
	}
}

func TestGateway_RateLimitHeaders(t *testing.T) {
	engine := setupTestRouter()

	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodGet, "/health")
	engine.ServeHTTP(w, req)

	limitHeader := w.Header().Get("X-RateLimit-Limit")
	if limitHeader == "" {
		t.Log("rate limit headers not present (rate limiting disabled in test)")
	}
}

func BenchmarkGatewayHealth(b *testing.B) {
	engine := setupTestRouter()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := newTestRequest(http.MethodGet, "/health")
		engine.ServeHTTP(w, req)
	}
}

func BenchmarkGatewayMetrics(b *testing.B) {
	engine := setupTestRouter()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := newTestRequest(http.MethodGet, "/metrics")
		engine.ServeHTTP(w, req)
	}
}

func BenchmarkGatewayUpstreams(b *testing.B) {
	engine := setupTestRouter()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := newTestRequest(http.MethodGet, "/upstreams")
		engine.ServeHTTP(w, req)
	}
}
