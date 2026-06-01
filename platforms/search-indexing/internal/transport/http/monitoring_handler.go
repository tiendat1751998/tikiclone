package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/search-indexing/internal/monitoring"
)

type reportMetricsRequest struct {
	IndexName       string                `json:"index_name"`
	ShardCount      int                   `json:"shard_count"`
	DocCount        int64                 `json:"doc_count"`
	SizeBytes       int64                 `json:"size_bytes"`
	IndexingRate    float64               `json:"indexing_rate"`
	SearchLatencyP99 float64              `json:"search_latency_p99"`
	Health          monitoring.HealthStatus `json:"health"`
}

func (h *Handler) ReportMetrics(c *gin.Context) {
	var req reportMetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	metric := &monitoring.IndexMetric{
		IndexName:        req.IndexName,
		ShardCount:       req.ShardCount,
		DocCount:         req.DocCount,
		SizeBytes:        req.SizeBytes,
		IndexingRate:     req.IndexingRate,
		SearchLatencyP99: req.SearchLatencyP99,
		Health:           req.Health,
	}
	if err := h.monitoring.ReportMetrics(c.Request.Context(), metric); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "metrics reported", "metric": metric})
}

func (h *Handler) GetClusterHealth(c *gin.Context) {
	health, err := h.monitoring.GetClusterHealth(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, health)
}

func (h *Handler) GetIndexMetrics(c *gin.Context) {
	metrics, err := h.monitoring.GetIndexMetrics(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"indexes": metrics})
}
