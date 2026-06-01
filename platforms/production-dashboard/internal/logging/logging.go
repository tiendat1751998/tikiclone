package logging

import (
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"
)

func Init(serviceName, level string) *zap.Logger {
	return observability.InitLogger(serviceName, level)
}

func WithService(logger *zap.Logger, serviceName string) *zap.Logger {
	return logger.With(zap.String("service_name", serviceName))
}

func WithActor(logger *zap.Logger, actor string) *zap.Logger {
	return logger.With(zap.String("actor", actor))
}

func WithIncidentID(logger *zap.Logger, incidentID string) *zap.Logger {
	return logger.With(zap.String("incident_id", incidentID))
}
