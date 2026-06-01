package transport

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/gateway/internal/discovery"
	"github.com/tikiclone/tiki/services/gateway/internal/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

type GRPCProxy struct {
	discovery *discovery.ServiceDiscovery
	pool      *GRPCConnPool
}

type GRPCConnPool struct {
	mu    sync.RWMutex
	conns map[string]*grpc.ClientConn
}

func NewGRPCConnPool() *GRPCConnPool {
	return &GRPCConnPool{conns: make(map[string]*grpc.ClientConn)}
}

func (p *GRPCConnPool) GetConn(target string) (*grpc.ClientConn, error) {
	p.mu.RLock()
	conn, ok := p.conns[target]
	p.mu.RUnlock()
	if ok {
		return conn, nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if conn, ok := p.conns[target]; ok {
		return conn, nil
	}

	conn, err := grpc.Dial(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithUnaryInterceptor(otelUnaryClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC %s: %w", target, err)
	}

	p.conns[target] = conn
	return conn, nil
}

func (p *GRPCConnPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for target, conn := range p.conns {
		conn.Close()
		delete(p.conns, target)
	}
}

func NewGRPCProxy(svcDiscovery *discovery.ServiceDiscovery) *GRPCProxy {
	return &GRPCProxy{
		discovery: svcDiscovery,
		pool:      NewGRPCConnPool(),
	}
}

type GRPCUpstreamTarget struct {
	ServiceName string
	Port        int
}

func (p *GRPCProxy) Invoke(ctx context.Context, target *GRPCUpstreamTarget, method string, req, resp interface{}, md map[string]string) error {
	ctx, span := otel.Tracer("tiki-gateway").Start(ctx,
		fmt.Sprintf("grpc_upstream.%s", target.ServiceName),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("upstream.service", target.ServiceName),
		attribute.String("upstream.method", method),
	)

	instance, err := p.discovery.GetInstance(target.ServiceName)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "no healthy upstream instances")
		return fmt.Errorf("no healthy instance for %s: %w", target.ServiceName, err)
	}

	addr := fmt.Sprintf("%s:%d", instance.Address, target.Port)
	conn, err := p.pool.GetConn(addr)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "connection failed")
		return err
	}

	mdOut := metadata.New(md)
	if traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String(); traceID != "" {
		mdOut.Set("x-trace-id", traceID)
	}

	propCtx := metadata.NewOutgoingContext(ctx, mdOut)

	if err := conn.Invoke(propCtx, method, req, resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		observability.BusinessErrorsTotal.WithLabelValues("gateway", "GRPC_UPSTREAM_ERROR").Inc()
		return err
	}

	return nil
}

func otelUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start)

		observability.GRPCRequestDuration.WithLabelValues(
			cc.Target(), method, statusCode(err),
		).Observe(duration.Seconds())

		return err
	}
}

func statusCode(err error) string {
	if err == nil {
		return "OK"
	}
	return "ERROR"
}

func ExtractGRPCMetadata(ctx context.Context) map[string]string {
	md := make(map[string]string)
	if incomingMD, ok := metadata.FromIncomingContext(ctx); ok {
		for k, v := range incomingMD {
			if len(v) > 0 {
				md[k] = v[0]
			}
		}
	}
	return md
}

func InjectGatewayMetadata(ctx context.Context, c *gin.Context) context.Context {
	md := metadata.Pairs()

	if userID, exists := c.Get(string(middleware.UserIDKey)); exists {
		md.Set("x-user-id", fmt.Sprintf("%v", userID))
	}
	if roles, exists := c.Get(string(middleware.UserRolesKey)); exists {
		if roleList, ok := roles.([]string); ok {
			for _, r := range roleList {
				md.Append("x-user-roles", r)
			}
		}
	}
	if corrID, exists := c.Get(string(middleware.CorrelationIDKey)); exists {
		md.Set("x-correlation-id", fmt.Sprintf("%v", corrID))
	}
	if reqID, exists := c.Get(string(middleware.RequestIDKey)); exists {
		md.Set("x-request-id", fmt.Sprintf("%v", reqID))
	}
	if deviceInfo, exists := c.Get(string(middleware.DeviceInfoKey)); exists {
		if info, ok := deviceInfo.(map[string]string); ok {
			for k, v := range info {
				md.Set(fmt.Sprintf("x-device-%s", k), v)
			}
		}
	}

	return metadata.NewOutgoingContext(ctx, md)
}
