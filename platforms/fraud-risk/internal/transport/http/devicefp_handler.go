package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/devicefp"
)

type identifyDeviceRequest struct {
	UserAgent    string `json:"user_agent" binding:"required"`
	ScreenWidth  int    `json:"screen_width"`
	ScreenHeight int    `json:"screen_height"`
	ColorDepth   int    `json:"color_depth"`
	Platform     string `json:"platform"`
	Language     string `json:"language"`
	Timezone     string `json:"timezone"`
}

func (h *Handler) IdentifyDevice(c *gin.Context) {
	var req identifyDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fp := &devicefp.Fingerprint{
		UserAgent:    req.UserAgent,
		ScreenWidth:  req.ScreenWidth,
		ScreenHeight: req.ScreenHeight,
		ColorDepth:   req.ColorDepth,
		Platform:     req.Platform,
		Language:     req.Language,
		Timezone:     req.Timezone,
	}

	profile, isNew, err := h.deviceSvc.IdentifyDevice(c.Request.Context(), fp)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profile": profile, "is_new": isNew})
}

type markSuspiciousRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}

func (h *Handler) MarkSuspicious(c *gin.Context) {
	var req markSuspiciousRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.deviceSvc.MarkSuspicious(c.Request.Context(), req.DeviceID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}
