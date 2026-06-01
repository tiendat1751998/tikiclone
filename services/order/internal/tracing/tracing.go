package tracing

import (
	"github.com/tikiclone/tiki/services/order/internal/config"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

func Init(cfg config.OTELConfig) (func(), error) {
	shutdown, err := observability.InitTracer(cfg.ServiceName, cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	Tracer = otel.Tracer(cfg.ServiceName)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return shutdown, nil
}
