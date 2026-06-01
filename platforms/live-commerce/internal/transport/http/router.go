package http

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/health"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/websocket"
)

type Router struct {
	handler *Handler
	health  *health.Checker
	hub     *websocket.Hub
}

func NewRouter(h *Handler, hc *health.Checker, hub *websocket.Hub) *Router {
	return &Router{handler: h, health: hc, hub: hub}
}

func (r *Router) Setup(e *gin.Engine) {
	e.Use(middleware.Recovery())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORS())
	e.Use(middleware.OTelMiddleware("tiki-live-commerce"))
	e.Use(observability.ObserveHTTPMetrics("tiki-live-commerce"))

	e.GET("/healthz", r.health.LivenessHandler())
	e.GET("/readyz", r.health.ReadinessHandler())
	e.GET("/metrics", observability.MetricsHandler())

	e.GET("/ws/:room_id", func(c *gin.Context) {
		userID := c.Query("user_id")
		username := c.Query("username")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user_id required"})
			return
		}
		r.hub.HandleWS(c.Writer, c.Request, c.Param("room_id"), userID, username)
	})

	v1 := e.Group("/api/v1")
	{
		v1.POST("/livestreams", r.handler.CreateLivestream)
		v1.GET("/livestreams", r.handler.ListActiveLivestreams)
		v1.GET("/livestreams/trending", r.handler.GetTrendingLivestreams)
		v1.GET("/livestreams/:id", r.handler.GetLivestream)
		v1.POST("/livestreams/:id/start", r.handler.StartLivestream)
		v1.POST("/livestreams/:id/end", r.handler.EndLivestream)

		v1.POST("/livestreams/:id/pin", r.handler.PinProduct)
		v1.DELETE("/livestreams/:id/pin/:product_id", r.handler.UnpinProduct)
		v1.GET("/livestreams/:id/pins", r.handler.GetPinnedProducts)

		v1.GET("/livestreams/:id/viewers", r.handler.GetViewerCount)
		v1.GET("/livestreams/:id/reactions", r.handler.GetReactionSummary)
		v1.GET("/livestreams/:id/gifts/leaderboard", r.handler.GetGiftLeaderboard)
		v1.GET("/livestreams/:id/chat", r.handler.GetChatHistory)

		v1.POST("/rooms/:room_id/chat", r.handler.SendMessage)
		v1.POST("/rooms/:room_id/reactions", r.handler.SendReaction)
		v1.POST("/rooms/:room_id/gifts", r.handler.SendGift)

		v1.POST("/moderation/action", r.handler.ModerateAction)

		seller := v1.Group("/sellers/:seller_id/livestreams")
		{
			seller.GET("", r.handler.ListSellerLivestreams)
		}
	}
}
