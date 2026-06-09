package health

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	sharedHealth "github.com/tikiclone/tiki/packages/go-shared/pkg/health"
)

type Checker struct {
	*sharedHealth.Checker
	db          *sql.DB
	redisClient *redis.Client
}

func NewChecker(appName, version string, db *sql.DB, redisClient *redis.Client) *Checker {
	c := sharedHealth.NewChecker(appName, version)
	c.AddCheck("database", func(ctx context.Context) error {
		return db.PingContext(ctx)
	})
	if redisClient != nil {
		c.AddCheck("redis", func(ctx context.Context) error {
			return redisClient.Ping(ctx).Err()
		})
	}
	return &Checker{
		Checker:     c,
		db:          db,
		redisClient: redisClient,
	}
}

func (c *Checker) LivenessHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "alive",
			"service": "tiki-rec-vector",
		})
	}
}

func (c *Checker) ReadinessHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := c.db.PingContext(ctx); err != nil {
			zap.L().Warn("readiness check: db ping failed", zap.Error(err))
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "reason": "database unavailable"})
			return
		}
		if c.redisClient != nil {
			pingCtx, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
			defer cancel()
			if err := c.redisClient.Ping(pingCtx).Err(); err != nil {
				zap.L().Warn("readiness check: redis ping failed", zap.Error(err))
			}
		}
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "tiki-rec-vector",
			"version": "1.0.0",
		})
	}
}
