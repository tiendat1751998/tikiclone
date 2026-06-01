package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/audience"
)

type createSegmentRequest struct {
	Name     string           `json:"name" binding:"required"`
	Criteria audience.Criteria `json:"criteria"`
}

type evaluateSegmentRequest struct {
	SegmentID string `json:"segment_id" binding:"required"`
	UserID    string `json:"user_id" binding:"required"`
}

func (h *Handler) CreateSegment(c *gin.Context) {
	var req createSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	seg, err := h.audienceSvc.CreateSegment(c.Request.Context(), &audience.CreateSegmentRequest{
		Name:     req.Name,
		Criteria: req.Criteria,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, seg)
}

func (h *Handler) ListSegments(c *gin.Context) {
	segments, err := h.audienceSvc.ListSegments(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"segments": segments})
}

func (h *Handler) EvaluateSegment(c *gin.Context) {
	var req evaluateSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	match, err := h.audienceSvc.EvaluateUser(c.Request.Context(), req.SegmentID, req.UserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"matches": match})
}
