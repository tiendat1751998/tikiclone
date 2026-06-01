package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/analytics/internal/dashboard"
)

type createDashboardRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description,omitempty"`
	IsPublic    bool     `json:"is_public"`
	Tags        []string `json:"tags,omitempty"`
}

type updateDashboardRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description,omitempty"`
	IsPublic    bool     `json:"is_public"`
	Tags        []string `json:"tags,omitempty"`
}

type addWidgetRequest struct {
	Title      string                 `json:"title" binding:"required"`
	Type       string                 `json:"type" binding:"required"`
	Width      int                    `json:"width"`
	Height     int                    `json:"height"`
	PositionX  int                    `json:"position_x"`
	PositionY  int                    `json:"position_y"`
	DataSource dashboard.DataSource   `json:"data_source" binding:"required"`
	Config     map[string]interface{} `json:"config,omitempty"`
}

func (h *Handler) CreateDashboard(c *gin.Context) {
	var req createDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d, err := h.dashboardSvc.CreateDashboard(c.Request.Context(), req.Title, req.Description, "org-1", "user-1", req.IsPublic, req.Tags)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, d)
}

func (h *Handler) ListDashboards(c *gin.Context) {
	dashboards, total, err := h.dashboardSvc.ListDashboards(c.Request.Context(), "", 0, 50)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dashboards, "total": total})
}

func (h *Handler) GetDashboard(c *gin.Context) {
	id := c.Param("id")
	d, err := h.dashboardSvc.GetDashboard(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, d)
}

func (h *Handler) UpdateDashboard(c *gin.Context) {
	id := c.Param("id")
	var req updateDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d, err := h.dashboardSvc.UpdateDashboard(c.Request.Context(), id, req.Title, req.Description, req.IsPublic, req.Tags)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, d)
}

func (h *Handler) AddWidget(c *gin.Context) {
	dashboardID := c.Param("id")
	var req addWidgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	w, err := h.dashboardSvc.AddWidget(c.Request.Context(), dashboardID, req.Title, req.Type, req.Width, req.Height, req.PositionX, req.PositionY, req.DataSource, req.Config)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, w)
}
