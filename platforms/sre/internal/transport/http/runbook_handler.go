package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/sre/internal/runbook"
)

type CreateRunbookReq struct {
	Title        string       `json:"title" binding:"required"`
	Service      string       `json:"service" binding:"required"`
	IncidentType string       `json:"incident_type" binding:"required"`
	Steps        []runbook.Step `json:"steps" binding:"required"`
}

func (h *Handler) CreateRunbook(c *gin.Context) {
	var req CreateRunbookReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rb, err := h.runbookSvc.Create(req.Title, req.Service, req.IncidentType, req.Steps)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rb)
}

func (h *Handler) ListRunbooks(c *gin.Context) {
	filter := runbook.Filter{
		Service:      c.Query("service"),
		IncidentType: c.Query("incident_type"),
	}
	runbooks, err := h.runbookSvc.List(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, runbooks)
}
