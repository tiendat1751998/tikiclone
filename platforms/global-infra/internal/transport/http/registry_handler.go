package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/registry"
)

type registerServiceRequest struct {
	ID             string            `json:"id" binding:"required"`
	Name           string            `json:"name" binding:"required"`
	Version        string            `json:"version"`
	Address        string            `json:"address" binding:"required"`
	Port           int               `json:"port" binding:"required"`
	Region         string            `json:"region"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	HealthEndpoint string            `json:"health_endpoint"`
}

type heartbeatRequest struct {
	ID string `json:"id" binding:"required"`
}

func (h *Handler) RegisterService(c *gin.Context) {
	var req registerServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	instance := &registry.ServiceInstance{
		ID:             req.ID,
		Name:           req.Name,
		Version:        req.Version,
		Address:        req.Address,
		Port:           req.Port,
		Region:         req.Region,
		Metadata:       req.Metadata,
		HealthEndpoint: req.HealthEndpoint,
	}

	created, err := h.RegistrySvc.Register(c.Request.Context(), instance)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, created)
}

func (h *Handler) Heartbeat(c *gin.Context) {
	var req heartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.RegistrySvc.Heartbeat(c.Request.Context(), req.ID); err != nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) DiscoverService(c *gin.Context) {
	name := c.Query("name")
	region := c.Query("region")
	if name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	instances, err := h.RegistrySvc.Discover(c.Request.Context(), name, region)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"instances": instances})
}
