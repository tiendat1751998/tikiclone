package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/riskscoring"
)

type calculateRiskScoreRequest struct {
	Factors []riskscoring.RiskFactor `json:"factors" binding:"required"`
}

func (h *Handler) CalculateRiskScore(c *gin.Context) {
	var req calculateRiskScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card := h.riskCalc.Calculate(c.Request.Context(), req.Factors)
	c.JSON(http.StatusOK, card)
}

func (h *Handler) GetRiskFactors(c *gin.Context) {
	factors := h.riskCalc.ListFactors(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"data": factors})
}
