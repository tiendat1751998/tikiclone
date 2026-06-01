package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/aiml/internal/featurestore"
)

type registerFeatureRequest struct {
	Name        string                `json:"name" binding:"required"`
	ValueType   string                `json:"value_type" binding:"required"`
	Entity      string                `json:"entity" binding:"required"`
	Source      string                `json:"source"`
	Description string                `json:"description"`
	IsOnline    bool                  `json:"is_online"`
}

type setFeatureValueRequest struct {
	FeatureName string      `json:"feature_name" binding:"required"`
	EntityID    string      `json:"entity_id" binding:"required"`
	Value       interface{} `json:"value" binding:"required"`
}

type batchGetRequest struct {
	FeatureNames []string `json:"feature_names" binding:"required"`
	EntityID     string   `json:"entity_id" binding:"required"`
}

func (h *Handler) RegisterFeature(c *gin.Context) {
	var req registerFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	feature := &featurestore.Feature{
		Name:        req.Name,
		ValueType:   featurestore.ValueType(req.ValueType),
		Entity:      featurestore.EntityType(req.Entity),
		Source:      req.Source,
		Description: req.Description,
		IsOnline:    req.IsOnline,
	}
	if err := h.featureSvc.RegisterFeature(c.Request.Context(), feature); err != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, feature)
}

func (h *Handler) ListFeatures(c *gin.Context) {
	features, err := h.featureSvc.ListFeatures(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, features)
}

func (h *Handler) SetFeatureValue(c *gin.Context) {
	var req setFeatureValueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	value := &featurestore.FeatureValue{
		FeatureName: req.FeatureName,
		EntityID:    req.EntityID,
		Value:       req.Value,
	}
	if err := h.featureSvc.SetFeatureValue(c.Request.Context(), value); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, value)
}

func (h *Handler) BatchGetFeatureValues(c *gin.Context) {
	var req batchGetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	values, err := h.featureSvc.BatchGet(c.Request.Context(), req.FeatureNames, req.EntityID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, values)
}
