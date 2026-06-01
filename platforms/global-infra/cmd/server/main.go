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
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"

	"github.com/tikiclone/tiki/platforms/global-infra/internal/config"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/configmanager"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/featureflag"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/health"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/logging"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/multiregion"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/ratelimit"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/registry"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/secrets"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/tracing"
	httptransport "github.com/tikiclone/tiki/platforms/global-infra/internal/transport/http"
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

	shutdownTracer, _ := tracing.Init(cfg.AppName)
	defer shutdownTracer()

	hc := health.NewChecker(cfg.AppName, version)

	configRepo := configmanager.NewInMemoryRepository()
	configPub := configmanager.NewNoOpPublisher()
	configSvc := configmanager.NewService(configRepo, configPub, logger)

	featureFlagRepo := featureflag.NewInMemoryRepository()
	featureFlagSvc := featureflag.NewService(featureFlagRepo)

	multiRegionRepo := multiregion.NewInMemoryRepository()
	multiRegionSvc := multiregion.NewService(multiRegionRepo)

	registryRepo := registry.NewInMemoryRepository()
	registryHC := registry.NewHealthChecker()
	registrySvc := registry.NewService(registryRepo, registryHC)

	secretRepo := secrets.NewInMemoryRepository()
	secretSvc := secrets.NewService(secretRepo)

	rateLimitRepo := ratelimit.NewInMemoryRepository()
	rateLimiter := ratelimit.NewRateLimiter(rateLimitRepo)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	handler := httptransport.NewHandler(configSvc, featureFlagSvc, multiRegionSvc, registrySvc, secretSvc, rateLimiter)
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
		logger.Info("starting global-infra service", zap.Int("http_port", cfg.HTTPPort))
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
