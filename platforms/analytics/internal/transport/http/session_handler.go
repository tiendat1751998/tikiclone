package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/analytics/internal/session"
)

func (h *Handler) GetSessions(c *gin.Context) {
	filter := &session.SessionFilter{}

	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = userID
	}
	if source := c.Query("source"); source != "" {
		filter.Source = source
	}
	if device := c.Query("device"); device != "" {
		filter.Device = device
	}
	if country := c.Query("country"); country != "" {
		filter.Country = country
	}
	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.StartTime = &t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.EndTime = &t
		}
	}
	filter.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "50"))
	filter.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))

	sessions, total, err := h.sessionSvc.GetSessions(c.Request.Context(), filter)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sessions, "total": total})
}
