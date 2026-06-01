package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/fraud/internal/verification"
)

type initiateVerificationRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Method string `json:"method" binding:"required"`
	Target string `json:"target" binding:"required"`
}

func (h *Handler) InitiateVerification(c *gin.Context) {
	var req initiateVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vReq, err := h.verificationSvc.InitiateVerification(
		c.Request.Context(),
		req.UserID,
		verification.VerificationMethod(req.Method),
		req.Target,
	)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"verification_id": vReq.ID,
		"method":          vReq.Method,
		"target":          vReq.Target,
		"status":          vReq.Status,
		"expires_at":      vReq.ExpiresAt,
	})
}

type verifyCodeRequest struct {
	VerificationID string `json:"verification_id" binding:"required"`
	Code           string `json:"code" binding:"required"`
}

func (h *Handler) VerifyCode(c *gin.Context) {
	var req verifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vReq, err := h.verificationSvc.VerifyCode(c.Request.Context(), req.VerificationID, req.Code)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   vReq.Status,
		"verified": vReq.Status == verification.StatusVerified,
	})
}
