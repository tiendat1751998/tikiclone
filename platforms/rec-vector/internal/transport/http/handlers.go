package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/similarity"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/vectorstore"
)

type insertVectorRequest struct {
	ID        string                 `json:"id"`
	Vector    []float64              `json:"vector" binding:"required"`
	Metadata  map[string]interface{} `json:"metadata"`
	Namespace string                 `json:"namespace" binding:"required"`
}

func (h *Handler) InsertVector(c *gin.Context) {
	var req insertVectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	record := &vectorstore.VectorRecord{
		ID:        req.ID,
		Vector:    req.Vector,
		Metadata:  req.Metadata,
		Namespace: req.Namespace,
	}
	if err := h.vectorStore.Insert(c.Request.Context(), record); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, record)
}

type batchInsertRequest struct {
	Records   []insertVectorRequest `json:"records" binding:"required"`
	Namespace string                `json:"namespace"`
}

func (h *Handler) BatchInsertVectors(c *gin.Context) {
	var req batchInsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	records := make([]*vectorstore.VectorRecord, len(req.Records))
	for i, r := range req.Records {
		ns := r.Namespace
		if ns == "" {
			ns = req.Namespace
		}
		records[i] = &vectorstore.VectorRecord{
			ID:        r.ID,
			Vector:    r.Vector,
			Metadata:  r.Metadata,
			Namespace: ns,
		}
	}
	if err := h.vectorStore.BatchInsert(c.Request.Context(), records); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, records)
}

type searchVectorsRequest struct {
	QueryEmbedding []float64 `json:"query_embedding" binding:"required"`
	Namespace      string    `json:"namespace" binding:"required"`
	TopK           int       `json:"top_k"`
}

func (h *Handler) SearchVectors(c *gin.Context) {
	var req searchVectorsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}
	results, err := h.vectorStore.Search(c.Request.Context(), req.QueryEmbedding, req.Namespace, topK)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func (h *Handler) DeleteVector(c *gin.Context) {
	id := c.Param("id")
	namespace := c.Query("namespace")
	if namespace == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "namespace query param required"})
		return
	}
	if err := h.vectorStore.Delete(c.Request.Context(), id, namespace); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

type generateUserEmbeddingRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	ModelVersion string `json:"model_version"`
}

func (h *Handler) GenerateUserEmbedding(c *gin.Context) {
	var req generateUserEmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mv := req.ModelVersion
	if mv == "" {
		mv = "v1"
	}
	emb, err := h.userEmbSvc.GenerateUserEmbedding(c.Request.Context(), req.UserID, mv)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, emb)
}

func (h *Handler) GetUserEmbedding(c *gin.Context) {
	userID := c.Param("id")
	emb, err := h.userEmbSvc.GetEmbedding(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, emb)
}

type updateUserEmbeddingRequest struct {
	Embedding    []float64 `json:"embedding" binding:"required"`
	ModelVersion string    `json:"model_version"`
}

func (h *Handler) UpdateUserEmbedding(c *gin.Context) {
	userID := c.Param("id")
	var req updateUserEmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mv := req.ModelVersion
	if mv == "" {
		mv = "v1"
	}
	emb, err := h.userEmbSvc.UpdateEmbedding(c.Request.Context(), userID, req.Embedding, mv)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, emb)
}

type generateItemEmbeddingRequest struct {
	ItemID       string   `json:"item_id" binding:"required"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	ModelVersion string   `json:"model_version"`
}

func (h *Handler) GenerateItemEmbedding(c *gin.Context) {
	var req generateItemEmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mv := req.ModelVersion
	if mv == "" {
		mv = "v1"
	}
	emb, err := h.itemEmbSvc.GenerateItemEmbedding(c.Request.Context(), req.ItemID, req.Category, req.Tags, mv)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, emb)
}

func (h *Handler) GetItemEmbedding(c *gin.Context) {
	itemID := c.Param("id")
	emb, err := h.itemEmbSvc.GetEmbedding(c.Request.Context(), itemID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, emb)
}

type similaritySearchRequest struct {
	QueryEmbedding []float64              `json:"query_embedding" binding:"required"`
	Namespace      string                 `json:"namespace" binding:"required"`
	TopK           int                    `json:"top_k"`
	MinScore       float64                `json:"min_score"`
	Filter         map[string]interface{} `json:"filter"`
}

func (h *Handler) SimilaritySearch(c *gin.Context) {
	var req similaritySearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}
	results, err := h.similaritySvc.Search(c.Request.Context(), &similarity.SimilarityRequest{
		QueryEmbedding: req.QueryEmbedding,
		Namespace:      req.Namespace,
		TopK:           topK,
		MinScore:       req.MinScore,
		Filter:         req.Filter,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

type hybridSearchRequest struct {
	QueryEmbedding []float64              `json:"query_embedding" binding:"required"`
	Namespace      string                 `json:"namespace" binding:"required"`
	Keyword        string                 `json:"keyword"`
	TopK           int                    `json:"top_k"`
	MinScore       float64                `json:"min_score"`
	Filter         map[string]interface{} `json:"filter"`
}

func (h *Handler) HybridSearch(c *gin.Context) {
	var req hybridSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}
	results, err := h.similaritySvc.HybridSearch(c.Request.Context(), &similarity.SimilarityRequest{
		QueryEmbedding: req.QueryEmbedding,
		Namespace:      req.Namespace,
		TopK:           topK,
		MinScore:       req.MinScore,
		Filter:         req.Filter,
		Keyword:        req.Keyword,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

type interactRequest struct {
	UserID          string `json:"user_id" binding:"required"`
	ItemID          string `json:"item_id" binding:"required"`
	InteractionType string `json:"interaction_type" binding:"required"`
}

func (h *Handler) RecordInteraction(c *gin.Context) {
	var req interactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.collabSvc.RecordInteraction(c.Request.Context(), req.UserID, req.ItemID, req.InteractionType); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "interaction recorded"})
}

type collabRecommendRequest struct {
	UserID string `json:"user_id" binding:"required"`
	TopK   int    `json:"top_k"`
}

func (h *Handler) CollaborativeRecommend(c *gin.Context) {
	var req collabRecommendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}
	if err := h.collabSvc.TrainFactorization(c.Request.Context(), 10, 10, 0.01); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	recs, err := h.collabSvc.RecommendByFactorization(c.Request.Context(), req.UserID, topK)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, recs)
}

type trackEventRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	SessionID string `json:"session_id" binding:"required"`
	EventType string `json:"event_type" binding:"required"`
	ItemID    string `json:"item_id"`
	Query     string `json:"query"`
}

func (h *Handler) TrackEvent(c *gin.Context) {
	var req trackEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session, err := h.realtimeSvc.TrackEvent(c.Request.Context(), req.UserID, req.SessionID, req.EventType, req.ItemID, req.Query)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, session)
}

type rtRecommendRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	TopK      int    `json:"top_k"`
}

func (h *Handler) RealtimeRecommend(c *gin.Context) {
	var req rtRecommendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}
	results, err := h.realtimeSvc.RecommendWithContext(c.Request.Context(), req.SessionID, req.Namespace, topK)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}
