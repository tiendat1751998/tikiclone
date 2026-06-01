package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/notification/internal/preferences"
)

type getPreferenceRequest struct {
	UserID string `form:"user_id" binding:"required"`
}

type updatePreferenceRequest struct {
	UserID      string                         `json:"user_id" binding:"required"`
	ChannelOptIn *preferences.ChannelOptIn      `json:"channel_opt_in,omitempty"`
	Categories  *preferences.CategoryPreferences `json:"categories,omitempty"`
	QuietHours  *preferences.QuietHours          `json:"quiet_hours,omitempty"`
	EmailDigest *bool                            `json:"email_digest,omitempty"`
	PushEnabled *bool                            `json:"push_enabled,omitempty"`
}

func (h *Handler) GetPreferences(c *gin.Context) {
	var req getPreferenceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pref, err := h.preferences.GetPreferences(c.Request.Context(), req.UserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pref)
}

func (h *Handler) UpdatePreferences(c *gin.Context) {
	var req updatePreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateReq := &preferences.UpdatePreferenceRequest{
		ChannelOptIn: req.ChannelOptIn,
		Categories:   req.Categories,
		QuietHours:   req.QuietHours,
		EmailDigest:  req.EmailDigest,
		PushEnabled:  req.PushEnabled,
	}

	pref, err := h.preferences.UpdatePreferences(c.Request.Context(), req.UserID, updateReq)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pref)
}
