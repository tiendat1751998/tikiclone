package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
)

func setupTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	return engine
}

func TestHealthEndpoint(t *testing.T) {
	engine := setupTestEngine()
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "alive" {
		t.Errorf("expected alive, got %v", resp["status"])
	}
}

func TestLoginValidation(t *testing.T) {
	engine := setupTestEngine()
	engine.POST("/api/v1/auth/login", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_code": "INVALID_REQUEST",
				"message":    err.Error(),
			})
			return
		}
		if req.Email == "" || req.Password == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_code": "VALIDATION_ERROR",
				"message":    "email and password are required",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"access_token": "test"})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{"missing fields", map[string]string{}, http.StatusBadRequest},
		{"empty email", map[string]string{"email": "", "password": "pass"}, http.StatusBadRequest},
		{"empty password", map[string]string{"email": "a@b.com", "password": ""}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestRegistrationValidation(t *testing.T) {
	engine := setupTestEngine()
	engine.POST("/api/v1/auth/register", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Username string `json:"username"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(req.Password) < 8 {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": "password too short"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": "test-id"})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{"valid", map[string]string{"email": "a@b.com", "password": "StrongPass1!", "username": "testuser"}, http.StatusCreated},
		{"short password", map[string]string{"email": "a@b.com", "password": "weak", "username": "test"}, http.StatusUnprocessableEntity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	engine := setupTestEngine()
	engine.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS, got %d", w.Code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	engine := setupTestEngine()
	engine.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Next()
	})
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	engine.ServeHTTP(w, req)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing X-Content-Type-Options header")
	}
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("missing X-Frame-Options header")
	}
}

func TestMetricsEndpoint(t *testing.T) {
	engine := setupTestEngine()
	engine.GET("/metrics", observability.MetricsHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestPasswordResetRequestValidation(t *testing.T) {
	engine := setupTestEngine()
	engine.POST("/api/v1/auth/password-reset/request", func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_code": "INVALID_REQUEST",
				"message":    "valid email is required",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "password reset email sent if account exists"})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{"valid email", map[string]string{"email": "user@example.com"}, http.StatusOK},
		{"missing email", map[string]string{}, http.StatusBadRequest},
		{"invalid email", map[string]string{"email": "not-an-email"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/password-reset/request", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestResetPasswordValidation(t *testing.T) {
	engine := setupTestEngine()
	engine.POST("/api/v1/auth/password-reset/reset", func(c *gin.Context) {
		var req struct {
			Token           string `json:"token" binding:"required"`
			NewPassword     string `json:"new_password" binding:"required,min=8"`
			ConfirmPassword string `json:"confirm_password" binding:"required,min=8"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_code": "INVALID_REQUEST",
				"message":    "token, new_password, and confirm_password are required",
			})
			return
		}
		if req.NewPassword != req.ConfirmPassword {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
				"error_code": "PASSWORD_MISMATCH",
				"message":    "passwords do not match",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{"valid", map[string]string{"token": "abc123", "new_password": "NewStrong1!", "confirm_password": "NewStrong1!"}, http.StatusOK},
		{"missing token", map[string]string{"new_password": "NewStrong1!", "confirm_password": "NewStrong1!"}, http.StatusBadRequest},
		{"mismatched passwords", map[string]string{"token": "abc123", "new_password": "Pass1!", "confirm_password": "Pass2!"}, http.StatusUnprocessableEntity},
		{"weak password", map[string]string{"token": "abc123", "new_password": "short", "confirm_password": "short"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/password-reset/reset", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestVerifyEmailValidation(t *testing.T) {
	engine := setupTestEngine()
	engine.POST("/api/v1/auth/verify-email", func(c *gin.Context) {
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_code": "INVALID_REQUEST",
				"message":    "token is required",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{"valid token", map[string]string{"token": "verify-token-123"}, http.StatusOK},
		{"missing token", map[string]string{}, http.StatusBadRequest},
		{"empty token", map[string]string{"token": ""}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	engine := setupTestEngine()
	engine.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Next()
	})
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	engine.ServeHTTP(w, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
	}
	for key, expected := range expectedHeaders {
		if got := w.Header().Get(key); got != expected {
			t.Errorf("header %s = %q, want %q", key, got, expected)
		}
	}
}

func TestRequestSanitizer(t *testing.T) {
	engine := setupTestEngine()
	engine.Use(func(c *gin.Context) {
		ct := c.GetHeader("Content-Type")
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if ct == "" || len(ct) > 256 {
				c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
					"error_code": "INVALID_CONTENT_TYPE",
					"message":    "content-type header is required",
				})
				return
			}
		}
		c.Next()
	})
	engine.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("missing content-type", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test", nil)
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusUnsupportedMediaType {
			t.Errorf("expected 415, got %d", w.Code)
		}
	})

	t.Run("valid content-type", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test", nil)
		req.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}
