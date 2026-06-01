package logging

import (
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"
)

func Init(serviceName, level string) *zap.Logger {
	return observability.InitLogger(serviceName, level)
}

func WithUserID(logger *zap.Logger, userID string) *zap.Logger {
	return logger.With(zap.String("user_id", userID))
}

func WithAction(logger *zap.Logger, action string) *zap.Logger {
	return logger.With(zap.String("action", action))
}

func WithIP(logger *zap.Logger, ip string) *zap.Logger {
	return logger.With(zap.String("ip", ip))
}
