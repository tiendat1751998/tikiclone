package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/sre/internal/slo"
)

type CreateSLOReq struct {
	Name             string    `json:"name" binding:"required"`
	Service          string    `json:"service" binding:"required"`
	SLIMetric        string    `json:"sli_metric" binding:"required"`
	TargetPercentage float64   `json:"target_percentage"`
	Window           slo.Window `json:"window"`
}

func (h *Handler) CreateSLO(c *gin.Context) {
	var req CreateSLOReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.TargetPercentage == 0 {
		req.TargetPercentage = 99.9
	}
	if req.Window == "" {
		req.Window = slo.Window28d
	}
	sl, err := h.sloSvc.CreateSLO(req.Name, req.Service, req.SLIMetric, req.TargetPercentage, req.Window)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sl)
}

func (h *Handler) ListSLOs(c *gin.Context) {
	slos, err := h.sloSvc.ListSLOs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, slos)
}

func (h *Handler) GetSLOReport(c *gin.Context) {
	id := c.Param("id")
	report, err := h.sloSvc.GetSLOReport(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}
