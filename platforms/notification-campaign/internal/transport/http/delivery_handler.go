package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/deliveryopt"
)

type optimizeTimeRequest struct {
	UserID  string `json:"user_id" binding:"required"`
	Channel string `json:"channel" binding:"required"`
}

type sendMessageRequest struct {
	UserID    string                 `json:"user_id" binding:"required"`
	Channel   string                 `json:"channel"`
	Subject   string                 `json:"subject"`
	Body      string                 `json:"body"`
	Priority  int                    `json:"priority"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

func (h *Handler) OptimizeSendTime(c *gin.Context) {
	var req optimizeTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.deliverySvc.OptimizeSendTime(c.Request.Context(), req.UserID, req.Channel)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) SendMessage(c *gin.Context) {
	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sendReq := &deliveryopt.SendRequest{
		UserID:    req.UserID,
		Channel:   req.Channel,
		Subject:   req.Subject,
		Body:      req.Body,
		Priority:  deliveryopt.PriorityLevel(req.Priority),
		Variables: req.Variables,
	}

	result, err := h.deliverySvc.SendWithFallback(c.Request.Context(), sendReq, nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
