package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/billing/internal/health"
)

type Router struct {
	handler *Handler
	health  *health.Checker
}

func NewRouter(h *Handler, hc *health.Checker) *Router {
	return &Router{handler: h, health: hc}
}

func (r *Router) Setup(e *gin.Engine) {
	e.Use(middleware.Recovery())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORS())
	e.Use(middleware.OTelMiddleware("tiki-billing"))
	e.Use(observability.ObserveHTTPMetrics("tiki-billing"))

	e.GET("/healthz", r.health.LivenessHandler())
	e.GET("/readyz", r.health.ReadinessHandler())
	e.GET("/metrics", observability.MetricsHandler())

	v1 := e.Group("/api/v1")
	{
		v1.POST("/wallets", r.handler.CreateWallet)
		v1.GET("/wallets/:id/balance", r.handler.GetWalletBalance)
		v1.POST("/wallets/deposit", r.handler.Deposit)
		v1.POST("/wallets/withdraw", r.handler.Withdraw)
		v1.POST("/wallets/transfer", r.handler.Transfer)

		v1.POST("/ledger/transactions", r.handler.PostLedgerTransaction)
		v1.GET("/ledger/transactions/:id", r.handler.GetTransaction)
		v1.POST("/ledger/transactions/:id/reverse", r.handler.ReverseTransaction)
		v1.GET("/ledger/accounts/:account_id/entries", r.handler.GetLedgerEntries)
	}
}
