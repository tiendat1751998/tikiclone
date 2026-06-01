package middleware

import (
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"
)

type SecurityConfig struct {
	MaxBodySize    int64
	AllowedHosts   []string
	TrustedProxies []string
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")

		c.Next()
	}
}

func BodySizeLimiter(maxBodySize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodySize)
		}
		c.Next()
	}
}

var (
	sqlInjectionPattern   = regexp.MustCompile(`(?i)(\b(ALTER|CREATE|DELETE|DROP|EXEC|INSERT|MERGE|SELECT|TRUNCATE|UPDATE|UNION)\b)`)
	xssPattern            = regexp.MustCompile(`(?i)(<script|<\/script|javascript:|onerror=|onload=|onclick=)`)
	pathTraversalPattern  = regexp.MustCompile(`\.\./|\.\.\\|%2e%2e%2f|%2e%2e%5c`)
	quickDangerPattern    = regexp.MustCompile(`(?i)(<script|javascript:|onerror=|\.\./)`)
)

func isWriteMethod(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete
}

func RequestSanitizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Request.URL.Query()
		if len(query) > 0 {
			if isWriteMethod(c.Request.Method) {
				for key, values := range query {
					for i, v := range values {
						values[i] = sanitizeInput(v)
					}
					query[key] = values
				}
				c.Request.URL.RawQuery = query.Encode()
			} else {
				for _, values := range query {
					for _, v := range values {
						if quickDangerPattern.MatchString(v) {
							c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
								"error_code": "INVALID_QUERY",
								"message":    "malformed query parameter detected",
							})
							return
						}
					}
				}
			}
		}

		for _, v := range c.Request.Header.Values("X-Forwarded-For") {
			if pathTraversalPattern.MatchString(v) {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error_code": "INVALID_HEADER",
					"message":    "malformed header detected",
				})
				return
			}
		}

		c.Next()
	}
}

func sanitizeInput(input string) string {
	input = html.EscapeString(input)
	input = sqlInjectionPattern.ReplaceAllString(input, "")
	input = xssPattern.ReplaceAllString(input, "")
	return input
}

func IPThrottler(trustedProxies []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		for _, proxy := range trustedProxies {
			if strings.HasPrefix(clientIP, proxy) {
				c.Next()
				return
			}
		}
		c.Next()
	}
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		bodySize := c.Writer.Size()
		userID, _ := c.Get(string(UserIDKey))

		log := observability.LogWithTrace(c.Request.Context())
		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("client_ip", clientIP),
			zap.Duration("latency", latency),
			zap.Int("body_size", bodySize),
		}
		if userID != nil {
			fields = append(fields, zap.String("user_id", fmt.Sprintf("%v", userID)))
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				fields = append(fields, zap.String("error", e.Err.Error()))
			}
			log.Error("request completed with errors", fields...)
		} else if status >= 500 {
			log.Error("server error", fields...)
		} else if status >= 400 {
			log.Warn("client error", fields...)
		} else {
			log.Info("request completed", fields...)
		}
	}
}

func AntiAbuse() gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := c.GetHeader("User-Agent")
		if userAgent == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_code": "MISSING_USER_AGENT",
				"message":    "User-Agent header is required",
			})
			return
		}
		if len(userAgent) > 500 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_code": "INVALID_USER_AGENT",
				"message":    "User-Agent header too long",
			})
			return
		}

		contentType := c.GetHeader("Content-Type")
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch {
			if contentType == "" {
				c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
					"error_code": "MISSING_CONTENT_TYPE",
					"message":    "Content-Type header is required for write operations",
				})
				return
			}
			if !strings.HasPrefix(contentType, "application/json") && !strings.HasPrefix(contentType, "multipart/form-data") && !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
				c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
					"error_code": "INVALID_CONTENT_TYPE",
					"message":    "unsupported Content-Type",
				})
				return
			}
		}

		query := c.Request.URL.RawQuery
		if len(query) > 2000 {
			c.AbortWithStatusJSON(http.StatusRequestURITooLong, gin.H{
				"error_code": "QUERY_TOO_LONG",
				"message":    "Query string exceeds maximum length",
			})
			return
		}

		c.Next()
	}
}

func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = c.GetHeader("Referer")
		}
		if origin == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error_code": "CSRF_ORIGIN_MISSING",
				"message":    "Origin or Referer header required for state-changing requests",
			})
			return
		}

		c.Next()
	}
}

func DeviceMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceInfo := map[string]string{
			"user_agent":      c.GetHeader("User-Agent"),
			"device_id":       c.GetHeader("X-Device-ID"),
			"device_type":     c.GetHeader("X-Device-Type"),
			"platform":        c.GetHeader("X-Platform"),
			"app_version":     c.GetHeader("X-App-Version"),
			"accept_language": c.GetHeader("Accept-Language"),
		}
		c.Set(string(DeviceInfoKey), deviceInfo)
		c.Next()
	}
}

func RequestValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := validateRequest(c); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_code": "VALIDATION_ERROR",
				"message":    err.Error(),
			})
			return
		}
		c.Next()
	}
}

func validateRequest(c *gin.Context) error {
	method := c.Request.Method
	path := c.Request.URL.Path

	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal detected")
	}
	if pathTraversalPattern.MatchString(path) {
		return fmt.Errorf("path traversal detected")
	}

	if method == http.MethodTrace {
		return fmt.Errorf("TRACE method not allowed")
	}

	if method == http.MethodConnect {
		return fmt.Errorf("CONNECT method not allowed")
	}

	if method == http.MethodOptions {
		return nil
	}

	host := c.Request.Host
	if host == "" {
		return fmt.Errorf("Host header is required")
	}

	contentLength := c.GetHeader("Content-Length")
	if contentLength != "" {
		length, err := strconv.ParseInt(contentLength, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid Content-Length header")
		}
		if length < 0 {
			return fmt.Errorf("negative Content-Length")
		}
		if length > 10*1024*1024 {
			return fmt.Errorf("Content-Length exceeds maximum allowed size")
		}
	}

	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		contentType := c.GetHeader("Content-Type")
		if contentType == "" {
			return fmt.Errorf("Content-Type header required for write operations")
		}
	}

	return nil
}
