package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/api-gateway/internal/ratelimit"
)

type createRateLimitRuleRequest struct {
	Key           string `json:"key" binding:"required"`
	MaxRequests   int    `json:"max_requests" binding:"required"`
	WindowSeconds int    `json:"window_seconds" binding:"required"`
	BurstSize     int    `json:"burst_size"`
}

type checkRateLimitRequest struct {
	Key string `json:"key" binding:"required"`
}

func (h *Handler) CreateRateLimitRule(c *gin.Context) {
	var req createRateLimitRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule := &ratelimit.RateLimitRule{
		Key:           req.Key,
		MaxRequests:   req.MaxRequests,
		WindowSeconds: req.WindowSeconds,
		BurstSize:     req.BurstSize,
	}

	if err := h.RateLimiter.CreateRule(rule); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

func (h *Handler) CheckRateLimit(c *gin.Context) {
	var req checkRateLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.RateLimiter.Check(req.Key)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
