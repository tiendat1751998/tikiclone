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
	sharedRedis "github.com/tikiclone/tiki/packages/go-shared/pkg/redis"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/cart/internal/application"
	"github.com/tikiclone/tiki/services/cart/internal/config"
	"github.com/tikiclone/tiki/services/cart/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/cart/internal/infrastructure/redis"
	httptransport "github.com/tikiclone/tiki/services/cart/internal/transport/http"
	kafkatransport "github.com/tikiclone/tiki/services/cart/internal/transport/kafka"
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

	cartRepo := mysql.NewCartRepository(db)
	itemRepo := mysql.NewCartItemRepository(db)
	snapshotRepo := mysql.NewCartSnapshotRepository(db)
	mergeRepo := mysql.NewCartMergeHistoryRepository(db)

	var publisher application.EventPublisher
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		producer := kafkatransport.NewProducer(cfg.Kafka.Brokers, cfg.AppName)
		defer producer.Close()
		publisher = producer
	}

	cartService := application.NewCartService(
		cartRepo, itemRepo, snapshotRepo, mergeRepo,
		redisStore, cfg.CartTTL, cfg.CheckoutPreviewTTL, cfg.MaxCartItems, cfg.MaxQuantityPerItem, publisher,
	)

	healthChecker := health.NewChecker(cfg.AppName, version)
	healthChecker.AddCheck("database", func(ctx context.Context) error { return db.Ping() })
	if redisClient != nil {
		healthChecker.AddCheck("redis", func(ctx context.Context) error { return redisClient.Ping(ctx).Err() })
	}

	gin.SetMode(getGinMode(cfg.AppEnv))
	engine := gin.New()

	handler := httptransport.NewHandler(cartService)
	httpRouter := httptransport.NewRouter(handler, healthChecker, cfg.JWT.AccessSecret)
	httpRouter.Setup(engine)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      engine,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("starting cart service", zap.Int("http_port", cfg.HTTPPort), zap.String("env", cfg.AppEnv))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server failed", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("shutting down cart service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown error", zap.Error(err))
	}
	if redisClient != nil {
		redisClient.Close()
	}

	logger.Info("cart service stopped")
}

func getGinMode(env string) string {
	switch env {
	case "production", "staging":
		return gin.ReleaseMode
	default:
		return gin.DebugMode
	}
}
