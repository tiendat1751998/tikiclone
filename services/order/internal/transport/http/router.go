package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/order/internal/transport/http/middleware"
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
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())

	if r.authMw != nil {
		engine.Use(r.authMw)
	}
	{
		engine.POST("/", r.handler.CreateOrder)
		engine.GET("/", r.handler.ListOrders)
		engine.GET("/:id", r.handler.GetOrder)
		engine.GET("/:id/status", r.handler.GetOrderStatus)
		engine.POST("/:id/cancel", r.handler.CancelOrder)
		engine.GET("/:id/history", r.handler.GetOrderHistory)
		engine.GET("/:id/reconciliation", r.handler.GetReconciliationStatus)
	}
}
