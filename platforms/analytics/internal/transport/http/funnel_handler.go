package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/analytics/internal/funnel"
)

type analyzeFunnelRequest struct {
	Name      string             `json:"name" binding:"required"`
	Steps     []funnelStepReq    `json:"steps" binding:"required"`
	TimeRange string             `json:"time_range"`
}

type funnelStepReq struct {
	Name      string `json:"name" binding:"required"`
	EventType string `json:"event_type" binding:"required"`
	Order     int    `json:"order"`
	Window    string `json:"window,omitempty"`
}

func (h *Handler) AnalyzeFunnel(c *gin.Context) {
	var req analyzeFunnelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var steps []funnel.FunnelStep
	for _, s := range req.Steps {
		steps = append(steps, funnel.FunnelStep{
			Name:      s.Name,
			EventType: s.EventType,
			Order:     s.Order,
			Window:    s.Window,
		})
	}

	definition := &funnel.FunnelDefinition{
		Name:      req.Name,
		Steps:     steps,
		TimeRange: req.TimeRange,
	}

	result, err := h.funnelSvc.BuildFunnel(c.Request.Context(), definition)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
