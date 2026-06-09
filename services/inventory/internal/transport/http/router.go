package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/inventory/internal/transport/http/middleware"
)

type Router struct {
	handler *Handler
	authMw  gin.HandlerFunc
}

func NewRouter(handler *Handler, authMw gin.HandlerFunc) *Router { return &Router{handler: handler, authMw: authMw} }

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Recovery())

	protected := engine.Group("/")
	if r.authMw != nil { protected.Use(r.authMw) }
	protected.POST("/reserve", r.handler.ReserveStock)
	protected.POST("/release/:id", r.handler.ReleaseStock)
	protected.GET("/stock/:sku_id", r.handler.GetStock)
}
