package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/health"
)

type Router struct {
	handler *Handler
	health  *health.Checker
}

func NewRouter(h *Handler, hc *health.Checker) *Router {
	return &Router{handler: h, health: hc}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CORS())
	engine.Use(middleware.RequestID())
	engine.Use(middleware.OTelMiddleware("tiki-rec-vector"))
	engine.Use(observability.ObserveHTTPMetrics("tiki-rec-vector"))

	engine.GET("/health/live", r.health.LivenessHandler())
	engine.GET("/health/ready", r.health.ReadinessHandler())
	engine.GET("/metrics", observability.MetricsHandler())

	v1 := engine.Group("/api/v1")
	{
		vectors := v1.Group("/vectors")
		{
			vectors.POST("", r.handler.InsertVector)
			vectors.POST("/batch", r.handler.BatchInsertVectors)
			vectors.POST("/search", r.handler.SearchVectors)
			vectors.DELETE("/:id", r.handler.DeleteVector)
		}

		users := v1.Group("/users")
		{
			users.POST("/embeddings", r.handler.GenerateUserEmbedding)
			users.GET("/:id/embeddings", r.handler.GetUserEmbedding)
			users.POST("/:id/embeddings", r.handler.UpdateUserEmbedding)
		}

		items := v1.Group("/items")
		{
			items.POST("/embeddings", r.handler.GenerateItemEmbedding)
			items.GET("/:id/embeddings", r.handler.GetItemEmbedding)
		}

		similarity := v1.Group("/similarity")
		{
			similarity.POST("/search", r.handler.SimilaritySearch)
			similarity.POST("/hybrid", r.handler.HybridSearch)
		}

		collab := v1.Group("/collaborative")
		{
			collab.POST("/interact", r.handler.RecordInteraction)
			collab.POST("/recommend", r.handler.CollaborativeRecommend)
		}

		realtime := v1.Group("/realtime")
		{
			realtime.POST("/track", r.handler.TrackEvent)
			realtime.POST("/recommend", r.handler.RealtimeRecommend)
		}
	}
}
