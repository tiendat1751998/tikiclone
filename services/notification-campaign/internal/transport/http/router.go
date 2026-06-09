package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/notification-campaign/internal/transport/http/middleware"
)

type Router struct {
	handler *Handler
	authMw  gin.HandlerFunc
}

func NewRouter(handler *Handler, authMw gin.HandlerFunc) *Router {
	return &Router{handler: handler, authMw: authMw}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Recovery())

	campaigns := engine.Group("/api/v1/notification/campaigns")
	if r.authMw != nil {
		campaigns.Use(r.authMw)
	}
	{
		campaigns.GET("", r.handler.ListCampaigns)
		campaigns.POST("", r.handler.CreateCampaign)
		campaigns.GET("/:id", r.handler.GetCampaign)
		campaigns.PUT("/:id", r.handler.UpdateCampaign)
		campaigns.DELETE("/:id", r.handler.DeleteCampaign)
		campaigns.POST("/:id/launch", r.handler.LaunchCampaign)
		campaigns.POST("/:id/pause", r.handler.PauseCampaign)
		campaigns.POST("/:id/recipients", r.handler.AddRecipient)
		campaigns.GET("/:id/recipients", r.handler.ListRecipients)
	}

	templates := engine.Group("/api/v1/notification/templates")
	if r.authMw != nil {
		templates.Use(r.authMw)
	}
	{
		templates.GET("", r.handler.ListTemplates)
		templates.POST("", r.handler.CreateTemplate)
		templates.GET("/:id", r.handler.GetTemplate)
		templates.PUT("/:id", r.handler.UpdateTemplate)
	}

	preferences := engine.Group("/api/v1/notification/preferences")
	if r.authMw != nil {
		preferences.Use(r.authMw)
	}
	{
		preferences.GET("/:user_id", r.handler.GetPreference)
		preferences.PUT("/:user_id", r.handler.UpdatePreference)
	}

	delivery := engine.Group("/api/v1/notification/delivery-logs")
	if r.authMw != nil {
		delivery.Use(r.authMw)
	}
	{
		delivery.GET("", r.handler.ListDeliveryLogs)
		delivery.GET("/:id", r.handler.GetDeliveryLog)
	}
}
