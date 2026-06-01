package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/checkout/internal/application"
	"github.com/tikiclone/tiki/services/checkout/internal/domain"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type Handler struct {
	service *application.CheckoutService
}

func NewHandler(service *application.CheckoutService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) InitiateCheckout(c *gin.Context) {
	ctx, _ := otel.Tracer("tiki-checkout").Start(c.Request.Context(), "http.initiate_checkout")

	var req application.InitiateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	// Extract user_id from JWT context, never trust request body for identity
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "UNAUTHORIZED", "message": "unauthorized"})
		return
	}
	uid, ok := userID.(string)
	if !ok || uid == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "UNAUTHORIZED", "message": "unauthorized"})
		return
	}
	req.UserID = uid

	checkout, err := h.service.InitiateCheckout(ctx, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, checkout)
}

func (h *Handler) GetCheckoutStatus(c *gin.Context) {
	ctx, _ := otel.Tracer("tiki-checkout").Start(c.Request.Context(), "http.get_status")

	checkoutID := c.Param("checkout_id")
	checkout, err := h.service.GetCheckoutStatus(ctx, checkoutID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, checkout)
}

func (h *Handler) RetryCheckout(c *gin.Context) {
	ctx, _ := otel.Tracer("tiki-checkout").Start(c.Request.Context(), "http.retry_checkout")

	checkoutID := c.Param("checkout_id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, ok := userID.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := h.service.RetryCheckout(ctx, checkoutID, uid); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "checkout retry initiated"})
}

var errorStatusMap = map[error]int{
	domain.ErrCheckoutNotFound:    http.StatusNotFound,
	domain.ErrCheckoutExpired:     http.StatusGone,
	domain.ErrCheckoutCompleted:   http.StatusConflict,
	domain.ErrIdempotencyConflict: http.StatusConflict,
	domain.ErrValidationFailed:    http.StatusUnprocessableEntity,
	domain.ErrPricingChanged:      http.StatusConflict,
	domain.ErrReservationFailed:   http.StatusServiceUnavailable,
	domain.ErrRollbackFailed:      http.StatusInternalServerError,
	domain.ErrMaxRetriesExceeded:  http.StatusTooManyRequests,
	domain.ErrUnauthorized:        http.StatusForbidden,
}

func handleError(c *gin.Context, err error) {
	for domainErr, status := range errorStatusMap {
		if errors.Is(err, domainErr) {
			c.AbortWithStatusJSON(status, gin.H{"error_code": domainErr.Error(), "message": err.Error()})
			return
		}
	}
	observability.LogWithTrace(c.Request.Context()).Error("unhandled error", zap.Error(err))
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error_code": "INTERNAL_ERROR", "message": "An unexpected error occurred"})
}
