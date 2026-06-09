package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/payment-ledger/internal/application"
	"github.com/tikiclone/tiki/services/payment-ledger/internal/domain"
	"go.uber.org/zap"
)

type Handler struct {
	svc *application.LedgerService
}

func NewHandler(svc *application.LedgerService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateAccount(c *gin.Context) {
	var req application.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a, err := h.svc.CreateAccount(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetAccount(c *gin.Context) {
	a, err := h.svc.GetAccount(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *Handler) ListAccounts(c *gin.Context) {
	items, err := h.svc.ListAccounts(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) PostEntry(c *gin.Context) {
	var req application.PostEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	e, err := h.svc.PostEntry(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, e)
}

func (h *Handler) GetEntry(c *gin.Context) {
	e, err := h.svc.GetEntry(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, e)
}

func (h *Handler) ListEntries(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	refType := c.Query("reference_type")
	refID := c.Query("reference_id")
	items, err := h.svc.ListEntries(c.Request.Context(), refType, refID, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) CreateWallet(c *gin.Context) {
	var req application.CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w, err := h.svc.CreateWallet(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, w)
}

func (h *Handler) GetWallet(c *gin.Context) {
	w, err := h.svc.GetWallet(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *Handler) GetWalletByUserType(c *gin.Context) {
	userID := c.Query("user_id")
	walletType := domain.WalletType(c.Query("wallet_type"))
	if userID == "" || walletType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and wallet_type are required"})
		return
	}
	w, err := h.svc.GetWalletByUserType(c.Request.Context(), userID, walletType)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *Handler) ListWallets(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	items, err := h.svc.ListWallets(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) CreateBatch(c *gin.Context) {
	var req application.CreateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b, err := h.svc.CreateBatch(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, b)
}

func (h *Handler) GetBatch(c *gin.Context) {
	b, err := h.svc.GetBatch(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *Handler) ListBatches(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	sellerID := c.Query("seller_id")
	items, err := h.svc.ListBatches(c.Request.Context(), sellerID, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) CreateReconciliation(c *gin.Context) {
	var req application.CreateReconciliationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r, err := h.svc.CreateReconciliation(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, r)
}

func (h *Handler) ListReconciliations(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	accountID := c.Query("account_id")
	items, err := h.svc.ListReconciliations(c.Request.Context(), accountID, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound),
		errors.Is(err, domain.ErrAccountNotFound),
		errors.Is(err, domain.ErrWalletNotFound),
		errors.Is(err, domain.ErrEntryNotFound),
		errors.Is(err, domain.ErrBatchNotFound),
		errors.Is(err, domain.ErrReconcileNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, domain.ErrInsufficientBalance),
		errors.Is(err, domain.ErrWalletFrozen),
		errors.Is(err, domain.ErrWalletClosed):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidAmount),
		errors.Is(err, domain.ErrDebitCreditMismatch),
		errors.Is(err, domain.ErrEntryAlreadyPosted),
		errors.Is(err, domain.ErrEntryAlreadyReversed):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrUnauthorized):
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
	default:
		zap.L().Error("unexpected error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
