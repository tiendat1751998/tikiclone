package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/gateway/internal/config"
	"go.uber.org/zap"
)

type RateLimiter struct {
	rdb     *redis.Client
	limiter *redis_rate.Limiter
	cfg     config.RateLimitConfig
	mu      sync.RWMutex
	overrides map[string]int
}

func NewRateLimiter(rdb *redis.Client, cfg config.RateLimitConfig) *RateLimiter {
	if rdb == nil {
		return &RateLimiter{
			cfg:       cfg,
			overrides: make(map[string]int),
		}
	}
	return &RateLimiter{
		rdb:       rdb,
		limiter:   redis_rate.NewLimiter(rdb),
		cfg:       cfg,
		overrides: make(map[string]int),
	}
}

func (l *RateLimiter) SetOverride(path string, rps int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.overrides[path] = rps
}

type LimitResult struct {
	Allowed   bool
	Remaining int
	ResetAfter time.Duration
	Limit     int
}

func (l *RateLimiter) Allow(ctx context.Context, key string, rate int) (*LimitResult, error) {
	return l.AllowN(ctx, key, rate, 1)
}

func (l *RateLimiter) AllowN(ctx context.Context, key string, rate int, n int) (*LimitResult, error) {
	if l.limiter == nil {
		return &LimitResult{
			Allowed:    true,
			Remaining:  rate,
			ResetAfter: l.cfg.WindowSize,
			Limit:      rate,
		}, nil
	}

	res, err := l.limiter.AllowN(ctx, key, redis_rate.PerSecond(rate), n)
	if err != nil {
		// [RELIABILITY] Fail open on Redis error — don't block traffic when Redis is down
		observability.GetLogger().Warn("rate limiter Redis error, failing open",
			zap.String("key", key),
			zap.Error(err),
		)
		return &LimitResult{
			Allowed:    true,
			Remaining:  rate,
			ResetAfter: l.cfg.WindowSize,
			Limit:      rate,
		}, nil
	}

	return &LimitResult{
		Allowed:    res.Allowed >= n,
		Remaining:  res.Remaining,
		ResetAfter: res.RetryAfter,
		Limit:      rate,
	}, nil
}

func (l *RateLimiter) IPRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.cfg.Enabled {
			c.Next()
			return
		}

		clientIP := forwardedClientIP(c)
		key := fmt.Sprintf("ip:%s", clientIP)

		result, err := l.Allow(c.Request.Context(), key, l.cfg.IPMaxRPS)
		if err != nil || !result.Allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error_code": "IP_RATE_LIMIT_EXCEEDED",
				"message":    "too many requests from this IP",
			})
			return
		}

		setRateLimitHeaders(c, result)
		c.Next()
	}
}

func (l *RateLimiter) GlobalMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.cfg.Enabled {
			c.Next()
			return
		}

		result, err := l.Allow(c.Request.Context(), "global:"+forwardedClientIP(c), l.cfg.GlobalMaxRPS)
		if err != nil || !result.Allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error_code": "GLOBAL_RATE_LIMIT_EXCEEDED",
				"message":    "global rate limit exceeded",
			})
			return
		}

		setRateLimitHeaders(c, result)
		c.Next()
	}
}

// forwardedClientIP extracts the real client IP from X-Forwarded-For (set by our proxy)
func forwardedClientIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		for i, ch := range xff {
			if ch == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	return c.ClientIP()
}

func (l *RateLimiter) AuthenticatedRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.cfg.Enabled {
			c.Next()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.Next()
			return
		}

		key := fmt.Sprintf("user:%s", userIDStr)

		rate := l.cfg.AuthenticatedRPS
		path := c.FullPath()
		if path != "" {
			l.mu.RLock()
			if override, ok := l.overrides[path]; ok {
				rate = override
			}
			l.mu.RUnlock()
		}

		result, err := l.Allow(c.Request.Context(), key, rate)
		if err != nil || !result.Allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error_code": "USER_RATE_LIMIT_EXCEEDED",
				"message":    "too many requests",
			})
			return
		}

		setRateLimitHeaders(c, result)
		c.Next()
	}
}

func (l *RateLimiter) PerEndpointRateLimit(rps int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.cfg.Enabled {
			c.Next()
			return
		}

		key := fmt.Sprintf("endpoint:%s:%s", c.Request.Method, c.FullPath())
		if userID, exists := c.Get("user_id"); exists {
			key = fmt.Sprintf("endpoint:%s:%s:%v", c.Request.Method, c.FullPath(), userID)
		}

		result, err := l.Allow(c.Request.Context(), key, rps)
		if err != nil || !result.Allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error_code": "ENDPOINT_RATE_LIMIT_EXCEEDED",
				"message":    "endpoint rate limit exceeded",
			})
			return
		}

		setRateLimitHeaders(c, result)
		c.Next()
	}
}

func setRateLimitHeaders(c *gin.Context, result *LimitResult) {
	c.Header("X-RateLimit-Limit", strconv.Itoa(result.Limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
	if result.ResetAfter > 0 {
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(result.ResetAfter).Unix(), 10))
	}
}

func KeyByPath(r *http.Request) string {
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return parts[0] + "/" + parts[1]
	}
	return path
}
