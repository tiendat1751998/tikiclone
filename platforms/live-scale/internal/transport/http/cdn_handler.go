package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/cdn"
)

func (h *Handler) PurgeCDNCache(c *gin.Context) {
	var req struct {
		URLs    []string `json:"urls"`
		Pattern string   `json:"pattern"`
		Tags    []string `json:"tags"`
		Reason  string   `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	purgeReq := &cdn.CDNPurgeRequest{
		URLs:    req.URLs,
		Pattern: req.Pattern,
		Tags:    req.Tags,
		Reason:  req.Reason,
	}
	if err := h.cdn.PurgeCache(c.Request.Context(), purgeReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "purge request submitted", "id": purgeReq.ID})
}
