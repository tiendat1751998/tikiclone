package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/payment/internal/application"
	"github.com/tikiclone/tiki/services/payment/internal/domain"
	"go.uber.org/zap"
)

type Handler struct {
	paymentService *application.PaymentService
}

func NewHandler(paymentService *application.PaymentService) *Handler {
	return &Handler{paymentService: paymentService}
}

// VNPayCreatePayment handles frontend VNPay payment initiation
func (h *Handler) VNPayCreatePayment(c *gin.Context) {
	ctx := c.Request.Context()
	var req struct {
		OrderID   string `json:"order_id" binding:"required"`
		Amount    int64  `json:"amount" binding:"required"`
		ClientIP  string `json:"client_ip" binding:"required"`
		ReturnURL string `json:"return_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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

	txnRef := req.OrderID
	vnpayReq, paymentID, err := h.paymentService.VNPayAuthorizePayment(ctx, req.OrderID, uid, req.Amount, txnRef, req.ClientIP, req.ReturnURL)
	if err != nil {
		handleError(c, err)
		return
	}

	// Build params map for hash generation
	params := map[string]string{
		"vnp_Version":    vnpayReq.Version,
		"vnp_Command":    vnpayReq.Command,
		"vnp_TmnCode":    vnpayReq.TmnCode,
		"vnp_Amount":     fmt.Sprintf("%d", vnpayReq.Amount),
		"vnp_CurrCode":   vnpayReq.CurrCode,
		"vnp_TxnRef":     vnpayReq.TxnRef,
		"vnp_OrderInfo":  vnpayReq.OrderInfo,
		"vnp_ReturnUrl":  vnpayReq.ReturnUrl,
		"vnp_IpAddr":     vnpayReq.IpAddr,
		"vnp_CreateDate": vnpayReq.CreateDate,
		"vnp_Locale":     vnpayReq.Locale,
	}
	if vnpayReq.BankCode != "" {
		params["vnp_BankCode"] = vnpayReq.BankCode
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_id":   paymentID,
		"vnpay_params": vnpayReq,
		"redirect_url": "/simulator/vnpay?" + domain.BuildVNPayQueryString(params),
	})
}

// VNPayCallback handles VNPay return callback after payment processing
func (h *Handler) VNPayCallback(c *gin.Context) {
	ctx := c.Request.Context()
	params := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	if err := h.paymentService.ProcessVNPayCallback(ctx, params); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "order_id": params["vnp_TxnRef"]})
}

func (h *Handler) AuthorizePayment(c *gin.Context) {
	ctx := c.Request.Context()
	var req application.AuthorizePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
	req.UserID = uid
	payment, err := h.paymentService.AuthorizePayment(ctx, &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, payment)
}

func (h *Handler) CapturePayment(c *gin.Context) {
	ctx := c.Request.Context()
	paymentID := c.Param("id")
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
	payment, err := h.paymentService.CapturePayment(ctx, paymentID, uid)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, payment)
}

func (h *Handler) RefundPayment(c *gin.Context) {
	ctx := c.Request.Context()
	paymentID := c.Param("id")
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
	var req struct {
		Reason         string `json:"reason" binding:"required"`
		Amount         int64  `json:"amount" binding:"required"`
		IdempotencyKey string `json:"idempotency_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	refund, err := h.paymentService.RefundPayment(ctx, paymentID, req.Reason, req.IdempotencyKey, req.Amount, uid)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, refund)
}

func (h *Handler) GetPayment(c *gin.Context) {
	ctx := c.Request.Context()
	paymentID := c.Param("id")
	payment, err := h.paymentService.GetPayment(ctx, paymentID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, payment)
}

func (h *Handler) HandleWebhook(c *gin.Context) {
	ctx := c.Request.Context()
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<10)

	pspProvider := c.Param("provider")
	eventType := c.Query("event_type")
	signature := c.GetHeader("X-Webhook-Signature")
	idempotencyKey := c.GetHeader("X-Idempotency-Key")
	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
		return
	}
	if err := h.paymentService.HandleWebhook(ctx, pspProvider, eventType, payload, signature, idempotencyKey); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

func handleError(c *gin.Context, err error) {
	switch err {
	case domain.ErrPaymentNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case domain.ErrDoubleChargeDetected:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case domain.ErrInvalidWebhookSignature:
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case domain.ErrWebhookReplayDetected:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case domain.ErrRefundNotAllowed:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case domain.ErrRefundAmountExceeded:
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	case domain.ErrPaymentAlreadyProcessed:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case domain.ErrInvalidPaymentState:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case domain.ErrUnauthorized, domain.ErrInsufficientPermissions:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case domain.ErrIdempotencyKeyExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case domain.ErrConcurrentModification:
		c.JSON(http.StatusConflict, gin.H{"error": "concurrent modification detected, please retry"})
	case domain.ErrFraudDetected:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		zap.L().Error("unexpected error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
