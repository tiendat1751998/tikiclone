package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/aiml/internal/health"
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
	engine.Use(middleware.OTelMiddleware("tiki-aiml"))
	engine.Use(observability.ObserveHTTPMetrics("tiki-aiml"))

	engine.GET("/health/live", r.health.LivenessHandler())
	engine.GET("/health/ready", r.health.ReadinessHandler())
	engine.GET("/metrics", observability.MetricsHandler())

	v1 := engine.Group("/api/v1")
	{
		features := v1.Group("/features")
		{
			features.POST("", r.handler.RegisterFeature)
			features.GET("", r.handler.ListFeatures)
			features.POST("/values", r.handler.SetFeatureValue)
			features.POST("/batch-get", r.handler.BatchGetFeatureValues)
		}

		models := v1.Group("/models")
		{
			models.POST("", r.handler.RegisterModel)
			models.GET("", r.handler.ListModels)
			models.POST("/:id/promote", r.handler.PromoteModel)
		}

		training := v1.Group("/training")
		{
			training.POST("/jobs", r.handler.CreateTrainingJob)
			training.GET("/jobs", r.handler.ListTrainingJobs)
			training.GET("/jobs/:id", r.handler.GetTrainingJob)
		}

		inference := v1.Group("/inference")
		{
			inference.POST("/predict", r.handler.Predict)
			inference.POST("/batch-predict", r.handler.BatchPredict)
		}

		embeddings := v1.Group("/embeddings")
		{
			embeddings.POST("/generate", r.handler.GenerateEmbedding)
			embeddings.POST("/similar", r.handler.FindSimilar)
		}

		experiments := v1.Group("/experiments")
		{
			experiments.POST("", r.handler.CreateExperiment)
			experiments.POST("/:id/assign", r.handler.AssignVariant)
			experiments.POST("/:id/record", r.handler.RecordExperimentResult)
			experiments.GET("/:id/results", r.handler.GetExperimentResults)
		}
	}
}
