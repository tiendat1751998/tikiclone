package http

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	vnpay "github.com/tikiclone/tiki/services/payment/internal/infrastructure/vnpay"
	"github.com/tikiclone/tiki/services/payment/internal/transport/http/middleware"
)

type Router struct {
	handler            *Handler
	authMw             gin.HandlerFunc
	webhookMiddlewares []gin.HandlerFunc
	vnpaySimulator     *vnpay.Simulator
}

func NewRouter(handler *Handler, authMw gin.HandlerFunc, webhookMiddlewares ...gin.HandlerFunc) *Router {
	return &Router{handler: handler, authMw: authMw, webhookMiddlewares: webhookMiddlewares}
}

func NewRouterWithVNPay(handler *Handler, authMw gin.HandlerFunc, vnpaySimulator *vnpay.Simulator, webhookMiddlewares ...gin.HandlerFunc) *Router {
	return &Router{handler: handler, authMw: authMw, vnpaySimulator: vnpaySimulator, webhookMiddlewares: webhookMiddlewares}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Recovery())
	engine.Use(gzip.Gzip(gzip.DefaultCompression))

	// VNPay simulator endpoints (development only - no auth, must be registered before auth middleware)
	if r.vnpaySimulator != nil {
		engine.Any("/simulator/vnpay", func(c *gin.Context) {
			r.vnpaySimulator.ServeHTTP(c.Writer, c.Request)
		})
		engine.Any("/simulator/vnpay/*any", func(c *gin.Context) {
			r.vnpaySimulator.ServeHTTP(c.Writer, c.Request)
		})
	}

	// VNPay callback (no auth - called by VNPay/bank)
	engine.GET("/vnpay/callback", r.handler.VNPayCallback)

	protected := engine.Group("/")
	if r.authMw != nil { protected.Use(r.authMw) }
	protected.POST("/", r.handler.AuthorizePayment)
	protected.GET("/:id", r.handler.GetPayment)
	protected.POST("/:id/capture", r.handler.CapturePayment)
	protected.POST("/:id/refund", r.handler.RefundPayment)

	// VNPay endpoints (for real VNPay integration)
	vnpayRoutes := engine.Group("/vnpay")
	vnpayRoutes.POST("/create", r.handler.VNPayCreatePayment)

	// Webhook endpoint (no auth, uses signature verification)
	webhooks := engine.Group("/webhooks")
	for _, mw := range r.webhookMiddlewares {
		webhooks.Use(mw)
	}
	webhooks.POST("/:provider", r.handler.HandleWebhook)
}
