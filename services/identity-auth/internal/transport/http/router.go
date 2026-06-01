package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/identity-auth/internal/transport/middleware"
)

func SetupRouter(engine *gin.Engine, handler *Handler) {
	engine.Use(
		middleware.Recovery(),
		middleware.ErrorHandler(),
		middleware.RequestID(),
		middleware.SecurityHeaders(),
		middleware.CORS(),
	)

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "alive",
			"service":   "identity-auth",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})
	engine.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "identity-auth",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	api := engine.Group("/api/v1/auth")
	{
		api.POST("/register", handler.Register)
		api.POST("/login", handler.Login)
		api.POST("/refresh", handler.Refresh)
		api.POST("/logout", middleware.AuthRequired(), handler.Logout)
		api.GET("/me", middleware.AuthRequired(), handler.Me)
		api.POST("/validate", handler.ValidateToken)
		api.GET("/users/:userId", middleware.AuthRequired(), handler.GetUserByID)
	}

	engine.GET("/.well-known/jwks.json", func(c *gin.Context) {
		c.String(http.StatusOK, "{}")
	})
	engine.GET("/.well-known/jwk.json", func(c *gin.Context) {
		c.String(http.StatusOK, "{}")
	})
}