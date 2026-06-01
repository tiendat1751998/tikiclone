package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/errors"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log := observability.LogWithTrace(c.Request.Context())
				log.Error("panic recovered",
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
				)

			observability.BusinessErrorsTotal.WithLabelValues(
				"recovery",
				"PANIC",
			).Inc()

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error_code": "INTERNAL_ERROR",
					"message":    "An unexpected error occurred",
				})
			}
		}()
		c.Next()
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			appErr := errors.FromError(err)
			log := observability.LogWithTrace(c.Request.Context())

			log.Error("request error",
				zap.String("error_code", string(appErr.Code)),
				zap.String("message", appErr.Message),
				zap.Int("status", appErr.HTTPStatus),
			)

			observability.BusinessErrorsTotal.WithLabelValues(
				c.Request.Method,
				string(appErr.Code),
			).Inc()

			c.AbortWithStatusJSON(appErr.HTTPStatus, gin.H{
				"error_code": appErr.Code,
				"message":    appErr.Message,
				"details":    appErr.Details,
			})
		}
	}
}
