package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/gateway/internal/config"
)

func TestKeyByPath(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"/api/v1/products", "api/v1"},
		{"/api/v1/products/123", "api/v1"},
		{"/health", "health"},
		{"/", ""},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			key := KeyByPath(req)
			if key != tt.expected {
				t.Errorf("got %q, want %q", key, tt.expected)
			}
		})
	}
}

func TestRateLimiter_NoRedis(t *testing.T) {
	cfg := config.RateLimitConfig{
		Enabled:       true,
		DefaultMaxRPS: 10,
		WindowSize:    time.Second,
	}

	rl := NewRateLimiter(nil, cfg)
	if rl == nil {
		t.Fatal("rate limiter should not be nil")
	}

	result, err := rl.Allow(context.Background(), "test-key", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("expected request to be allowed when no redis")
	}
}

func TestRateLimiter_GlobalMiddleware_Disabled(t *testing.T) {
	cfg := config.RateLimitConfig{Enabled: false}
	rl := NewRateLimiter(nil, cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	rl.GlobalMiddleware()(c)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 when disabled, got %d", w.Code)
	}
}

func TestRateLimiter_IPRateLimit_Disabled(t *testing.T) {
	cfg := config.RateLimitConfig{Enabled: false}
	rl := NewRateLimiter(nil, cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	rl.IPRateLimit()(c)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 when disabled, got %d", w.Code)
	}
}

func TestSetOverride(t *testing.T) {
	cfg := config.RateLimitConfig{Enabled: true, AuthenticatedRPS: 100}
	rl := NewRateLimiter(nil, cfg)

	rl.SetOverride("/api/v1/auth/login", 5)
	rl.mu.RLock()
	override, exists := rl.overrides["/api/v1/auth/login"]
	rl.mu.RUnlock()

	if !exists {
		t.Error("override should exist")
	}
	if override != 5 {
		t.Errorf("expected override 5, got %d", override)
	}
}
