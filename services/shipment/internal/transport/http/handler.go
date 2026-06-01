package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/shipment/internal/application"
	"github.com/tikiclone/tiki/services/shipment/internal/domain"
	"go.uber.org/zap"
)

type Handler struct {
	shipmentService *application.ShipmentService
}

func NewHandler(svc *application.ShipmentService) *Handler { return &Handler{shipmentService: svc} }

func (h *Handler) CreateShipment(c *gin.Context) {
	var req application.CreateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id not found in context"})
		return
	}
	uid, ok := userID.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id in context"})
		return
	}
	req.UserID = uid
	shipment, err := h.shipmentService.CreateShipment(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, shipment)
}

func (h *Handler) GetShipment(c *gin.Context) {
	shipmentID := c.Param("id")
	shipment, err := h.shipmentService.GetShipment(c.Request.Context(), shipmentID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, shipment)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	shipmentID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists || userID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id not found in context"})
		return
	}
	uid, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is not a valid string"})
		return
	}
	shipment, err := h.shipmentService.UpdateStatus(c.Request.Context(), shipmentID, domain.ShipmentStatus(req.Status), uid, req.Reason)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, shipment)
}

func (h *Handler) GetTracking(c *gin.Context) {
	shipmentID := c.Param("id")
	history, err := h.shipmentService.GetTrackingHistory(c.Request.Context(), shipmentID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"shipment_id": shipmentID, "tracking": history})
}

func (h *Handler) HandleWebhook(c *gin.Context) {
	provider := c.Param("provider")
	eventType := c.Query("event_type")
	signature := c.GetHeader("X-Webhook-Signature")
	idempotencyKey := c.GetHeader("X-Idempotency-Key")
	payload, _ := c.GetRawData()
	if err := h.shipmentService.HandleWebhook(c.Request.Context(), provider, eventType, payload, signature, idempotencyKey); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

func handleError(c *gin.Context, err error) {
	switch err {
	case domain.ErrShipmentNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case domain.ErrShipmentAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case domain.ErrInvalidShipmentState:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case domain.ErrCarrierUnavailable:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	case domain.ErrInvalidWebhookSignature:
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case domain.ErrWebhookReplayDetected:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case domain.ErrTrackingNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case domain.ErrUnauthorized:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case domain.ErrIdempotencyKeyExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case domain.ErrConcurrentModification:
		c.JSON(http.StatusConflict, gin.H{"error": "concurrent modification detected, please retry"})
	default:
		zap.L().Error("unexpected error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
