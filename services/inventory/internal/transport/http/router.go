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

	if r.authMw != nil { engine.Use(r.authMw) }
	{
		engine.POST("/reserve", r.handler.ReserveStock)
		engine.POST("/release/:id", r.handler.ReleaseStock)
		engine.GET("/stock/:sku_id", r.handler.GetStock)
	}
}
