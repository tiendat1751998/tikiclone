package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/sre/internal/incident"
)

type CreateIncidentReq struct {
	Title       string            `json:"title" binding:"required"`
	Severity    incident.Severity `json:"severity" binding:"required"`
	Service     string            `json:"service" binding:"required"`
	Region      string            `json:"region"`
	Description string            `json:"description"`
	Assignee    string            `json:"assignee"`
}

func (h *Handler) CreateIncident(c *gin.Context) {
	var req CreateIncidentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	inc, err := h.incidentSvc.Create(req.Title, req.Severity, req.Service, req.Region, req.Description, req.Assignee)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, inc)
}

func (h *Handler) ListIncidents(c *gin.Context) {
	filter := incident.Filter{
		Status:   incident.Status(c.Query("status")),
		Severity: incident.Severity(c.Query("severity")),
		Service:  c.Query("service"),
	}
	incidents, err := h.incidentSvc.List(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, incidents)
}

func (h *Handler) AcknowledgeIncident(c *gin.Context) {
	id := c.Param("id")
	inc, err := h.incidentSvc.Acknowledge(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inc)
}

func (h *Handler) ResolveIncident(c *gin.Context) {
	id := c.Param("id")
	inc, err := h.incidentSvc.Resolve(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inc)
}
