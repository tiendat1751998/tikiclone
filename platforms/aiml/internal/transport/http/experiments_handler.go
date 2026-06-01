package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/aiml/internal/experiments"
)

type createExperimentRequest struct {
	Name       string  `json:"name" binding:"required"`
	ModelA     string  `json:"model_a" binding:"required"`
	ModelB     string  `json:"model_b" binding:"required"`
	TrafficPct float64 `json:"traffic_percentage"`
	Metric     string  `json:"metric" binding:"required"`
}

type assignVariantRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type recordResultRequest struct {
	Variant string  `json:"variant" binding:"required"`
	UserID  string  `json:"user_id" binding:"required"`
	Value   float64 `json:"value" binding:"required"`
}

func (h *Handler) CreateExperiment(c *gin.Context) {
	var req createExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	trafficPct := req.TrafficPct
	if trafficPct <= 0 {
		trafficPct = 50
	}
	exp := &experiments.Experiment{
		ID:         uuid.New().String(),
		Name:       req.Name,
		ModelA:     req.ModelA,
		ModelB:     req.ModelB,
		TrafficPct: trafficPct,
		Metric:     experiments.MetricType(req.Metric),
	}
	if err := h.experimentSvc.CreateExperiment(c.Request.Context(), exp); err != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, exp)
}

func (h *Handler) AssignVariant(c *gin.Context) {
	experimentID := c.Param("id")
	var req assignVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	assignment, err := h.experimentSvc.AssignVariant(c.Request.Context(), experimentID, req.UserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assignment)
}

func (h *Handler) RecordExperimentResult(c *gin.Context) {
	experimentID := c.Param("id")
	var req recordResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.experimentSvc.RecordResult(c.Request.Context(), experimentID, req.Variant, req.UserID, req.Value); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

func (h *Handler) GetExperimentResults(c *gin.Context) {
	experimentID := c.Param("id")
	results, err := h.experimentSvc.GetExperimentResults(c.Request.Context(), experimentID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}
