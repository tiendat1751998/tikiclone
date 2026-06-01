package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/featureflag"
)

type createFlagRequest struct {
	Name              string `json:"name" binding:"required"`
	Enabled           bool   `json:"enabled"`
	PercentageRollout int    `json:"percentage_rollout"`
	UserSegment       string `json:"user_segment"`
}

func (h *Handler) CreateFlag(c *gin.Context) {
	var req createFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	flag := &featureflag.FeatureFlag{
		Name:              req.Name,
		Enabled:           req.Enabled,
		PercentageRollout: req.PercentageRollout,
		UserSegment:       featureflag.Segment(req.UserSegment),
	}

	created, err := h.FeatureFlagSvc.Create(c.Request.Context(), flag)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, created)
}

func (h *Handler) ListFlags(c *gin.Context) {
	flags, err := h.FeatureFlagSvc.List(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"flags": flags})
}

type evaluateFlagRequest struct {
	FlagName string `json:"flag_name" binding:"required"`
	UserID   string `json:"user_id" binding:"required"`
}

func (h *Handler) EvaluateFlag(c *gin.Context) {
	var req evaluateFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.FeatureFlagSvc.Evaluate(c.Request.Context(), req.FlagName, req.UserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
