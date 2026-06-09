package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/rec-vector/internal/application"
	"github.com/tikiclone/tiki/services/rec-vector/internal/domain"
	"go.uber.org/zap"
)

type Handler struct {
	svc *application.RecService
}

func NewHandler(svc *application.RecService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetRecommendations(c *gin.Context) {
	userID := c.Query("user_id")
	collectionID := c.Query("collection_id")
	algorithm := c.Query("algorithm")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if uid, exists := c.Get("user_id"); exists {
		if u, ok := uid.(string); ok && u != "" {
			userID = u
		}
	}

	recs, err := h.svc.GetRecommendations(c.Request.Context(), userID, collectionID, domain.RecommendAlgorithm(algorithm), limit)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": recs})
}

func (h *Handler) ListCollections(c *gin.Context) {
	cols, err := h.svc.ListCollections(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": cols})
}

func (h *Handler) CreateCollection(c *gin.Context) {
	var req application.CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	col, err := h.svc.CreateCollection(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, col)
}

func (h *Handler) GetCollection(c *gin.Context) {
	id := c.Param("id")
	col, err := h.svc.GetCollection(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, col)
}

func (h *Handler) AddMetadata(c *gin.Context) {
	collectionID := c.Param("id")
	var req application.AddMetadataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.CollectionID = collectionID
	m, err := h.svc.AddMetadata(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *Handler) ListMetadata(c *gin.Context) {
	collectionID := c.Param("id")
	items, err := h.svc.ListMetadata(c.Request.Context(), collectionID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) ListModels(c *gin.Context) {
	models, err := h.svc.ListModels(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": models})
}

func (h *Handler) GetModel(c *gin.Context) {
	id := c.Param("id")
	m, err := h.svc.GetModel(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *Handler) ListJobs(c *gin.Context) {
	jobs, err := h.svc.ListJobs(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": jobs})
}

func (h *Handler) CreateJob(c *gin.Context) {
	var req struct {
		CollectionID string `json:"collection_id" binding:"required"`
		JobType      string `json:"job_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.JobType == "" {
		req.JobType = "incremental_update"
	}
	job, err := h.svc.CreateJob(c.Request.Context(), req.CollectionID, req.JobType)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, job)
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrCollectionNotFound), errors.Is(err, domain.ErrModelNotFound), errors.Is(err, domain.ErrJobNotFound), errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, domain.ErrUnauthorized):
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
	default:
		zap.L().Error("unexpected error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
