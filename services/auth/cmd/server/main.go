package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	sharedRedis "github.com/tikiclone/tiki/packages/go-shared/pkg/redis"
	"github.com/tikiclone/tiki/services/auth/internal/application"
	"github.com/tikiclone/tiki/services/auth/internal/config"
	"github.com/tikiclone/tiki/services/auth/internal/infrastructure/hash"
	"github.com/tikiclone/tiki/services/auth/internal/infrastructure/jwt"
	"github.com/tikiclone/tiki/services/auth/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/auth/internal/infrastructure/redis"
	"github.com/tikiclone/tiki/services/auth/internal/security"
	"github.com/tikiclone/tiki/services/auth/internal/tracing"
	grpctransport "github.com/tikiclone/tiki/services/auth/internal/transport/grpc"
	httptransport "github.com/tikiclone/tiki/services/auth/internal/transport/http"
	pb "github.com/tikiclone/tiki/services/auth/proto/auth/v1"
	"github.com/jmoiron/sqlx"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var version = "1.0.0"

func init() {
	// Tune GC for low-latency: more frequent GCs, less heap growth
	if gogc := os.Getenv("GOGC"); gogc == "" {
		// GOGC=50 is a good default for high-throughput services
		// Lower values = more frequent GC, less latency spikes
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

	db, err := mysql.NewDB(cfg.MySQL)
	if err != nil {
		logger.Fatal("failed to connect to mysql", zap.Error(err))
	}
	defer db.Close()

	redisClient, err := sharedRedis.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Warn("redis not available, continuing without redis", zap.Error(err))
		redisClient = nil
	}

	userRepo := mysql.NewUserRepository(db)
	sessionRepo := mysql.NewSessionRepository(db)
	auditRepo := mysql.NewAuditRepository(context.Background(), db, cfg.Audit)
	defer auditRepo.Stop()

	redisStore := redisinfra.NewStore(redisClient, cfg.Redis)

	hashService := hash.NewService(cfg.Password)

	jwtService := jwt.NewService(cfg.JWT, redisStore)
	if redisClient == nil {
		jwtService = jwt.NewService(cfg.JWT, nil)
	}

	rateLimiter := security.NewRateLimiter(redisClient, cfg.RateLimit)
	suspiciousDetector := security.NewSuspiciousDetector(redisClient, cfg.Security)

	authService := application.NewAuthService(
		cfg, userRepo, sessionRepo, auditRepo, redisStore,
		rateLimiter, suspiciousDetector, jwtService, hashService,
	)

	gin.SetMode(getGinMode(cfg.AppEnv))
	engine := gin.New()

	handler := httptransport.NewHandler(authService)

	hc := health.NewChecker(cfg.AppName, version)
	hc.AddCheck("mysql", func(ctx context.Context) error { return db.PingContext(ctx) })
	if redisClient != nil {
		hc.AddCheck("redis", func(ctx context.Context) error { return redisClient.Ping(ctx).Err() })
	}
	httpRouter := httptransport.NewRouter(handler, hc, redisClient)
	httpRouter.Setup(engine)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           engine,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	grpcServer := grpc.NewServer()
	grpcAuthServer := grpctransport.NewAuthGRPCServer(authService)
	pb.RegisterAuthServiceServer(grpcServer, grpcAuthServer)
	grpc_health_v1.RegisterHealthServer(grpcServer, &grpcHealthServer{db: db, redis: redisClient})
	reflection.Register(grpcServer)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("failed to listen grpc", zap.Int("port", cfg.GRPCPort), zap.Error(err))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("starting auth service",
			zap.Int("http_port", cfg.HTTPPort),
			zap.Int("grpc_port", cfg.GRPCPort),
			zap.String("env", cfg.AppEnv),
		)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server failed", zap.Error(err))
		}
	}()

	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Fatal("grpc server failed", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("shutting down auth service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	grpcServer.GracefulStop()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown error", zap.Error(err))
	}

	if redisClient != nil {
		redisClient.Close()
	}

	logger.Info("auth service stopped")
}

type grpcHealthServer struct {
	db    *sqlx.DB
	redis *redis.Client
}

func (s *grpcHealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	status := grpc_health_v1.HealthCheckResponse_SERVING
	if s.db != nil {
		if err := s.db.PingContext(ctx); err != nil {
			status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
		}
	}
	if status == grpc_health_v1.HealthCheckResponse_SERVING && s.redis != nil {
		if err := s.redis.Ping(ctx).Err(); err != nil {
			status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
		}
	}
	return &grpc_health_v1.HealthCheckResponse{Status: status}, nil
}

func (s *grpcHealthServer) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return stream.Send(&grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING})
}

func getGinMode(env string) string {
	switch env {
	case "production", "staging":
		return gin.ReleaseMode
	default:
		return gin.DebugMode
	}
}
