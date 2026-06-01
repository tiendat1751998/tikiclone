package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tikiclone/tiki/services/inventory/internal/config"
)

// JWTAuth validates JWT tokens and extracts user context.
// It rejects tokens with invalid signatures, expired tokens, and tokens
// with unexpected signing methods (algorithm confusion protection).
func JWTAuth(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		bearer := c.GetHeader("Authorization")
		if bearer == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(bearer, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		tokenString := parts[1]

		// [SECURITY] Only allow HMAC signing when using shared secret.
		// This prevents algorithm confusion attacks where an attacker
		// could forge tokens using the public key as HMAC secret.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// [SECURITY] Reject non-HMAC algorithms when using shared secret
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.AccessSecret), nil
		},
			jwt.WithValidMethods([]string{"HS256", "HS384", "HS512"}),
			jwt.WithLeeway(30*time.Second), // 30 seconds leeway for clock skew
		)

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		// Extract user_id from standard "sub" claim or custom "user_id" claim
		userID, _ := claims["sub"].(string)
		if userID == "" {
			userID, _ = claims["user_id"].(string)
		}
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token missing user identifier"})
			return
		}

		// Extract role with safe default
		role := "buyer"
		if r, ok := claims["role"].(string); ok && r != "" {
			role = r
		}

		// Store authenticated user context
		c.Set("user_id", userID)
		c.Set("role", role)

		c.Next()
	}
}

// RequireRole creates middleware that checks if the authenticated user has the required role.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "role not found in context"})
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid role format"})
			return
		}

		for _, allowed := range roles {
			if roleStr == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("role '%s' does not have permission", roleStr),
		})
	}
}
