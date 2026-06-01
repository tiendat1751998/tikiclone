package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/auth/internal/application"
	"github.com/tikiclone/tiki/services/auth/internal/domain"
)

type AuthMiddleware struct {
	authService *application.AuthService
	publicPaths  []string
}

func NewAuthMiddleware(authService *application.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		publicPaths: []string{
			"/health", "/ready", "/metrics",
			"/api/v1/auth/register", "/api/v1/auth/login", "/api/v1/auth/refresh",
		},
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, path := range m.publicPaths {
			if c.Request.URL.Path == path || strings.HasPrefix(c.Request.URL.Path, path+"/") {
				c.Next()
				return
			}
		}

		tokenString := extractToken(c)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error_code": "MISSING_TOKEN",
				"message":    "authorization header required",
			})
			return
		}

		claims, err := m.authService.ValidateAccessToken(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error_code": "INVALID_TOKEN",
				"message":    err.Error(),
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("session_id", claims.SessionID)

		c.Request.Header.Set("X-User-ID", claims.UserID)
		c.Request.Header.Set("X-User-Roles", joinRoles(rolesToStrings(claims.Roles)))
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	bearer := c.GetHeader("Authorization")
	if len(bearer) > 7 && bearer[:7] == "Bearer " {
		return bearer[7:]
	}
	return ""
}

func rolesToStrings(roles []domain.Role) []string {
	s := make([]string, len(roles))
	for i, r := range roles {
		s[i] = string(r)
	}
	return s
}

func joinRoles(roles []string) string {
	return strings.Join(roles, ",")
}
