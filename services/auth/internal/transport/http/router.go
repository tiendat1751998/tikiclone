package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
)

type Router struct {
	handler *Handler
	health  *health.Checker
	redis   *redis.Client
}

func NewRouter(handler *Handler, healthChecker *health.Checker, redisClient *redis.Client) *Router {
	return &Router{
		handler: handler,
		health:  healthChecker,
		redis:   redisClient,
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Next()
	}
}

func RequestSanitizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		ct := c.GetHeader("Content-Type")
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if ct == "" || len(ct) > 256 {
				c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
					"error_code": "INVALID_CONTENT_TYPE",
					"message":    "content-type header is required",
				})
				return
			}
		}
		c.Next()
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(
		middleware.Recovery(),
		middleware.ErrorHandler(),
		middleware.RequestID(),
		middleware.CORS(),
		middleware.OTelMiddleware("tiki-auth"),
		observability.ObserveHTTPMetrics("tiki-auth"),
		SecurityHeaders(),
		RequestSanitizer(),
	)

	engine.GET("/health", r.health.LivenessHandler())
	engine.GET("/ready", r.health.ReadinessHandler())
	engine.GET("/startup", r.health.LivenessHandler())
	engine.GET("/metrics", observability.MetricsHandler())

	api := engine.Group("/api/v1/auth")
	{
		api.POST("/register", r.handler.Register)
		api.POST("/login",
			middleware.RedisSlidingWindowLimiter(r.redis, 5, 5*time.Minute, "auth_login"),
			r.handler.Login,
		)
		api.POST("/refresh", r.handler.RefreshToken)
		api.POST("/logout", r.handler.Logout)
		api.POST("/logout/all", r.handler.LogoutAll)
		api.GET("/sessions", r.handler.GetSessions)
		api.DELETE("/sessions/:session_id", r.handler.RevokeSession)
		api.GET("/profile", r.handler.GetProfile)
		api.GET("/me", r.handler.GetProfile)
		api.POST("/validate", r.handler.ValidateToken)

		api.POST("/password-reset/request", r.handler.RequestPasswordReset)
		api.POST("/password-reset/reset", r.handler.ResetPassword)
		api.POST("/verify-email", r.handler.VerifyEmail)
		api.POST("/verify-email/send", r.handler.SendVerificationEmail)
	}
}
