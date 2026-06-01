package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/gateway/internal/middleware"
)

type AuthMiddleware struct {
	validator  *JWTValidator
	publicPaths []string
}

func NewAuthMiddleware(validator *JWTValidator) *AuthMiddleware {
	return &AuthMiddleware{
		validator: validator,
		publicPaths: []string{
			"/health",
			"/ready",
			"/metrics",
			"/upstreams",
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/api/v1/auth/refresh",
			"/api/v1/auth/jwks",
			"/api/v1/products",
			"/api/v1/search",
			"/api/v1/categories",
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

		tokenString := ExtractToken(c.Request)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error_code": "MISSING_TOKEN",
				"message":    "authorization header is required",
			})
			return
		}

		claims, err := m.validator.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error_code": "INVALID_TOKEN",
				"message":    err.Error(),
			})
			return
		}

		c.Set(string(middleware.UserIDKey), claims.UserID)
		c.Set(string(middleware.UserRolesKey), claims.Roles)

		for key, value := range extractHeaders(c.Request) {
			c.Request.Header.Set(key, value)
		}

		c.Next()
	}
}

func (m *AuthMiddleware) RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get(string(middleware.UserRolesKey))
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error_code": "FORBIDDEN",
				"message":    "access denied",
			})
			return
		}

		roleList, ok := userRoles.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error_code": "FORBIDDEN",
				"message":    "invalid roles format",
			})
			return
		}

		roleSet := make(map[string]bool, len(roleList))
		for _, r := range roleList {
			roleSet[r] = true
		}

		for _, required := range roles {
			if !roleSet[required] {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error_code": "INSUFFICIENT_ROLE",
					"message":    "insufficient permissions",
				})
				return
			}
		}

		c.Next()
	}
}

func extractHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		headers["X-User-ID"] = userID
	}
	if roles := r.Header.Get("X-User-Roles"); roles != "" {
		headers["X-User-Roles"] = roles
	}
	if deviceID := r.Header.Get("X-Device-ID"); deviceID != "" {
		headers["X-Device-ID"] = deviceID
	}
	headers["X-Correlation-ID"] = r.Header.Get("X-Correlation-ID")
	return headers
}

func (m *AuthMiddleware) AddPublicPath(path string) {
	m.publicPaths = append(m.publicPaths, path)
}
