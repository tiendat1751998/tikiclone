package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/gateway/internal/config"
)

func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	allowMethods := strings.Join(cfg.AllowedMethods, ", ")
	allowHeaders := strings.Join(cfg.AllowedHeaders, ", ")
	exposeHeaders := strings.Join(cfg.ExposedHeaders, ", ")
	maxAge := strconv.Itoa(int(cfg.MaxAge.Seconds()))

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if origin != "" {
			allowedOrigin := ""

			if len(cfg.AllowedOrigins) > 0 && cfg.AllowedOrigins[0] == "*" {
				// [SECURITY] Wildcard origin only allowed without credentials
				if !cfg.AllowCredentials {
					allowedOrigin = "*"
				}
				// When credentials are required, wildcard is NOT allowed
				// Fall through to explicit origin matching below
			}

			// Explicit origin matching (always performed for security)
			if allowedOrigin == "" {
				for _, allowed := range cfg.AllowedOrigins {
					if allowed == origin {
						allowedOrigin = origin
						break
					}
				}
			}

			if allowedOrigin != "" {
				c.Header("Access-Control-Allow-Origin", allowedOrigin)
			}

			c.Header("Access-Control-Allow-Methods", allowMethods)
			c.Header("Access-Control-Allow-Headers", allowHeaders)
			c.Header("Access-Control-Expose-Headers", exposeHeaders)
			c.Header("Access-Control-Max-Age", maxAge)

			// [SECURITY] Only set credentials header when origin is explicitly matched
			// Never set credentials: true with wildcard origin
			if cfg.AllowCredentials && allowedOrigin != "" && allowedOrigin != "*" {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
