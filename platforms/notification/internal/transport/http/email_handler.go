package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/notification/internal/email"
)

type sendEmailRequest struct {
	To          []string               `json:"to" binding:"required"`
	CC          []string               `json:"cc,omitempty"`
	BCC         []string               `json:"bcc,omitempty"`
	Subject     string                 `json:"subject" binding:"required"`
	PlainText   string                 `json:"plain_text"`
	HTML        string                 `json:"html"`
	ReplyTo     string                 `json:"reply_to,omitempty"`
	Attachments []email.EmailAttachment `json:"attachments,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

func (h *Handler) SendEmail(c *gin.Context) {
	var req sendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	emailReq := &email.SendEmailRequest{
		To:          req.To,
		CC:          req.CC,
		BCC:         req.BCC,
		Subject:     req.Subject,
		PlainText:   req.PlainText,
		HTML:        req.HTML,
		ReplyTo:     req.ReplyTo,
		Attachments: req.Attachments,
		Metadata:    req.Metadata,
	}

	msg, err := h.email.SendEmail(c.Request.Context(), emailReq)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, msg)
}
