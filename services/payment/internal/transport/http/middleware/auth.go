package middleware

import (
	"github.com/gin-gonic/gin"
	authpkg "github.com/tikiclone/tiki/packages/go-shared/pkg/auth"
	"github.com/tikiclone/tiki/services/payment/internal/config"
)

func JWTAuth(cfg config.JWTConfig) gin.HandlerFunc {
	return authpkg.GinJWTAuth(cfg.AccessSecret)
}
