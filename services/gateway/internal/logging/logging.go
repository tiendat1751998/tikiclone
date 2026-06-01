package logging

import (
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/gateway/internal/config"
	"go.uber.org/zap"
)

func Init(cfg config.Config) *zap.Logger {
	return observability.InitLogger(cfg.AppName, cfg.LogLevel)
}

func WithRequestID(logger *zap.Logger, requestID string) *zap.Logger {
	return logger.With(zap.String("request_id", requestID))
}

func WithUserID(logger *zap.Logger, userID string) *zap.Logger {
	return logger.With(zap.String("user_id", userID))
}

func WithService(logger *zap.Logger, service string) *zap.Logger {
	return logger.With(zap.String("upstream_service", service))
}

func WithError(logger *zap.Logger, err error) *zap.Logger {
	return logger.With(zap.Error(err))
}
