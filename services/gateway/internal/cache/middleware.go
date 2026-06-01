package cache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
)

// CacheMiddleware provides Redis-backed response caching for GET requests.
type CacheMiddleware struct {
	rdb        *redis.Client
	defaultTTL time.Duration
	enabled    bool
}

func NewCacheMiddleware(rdb *redis.Client, defaultTTL time.Duration) *CacheMiddleware {
	if rdb == nil {
		return &CacheMiddleware{enabled: false}
	}
	return &CacheMiddleware{
		rdb:        rdb,
		defaultTTL: defaultTTL,
		enabled:    true,
	}
}

type cachedResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
}

// CacheByPath returns a gin middleware that caches responses for specified path prefixes.
func (c *CacheMiddleware) CacheByPath(ttl time.Duration, prefixes ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !c.enabled || ctx.Request.Method != "GET" {
			ctx.Next()
			return
		}

		shouldCache := len(prefixes) == 0
		for _, p := range prefixes {
			if strings.HasPrefix(ctx.Request.URL.Path, p) {
				shouldCache = true
				break
			}
		}
		if !shouldCache {
			ctx.Next()
			return
		}

		key := c.cacheKey(ctx.Request)

		// Try cache
		if resp, ok := c.get(ctx.Request.Context(), key); ok {
			for k, v := range resp.Headers {
				ctx.Header(k, v)
			}
			ctx.Header("X-Cache", "HIT")
			ctx.Data(resp.StatusCode, "application/json", resp.Body)
			ctx.Abort()
			return
		}

		// Capture response
		recorder := &responseCapture{ResponseWriter: ctx.Writer, statusCode: http.StatusOK}
		ctx.Writer = recorder

		ctx.Next()

		if recorder.statusCode == http.StatusOK {
			stored := cachedResponse{
				StatusCode: recorder.statusCode,
				Headers:    recorder.capturedHeaders,
				Body:       recorder.body.Bytes(),
			}
			if data, err := json.Marshal(stored); err == nil {
				c.set(ctx.Request.Context(), key, data, ttl)
			}
		}
	}
}

func (c *CacheMiddleware) cacheKey(r *http.Request) string {
	raw := r.Method + ":" + r.URL.Path + "?" + r.URL.RawQuery
	hash := sha256.Sum256([]byte(raw))
	return "gw:cache:" + hex.EncodeToString(hash[:16])
}

func (c *CacheMiddleware) get(ctx context.Context, key string) (*cachedResponse, bool) {
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	var resp cachedResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, false
	}
	return &resp, true
}

func (c *CacheMiddleware) set(ctx context.Context, key string, data []byte, ttl time.Duration) {
	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		observability.GetLogger().Warn("cache set failed",
			zap.String("key", key),
			zap.Error(err),
		)
	}
}

type responseCapture struct {
	gin.ResponseWriter
	statusCode      int
	body            bytes.Buffer
	capturedHeaders map[string]string
}

func (r *responseCapture) WriteHeader(code int) {
	r.statusCode = code
	r.capturedHeaders = make(map[string]string)
	for k := range r.Header() {
		r.capturedHeaders[k] = r.Header().Get(k)
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseCapture) Write(p []byte) (int, error) {
	r.body.Write(p)
	return r.ResponseWriter.Write(p)
}
