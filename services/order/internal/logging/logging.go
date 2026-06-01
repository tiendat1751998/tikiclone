package logging

import (
	"context"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func Logger() *zap.Logger {
	return observability.GetLogger()
}

func ExtractTraceFields(ctx context.Context) []zap.Field {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return []zap.Field{
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		}
	}
	return nil
}
