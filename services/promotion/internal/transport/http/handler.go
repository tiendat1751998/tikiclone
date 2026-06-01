package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/promotion/internal/application"
	"github.com/tikiclone/tiki/services/promotion/internal/domain"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type Handler struct {
	service *application.PromotionService
}

func NewHandler(service *application.PromotionService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ValidateVoucher(c *gin.Context) {
	ctx, _ := otel.Tracer("tiki-promotion").Start(c.Request.Context(), "http.validate_voucher")

	var req struct {
		Code          string `json:"code" binding:"required"`
		UserID        string `json:"user_id" binding:"required"`
		Subtotal      int64  `json:"subtotal" binding:"required,min=0"`
		ShopID        string `json:"shop_id"`
		CategoryID    string `json:"category_id"`
		SKU           string `json:"sku"`
		Region        string `json:"region"`
		PaymentMethod string `json:"payment_method"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	voucher, err := h.service.ValidateVoucher(ctx, req.Code, req.UserID, req.Subtotal, req.ShopID, req.CategoryID, req.SKU, req.Region, req.PaymentMethod)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, voucher)
}

func (h *Handler) RedeemVoucher(c *gin.Context) {
	ctx, _ := otel.Tracer("tiki-promotion").Start(c.Request.Context(), "http.redeem_voucher")

	var req struct {
		Code           string `json:"code" binding:"required"`
		UserID         string `json:"user_id" binding:"required"`
		OrderID        string `json:"order_id" binding:"required"`
		IdempotencyKey string `json:"idempotency_key"`
		Subtotal       int64  `json:"subtotal" binding:"required,min=0"`
		ShopID         string `json:"shop_id"`
		CategoryID     string `json:"category_id"`
		SKU            string `json:"sku"`
		Region         string `json:"region"`
		PaymentMethod  string `json:"payment_method"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	result, err := h.service.RedeemVoucher(ctx, req.Code, req.UserID, req.OrderID, req.IdempotencyKey, req.Subtotal, req.ShopID, req.CategoryID, req.SKU, req.Region, req.PaymentMethod)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) EvaluatePromotions(c *gin.Context) {
	ctx, _ := otel.Tracer("tiki-promotion").Start(c.Request.Context(), "http.evaluate_promotions")

	var req struct {
		UserID        string `json:"user_id" binding:"required"`
		Subtotal      int64  `json:"subtotal" binding:"required,min=0"`
		ShopID        string `json:"shop_id"`
		CategoryID    string `json:"category_id"`
		SKU           string `json:"sku"`
		Region        string `json:"region"`
		PaymentMethod string `json:"payment_method"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	results, err := h.service.EvaluatePromotions(ctx, req.UserID, req.Subtotal, req.ShopID, req.CategoryID, req.SKU, req.Region, req.PaymentMethod)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"promotions": results})
}

func (h *Handler) GetActiveCampaigns(c *gin.Context) {
	ctx, _ := otel.Tracer("tiki-promotion").Start(c.Request.Context(), "http.get_campaigns")
	campaigns, err := h.service.GetActiveCampaigns(ctx)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"campaigns": campaigns})
}

func (h *Handler) CreateVoucher(c *gin.Context) {
	ctx, _ := otel.Tracer("tiki-promotion").Start(c.Request.Context(), "http.create_voucher")
	var voucher domain.Voucher
	if err := c.ShouldBindJSON(&voucher); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	if err := h.service.CreateVoucher(ctx, &voucher); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, voucher)
}

var errorStatusMap = map[error]int{
	domain.ErrVoucherInvalid:       http.StatusBadRequest,
	domain.ErrVoucherExpired:       http.StatusGone,
	domain.ErrVoucherExhausted:     http.StatusGone,
	domain.ErrVoucherMinSpend:      http.StatusUnprocessableEntity,
	domain.ErrVoucherScope:         http.StatusUnprocessableEntity,
	domain.ErrVoucherRegion:        http.StatusUnprocessableEntity,
	domain.ErrVoucherPaymentMethod: http.StatusUnprocessableEntity,
	domain.ErrVoucherUserLimit:     http.StatusTooManyRequests,
	domain.ErrDuplicateRedemption:  http.StatusConflict,
	domain.ErrCampaignNotFound:     http.StatusNotFound,
	domain.ErrPromotionNotFound:    http.StatusNotFound,
	domain.ErrStackingConflict:     http.StatusConflict,
}

func handleError(c *gin.Context, err error) {
	// Check exact match first
	for domainErr, status := range errorStatusMap {
		if err == domainErr {
			c.AbortWithStatusJSON(status, gin.H{"error_code": domainErr.Error(), "message": err.Error()})
			return
		}
	}
	// Check for wrapped errors using errors.Is
	for domainErr, status := range errorStatusMap {
		if errors.Is(err, domainErr) {
			c.AbortWithStatusJSON(status, gin.H{"error_code": domainErr.Error(), "message": err.Error()})
			return
		}
	}
	observability.LogWithTrace(c.Request.Context()).Error("unhandled error", zap.Error(err))
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error_code": "INTERNAL_ERROR", "message": "An unexpected error occurred"})
}
