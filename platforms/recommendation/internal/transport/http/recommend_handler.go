package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/events"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/metrics"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/recommender"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/types"
)

type recommendRequest struct {
	UserID    string `json:"user_id"`
	Type      string `json:"type"`
	ProductID string `json:"product_id"`
	SessionID string `json:"session_id"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

type recommendResponse struct {
	Recommendations []types.ProductRecommendation `json:"recommendations"`
	TookMs          int64                          `json:"took_ms"`
}

type batchRecommendRequest struct {
	Requests []recommendRequest `json:"requests"`
}

type batchRecommendResponse struct {
	Results []recommendResponse `json:"results"`
}

func (h *Handler) GetRecommendations(c *gin.Context) {
	start := time.Now()

	var req recommendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	recCtx := recommender.RecommendationContext{
		UserID:    req.UserID,
		Type:      recommender.RecommendationType(req.Type),
		ProductID: req.ProductID,
		SessionID: req.SessionID,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	recs, err := h.recommender.GetRecommendations(c.Request.Context(), recCtx)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tookMs := time.Since(start).Milliseconds()
	metrics.RecLatency.Observe(float64(tookMs) / 1000.0)
	metrics.RecRequestsTotal.Inc()

	h.publisher.Publish(c.Request.Context(), events.EventRecommendationRequested, events.RecommendationRequested{
		UserID:    req.UserID,
		Type:      req.Type,
		Limit:     req.Limit,
		Timestamp: time.Now(),
	})

	c.JSON(http.StatusOK, recommendResponse{
		Recommendations: recs,
		TookMs:          tookMs,
	})
}

func (h *Handler) BatchRecommendations(c *gin.Context) {
	start := time.Now()

	var req batchRecommendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	results := make([]recommendResponse, 0, len(req.Requests))
	for _, r := range req.Requests {
		if r.Limit <= 0 {
			r.Limit = 20
		}
		recCtx := recommender.RecommendationContext{
			UserID:    r.UserID,
			Type:      recommender.RecommendationType(r.Type),
			ProductID: r.ProductID,
			SessionID: r.SessionID,
			Limit:     r.Limit,
			Offset:    r.Offset,
		}
		recs, err := h.recommender.GetRecommendations(c.Request.Context(), recCtx)
		if err != nil {
			results = append(results, recommendResponse{Recommendations: []types.ProductRecommendation{}, TookMs: 0})
			continue
		}
		results = append(results, recommendResponse{
			Recommendations: recs,
			TookMs:          time.Since(start).Milliseconds(),
		})
	}

	tookMs := time.Since(start).Milliseconds()
	metrics.RecLatency.Observe(float64(tookMs) / 1000.0)
	metrics.RecRequestsTotal.Inc()

	c.JSON(http.StatusOK, batchRecommendResponse{Results: results})
}

func (h *Handler) GetTrending(c *gin.Context) {
	start := time.Now()

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	trendingScores, err := h.trending.GetTrending(c.Request.Context(), limit)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	recs := make([]types.ProductRecommendation, 0, len(trendingScores))
	for _, ts := range trendingScores {
		recs = append(recs, types.ProductRecommendation{
			ProductID: ts.ProductID,
			Score:     ts.Score,
			Type:      types.RecTypeTrending,
			Reason:    string(types.ReasonTrendingNow),
		})
	}

	tookMs := time.Since(start).Milliseconds()
	metrics.RecLatency.Observe(float64(tookMs) / 1000.0)
	metrics.RecRequestsTotal.Inc()

	c.JSON(http.StatusOK, recommendResponse{
		Recommendations: recs,
		TookMs:          tookMs,
	})
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return def
}
