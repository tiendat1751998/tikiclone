package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/aiml/internal/inference"
)

type predictRequest struct {
	ModelName    string                 `json:"model_name" binding:"required"`
	ModelVersion string                 `json:"model_version"`
	Input        map[string]interface{} `json:"input" binding:"required"`
	Features     map[string]interface{} `json:"features,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

type batchPredictRequest struct {
	Requests []predictRequest `json:"requests" binding:"required"`
}

func (h *Handler) Predict(c *gin.Context) {
	var req predictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.inferenceSvc.Predict(c.Request.Context(), &inference.InferenceRequest{
		ModelName:    req.ModelName,
		ModelVersion: req.ModelVersion,
		Input:        req.Input,
		Features:     req.Features,
		Context:      req.Context,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) BatchPredict(c *gin.Context) {
	var req batchPredictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requests := make([]*inference.InferenceRequest, len(req.Requests))
	for i, r := range req.Requests {
		requests[i] = &inference.InferenceRequest{
			ModelName:    r.ModelName,
			ModelVersion: r.ModelVersion,
			Input:        r.Input,
			Features:     r.Features,
			Context:      r.Context,
		}
	}
	results, err := h.inferenceSvc.BatchPredict(c.Request.Context(), requests)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}
