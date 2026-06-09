package tracing

import (
	"fmt"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/rec-vector/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

func Init(cfg config.OTELConfig) (func(), error) {
	shutdown, err := observability.InitTracer(cfg.ServiceName, cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("init tracer: %w", err)
	}

	Tracer = otel.Tracer(cfg.ServiceName)
	return shutdown, nil
}
