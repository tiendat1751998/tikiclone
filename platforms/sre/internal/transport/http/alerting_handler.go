package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/sre/internal/alerting"
)

type CreateRuleReq struct {
	Name            string  `json:"name" binding:"required"`
	MetricName      string  `json:"metric_name" binding:"required"`
	Operator        string  `json:"operator" binding:"required"`
	Threshold       float64 `json:"threshold" binding:"required"`
	DurationSeconds int     `json:"duration_seconds"`
	CooldownSeconds int     `json:"cooldown_seconds"`
}

type EvaluateReq struct {
	Metrics []alerting.MetricValue `json:"metrics" binding:"required"`
}

func (h *Handler) CreateAlertRule(c *gin.Context) {
	var req CreateRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule := &alerting.Rule{
		Name:            req.Name,
		MetricName:      req.MetricName,
		Operator:        req.Operator,
		Threshold:       req.Threshold,
		DurationSeconds: req.DurationSeconds,
		CooldownSeconds: req.CooldownSeconds,
	}
	if err := h.alertingSvc.CreateRule(rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rule)
}

func (h *Handler) ListAlertRules(c *gin.Context) {
	rules, err := h.alertingSvc.ListRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func (h *Handler) EvaluateAlerts(c *gin.Context) {
	var req EvaluateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for i := range req.Metrics {
		req.Metrics[i].Timestamp = time.Now()
	}
	rules, err := h.alertingSvc.ListRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	alerts, err := h.alertingSvc.Evaluate(rules, req.Metrics)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

func (h *Handler) ListAlerts(c *gin.Context) {
	alerts, err := h.alertingSvc.ListAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}
