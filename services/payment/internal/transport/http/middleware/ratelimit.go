package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	redisinfra "github.com/tikiclone/tiki/services/payment/internal/infrastructure/redis"
)

func RateLimit(store *redisinfra.Store, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !store.IsAvailable() {
			c.Next()
			return
		}
		key := "ratelimit:webhook:" + c.ClientIP()
		count, err := store.IncrementCounter(c.Request.Context(), key, window)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "ratelimit error"})
			return
		}
		if count > int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
