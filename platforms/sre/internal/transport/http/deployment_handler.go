package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/sre/internal/deployment"
)

type CreateDeploymentReq struct {
	Service  string              `json:"service" binding:"required"`
	Version  string              `json:"version" binding:"required"`
	Strategy deployment.Strategy `json:"strategy"`
}

func (h *Handler) CreateDeployment(c *gin.Context) {
	var req CreateDeploymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Strategy == "" {
		req.Strategy = deployment.StrategyRolling
	}
	d, err := h.deploymentSvc.Create(req.Service, req.Version, req.Strategy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *Handler) ListDeployments(c *gin.Context) {
	deployments, err := h.deploymentSvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deployments)
}

func (h *Handler) ApproveDeployment(c *gin.Context) {
	id := c.Param("id")
	d, err := h.deploymentSvc.Approve(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *Handler) RollbackDeployment(c *gin.Context) {
	id := c.Param("id")
	d, err := h.deploymentSvc.Rollback(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, d)
}
