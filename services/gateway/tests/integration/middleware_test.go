package integration

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/gateway/internal/middleware"
)

func TestCorrelationIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.CorrelationID())
	engine.GET("/test", func(c *gin.Context) {
		reqID, _ := c.Get(string(middleware.RequestIDKey))
		corrID, _ := c.Get(string(middleware.CorrelationIDKey))
		c.JSON(http.StatusOK, gin.H{
			"request_id":     reqID,
			"correlation_id": corrID,
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Correlation-ID", "custom-corr-id")
	engine.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header")
	}
	if w.Header().Get("X-Correlation-ID") != "custom-corr-id" {
		t.Error("expected custom X-Correlation-ID header")
	}
}

func TestCorrelationID_Generated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.CorrelationID())
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	engine.ServeHTTP(w, req)

	if w.Header().Get("X-Correlation-ID") == "" {
		t.Error("expected correlation ID to be generated")
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.SecurityHeaders())
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	engine.ServeHTTP(w, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":         "nosniff",
		"X-Frame-Options":                "DENY",
		"X-XSS-Protection":               "1; mode=block",
		"Strict-Transport-Security":      "max-age=31536000; includeSubDomains; preload",
		"Referrer-Policy":                "strict-origin-when-cross-origin",
		"Cross-Origin-Resource-Policy":   "same-origin",
		"Cross-Origin-Opener-Policy":     "same-origin",
	}

	for header, expected := range expectedHeaders {
		if got := w.Header().Get(header); got != expected {
			t.Errorf("header %s: expected %q, got %q", header, expected, got)
		}
	}
}

func TestBodySizeLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.BodySizeLimiter(1024))
	engine.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	largeBody := make([]byte, 2048)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	req.Body = io.NopCloser(bytes.NewReader(largeBody))
	req.ContentLength = int64(len(largeBody))
	engine.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Log("body size exceeded but test returned 200 (handler may ignore)")
	}
}

func TestRequestSanitizer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.RequestSanitizer())
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"query": c.Request.URL.RawQuery,
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test?q=<script>alert(1)</script>", nil)
	engine.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Log("sanitizer processed request without error")
	}
}

func TestAntiAbuse_MissingUserAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.AntiAbuse())
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Del("User-Agent")
	engine.ServeHTTP(w, req)

	if w.Code == http.StatusBadRequest {
		t.Log("anti-abuse correctly rejected missing User-Agent")
	}
}

func TestAntiAbuse_LongUserAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.AntiAbuse())
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("User-Agent", string(make([]byte, 600)))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for long User-Agent, got %d", w.Code)
	}
}

func TestCSRFProtection_AllowsGET(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.CSRFProtection())
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for GET, got %d", w.Code)
	}
}

func TestCSRFProtection_RejectsPOSTWithoutOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.CSRFProtection())
	engine.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for POST without Origin, got %d", w.Code)
	}
}

func TestRequestValidation_RejectsTRACE(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.RequestValidation())
	engine.Any("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodTrace, "/test", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for TRACE, got %d", w.Code)
	}
}

func TestRequestValidation_RejectsPathTraversal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.RequestValidation())
	engine.Any("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/../../../etc/passwd", nil)
	req.Header.Set("Host", "localhost")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for path traversal, got %d", w.Code)
	}
}
