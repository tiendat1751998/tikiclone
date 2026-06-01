package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/decisionlog"
)

type logDecisionRequest struct {
	EventID        string                    `json:"event_id" binding:"required"`
	EventType      string                    `json:"event_type" binding:"required"`
	UserID         string                    `json:"user_id" binding:"required"`
	Decision       decisionlog.DecisionType  `json:"decision" binding:"required"`
	RiskScore      float64                   `json:"risk_score"`
	TriggeredRules []string                  `json:"triggered_rules"`
}

func (h *Handler) LogDecision(c *gin.Context) {
	var req logDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log := &decisionlog.DecisionLog{
		EventID:        req.EventID,
		EventType:      req.EventType,
		UserID:         req.UserID,
		Decision:       req.Decision,
		RiskScore:      req.RiskScore,
		TriggeredRules: req.TriggeredRules,
	}

	if err := h.decLogger.LogDecision(c.Request.Context(), log); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, log)
}

func (h *Handler) GetDecisions(c *gin.Context) {
	logs, err := h.decLogger.GetDecisionHistory(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func (h *Handler) GetDecisionStats(c *gin.Context) {
	stats, err := h.decLogger.GetStats(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
