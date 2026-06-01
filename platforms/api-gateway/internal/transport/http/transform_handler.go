package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/api-gateway/internal/transform"
)

type createTransformRuleRequest struct {
	ID             string                `json:"id" binding:"required"`
	MatchCondition transform.MatchCondition `json:"match_condition"`
	Actions        []transform.Action       `json:"actions" binding:"required"`
}

type applyTransformRequest struct {
	RuleID string                   `json:"rule_id" binding:"required"`
	Req    *transform.TransformRequest `json:"request" binding:"required"`
}

func (h *Handler) CreateTransformRule(c *gin.Context) {
	var req createTransformRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule := &transform.Rule{
		ID:             req.ID,
		MatchCondition: req.MatchCondition,
		Actions:        req.Actions,
	}

	if err := h.Transformer.CreateRule(rule); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

func (h *Handler) ApplyTransform(c *gin.Context) {
	var req applyTransformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Transformer.Apply(req.RuleID, req.Req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
