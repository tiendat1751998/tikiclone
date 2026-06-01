package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/notification/internal/notifier"
)

type sendNotificationRequest struct {
	UserID   string                 `json:"user_id" binding:"required"`
	Type     string                 `json:"type" binding:"required"`
	Channel  string                 `json:"channel" binding:"required"`
	Title    string                 `json:"title"`
	Body     string                 `json:"body"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Priority int                    `json:"priority,omitempty"`
}

type listNotificationsRequest struct {
	UserID string `form:"user_id" binding:"required"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

func (h *Handler) SendNotification(c *gin.Context) {
	var req sendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notifReq := &notifier.SendNotificationRequest{
		UserID:   req.UserID,
		Type:     notifier.NotificationType(req.Type),
		Channel:  notifier.Channel(req.Channel),
		Title:    req.Title,
		Body:     req.Body,
		Data:     req.Data,
		Priority: notifier.Priority(req.Priority),
	}

	n, err := h.notifier.Send(c.Request.Context(), notifReq)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, n)
}

func (h *Handler) ListNotifications(c *gin.Context) {
	var req listNotificationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	notifications, err := h.notifier.GetNotifications(c.Request.Context(), req.UserID, req.Limit, req.Offset)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

func (h *Handler) MarkNotificationRead(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.notifier.MarkRead(c.Request.Context(), id); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "read"})
}

func (h *Handler) DeleteNotification(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.notifier.Delete(c.Request.Context(), id); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
