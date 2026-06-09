package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/oms-fulfillment/internal/application"
	"github.com/tikiclone/tiki/services/oms-fulfillment/internal/domain"
	"go.uber.org/zap"
)

type Handler struct {
	svc *application.FulfillmentService
}

func NewHandler(svc *application.FulfillmentService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateFulfillment(c *gin.Context) {
	var req application.CreateFulfillmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, exists := c.Get("user_id")
	if exists {
		if uid, ok := userID.(string); ok {
			req.SellerID = uid
		}
	}
	f, err := h.svc.CreateFulfillment(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, f)
}

func (h *Handler) GetFulfillment(c *gin.Context) {
	id := c.Param("id")
	f, err := h.svc.GetFulfillment(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, f)
}

func (h *Handler) ListFulfillments(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := h.svc.ListFulfillments(c.Request.Context(), status, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	totalPages := (total + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

func (h *Handler) TransitionStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status  string `json:"status" binding:"required"`
		ActorID string `json:"actor_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	target := domain.FulfillmentStatus(req.Status)
	if !target.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}
	actorID := req.ActorID
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok && uid != "" {
			actorID = uid
		}
	}
	f, err := h.svc.TransitionStatus(c.Request.Context(), id, target, actorID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, f)
}

func (h *Handler) GetFulfillmentEvents(c *gin.Context) {
	id := c.Param("id")
	events, err := h.svc.GetFulfillmentEvents(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (h *Handler) ListWarehouses(c *gin.Context) {
	warehouses, err := h.svc.ListWarehouses(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": warehouses})
}

func (h *Handler) GetWarehouse(c *gin.Context) {
	id := c.Param("id")
	w, err := h.svc.GetWarehouse(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *Handler) GetWarehouseInventory(c *gin.Context) {
	id := c.Param("id")
	inv, err := h.svc.GetWarehouseInventory(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": inv})
}

func (h *Handler) CreateReturn(c *gin.Context) {
	var req application.CreateReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, exists := c.Get("user_id")
	if exists {
		if uid, ok := userID.(string); ok {
			req.UserID = uid
		}
	}
	r, err := h.svc.CreateReturn(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetReturn(c *gin.Context) {
	id := c.Param("id")
	r, err := h.svc.GetReturn(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *Handler) ListReturns(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := h.svc.ListReturns(c.Request.Context(), status, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	totalPages := (total + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

func (h *Handler) ApproveReturn(c *gin.Context) {
	id := c.Param("id")
	approvedBy := ""
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			approvedBy = uid
		}
	}
	r, err := h.svc.ApproveReturn(c.Request.Context(), id, approvedBy)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *Handler) RejectReturn(c *gin.Context) {
	id := c.Param("id")
	rejectedBy := ""
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			rejectedBy = uid
		}
	}
	r, err := h.svc.RejectReturn(c.Request.Context(), id, rejectedBy)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrFulfillmentNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "fulfillment not found"})
	case errors.Is(err, domain.ErrFulfillmentNotModifiable):
		c.JSON(http.StatusConflict, gin.H{"error": "fulfillment cannot be modified"})
	case errors.Is(err, domain.ErrInvalidStateTransition):
		c.JSON(http.StatusConflict, gin.H{"error": "invalid state transition"})
	case errors.Is(err, domain.ErrReturnNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "return not found"})
	case errors.Is(err, domain.ErrReturnNotModifiable):
		c.JSON(http.StatusConflict, gin.H{"error": "return cannot be modified"})
	case errors.Is(err, domain.ErrWarehouseNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "warehouse not found"})
	case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, domain.ErrInsufficientPermissions):
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
	case errors.Is(err, domain.ErrConcurrentModification):
		c.JSON(http.StatusConflict, gin.H{"error": "concurrent modification detected, please retry"})
	default:
		zap.L().Error("unexpected error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
