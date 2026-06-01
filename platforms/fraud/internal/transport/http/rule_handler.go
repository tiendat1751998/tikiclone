package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/fraud/internal/rules"
)

type createRuleRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Condition   string  `json:"condition" binding:"required"`
	Severity    int     `json:"severity" binding:"required"`
	Weight      float64 `json:"weight" binding:"required"`
	IsActive    bool    `json:"is_active"`
	Cooldown    int     `json:"cooldown"`
}

func (h *Handler) CreateRule(c *gin.Context) {
	var req createRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule := &rules.RuleDefinition{
		Name:        req.Name,
		Description: req.Description,
		Condition:   req.Condition,
		Severity:    req.Severity,
		Weight:      req.Weight,
		IsActive:    req.IsActive,
		Cooldown:    req.Cooldown,
	}

	if err := h.ruleSvc.CreateRule(c.Request.Context(), rule); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

func (h *Handler) ListRules(c *gin.Context) {
	rules, err := h.ruleSvc.ListRules(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rules})
}

func (h *Handler) UpdateRule(c *gin.Context) {
	var req createRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule := &rules.RuleDefinition{
		ID:          c.Param("id"),
		Name:        req.Name,
		Description: req.Description,
		Condition:   req.Condition,
		Severity:    req.Severity,
		Weight:      req.Weight,
		IsActive:    req.IsActive,
		Cooldown:    req.Cooldown,
	}

	if err := h.ruleSvc.UpdateRule(c.Request.Context(), rule); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

func (h *Handler) ToggleRule(c *gin.Context) {
	rule, err := h.ruleSvc.ToggleRule(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}
