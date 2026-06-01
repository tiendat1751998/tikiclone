package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/aiml/internal/training"
)

type createTrainingJobRequest struct {
	Name            string            `json:"name" binding:"required"`
	ModelName       string            `json:"model_name" binding:"required"`
	Dataset         string            `json:"dataset" binding:"required"`
	Hyperparameters map[string]string `json:"hyperparameters"`
}

func (h *Handler) CreateTrainingJob(c *gin.Context) {
	var req createTrainingJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job := &training.TrainingJob{
		ID:              uuid.New().String(),
		Name:            req.Name,
		ModelName:       req.ModelName,
		Dataset:         req.Dataset,
		Hyperparameters: req.Hyperparameters,
	}
	if err := h.trainingSvc.Create(c.Request.Context(), job); err != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, job)
}

func (h *Handler) ListTrainingJobs(c *gin.Context) {
	jobs, err := h.trainingSvc.List(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func (h *Handler) GetTrainingJob(c *gin.Context) {
	id := c.Param("id")
	job, err := h.trainingSvc.Get(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, job)
}
