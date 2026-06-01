package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/aiml/internal/modelregistry"
)

type registerModelRequest struct {
	Name         string  `json:"name" binding:"required"`
	Version      string  `json:"version" binding:"required"`
	Type         string  `json:"type" binding:"required"`
	Framework    string  `json:"framework" binding:"required"`
	ArtifactPath string  `json:"artifact_path"`
}

type promoteModelRequest struct {
	Stage string `json:"stage" binding:"required"`
}

func (h *Handler) RegisterModel(c *gin.Context) {
	var req registerModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	model := &modelregistry.Model{
		Name:         req.Name,
		Version:      req.Version,
		Type:         modelregistry.ModelType(req.Type),
		Framework:    modelregistry.Framework(req.Framework),
		Status:       modelregistry.StageDevelopment,
		ArtifactPath: req.ArtifactPath,
	}
	if err := h.modelSvc.Register(c.Request.Context(), model); err != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, model)
}

func (h *Handler) ListModels(c *gin.Context) {
	models, err := h.modelSvc.List(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models)
}

func (h *Handler) PromoteModel(c *gin.Context) {
	id := c.Param("id")
	var req promoteModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.modelSvc.Promote(c.Request.Context(), id, modelregistry.Stage(req.Stage)); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "promoted"})
}
