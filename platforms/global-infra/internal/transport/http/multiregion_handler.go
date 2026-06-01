package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/multiregion"
)

type createRegionRequest struct {
	Name           string            `json:"name" binding:"required"`
	Code           string            `json:"code" binding:"required"`
	IsActive       bool              `json:"is_active"`
	FailoverRegion string            `json:"failover_region,omitempty"`
	Endpoints      map[string]string `json:"endpoints,omitempty"`
}

func (h *Handler) ListRegions(c *gin.Context) {
	var req createRegionRequest
	if err := c.ShouldBindJSON(&req); err == nil && req.Code != "" {
		region := &multiregion.Region{
			Name:           req.Name,
			Code:           req.Code,
			IsActive:       req.IsActive,
			FailoverRegion: req.FailoverRegion,
			Endpoints:      req.Endpoints,
		}
		created, err := h.MultiRegionSvc.Create(c.Request.Context(), region)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, created)
		return
	}

	regions, err := h.MultiRegionSvc.List(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"regions": regions})
}

func (h *Handler) GetFailover(c *gin.Context) {
	code := c.Query("region")
	serviceName := c.Query("service")
	if code == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "region is required"})
		return
	}
	if serviceName == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "service is required"})
		return
	}

	result, err := h.MultiRegionSvc.GetFailoverStrategy(c.Request.Context(), code, serviceName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
