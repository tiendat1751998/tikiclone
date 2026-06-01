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
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"

	"github.com/tikiclone/tiki/platforms/live-scale/internal/cdn"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/config"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/health"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/logging"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/region"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/sfu"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/stream_health"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/transcoding"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/websocket_cluster"
	httpTransport "github.com/tikiclone/tiki/platforms/live-scale/internal/transport/http"
)

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


	logger := logging.NewLogger(cfg.AppEnv)
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracerShutdown, err := observability.InitTracer(cfg.OTEL.ServiceName, cfg.OTEL.Endpoint)
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer tracerShutdown()

	sfuRepo := sfu.NewInMemoryRepository()
	cdnRepo := cdn.NewInMemoryRepository()
	clusterRepo := websocket_cluster.NewInMemoryRepository()
	healthRepo := stream_health.NewInMemoryRepository()
	regionRepo := region.NewInMemoryRepository()
	transcodeRepo := transcoding.NewInMemoryRepository()

	sfuSvc := sfu.NewService(sfuRepo, nil)
	cdnSvc := cdn.NewService(cdnRepo, nil)
	clusterSvc := websocket_cluster.NewService(clusterRepo, nil)
	healthSvc := stream_health.NewService(healthRepo, nil)
	regionSvc := region.NewService(regionRepo, nil)
	transcodeSvc := transcoding.NewService(transcodeRepo)

	handler := httpTransport.NewHandler(sfuSvc, cdnSvc, clusterSvc, healthSvc, regionSvc, transcodeSvc)
	hc := health.NewChecker(cfg.AppName, "1.0.0")
	router := httpTransport.NewRouter(handler, hc)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	router.Setup(engine)

	httpPort := cfg.HTTPPort
	if httpPort == 0 {
		httpPort = 8081
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", httpPort),
		Handler:      engine,
		ReadTimeout:       5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		logger.Info("starting server", zap.Int("port", httpPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sfuSvc.CheckNodeHealth(ctx)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	srv.Shutdown(shutdownCtx)
	cancel()
}
