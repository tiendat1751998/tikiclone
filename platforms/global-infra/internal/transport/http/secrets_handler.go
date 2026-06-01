package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/secrets"
)

type createSecretRequest struct {
	Name           string `json:"name" binding:"required"`
	Value          string `json:"value" binding:"required"`
	ServiceName    string `json:"service_name" binding:"required"`
	RotationPeriod int    `json:"rotation_period"`
}

func (h *Handler) CreateSecret(c *gin.Context) {
	var req createSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	secret := &secrets.Secret{
		Name:           req.Name,
		Value:          req.Value,
		ServiceName:    req.ServiceName,
		RotationPeriod: req.RotationPeriod,
	}

	created, err := h.SecretSvc.Create(c.Request.Context(), secret)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, created)
}

func (h *Handler) ListSecrets(c *gin.Context) {
	serviceName := c.Query("service_name")

	secretsList, err := h.SecretSvc.List(c.Request.Context(), serviceName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"secrets": secretsList})
}

func (h *Handler) RotateSecret(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	rotated, err := h.SecretSvc.Rotate(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rotated)
}
