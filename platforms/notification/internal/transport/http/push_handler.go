package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/notification/internal/push"
)

type registerDeviceRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform" binding:"required"`
}

type sendPushRequest struct {
	UserID string            `json:"user_id" binding:"required"`
	Title  string            `json:"title" binding:"required"`
	Body   string            `json:"body" binding:"required"`
	Data   map[string]string `json:"data,omitempty"`
}

func (h *Handler) RegisterDevice(c *gin.Context) {
	var req registerDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device, err := h.push.RegisterDevice(c.Request.Context(), req.UserID, req.Token, push.Platform(req.Platform))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, device)
}

func (h *Handler) SendPush(c *gin.Context) {
	var req sendPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pushReq := &push.PushNotificationRequest{
		UserID: req.UserID,
		Title:  req.Title,
		Body:   req.Body,
		Data:   req.Data,
	}

	result, err := h.push.SendPush(c.Request.Context(), pushReq)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
