package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"

	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/audience"
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/campaign"
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/content"
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/deliveryopt"
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/health"
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/reporting"
	httpTransport "github.com/tikiclone/tiki/platforms/notification-campaign/internal/transport/http"
)

func init() {
	// Tune GC for low-latency: more frequent GCs, less heap growth
	if gogc := os.Getenv("GOGC"); gogc == "" {
		os.Setenv("GOGC", "50")
	}
}

func main() {
	logger, _ := zap.NewProduction()
	// Auto-tune GOMAXPROCS for container environments
	_, _ = automaxprocs.Set()

	defer logger.Sync()

	campaignRepo := campaign.NewInMemoryRepository()
	audienceRepo := audience.NewInMemoryRepository()
	contentRepo := content.NewInMemoryRepository()
	deliveryRepo := deliveryopt.NewInMemoryRepository()
	reportingRepo := reporting.NewInMemoryRepository()

	campaignSvc := campaign.NewService(campaignRepo)
	audienceSvc := audience.NewService(audienceRepo)
	contentSvc := content.NewService(contentRepo)
	deliverySvc := deliveryopt.NewService(deliveryRepo)
	reportingSvc := reporting.NewService(reportingRepo)

	handler := httpTransport.NewHandler(campaignSvc, audienceSvc, contentSvc, deliverySvc, reportingSvc)
	hc := health.NewChecker("notification-campaign", "1.0.0")
	router := httpTransport.NewRouter(handler, hc)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	router.Setup(engine)

	srv := &http.Server{
		Addr:         ":8090",
		Handler:      engine,
		ReadTimeout:       5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		logger.Info("starting notification-campaign server", zap.String("port", "8090"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	srv.Shutdown(shutdownCtx)
}
