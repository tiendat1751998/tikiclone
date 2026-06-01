package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	sharedRedis "github.com/tikiclone/tiki/packages/go-shared/pkg/redis"
	"github.com/tikiclone/tiki/services/product-catalog/internal/application"
	"github.com/tikiclone/tiki/services/product-catalog/internal/config"
	"github.com/tikiclone/tiki/services/product-catalog/internal/health"
	"github.com/tikiclone/tiki/services/product-catalog/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/product-catalog/internal/infrastructure/redis"
	httptransport "github.com/tikiclone/tiki/services/product-catalog/internal/transport/http"
	kafkatransport "github.com/tikiclone/tiki/services/product-catalog/internal/transport/kafka"
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
	redisStore := redisinfra.NewStore(redisClient, cfg.Redis)

	pr := mysql.NewProductRepository(db)
	sr := mysql.NewSKURepository(db)
	cr := mysql.NewCategoryRepository(db)
	ar := mysql.NewAttributeRepository(db)
	mr := mysql.NewProductMediaRepository(db)

	var pub application.EventPublisher
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		pub = kafkatransport.NewProducer(cfg.Kafka.Brokers, cfg.AppName)
	}

	catService := application.NewCatalogService(pr, sr, cr, ar, mr, redisStore, cfg.ProductCacheTTL, cfg.CategoryCacheTTL, pub)

	hc := health.NewChecker(cfg.AppName, version, db, redisClient)

	gin.SetMode(getGinMode(cfg.AppEnv))
	engine := gin.New()

	handler := httptransport.NewHandler(catService)
	router := httptransport.NewRouter(handler, hc, cfg.JWTConfig.AccessSecret)
	router.Setup(engine)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      engine,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var bgWg sync.WaitGroup

	go func() {
		logger.Info("starting product catalog service", zap.Int("http_port", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server failed", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("shutting down product catalog service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown error", zap.Error(err))
	}

	bgWg.Wait()

	if pub != nil {
		if p, ok := pub.(*kafkatransport.Producer); ok {
			p.Close()
		}
	}

	if redisClient != nil {
		redisClient.Close()
	}

	logger.Info("product catalog service stopped")
}

func getGinMode(env string) string {
	if env == "production" || env == "staging" {
		return gin.ReleaseMode
	}
	return gin.DebugMode
}
