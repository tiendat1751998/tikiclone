package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/fraud/internal/core"
)

type evaluateRequest struct {
	Type      string                 `json:"type" binding:"required"`
	UserID    string                 `json:"user_id" binding:"required"`
	IP        string                 `json:"ip"`
	DeviceID  string                 `json:"device_id"`
	Amount    float64                `json:"amount"`
	Currency  string                 `json:"currency"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

func (h *Handler) Evaluate(c *gin.Context) {
	var req evaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event := &core.FraudEvent{
		ID:        uuid.New().String(),
		Type:      core.EventType(req.Type),
		UserID:    req.UserID,
		IP:        req.IP,
		DeviceID:  req.DeviceID,
		Amount:    req.Amount,
		Currency:  req.Currency,
		Timestamp: time.Now(),
		Metadata:  req.Metadata,
	}

	score, err := h.detectSvc.Evaluate(c.Request.Context(), event)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, score)
}

func (h *Handler) ListAlerts(c *gin.Context) {
	status := c.DefaultQuery("status", "")
	riskLevel := core.RiskLevel(c.DefaultQuery("risk_level", ""))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	alerts, total, err := h.detectSvc.ListAlerts(c.Request.Context(), status, riskLevel, offset, limit)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": alerts, "total": total})
}

func (h *Handler) ResolveAlert(c *gin.Context) {
	var req struct {
		ResolvedBy string `json:"resolved_by" binding:"required"`
		Resolution string `json:"resolution" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.detectSvc.ResolveAlert(c.Request.Context(), c.Param("id"), req.ResolvedBy, req.Resolution); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "resolved"})
}
