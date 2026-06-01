package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/dispute"
	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/ledger"
	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/payment"
	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/payout"
	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/reconciliation"
)

type Handler struct {
	paymentService       *payment.Service
	ledgerService        *ledger.Service
	payoutService        *payout.Service
	reconciliationService *reconciliation.Service
	disputeService       *dispute.Service
}

func NewHandler(ps *payment.Service, ls *ledger.Service, pos *payout.Service, rs *reconciliation.Service, ds *dispute.Service) *Handler {
	return &Handler{
		paymentService:       ps,
		ledgerService:        ls,
		payoutService:        pos,
		reconciliationService: rs,
		disputeService:       ds,
	}
}

func (h *Handler) ProcessPayment(c *gin.Context) {
	var req struct {
		OrderID  string `json:"order_id" binding:"required"`
		UserID   string `json:"user_id" binding:"required"`
		Amount   int64  `json:"amount" binding:"required"`
		Fee      int64  `json:"fee"`
		Currency string `json:"currency" binding:"required"`
		Method   string `json:"method" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.paymentService.Process(c.Request.Context(), req.OrderID, req.UserID, req.Amount, req.Fee, req.Currency, payment.PaymentMethod(req.Method))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *Handler) RefundPayment(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Amount *int64 `json:"amount"`
	}
	c.ShouldBindJSON(&req)

	if req.Amount != nil {
		p, err := h.paymentService.PartialRefund(c.Request.Context(), id, *req.Amount)
		if err != nil {
			handlePaymentError(c, err)
			return
		}
		c.JSON(http.StatusOK, p)
		return
	}

	p, err := h.paymentService.Refund(c.Request.Context(), id)
	if err != nil {
		handlePaymentError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) GetPayment(c *gin.Context) {
	p, err := h.paymentService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handlePaymentError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) ListPayments(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	items, total, err := h.paymentService.List(c.Request.Context(), offset, limit)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

func (h *Handler) GetPaymentsByOrder(c *gin.Context) {
	items, err := h.paymentService.GetByOrder(c.Request.Context(), c.Param("orderId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *Handler) PostLedgerTransaction(c *gin.Context) {
	var req struct {
		DebitAccountID  string `json:"debit_account_id" binding:"required"`
		CreditAccountID string `json:"credit_account_id" binding:"required"`
		Amount          int64  `json:"amount" binding:"required"`
		EntryType       string `json:"entry_type"`
		ReferenceType   string `json:"reference_type"`
		ReferenceID     string `json:"reference_id"`
		Description     string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	txn, err := h.ledgerService.PostTransaction(c.Request.Context(), req.DebitAccountID, req.CreditAccountID, req.Amount, req.EntryType, req.ReferenceType, req.ReferenceID, req.Description)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, txn)
}

func (h *Handler) GetLedgerAccount(c *gin.Context) {
	a, err := h.ledgerService.GetAccount(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *Handler) GetAccountStatement(c *gin.Context) {
	entries, err := h.ledgerService.GetAccountStatement(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": entries})
}

func (h *Handler) GetAccountBalance(c *gin.Context) {
	a, err := h.ledgerService.GetAccountBalance(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"account_id": a.ID, "balance": a.Balance, "currency": a.Currency})
}

func (h *Handler) CreatePayout(c *gin.Context) {
	var req struct {
		SellerID      string `json:"seller_id" binding:"required"`
		Amount        int64  `json:"amount" binding:"required"`
		Fee           int64  `json:"fee"`
		PaymentMethod string `json:"payment_method" binding:"required"`
		PeriodStart   string `json:"period_start" binding:"required"`
		PeriodEnd     string `json:"period_end" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.payoutService.CreatePayout(c.Request.Context(), req.SellerID, req.Amount, req.Fee, payout.PaymentMethod(req.PaymentMethod), req.PeriodStart, req.PeriodEnd)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *Handler) BatchPayout(c *gin.Context) {
	var req struct {
		PayoutIDs []string `json:"payout_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var payouts []*payout.Payout
	for _, id := range req.PayoutIDs {
		p, err := h.payoutService.GetByID(c.Request.Context(), id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "payout not found: " + id})
			return
		}
		payouts = append(payouts, p)
	}
	batch, err := h.payoutService.BatchPayout(c.Request.Context(), payouts)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, batch)
}

func (h *Handler) GetPayout(c *gin.Context) {
	p, err := h.payoutService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) ListPayouts(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	items, total, err := h.payoutService.List(c.Request.Context(), offset, limit)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

func (h *Handler) RunReconciliation(c *gin.Context) {
	var req struct {
		Date     string                       `json:"date" binding:"required"`
		Payments []reconciliation.PaymentRecord `json:"payments"`
		Ledgers  []reconciliation.LedgerRecord  `json:"ledgers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	run, err := h.reconciliationService.RunReconciliation(c.Request.Context(), req.Date, req.Payments, req.Ledgers)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *Handler) ListReconciliationRuns(c *gin.Context) {
	runs, err := h.reconciliationService.ListRuns(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": runs})
}

func (h *Handler) GetUnmatchedItems(c *gin.Context) {
	items, err := h.reconciliationService.GetUnmatched(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *Handler) OpenDispute(c *gin.Context) {
	var req struct {
		TransactionID string `json:"transaction_id" binding:"required"`
		PaymentID     string `json:"payment_id" binding:"required"`
		UserID        string `json:"user_id" binding:"required"`
		Reason        string `json:"reason" binding:"required"`
		Amount        int64  `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	d, err := h.disputeService.OpenDispute(c.Request.Context(), req.TransactionID, req.PaymentID, req.UserID, req.Reason, req.Amount)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *Handler) ResolveDispute(c *gin.Context) {
	var req struct {
		Resolution string `json:"resolution" binding:"required"`
		Notes      string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	d, err := h.disputeService.Resolve(c.Request.Context(), c.Param("id"), dispute.Resolution(req.Resolution), req.Notes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *Handler) ListDisputes(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	items, total, err := h.disputeService.List(c.Request.Context(), offset, limit)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

func handlePaymentError(c *gin.Context, err error) {
	switch err {
	case payment.ErrPaymentNotFound:
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case payment.ErrInvalidAmount, payment.ErrInvalidMethod:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case payment.ErrInvalidStatus:
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": err.Error()})
	case payment.ErrAlreadyRefunded:
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": err.Error()})
	case payment.ErrRefundExceeds:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
