package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/transcoding"
)

func (h *Handler) CreateTranscodeJob(c *gin.Context) {
	var req struct {
		StreamID string   `json:"stream_id" binding:"required"`
		InputURL string   `json:"input_url" binding:"required"`
		Profiles []string `json:"profiles"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var profiles []transcoding.VideoProfile
	for _, p := range req.Profiles {
		profiles = append(profiles, transcoding.VideoProfile(p))
	}
	job, err := h.transcode.CreateJob(c.Request.Context(), req.StreamID, req.InputURL, profiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, job)
}

func (h *Handler) GetTranscodeJob(c *gin.Context) {
	id := c.Param("id")
	job, err := h.transcode.GetJobStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transcode job not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}
