package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/notification/internal/health"
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
	engine.Use(middleware.OTelMiddleware("notification"))
	engine.Use(observability.ObserveHTTPMetrics("notification"))

	api := engine.Group("/api/v1")

	notifications := api.Group("/notifications")
	notifications.POST("/send", r.handler.SendNotification)
	notifications.GET("", r.handler.ListNotifications)
	notifications.PUT("/:id/read", r.handler.MarkNotificationRead)
	notifications.DELETE("/:id", r.handler.DeleteNotification)

	devices := api.Group("/devices")
	devices.POST("", r.handler.RegisterDevice)

	pushGroup := api.Group("/push")
	pushGroup.POST("/send", r.handler.SendPush)

	emailGroup := api.Group("/email")
	emailGroup.POST("/send", r.handler.SendEmail)

	smsGroup := api.Group("/sms")
	smsGroup.POST("/send", r.handler.SendSMS)

	prefs := api.Group("/preferences")
	prefs.GET("", r.handler.GetPreferences)
	prefs.PUT("", r.handler.UpdatePreferences)

	templates := api.Group("/templates")
	templates.POST("", r.handler.CreateTemplate)
	templates.GET("", r.handler.ListTemplates)
	templates.GET("/:id", r.handler.GetTemplate)
	templates.PUT("/:id", r.handler.UpdateTemplate)
	templates.GET("/:id/versions", r.handler.ListTemplateVersions)

	engine.GET("/health/live", r.health.LivenessHandler())
	engine.GET("/health/ready", r.health.ReadinessHandler())
	engine.GET("/metrics", observability.MetricsHandler())
}
