package grpc

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func otelUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx, span := otel.Tracer("tiki-clone").Start(ctx, method,
			trace.WithSpanKind(trace.SpanKindClient),
		)
		defer span.End()

		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", method),
		)

		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start)

		observability.GRPCRequestDuration.WithLabelValues(
			cc.Target(),
			method,
			status.Code(err).String(),
		).Observe(duration.Seconds())

		observability.GRPCRequestsTotal.WithLabelValues(
			cc.Target(),
			method,
			status.Code(err).String(),
		).Inc()

		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			observability.LogWithTrace(ctx).Error("grpc call failed",
				zap.String("method", method),
				zap.Error(err),
				zap.Duration("duration", duration),
			)
		}

		return err
	}
}

func retryUnaryClientInterceptor(maxRetries int) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		var lastErr error
		for i := 0; i <= maxRetries; i++ {
			if i > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Duration(100*(1<<i)) * time.Millisecond):
				}
			}
			err := invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}
			lastErr = err
			if st, ok := status.FromError(err); ok {
				code := st.Code()
				// Only retry on transient errors
				if code != codes.Unavailable && code != codes.ResourceExhausted &&
					code != codes.DeadlineExceeded && code != codes.Aborted {
					return err
				}
			}
		}
		return lastErr
	}
}
