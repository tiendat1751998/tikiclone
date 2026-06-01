package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"

	"github.com/tikiclone/tiki/platforms/api-gateway/internal/auth"
	"github.com/tikiclone/tiki/platforms/api-gateway/internal/circuitbreaker"
	"github.com/tikiclone/tiki/platforms/api-gateway/internal/edgecache"
	"github.com/tikiclone/tiki/platforms/api-gateway/internal/ratelimit"
	"github.com/tikiclone/tiki/platforms/api-gateway/internal/routes"
	"github.com/tikiclone/tiki/platforms/api-gateway/internal/transform"
	httptransport "github.com/tikiclone/tiki/platforms/api-gateway/internal/transport/http"
)

func init() {
	// Tune GC for low-latency: more frequent GCs, less heap growth
	if gogc := os.Getenv("GOGC"); gogc == "" {
		os.Setenv("GOGC", "50")
	}
}

func main() {
	routeRepo := routes.NewInMemoryRepository()
	// Auto-tune GOMAXPROCS for container environments
	_, _ = automaxprocs.Set()

	routeService := routes.NewService(routeRepo)

	rateLimitRepo := ratelimit.NewInMemoryRepository()
	rateLimiter := ratelimit.NewRateLimiter(rateLimitRepo)

	apiKeyStore := auth.NewAPIKeyStore()
	apiKeyValidator := auth.NewAPIKeyValidator(apiKeyStore)
	jwtHandler := auth.NewJWTHandler("api-gateway-secret-key-2024")
	keyRateLimiter := auth.NewKeyRateLimiter()

	transformRepo := transform.NewInMemoryRepository()
	transformer := transform.NewTransformer(transformRepo)
	composer := transform.NewComposer([]*transform.Transformer{transformer})

	cbRepo := circuitbreaker.NewInMemoryRepository()
	cbSvc := circuitbreaker.NewService(cbRepo)

	edgeCache := edgecache.NewCache()

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	handler := httptransport.NewHandler(routeService, rateLimiter, apiKeyStore, apiKeyValidator, jwtHandler, keyRateLimiter, transformer, composer, cbSvc, edgeCache)
	router := httptransport.NewRouter(handler)
	router.Setup(engine)

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      engine,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("starting api-gateway", zap.String("addr", ":8080"))
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
