package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tikiclone/tiki/services/order/internal/config"
)

// JWTAuth validates JWT tokens and extracts user context.
// [SECURITY] Only allows HMAC signing to prevent algorithm confusion attacks.
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

		// [SECURITY] Only allow HMAC signing when using shared secret
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.AccessSecret), nil
		},
			jwt.WithValidMethods([]string{"HS256", "HS384", "HS512"}),
			jwt.WithLeeway(30*time.Second), // [FIX A6] 30 seconds, not 30 days
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

		// [FIX C3] Safe type assertion with ok check
		userID, _ := claims["sub"].(string)
		if userID == "" {
			if uid, ok := claims["user_id"].(string); ok {
				userID = uid
			}
		}
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token missing user identifier"})
			return
		}

		role := "buyer"
		if r, ok := claims["role"].(string); ok && r != "" {
			role = r
		} else if roles, ok := claims["roles"].([]interface{}); ok && len(roles) > 0 {
			if r, ok := roles[0].(string); ok {
				role = r
			}
		}

		// [FIX C3] Safe type assertion for email
		if email, ok := claims["email"].(string); ok {
			c.Set("email", email)
		}

		c.Set("user_id", userID)
		c.Set("role", role)

		c.Next()
	}
}
