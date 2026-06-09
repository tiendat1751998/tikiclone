package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/payment-ledger/internal/transport/http/middleware"
)

type Router struct {
	handler *Handler
	authMw  gin.HandlerFunc
}

func NewRouter(handler *Handler, authMw gin.HandlerFunc) *Router {
	return &Router{
		handler: handler,
		authMw:  authMw,
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Recovery())

	accounts := engine.Group("/api/v1/ledger/accounts")
	if r.authMw != nil {
		accounts.Use(r.authMw)
	}
	{
		accounts.GET("", r.handler.ListAccounts)
		accounts.POST("", r.handler.CreateAccount)
		accounts.GET("/:id", r.handler.GetAccount)
	}

	entries := engine.Group("/api/v1/ledger/entries")
	if r.authMw != nil {
		entries.Use(r.authMw)
	}
	{
		entries.GET("", r.handler.ListEntries)
		entries.POST("", r.handler.PostEntry)
		entries.GET("/:id", r.handler.GetEntry)
	}

	wallets := engine.Group("/api/v1/ledger/wallets")
	if r.authMw != nil {
		wallets.Use(r.authMw)
	}
	{
		wallets.GET("", r.handler.ListWallets)
		wallets.POST("", r.handler.CreateWallet)
		wallets.GET("/by-user", r.handler.GetWalletByUserType)
		wallets.GET("/:id", r.handler.GetWallet)
	}

	walletTxs := engine.Group("/api/v1/ledger/wallets/:id/transactions")
	if r.authMw != nil {
		walletTxs.Use(r.authMw)
	}
	{
		walletTxs.GET("", r.handler.ListWalletTransactions)
	}

	batches := engine.Group("/api/v1/ledger/settlements")
	if r.authMw != nil {
		batches.Use(r.authMw)
	}
	{
		batches.GET("", r.handler.ListBatches)
		batches.POST("", r.handler.CreateBatch)
		batches.GET("/:id", r.handler.GetBatch)
	}

	reconcile := engine.Group("/api/v1/ledger/reconciliations")
	if r.authMw != nil {
		reconcile.Use(r.authMw)
	}
	{
		reconcile.GET("", r.handler.ListReconciliations)
		reconcile.POST("", r.handler.CreateReconciliation)
	}
}

func (h *Handler) ListWalletTransactions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	walletID := c.Param("id")
	items, err := h.svc.ListWalletTransactions(c.Request.Context(), walletID, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
