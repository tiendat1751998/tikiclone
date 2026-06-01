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
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/application"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/cache"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/config"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/engagement"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/fanout"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/health"
	ch "github.com/tikiclone/tiki/platforms/live-commerce/internal/infrastructure/clickhouse"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/infrastructure/postgres"
	redi "github.com/tikiclone/tiki/platforms/live-commerce/internal/infrastructure/redis"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/moderation"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/recommendations"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/replay"
	httptransport "github.com/tikiclone/tiki/platforms/live-commerce/internal/transport/http"
	kafkatransport "github.com/tikiclone/tiki/platforms/live-commerce/internal/transport/kafka"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/websocket"
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"
)

var version = "1.0.0"

func init() {
	if gogc := os.Getenv("GOGC"); gogc == "" {
		os.Setenv("GOGC", "50")
	}
	// Tune net/http defaults for flash-sale traffic
	http.DefaultTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          256,
		MaxIdleConnsPerHost:   64,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func main() {
	cfg := config.Load()

	logger := observability.InitLogger(cfg.AppName, cfg.LogLevel)

	if _, err := automaxprocs.Set(); err != nil {
		logger.Warn("failed to set automaxprocs", zap.Error(err))
	}
	defer observability.Sync()

	shutdownTracer, err := observability.InitTracer(cfg.OpenTelemetry.ServiceName, cfg.OpenTelemetry.Endpoint)
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer shutdownTracer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	redisClient, err := sharedRedis.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Warn("redis not available, running without cache", zap.Error(err))
		redisClient = nil
	}

	pgPool, err := postgres.NewPool(ctx, cfg.Postgres.DSN)
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer pgPool.Close()

	var clickhouseConn *ch.Conn
	if cfg.ClickHouse.Enabled {
		conn, err := ch.NewConn(ctx, cfg.ClickHouse.DSN)
		if err != nil {
			logger.Warn("clickhouse not available", zap.Error(err))
		} else {
			clickhouseConn = conn
			defer clickhouseConn.Close()
		}
	}

	var redisStore *redi.Store
	var cacheStore *cache.Store
	if redisClient != nil {
		redisStore = redi.NewStore(redisClient)
		cacheStore = cache.NewStore(redisClient)
	}

	hub := websocket.NewHub()
	go hub.Run()

	broadcaster := fanout.NewBroadcaster(hub)

	var eventPub application.EventPublisher
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		kp := kafkatransport.NewProducer(cfg.Kafka.Brokers, cfg.AppName)
		defer kp.Close()
		eventPub = kp

		consumer := kafkatransport.NewConsumer(cfg.Kafka.Brokers, "live-commerce-group", kafkatransport.TopicLiveEvents, nil)
		consumer.Start(ctx)
		defer consumer.Close()
	}

	modFilter := moderation.NewFilter()
	modQueue := moderation.NewQueue()
	engCounters := engagement.NewCounters(redisStore, broadcaster)
	replayBuffer := replay.NewEventBuffer(200, 30*time.Minute)
	recEngine := recommendations.NewEngine()

	livestreamRepo := postgres.NewLivestreamRepo(pgPool)
	messageRepo := postgres.NewMessageRepo(pgPool)
	reactionRepo := postgres.NewReactionRepo(pgPool)
	giftRepo := postgres.NewGiftRepo(pgPool)
	pinnedRepo := postgres.NewPinnedProductRepo(pgPool)
	modRepo := postgres.NewModerationRepo(pgPool)

	liveService := application.NewLiveCommerceService(
		livestreamRepo, messageRepo, reactionRepo, giftRepo, pinnedRepo, modRepo,
		eventPub, redisStore, cacheStore, broadcaster, engCounters,
		modFilter, modQueue, replayBuffer, recEngine, clickhouseConn,
	)

	modQueue.StartWorker(ctx, func(ctx context.Context, action *domain.ModerationAction) error {
		return nil
	})

	hc := health.NewChecker(cfg.AppName, version, redisClient)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	handler := httptransport.NewHandler(liveService)
	router := httptransport.NewRouter(handler, hc, hub)
	router.Setup(engine)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           engine,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("starting live commerce service",
			zap.Int("http_port", cfg.HTTPPort),
			zap.Int("grpc_port", cfg.GRPCPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server failed", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				liveService.UpdateTrending(ctx)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	httpServer.Shutdown(shutdownCtx)
	hub.Stop()
	broadcaster.Stop()

	if redisClient != nil {
		redisClient.Close()
	}

	cancel()
	wg.Wait()
	logger.Info("stopped")
}
