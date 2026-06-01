package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/configmanager"
)

type createConfigRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Environment string `json:"environment"`
	ServiceName string `json:"service_name" binding:"required"`
	IsEncrypted bool   `json:"is_encrypted"`
}

func (h *Handler) CreateConfig(c *gin.Context) {
	var req createConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry := &configmanager.ConfigEntry{
		Key:         req.Key,
		Value:       req.Value,
		Environment: configmanager.Environment(req.Environment),
		ServiceName: req.ServiceName,
		IsEncrypted: req.IsEncrypted,
	}

	created, err := h.ConfigSvc.Create(c.Request.Context(), entry)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, created)
}

func (h *Handler) ListConfigs(c *gin.Context) {
	serviceName := c.Query("service_name")
	env := c.Query("environment")

	configs, err := h.ConfigSvc.List(c.Request.Context(), serviceName, configmanager.Environment(env))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}
