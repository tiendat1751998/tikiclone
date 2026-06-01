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
	sharedRedis "github.com/tikiclone/tiki/packages/go-shared/pkg/redis"
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"

	"github.com/tikiclone/tiki/platforms/recommendation/internal/collaborative"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/config"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/content"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/events"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/health"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/logging"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/personalization"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/recommender"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/reranker"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/tracing"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/trending"
	httptransport "github.com/tikiclone/tiki/platforms/recommendation/internal/transport/http"
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
	// Auto-tune GOMAXPROCS for container environments
	_, _ = automaxprocs.Set()


	logger := logging.Init(cfg.AppName, cfg.LogLevel)
	defer logger.Sync()

	shutdownTracer, err := tracing.Init(cfg.OpenTelemetry)
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer shutdownTracer()

	redisClient, err := sharedRedis.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Warn("redis not available, running without cache", zap.Error(err))
		redisClient = nil
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	var pub events.Publisher
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		p := events.NewKafkaProducer(cfg.Kafka.Brokers, cfg.AppName)
		defer p.Close()
		pub = p
	} else {
		pub = events.NewNoOpPublisher()
	}

	collabRepo := collaborative.NewInMemoryRepository()
	collabSvc := collaborative.NewService(collabRepo)

	contentRepo := content.NewInMemoryRepository()
	contentSvc := content.NewService(contentRepo)

	trendingRepo := trending.NewInMemoryRepository()
	trendingSvc := trending.NewService(trendingRepo)

	personalRepo := personalization.NewInMemoryRepository()
	personalSvc := personalization.NewService(personalRepo)

	rerankerRepo := reranker.NewInMemoryRepository()
	rerankerSvc := reranker.NewService(rerankerRepo)

	recRepo := recommender.NewInMemoryRepository()
	recSvc := recommender.NewService(recRepo, collabSvc, contentSvc, trendingSvc, personalSvc, rerankerSvc)

	hc := health.NewChecker(cfg.AppName, version, redisClient)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	handler := httptransport.NewHandler(recSvc, trendingSvc, collabSvc, pub)
	router := httptransport.NewRouter(handler, hc)
	router.Setup(engine)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      engine,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("starting recommendation service", zap.Int("http_port", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	httpServer.Shutdown(shutdownCtx)
	wg.Wait()
	logger.Info("stopped")
}
