package http

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/billing/internal/domain"
	"github.com/tikiclone/tiki/platforms/billing/internal/ledger"
	"github.com/tikiclone/tiki/platforms/billing/internal/settlements"
	"github.com/tikiclone/tiki/platforms/billing/internal/wallets"
)

type Handler struct {
	walletService    *wallets.Service
	ledgerEngine     *ledger.Engine
	settlementService *settlements.Service
}

func NewHandler(ws *wallets.Service, le *ledger.Engine, ss *settlements.Service) *Handler {
	return &Handler{walletService: ws, ledgerEngine: le, settlementService: ss}
}

func (h *Handler) CreateWallet(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		Type     string `json:"type" binding:"required"`
		Currency string `json:"currency" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w, err := h.walletService.CreateWallet(c.Request.Context(), req.UserID, req.Type, req.Currency)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, w)
}

func (h *Handler) Deposit(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		WalletType  string `json:"wallet_type" binding:"required"`
		Amount      int64  `json:"amount" binding:"required"`
		Currency    string `json:"currency" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	txn, err := h.walletService.Deposit(c.Request.Context(), req.UserID, req.WalletType, req.Amount, req.Currency, req.Description)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, txn)
}

func (h *Handler) Withdraw(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		WalletType  string `json:"wallet_type" binding:"required"`
		Amount      int64  `json:"amount" binding:"required"`
		Currency    string `json:"currency" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	txn, err := h.walletService.Withdraw(c.Request.Context(), req.UserID, req.WalletType, req.Amount, req.Currency, req.Description)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, txn)
}

func (h *Handler) Transfer(c *gin.Context) {
	var req struct {
		FromUserID  string `json:"from_user_id" binding:"required"`
		ToUserID    string `json:"to_user_id" binding:"required"`
		Amount      int64  `json:"amount" binding:"required"`
		Currency    string `json:"currency" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	txn, err := h.walletService.Transfer(c.Request.Context(), req.FromUserID, req.ToUserID, req.Amount, req.Currency, req.Description)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, txn)
}

func (h *Handler) GetWalletBalance(c *gin.Context) {
	balance, frozen, pending, err := h.walletService.GetBalance(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": balance, "frozen": frozen, "pending": pending})
}

func (h *Handler) PostLedgerTransaction(c *gin.Context) {
	var req struct {
		Type           string `json:"type" binding:"required"`
		DebitAccountID  string `json:"debit_account_id" binding:"required"`
		CreditAccountID string `json:"credit_account_id" binding:"required"`
		Amount         int64  `json:"amount" binding:"required"`
		Currency       string `json:"currency" binding:"required"`
		Description    string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	txn, err := h.ledgerEngine.PostTransaction(c.Request.Context(), domain.TransactionType(req.Type), req.DebitAccountID, req.CreditAccountID, req.Amount, req.Currency, req.Description)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, txn)
}

func (h *Handler) GetTransaction(c *gin.Context) {
	txn, err := h.ledgerEngine.GetTransaction(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, txn)
}

func (h *Handler) GetLedgerEntries(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	entries, total, err := h.ledgerEngine.GetLedgerEntries(c.Request.Context(), c.Param("account_id"), offset, limit)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": entries, "total": total})
}

func (h *Handler) ReverseTransaction(c *gin.Context) {
	txn, err := h.ledgerEngine.ReverseTransaction(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, txn)
}

var errorStatusMap = map[string]int{
	domain.ErrInsufficientBalance.Error(): http.StatusPaymentRequired,
	domain.ErrAccountNotFound.Error():     http.StatusNotFound,
	domain.ErrWalletNotFound.Error():      http.StatusNotFound,
	domain.ErrDuplicateTxn.Error():        http.StatusConflict,
	domain.ErrInvalidAmount.Error():       http.StatusBadRequest,
	domain.ErrAccountFrozen.Error():       http.StatusForbidden,
	domain.ErrCurrencyMismatch.Error():    http.StatusBadRequest,
	domain.ErrLedgerImbalance.Error():     http.StatusInternalServerError,
}

func handleError(c *gin.Context, err error) {
	if code, ok := errorStatusMap[err.Error()]; ok {
		c.AbortWithStatusJSON(code, gin.H{"error_code": err.Error(), "message": err.Error()})
		return
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error_code": "INTERNAL_ERROR", "message": err.Error()})
}
