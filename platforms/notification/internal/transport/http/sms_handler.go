package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/notification/internal/sms"
)

type sendSMSRequest struct {
	To   string `json:"to" binding:"required"`
	Body string `json:"body" binding:"required"`
}

func (h *Handler) SendSMS(c *gin.Context) {
	var req sendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	smsReq := &sms.SendSMSRequest{
		To:   req.To,
		Body: req.Body,
	}

	msg, err := h.sms.SendSMS(c.Request.Context(), smsReq)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, msg)
}
