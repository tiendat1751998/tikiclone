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
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	sharedRedis "github.com/tikiclone/tiki/packages/go-shared/pkg/redis"
	"github.com/tikiclone/tiki/services/checkout/internal/application"
	"github.com/tikiclone/tiki/services/checkout/internal/config"
	"github.com/tikiclone/tiki/services/checkout/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/checkout/internal/infrastructure/redis"
	httptransport "github.com/tikiclone/tiki/services/checkout/internal/transport/http"
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"
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

	redisClient, err := sharedRedis.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Warn("redis not available", zap.Error(err))
		redisClient = nil
	}
	var redisStore *redisinfra.Store
	if redisClient != nil {
		redisStore = redisinfra.NewStore(redisClient, cfg.Redis)
	}

	checkoutRepo := mysql.NewCheckoutRepository(db)
	stepLogRepo := mysql.NewCheckoutStepLogRepository(db)
	pricingRepo := mysql.NewPricingSnapshotRepository(db)
	reservationRepo := mysql.NewReservationOrchestrationRepository(db)
	reconcileRepo := mysql.NewReconciliationJobRepository(db)

	checkoutService := application.NewCheckoutService(
		checkoutRepo, stepLogRepo, pricingRepo, reservationRepo, reconcileRepo,
		redisStore, cfg.SnapshotTTL, cfg.ReservationTimeout, cfg.IdempotencyTTL, cfg.MaxRetries, nil,
	)

	healthChecker := health.NewChecker(cfg.AppName, version)
	healthChecker.AddCheck("database", func(ctx context.Context) error { return db.Ping() })
	if redisClient != nil {
		healthChecker.AddCheck("redis", func(ctx context.Context) error { return redisClient.Ping(ctx).Err() })
	}

	gin.SetMode(getGinMode(cfg.AppEnv))
	engine := gin.New()
	handler := httptransport.NewHandler(checkoutService)
	httpRouter := httptransport.NewRouter(handler, healthChecker, cfg.JWTConfig.AccessSecret, redisClient)
	httpRouter.Setup(engine)

	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.HTTPPort), Handler: engine,
		ReadTimeout:       5 * time.Second, WriteTimeout:      10 * time.Second, IdleTimeout:       120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("starting checkout service", zap.Int("http_port", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server failed", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("shutting down checkout service...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	httpServer.Shutdown(shutdownCtx)
	if redisClient != nil {
		redisClient.Close()
	}
	logger.Info("checkout service stopped")
}

func getGinMode(env string) string {
	if env == "production" || env == "staging" {
		return gin.ReleaseMode
	}
	return gin.DebugMode
}
