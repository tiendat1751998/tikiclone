package health

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/gateway/internal/discovery"
	"go.uber.org/zap"
)

type GatewayHealth struct {
	discovery *discovery.ServiceDiscovery
	redis     *redis.Client
}

func NewGatewayHealth(d *discovery.ServiceDiscovery, rdb *redis.Client) *GatewayHealth {
	return &GatewayHealth{discovery: d, redis: rdb}
}

type UpstreamHealth struct {
	Service string `json:"service"`
	Healthy bool   `json:"healthy"`
	Error   string `json:"error,omitempty"`
}

func (h *GatewayHealth) UpstreamsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		services := h.discovery.GetAllServices()
		results := make([]UpstreamHealth, 0, len(services))

		for _, svc := range services {
			instances := h.discovery.GetInstances(svc)
			healthy := len(instances) > 0
			uh := UpstreamHealth{
				Service: svc,
				Healthy: healthy,
			}
			if !healthy {
				uh.Error = "no healthy instances"
			}
			results = append(results, uh)
		}

		c.JSON(http.StatusOK, gin.H{
			"timestamp": time.Now().UTC(),
			"services":  results,
			"total":     len(results),
			"healthy":   countHealthy(results),
		})
	}
}

func countHealthy(results []UpstreamHealth) int {
	count := 0
	for _, r := range results {
		if r.Healthy {
			count++
		}
	}
	return count
}

// [FIX A5] RedisHealthCheck now actually checks Redis connectivity.
// Previously it always returned nil (healthy), even when Redis was down.
// This caused load balancers to never detect Redis failures.
func RedisHealthCheck(addr string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		if addr == "" {
			return fmt.Errorf("redis address is empty")
		}

		// Create a temporary client for health check
		client := redis.NewClient(&redis.Options{
			Addr:        addr,
			DialTimeout: 5 * time.Second,
		})
		defer client.Close()

		// Actually ping Redis
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := client.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("redis ping failed: %w", err)
		}
		return nil
	}
}

func logHealthState(service string, healthy bool) {
	if healthy {
		observability.GetLogger().Info("service health check passed",
			zap.String("service", service))
	} else {
		observability.GetLogger().Warn("service health check failed",
			zap.String("service", service))
	}
}
