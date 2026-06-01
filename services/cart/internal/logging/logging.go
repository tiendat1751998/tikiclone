package logging

import (
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"
)

func Init(serviceName, level string) *zap.Logger {
	return observability.InitLogger(serviceName, level)
}

func WithCartID(logger *zap.Logger, cartID string) *zap.Logger {
	return logger.With(zap.String("cart_id", cartID))
}

func WithUserID(logger *zap.Logger, userID string) *zap.Logger {
	return logger.With(zap.String("user_id", userID))
}
