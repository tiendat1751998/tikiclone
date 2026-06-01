package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/search/internal/search"
)

type indexDocumentRequest struct {
	DocumentID     string   `json:"document_id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Category       string   `json:"category"`
	SellerID       string   `json:"seller_id"`
	Price          float64  `json:"price"`
	Rating         float64  `json:"rating"`
	Stock          int      `json:"stock"`
	Tags           []string `json:"tags"`
	ImageURLs      []string `json:"image_urls"`
	IdempotencyKey string   `json:"idempotency_key"`
}

func (h *Handler) IndexDocument(c *gin.Context) {
	var req indexDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	doc := &search.ProductDocument{
		ID:          req.DocumentID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		SellerID:    req.SellerID,
		Price:       req.Price,
		Rating:      req.Rating,
		Stock:       req.Stock,
		Tags:        req.Tags,
		ImageURLs:   req.ImageURLs,
	}

	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}

	task, err := h.indexing.IndexDocument(c.Request.Context(), doc, req.IdempotencyKey)
	if err != nil {
		if err.Error() == "duplicate idempotency key" {
			c.JSON(http.StatusConflict, gin.H{"error": "duplicate idempotency key", "task_id": task.ID})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "document indexed",
		"task_id": task.ID,
		"status":  task.Status,
	})
}

type bulkIndexRequest struct {
	Documents []indexDocumentRequest `json:"documents"`
}

func (h *Handler) BulkIndex(c *gin.Context) {
	var req bulkIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	docs := make([]*search.ProductDocument, len(req.Documents))
	for i, d := range req.Documents {
		doc := &search.ProductDocument{
			ID:          d.DocumentID,
			Title:       d.Title,
			Description: d.Description,
			Category:    d.Category,
			SellerID:    d.SellerID,
			Price:       d.Price,
			Rating:      d.Rating,
			Stock:       d.Stock,
			Tags:        d.Tags,
			ImageURLs:   d.ImageURLs,
		}
		if doc.ID == "" {
			doc.ID = uuid.New().String()
		}
		docs[i] = doc
	}

	result, err := h.indexing.BulkIndex(c.Request.Context(), docs)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListIndexTasks(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	tasks, err := h.indexing.ListTasks(c.Request.Context(), limit, offset)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}
