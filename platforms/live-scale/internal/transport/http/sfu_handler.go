package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/sfu"
)

func (h *Handler) RegisterSFUNode(c *gin.Context) {
	var req struct {
		ID       string `json:"id" binding:"required"`
		Address  string `json:"address" binding:"required"`
		Region   string `json:"region" binding:"required"`
		Capacity int    `json:"capacity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	node := &sfu.SFUNode{
		ID:       req.ID,
		Address:  req.Address,
		Region:   req.Region,
		Capacity: req.Capacity,
	}
	if err := h.sfu.RegisterNode(c.Request.Context(), node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, node)
}

func (h *Handler) GetOptimalSFUNode(c *gin.Context) {
	streamID := c.Query("stream_id")
	region := c.Query("region")
	if streamID == "" || region == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream_id and region required"})
		return
	}
	node, err := h.sfu.SelectOptimalNode(c.Request.Context(), streamID, region)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, node)
}
