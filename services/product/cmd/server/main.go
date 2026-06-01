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
	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	segmentioKafka "github.com/segmentio/kafka-go"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/product/internal/application"
	"github.com/tikiclone/tiki/services/product/internal/config"
	catalogGrpc "github.com/tikiclone/tiki/services/product/internal/transport/grpc"
	catalogHttp "github.com/tikiclone/tiki/services/product/internal/transport/http"
	"github.com/tikiclone/tiki/services/product/internal/infrastructure/kafka"
	"github.com/tikiclone/tiki/services/product/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/product/internal/infrastructure/redis"
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var version = "0.1.0"

func init() {
	// Tune GC for low-latency: more frequent GCs, less heap growth
	if gogc := os.Getenv("GOGC"); gogc == "" {
		os.Setenv("GOGC", "50")
	}
}

func main() {
	cfg := config.Load()

	logger := observability.InitLogger(cfg.ServiceName, cfg.LogLevel)

	// Auto-tune GOMAXPROCS for container environments
	if _, err := automaxprocs.Set(); err != nil {
		logger.Warn("failed to set automaxprocs", zap.Error(err))
	}

	shutdownTracer, err := observability.InitTracer(cfg.ServiceName, cfg.OpenTelemetry.Endpoint)
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer shutdownTracer()
	defer observability.Sync()

	// MySQL
	db, err := sqlx.Connect("mysql", cfg.MySQL.DSN())
	if err != nil {
		logger.Fatal("failed to connect to mysql", zap.Error(err))
	}
	defer db.Close()
	db.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MySQL.MaxLifetime)

	// Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})
	defer redisClient.Close()

	// Kafka
	kafkaProducer := kafka.NewProducer(cfg.Kafka.Brokers)
	defer kafkaProducer.Close()

	// Repositories
	productRepo := mysql.NewProductRepo(db)
	categoryRepo := mysql.NewCategoryRepo(db)
	attributeRepo := mysql.NewAttributeRepo(db)
	moderationRepo := mysql.NewModerationRepo(db)

	// Cache
	cache := redisinfra.NewCache(redisClient)

	// Services
	productService := application.NewProductService(productRepo, cache, kafkaProducer)
	categoryService := application.NewCategoryService(categoryRepo, cache.AsCategoryCache(), kafkaProducer)
	attributeService := application.NewAttributeService(attributeRepo, cache.AsAttributeCache(), kafkaProducer)
	_ = moderationRepo

	// Health checker with deep checks
	healthChecker := health.NewChecker(cfg.ServiceName, version)
	healthChecker.AddCheck("mysql", func(ctx context.Context) error {
		return db.PingContext(ctx)
	})
	healthChecker.AddCheck("redis", func(ctx context.Context) error {
		return redisClient.Ping(ctx).Err()
	})
	// [RELIABILITY] Deep health check — verify Kafka connectivity
	healthChecker.AddCheck("kafka", func(ctx context.Context) error {
		// Try to read metadata from Kafka to verify connectivity
		conn, err := segmentioKafka.DialContext(ctx, "tcp", cfg.Kafka.Brokers[0])
		if err != nil {
			return fmt.Errorf("kafka dial failed: %w", err)
		}
		defer conn.Close()
		_, err = conn.Brokers()
		if err != nil {
			return fmt.Errorf("kafka brokers check failed: %w", err)
		}
		return nil
	})

	// HTTP server
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(
		middleware.Recovery(),
		middleware.ErrorHandler(),
		middleware.RequestID(),
		middleware.CORS(),
		middleware.OTelMiddleware(cfg.ServiceName),
		observability.ObserveHTTPMetrics(cfg.ServiceName),
	)

	router.GET("/health", healthChecker.LivenessHandler())
	router.GET("/ready", healthChecker.ReadinessHandler())
	router.GET("/metrics", observability.MetricsHandler())

	api := router.Group("/api/v1")
	httpHandler := catalogHttp.NewHandler(productService, categoryService, attributeService)
	httpHandler.RegisterRoutes(api)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(catalogGrpc.OTelUnaryServerInterceptor()),
	)
	grpc_health_v1.RegisterHealthServer(grpcServer, &catalogGrpc.HealthServer{})
	reflection.Register(grpcServer)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("failed to listen grpc", zap.Int("port", cfg.GRPCPort), zap.Error(err))
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("starting http server", zap.Int("port", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server failed", zap.Error(err))
		}
	}()

	go func() {
		logger.Info("starting grpc server", zap.Int("port", cfg.GRPCPort))
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Fatal("grpc server failed", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("shutting down servers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	grpcServer.GracefulStop()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown error", zap.Error(err))
	}

	logger.Info("servers stopped")
}
