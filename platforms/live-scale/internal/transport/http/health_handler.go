package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/stream_health"
)

func (h *Handler) ReportStreamHealth(c *gin.Context) {
	var req struct {
		StreamID   string  `json:"stream_id" binding:"required"`
		NodeID     string  `json:"node_id"`
		Region     string  `json:"region"`
		Bitrate    int     `json:"bitrate"`
		FrameRate  float64 `json:"frame_rate"`
		LatencyMs  int     `json:"latency_ms"`
		PacketLoss float64 `json:"packet_loss"`
		Viewers    int     `json:"viewers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	metric := stream_health.HealthMetric{
		Bitrate:    req.Bitrate,
		FrameRate:  req.FrameRate,
		LatencyMs:  req.LatencyMs,
		PacketLoss: req.PacketLoss,
		Viewers:    req.Viewers,
		RecordedAt: time.Now().UTC(),
	}
	health, err := h.health.ReportHealth(c.Request.Context(), req.StreamID, metric)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stream_id": health.StreamID, "status": health.Status})
}

func (h *Handler) GetStreamHealth(c *gin.Context) {
	id := c.Param("id")
	health, err := h.health.GetStreamHealth(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stream health not found"})
		return
	}
	c.JSON(http.StatusOK, health)
}
