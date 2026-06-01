package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tikiclone/tiki/services/shipment/internal/config"
)

func JWTAuth(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		bearer := c.GetHeader("Authorization")
		if bearer == "" { c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"}); return }
		parts := strings.SplitN(bearer, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") { c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid format"}); return }
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) { return []byte(cfg.AccessSecret), nil })
		if err != nil || !token.Valid { c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"}); return }
		claims, _ := token.Claims.(jwt.MapClaims)
		userID, _ := claims["sub"].(string)
		if userID == "" { userID, _ = claims["user_id"].(string) }
		role := "buyer"
		if r, ok := claims["role"].(string); ok { role = r }
		c.Set("user_id", userID); c.Set("role", role)
		c.Next()
	}
}
