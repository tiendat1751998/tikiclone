package grpc

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func OTelUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx, span := otel.Tracer("catalog-product").Start(ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", info.FullMethod),
		)

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		observability.GRPCRequestDuration.WithLabelValues(
			"catalog-product",
			info.FullMethod,
			status.Code(err).String(),
		).Observe(duration.Seconds())

		observability.GRPCRequestsTotal.WithLabelValues(
			"catalog-product",
			info.FullMethod,
			status.Code(err).String(),
		).Inc()

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			observability.LogWithTrace(ctx).Error("grpc request failed",
				zap.String("method", info.FullMethod),
				zap.Error(err),
				zap.Duration("duration", duration),
			)
		}

		return resp, err
	}
}
