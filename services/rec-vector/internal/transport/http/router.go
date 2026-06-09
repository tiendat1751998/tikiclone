package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/rec-vector/internal/transport/http/middleware"
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

	api := engine.Group("/api/v1/recommendations")
	if r.authMw != nil {
		api.Use(r.authMw)
	}
	{
		api.GET("", r.handler.GetRecommendations)
	}

	collections := engine.Group("/api/v1/collections")
	if r.authMw != nil {
		collections.Use(r.authMw)
	}
	{
		collections.GET("", r.handler.ListCollections)
		collections.POST("", r.handler.CreateCollection)
		collections.GET("/:id", r.handler.GetCollection)
		collections.POST("/:id/metadata", r.handler.AddMetadata)
		collections.GET("/:id/metadata", r.handler.ListMetadata)
	}

	models := engine.Group("/api/v1/models")
	if r.authMw != nil {
		models.Use(r.authMw)
	}
	{
		models.GET("", r.handler.ListModels)
		models.GET("/:id", r.handler.GetModel)
	}

	jobs := engine.Group("/api/v1/jobs")
	if r.authMw != nil {
		jobs.Use(r.authMw)
	}
	{
		jobs.GET("", r.handler.ListJobs)
		jobs.POST("", r.handler.CreateJob)
	}
}
