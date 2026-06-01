package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/production-dashboard/internal/application"
	"github.com/tikiclone/tiki/platforms/production-dashboard/internal/config"
	"github.com/tikiclone/tiki/platforms/production-dashboard/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/platforms/production-dashboard/internal/infrastructure/redis"
	httptransport "github.com/tikiclone/tiki/platforms/production-dashboard/internal/transport/http"
	kafkatransport "github.com/tikiclone/tiki/platforms/production-dashboard/internal/transport/kafka"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

var version = "1.0.0"

func main() {
	cfg := config.Load()
	logger := observability.InitLogger(cfg.AppName, cfg.LogLevel)

	shutdownTracer, err := observability.InitTracer(cfg.OpenTelemetry.ServiceName, cfg.OpenTelemetry.Endpoint)
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

	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		MaxRetries:   cfg.Redis.MaxRetries,
	})

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		logger.Warn("redis not available", zap.Error(err))
		redisClient = nil
	}
	var cacheStore application.CacheProvider
	if redisClient != nil {
		cacheStore = redisinfra.NewStore(redisClient, cfg.Redis)
	}

	healthRepo := mysql.NewServiceHealthRepository(db)
	deployRepo := mysql.NewDeploymentRepository(db)
	incidentRepo := mysql.NewIncidentRepository(db)
	alertRepo := mysql.NewAlertRuleRepository(db)
	auditRepo := mysql.NewAuditLogRepository(db)
	depRepo := mysql.NewServiceDependencyRepository(db)
	capacityRepo := mysql.NewCapacityMetricRepository(db)

	var publisher application.EventPublisher
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		producer := kafkatransport.NewProducer(cfg.Kafka.Brokers, cfg.AppName)
		defer producer.Close()
		publisher = producer
	}

	dashboardService := application.NewDashboardService(
		healthRepo, deployRepo, incidentRepo, alertRepo,
		auditRepo, depRepo, capacityRepo, publisher,
		time.Duration(cfg.Dashboard.IncidentRetentionDays)*24*time.Hour,
		cacheStore, 30*time.Second,
	)

	healthChecker := health.NewChecker(cfg.AppName, version)
	healthChecker.AddCheck("database", func(ctx context.Context) error { return db.Ping() })
	if redisClient != nil {
		healthChecker.AddCheck("redis", func(ctx context.Context) error { return redisClient.Ping(ctx).Err() })
	}

	gin.SetMode(getGinMode(cfg.AppEnv))
	engine := gin.New()

	handler := httptransport.NewHandler(dashboardService)
	httpRouter := httptransport.NewRouter(handler, healthChecker)
	httpRouter.Setup(engine)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("starting production dashboard",
			zap.Int("http_port", cfg.HTTPPort),
			zap.String("env", cfg.AppEnv),
		)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server failed", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("shutting down production dashboard...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown error", zap.Error(err))
	}
	if redisClient != nil {
		redisClient.Close()
	}

	logger.Info("production dashboard stopped")
}

func getGinMode(env string) string {
	switch env {
	case "production", "staging":
		return gin.ReleaseMode
	default:
		return gin.DebugMode
	}
}
