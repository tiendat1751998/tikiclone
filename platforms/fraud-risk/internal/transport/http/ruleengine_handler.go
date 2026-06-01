package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/core"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/ruleengine"
)

type createRuleRequest struct {
	Name            string  `json:"name" binding:"required"`
	ConditionExpr   string  `json:"condition_expression" binding:"required"`
	Priority        int     `json:"priority"`
	Weight          float64 `json:"weight"`
	CooldownSeconds int     `json:"cooldown_seconds"`
	IsActive        bool    `json:"is_active"`
	ScoreDelta      float64 `json:"score_delta"`
}

func (h *Handler) CreateRule(c *gin.Context) {
	var req createRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule := &ruleengine.Rule{
		Name:            req.Name,
		ConditionExpr:   req.ConditionExpr,
		Priority:        req.Priority,
		Weight:          req.Weight,
		CooldownSeconds: req.CooldownSeconds,
		IsActive:        req.IsActive,
		ScoreDelta:      req.ScoreDelta,
	}

	if err := h.ruleEngine.CreateRule(c.Request.Context(), rule); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

func (h *Handler) ListRules(c *gin.Context) {
	rules, err := h.ruleEngine.ListRules(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rules})
}

type evaluateEventRequest struct {
	EventType string                `json:"event_type" binding:"required"`
	UserID    string                `json:"user_id" binding:"required"`
	IP        string                `json:"ip"`
	DeviceID  string                `json:"device_id"`
	Amount    float64               `json:"amount"`
	Currency  string                `json:"currency"`
	Metadata  map[string]interface{} `json:"metadata"`
}

func (h *Handler) EvaluateEvent(c *gin.Context) {
	var req evaluateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ev := &core.Event{
		Type:     core.EventType(req.EventType),
		UserID:   req.UserID,
		IP:       req.IP,
		DeviceID: req.DeviceID,
		Amount:   req.Amount,
		Currency: req.Currency,
		Metadata: req.Metadata,
	}

	results, err := h.ruleEngine.EvaluateEvent(c.Request.Context(), ev)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"evaluations": results})
}

type createRuleSetRequest struct {
	Name     string              `json:"name" binding:"required"`
	Rules    []ruleengine.Rule   `json:"rules"`
	Strategy ruleengine.Strategy `json:"strategy" binding:"required"`
}

func (h *Handler) CreateRuleSet(c *gin.Context) {
	var req createRuleSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rs := &ruleengine.RuleSet{
		Name:     req.Name,
		Rules:    req.Rules,
		Strategy: req.Strategy,
	}

	if err := h.ruleEngine.CreateRuleSet(c.Request.Context(), rs); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rs)
}

type evaluateRuleSetRequest struct {
	RuleSetID string                `json:"ruleset_id" binding:"required"`
	EventType string                `json:"event_type" binding:"required"`
	UserID    string                `json:"user_id" binding:"required"`
	IP        string                `json:"ip"`
	DeviceID  string                `json:"device_id"`
	Amount    float64               `json:"amount"`
	Currency  string                `json:"currency"`
	Metadata  map[string]interface{} `json:"metadata"`
}

func (h *Handler) EvaluateRuleSet(c *gin.Context) {
	var req evaluateRuleSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ev := &core.Event{
		Type:     core.EventType(req.EventType),
		UserID:   req.UserID,
		IP:       req.IP,
		DeviceID: req.DeviceID,
		Amount:   req.Amount,
		Currency: req.Currency,
		Metadata: req.Metadata,
	}

	result, err := h.ruleEngine.EvaluateRuleSet(c.Request.Context(), req.RuleSetID, ev)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
