package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/search-indexing/internal/pipeline"
)

type createPipelineRequest struct {
	Name      string                   `json:"name"`
	IndexName string                   `json:"index_name"`
	Stages    []pipeline.PipelineStage `json:"stages"`
}

func (h *Handler) CreatePipeline(c *gin.Context) {
	var req createPipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	p, err := h.pipeline.CreatePipeline(c.Request.Context(), req.Name, req.IndexName, req.Stages)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

type processDocumentRequest struct {
	PipelineID string                   `json:"pipeline_id"`
	Document   *pipeline.Document       `json:"document"`
}

func (h *Handler) ProcessDocument(c *gin.Context) {
	var req processDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	doc, err := h.pipeline.ProcessDocument(c.Request.Context(), req.PipelineID, req.Document)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, doc)
}

func (h *Handler) ListPipelines(c *gin.Context) {
	pipelines, err := h.pipeline.ListPipelines(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pipelines": pipelines})
}
