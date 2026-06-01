package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/search-indexing/internal/coordinator"
)

type registerNodeRequest struct {
	ID              string  `json:"id"`
	Address         string  `json:"address"`
	Region          string  `json:"region"`
	AvailableShards int     `json:"available_shards"`
	LoadPercentage  float64 `json:"load_percentage"`
}

func (h *Handler) RegisterNode(c *gin.Context) {
	var req registerNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	node := &coordinator.IndexNode{
		ID:              req.ID,
		Address:         req.Address,
		Region:          req.Region,
		AvailableShards: req.AvailableShards,
		LoadPercentage:  req.LoadPercentage,
	}
	result, err := h.coordinator.RegisterNode(c.Request.Context(), node)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListNodes(c *gin.Context) {
	nodes, err := h.coordinator.ListNodes(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}

type assignShardRequest struct {
	ID          string `json:"id"`
	IndexName   string `json:"index_name"`
	ShardNumber int    `json:"shard_number"`
	DocCount    int64  `json:"doc_count"`
	SizeBytes   int64  `json:"size_bytes"`
}

func (h *Handler) AssignShard(c *gin.Context) {
	var req assignShardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	shard := &coordinator.IndexShard{
		ID:          req.ID,
		IndexName:   req.IndexName,
		ShardNumber: req.ShardNumber,
		DocCount:    req.DocCount,
		SizeBytes:   req.SizeBytes,
	}
	result, err := h.coordinator.AssignShard(c.Request.Context(), shard)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) Rebalance(c *gin.Context) {
	if err := h.coordinator.RebalanceShards(c.Request.Context()); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "rebalance completed"})
}
