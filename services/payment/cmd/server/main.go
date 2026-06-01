package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	sharedRedis "github.com/tikiclone/tiki/packages/go-shared/pkg/redis"
	"github.com/tikiclone/tiki/services/payment/internal/application"
	"github.com/tikiclone/tiki/services/payment/internal/config"
	"github.com/tikiclone/tiki/services/payment/internal/health"
	"github.com/tikiclone/tiki/services/payment/internal/infrastructure/fraud"
	"github.com/tikiclone/tiki/services/payment/internal/infrastructure/kafka"
	"github.com/tikiclone/tiki/services/payment/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/payment/internal/infrastructure/redis"
	"github.com/tikiclone/tiki/services/payment/internal/tracing"
	grpctransport "github.com/tikiclone/tiki/services/payment/internal/transport/grpc"
	httptransport "github.com/tikiclone/tiki/services/payment/internal/transport/http"
	"github.com/tikiclone/tiki/services/payment/internal/transport/http/middleware"
	pb "github.com/tikiclone/tiki/services/payment/proto/payment/v1"
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

	db, err := mysql.NewDB(cfg.MySQL)
	if err != nil {
		logger.Fatal("failed to connect to mysql", zap.Error(err))
	}
	defer db.Close()

	redisClient, err := sharedRedis.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Warn("redis not available", zap.Error(err))
		redisClient = nil
	}

	paymentRepo := mysql.NewPaymentRepository(db)
	redisStore := redisinfra.NewStore(redisClient, cfg.Redis)

	var kafkaProducer *kafka.Producer
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		kafkaProducer = kafka.NewProducer(cfg.Kafka)
	}

	fraudDetector := fraud.NewDetector(fraud.DetectorConfig{
		RiskThreshold: cfg.Payment.FraudRiskThreshold,
	})

	paymentService := application.NewPaymentService(cfg, paymentRepo, redisStore, kafkaProducer, fraudDetector)

	gin.SetMode(getGinMode(cfg.AppEnv))
	engine := gin.New()

	var authMw gin.HandlerFunc
	if cfg.JWT.AccessSecret != "" && cfg.JWT.AccessSecret != "change-me-in-production" {
		authMw = middleware.JWTAuth(cfg.JWT)
	}

	handler := httptransport.NewHandler(paymentService)

	var webhookMiddlewares []gin.HandlerFunc
	if redisClient != nil {
		webhookMiddlewares = append(webhookMiddlewares, middleware.RateLimit(redisStore, 100, time.Minute))
	}
	router := httptransport.NewRouter(handler, authMw, webhookMiddlewares...)
	router.Setup(engine)

	healthChecker := health.NewChecker(cfg.AppName, version, db, redisClient)
	engine.GET("/health/live", healthChecker.LivenessHandler())
	engine.GET("/health/ready", healthChecker.ReadinessHandler())

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           engine,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	grpcServer := grpc.NewServer()
	grpcPaymentServer := grpctransport.NewPaymentGRPCServer(paymentService)
	pb.RegisterPaymentServiceServer(grpcServer, grpcPaymentServer)
	grpc_health_v1.RegisterHealthServer(grpcServer, &grpcHealthServer{})
	reflection.Register(grpcServer)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("failed to listen grpc", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var bgWg sync.WaitGroup

	go func() {
		logger.Info("starting payment service", zap.Int("http_port", cfg.HTTPPort), zap.Int("grpc_port", cfg.GRPCPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server failed", zap.Error(err))
		}
	}()

	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Warn("grpc server stopped", zap.Error(err))
		}
	}()

	bgWg.Add(1)
	go func() {
		defer bgWg.Done()
		defer func() {
			if r := recover(); r != nil {
				zap.L().Error("panic in payment outbox worker", zap.Any("recover", r))
			}
		}()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				func() {
					oCtx, oCancel := context.WithTimeout(ctx, 30*time.Second)
					defer oCancel()
					if err := paymentService.ProcessOutboxEvents(oCtx); err != nil {
						zap.L().Warn("outbox processing failed", zap.Error(err))
					}
				}()
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down payment service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	httpServer.Shutdown(shutdownCtx)
	bgWg.Wait()

	if redisClient != nil {
		redisClient.Close()
	}
	if kafkaProducer != nil {
		kafkaProducer.Close()
	}

	logger.Info("payment service stopped")
}

type grpcHealthServer struct{}

func (s *grpcHealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
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
