package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/core"
)

type analyzeBehaviorRequest struct {
	EventType string                `json:"event_type" binding:"required"`
	UserID    string                `json:"user_id" binding:"required"`
	IP        string                `json:"ip"`
	DeviceID  string                `json:"device_id"`
	Amount    float64               `json:"amount"`
	Currency  string                `json:"currency"`
	Metadata  map[string]interface{} `json:"metadata"`
}

func (h *Handler) AnalyzeBehavior(c *gin.Context) {
	var req analyzeBehaviorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ev := &core.Event{
		Type:     core.EventType(req.EventType),
		UserID:   req.UserID,
		IP:       req.IP,
		DeviceID: req.DeviceID,
		Amount:   req.Amount,
		Currency: req.Currency,
		Metadata: req.Metadata,
	}

	result, err := h.behavAnalyzer.AnalyzeBehavior(c.Request.Context(), ev)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

type buildProfileRequest struct {
	UserID         string `json:"user_id" binding:"required"`
	TypicalLoginHour int    `json:"typical_login_hour"`
	TypicalIPRange string `json:"typical_ip_range"`
	TypicalDevice  string `json:"typical_device"`
}

func (h *Handler) BuildProfile(c *gin.Context) {
	var req buildProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.behavAnalyzer.BuildProfile(
		c.Request.Context(),
		req.UserID,
		req.TypicalLoginHour,
		req.TypicalIPRange,
		req.TypicalDevice,
	)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}
