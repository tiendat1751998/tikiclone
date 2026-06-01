package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	sharedRedis "github.com/tikiclone/tiki/packages/go-shared/pkg/redis"
	"github.com/tikiclone/tiki/services/gateway/internal/auth"
	"github.com/tikiclone/tiki/services/gateway/internal/config"
	"github.com/tikiclone/tiki/services/gateway/internal/discovery"
	"github.com/tikiclone/tiki/services/gateway/internal/ratelimit"
	"github.com/tikiclone/tiki/services/gateway/internal/resilience"
	"github.com/tikiclone/tiki/services/gateway/internal/routing"
	"github.com/tikiclone/tiki/services/gateway/internal/tracing"
	"github.com/tikiclone/tiki/services/gateway/internal/transport"
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var version = "1.0.0"

func init() {
	// Tune GC for low-latency: more frequent GCs, less heap growth
	if gogc := os.Getenv("GOGC"); gogc == "" {
		os.Setenv("GOGC", "50")
	}
}

func main() {
	cfg := config.Load()

	logger := observability.InitLogger(cfg.AppName, cfg.LogLevel)

	// Auto-tune GOMAXPROCS for container environments
	if _, err := automaxprocs.Set(); err != nil {
		logger.Warn("failed to set automaxprocs", zap.Error(err))
	}

	shutdownTracer, err := tracing.Init(cfg.OpenTelemetry)
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer shutdownTracer()
	defer observability.Sync()

	redisClient, err := sharedRedis.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Warn("redis not available, continuing without redis", zap.Error(err))
		redisClient = nil
	}

	jwtValidator := auth.NewJWTValidator(cfg.Auth, redisClient)
	authMiddleware := auth.NewAuthMiddleware(jwtValidator)

	rateLimiter := ratelimit.NewRateLimiter(redisClient, cfg.RateLimit)

	svcDiscovery := discovery.NewServiceDiscovery()
	registerUpstreams(cfg, svcDiscovery)

	proxy := transport.NewProxy(
		svcDiscovery,
		cfg.Upstreams.MaxIdleConns,
		cfg.Upstreams.IdleConnTimeout,
	)

	registerProxyOptions(cfg, proxy)

	grpcProxy := transport.NewGRPCProxy(svcDiscovery)

	healthChecker := health.NewChecker(cfg.AppName, version)
	registerHealthChecks(healthChecker, redisClient)

	gin.SetMode(getGinMode(cfg.AppEnv))
	engine := gin.New()

	router := routing.NewRouter(cfg, proxy, grpcProxy, rateLimiter, authMiddleware, svcDiscovery, healthChecker, redisClient)
	router.Setup(engine)

	httpServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:        engine,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(otelUnaryServerInterceptor()),
	)
	grpc_health_v1.RegisterHealthServer(grpcServer, &gatewayHealthServer{})
	reflection.Register(grpcServer)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("failed to listen gRPC", zap.Int("port", cfg.GRPCPort), zap.Error(err))
	}

	go func() {
		logger.Info("starting gateway",
			zap.Int("http_port", cfg.HTTPPort),
			zap.Int("grpc_port", cfg.GRPCPort),
			zap.String("env", cfg.AppEnv),
			zap.String("version", version),
		)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("gateway http server failed", zap.Error(err))
		}
	}()

	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Fatal("gateway grpc server failed", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("shutting down gateway...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	grpcServer.GracefulStop()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("gateway http shutdown error", zap.Error(err))
	}

	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			logger.Error("redis close error", zap.Error(err))
		}
	}

	logger.Info("gateway stopped")
}

func registerUpstreams(cfg *config.Config, svcDiscovery *discovery.ServiceDiscovery) {
	upstreams := map[string][]*discovery.ServiceInstance{
		"auth": {
			{ID: "auth-1", Name: "auth", Address: extractHost(cfg.Upstreams.AuthService), Port: extractPort(cfg.Upstreams.AuthService), Weight: 10},
		},
		"catalog": {
			{ID: "catalog-1", Name: "catalog", Address: extractHost(cfg.Upstreams.CatalogService), Port: extractPort(cfg.Upstreams.CatalogService), Weight: 10},
			{ID: "catalog-2", Name: "catalog", Address: extractHost(cfg.Upstreams.CatalogServiceReplica2), Port: extractPort(cfg.Upstreams.CatalogServiceReplica2), Weight: 10},
			{ID: "catalog-3", Name: "catalog", Address: extractHost(cfg.Upstreams.CatalogServiceReplica3), Port: extractPort(cfg.Upstreams.CatalogServiceReplica3), Weight: 10},
		},
		"cart": {
			{ID: "cart-1", Name: "cart", Address: extractHost(cfg.Upstreams.CartService), Port: extractPort(cfg.Upstreams.CartService), Weight: 10},
		},
		"order": {
			{ID: "order-1", Name: "order", Address: extractHost(cfg.Upstreams.OrderService), Port: extractPort(cfg.Upstreams.OrderService), Weight: 10},
		},
		"inventory": {
			{ID: "inventory-1", Name: "inventory", Address: extractHost(cfg.Upstreams.InventoryService), Port: extractPort(cfg.Upstreams.InventoryService), Weight: 10},
		},
		"payment": {
			{ID: "payment-1", Name: "payment", Address: extractHost(cfg.Upstreams.PaymentService), Port: extractPort(cfg.Upstreams.PaymentService), Weight: 10},
		},
		"search": {
			{ID: "search-1", Name: "search", Address: extractHost(cfg.Upstreams.SearchService), Port: extractPort(cfg.Upstreams.SearchService), Weight: 10},
		},
		"recommendation": {
			{ID: "rec-1", Name: "recommendation", Address: extractHost(cfg.Upstreams.RecommendationService), Port: extractPort(cfg.Upstreams.RecommendationService), Weight: 5},
		},
		"delivery": {
			{ID: "delivery-1", Name: "delivery", Address: extractHost(cfg.Upstreams.DeliveryService), Port: extractPort(cfg.Upstreams.DeliveryService), Weight: 10},
		},
	}

	for name, instances := range upstreams {
		svcDiscovery.RegisterStatic(name, instances)
	}
}

func registerProxyOptions(cfg *config.Config, proxy *transport.Proxy) {
	services := []string{"auth", "catalog", "cart", "order", "inventory", "payment", "search", "recommendation", "delivery"}
	opts := make([]transport.ProxyOption, 0, len(services))

	timeouts := map[string]time.Duration{
		"auth":           5 * time.Second,
		"catalog":        5 * time.Second,
		"cart":           3 * time.Second,
		"order":          10 * time.Second,
		"inventory":      3 * time.Second,
		"payment":        10 * time.Second,
		"search":         5 * time.Second,
		"recommendation": 5 * time.Second,
		"delivery":       5 * time.Second,
	}

	for _, svc := range services {
		opt := transport.ProxyOption{
			ServiceName: svc,
			Timeout:     timeouts[svc],
			RetryConfig: resilience.RetryConfig{
				MaxAttempts:     cfg.Upstreams.MaxRetries + 1,
				InitialInterval: 50 * time.Millisecond,
				MaxInterval:     5 * time.Second,
				Multiplier:      2.0,
				JitterFactor:    0.1,
				RetryableErrors: resilience.DefaultRetryableCheck,
			},
		}

		if cfg.Upstreams.CircuitBreaker.Enabled {
			opt.CircuitBreaker = resilience.CircuitBreakerOptions{
				MaxRequests:  cfg.Upstreams.CircuitBreaker.MaxRequests,
				Interval:     cfg.Upstreams.CircuitBreaker.Interval,
				Timeout:      cfg.Upstreams.CircuitBreaker.Timeout,
				FailureRatio: cfg.Upstreams.CircuitBreaker.FailureRatio,
				MinSamples:   cfg.Upstreams.CircuitBreaker.MinSamples,
			}
		}

		opts = append(opts, opt)
	}

	proxy.Configure(opts)
}

func registerHealthChecks(healthChecker *health.Checker, redisClient *redis.Client) {
	if redisClient != nil {
		healthChecker.AddCheck("redis", func(ctx context.Context) error {
			return redisClient.Ping(ctx).Err()
		})
	}
}

func otelUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx, span := tracing.Tracer.Start(ctx, info.FullMethod)
		defer span.End()
		return handler(ctx, req)
	}
}

type gatewayHealthServer struct{}

func (s *gatewayHealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (s *gatewayHealthServer) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return stream.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
}

func getGinMode(env string) string {
	switch env {
	case "production", "staging":
		return gin.ReleaseMode
	default:
		return gin.DebugMode
	}
}

func extractHost(addr string) string {
	parts := splitHostPort(addr)
	return parts[0]
}

func extractPort(addr string) int {
	parts := splitHostPort(addr)
	if len(parts) >= 2 {
		p, err := strconv.Atoi(parts[1])
		if err == nil && p > 0 {
			return p
		}
	}
	return 8080
}

func splitHostPort(addr string) []string {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return []string{addr[:i], addr[i+1:]}
		}
	}
	return []string{addr}
}
