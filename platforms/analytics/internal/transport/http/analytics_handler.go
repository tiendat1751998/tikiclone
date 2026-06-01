package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/analytics/internal/analytics"
)

type queryRequest struct {
	Metrics    []metricReq          `json:"metrics" binding:"required"`
	Dimensions []string             `json:"dimensions,omitempty"`
	TimeRange  timeRangeReq         `json:"time_range" binding:"required"`
	Filters    []analytics.QueryFilter `json:"filters,omitempty"`
	GroupBy    []string             `json:"group_by,omitempty"`
	OrderBy    string               `json:"order_by,omitempty"`
	OrderDir   string               `json:"order_dir,omitempty"`
	Limit      int                  `json:"limit,omitempty"`
	Offset     int                  `json:"offset,omitempty"`
}

type metricReq struct {
	Name       string `json:"name" binding:"required"`
	Aggregation string `json:"aggregation"`
	Alias      string `json:"alias,omitempty"`
}

type timeRangeReq struct {
	Type    string `json:"type" binding:"required"`
	StartAt string `json:"start_at,omitempty"`
	EndAt   string `json:"end_at,omitempty"`
}

func (h *Handler) RunQuery(c *gin.Context) {
	var req queryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tr := analytics.TimeRange{Type: analytics.TimeRangeType(req.TimeRange.Type)}
	if req.TimeRange.StartAt != "" {
		t, err := time.Parse(time.RFC3339, req.TimeRange.StartAt)
		if err == nil {
			tr.StartAt = &t
		}
	}
	if req.TimeRange.EndAt != "" {
		t, err := time.Parse(time.RFC3339, req.TimeRange.EndAt)
		if err == nil {
			tr.EndAt = &t
		}
	}

	var metrics []analytics.Metric
	for _, m := range req.Metrics {
		agg := analytics.AggregationType(m.Aggregation)
		if agg == "" {
			agg = analytics.AggSum
		}
		metrics = append(metrics, analytics.Metric{
			Name:       analytics.MetricType(m.Name),
			Aggregation: agg,
			Alias:      m.Alias,
		})
	}

	var dimensions []analytics.DimensionType
	for _, d := range req.Dimensions {
		dimensions = append(dimensions, analytics.DimensionType(d))
	}

	query := &analytics.AnalyticsQuery{
		Metrics:    metrics,
		Dimensions: dimensions,
		TimeRange:  tr,
		Filters:    req.Filters,
		GroupBy:    req.GroupBy,
		OrderBy:    req.OrderBy,
		OrderDir:   req.OrderDir,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	result, err := h.analyticsSvc.RunQuery(c.Request.Context(), query)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetMetrics(c *gin.Context) {
	timeRangeStr := c.DefaultQuery("time_range", "last_30d")
	tr := analytics.TimeRange{Type: analytics.TimeRangeType(timeRangeStr)}

	metrics, err := h.analyticsSvc.GetKeyMetrics(c.Request.Context(), tr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": metrics})
}
