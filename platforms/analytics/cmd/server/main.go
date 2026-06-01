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
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"

	"github.com/tikiclone/tiki/platforms/analytics/internal/analytics"
	"github.com/tikiclone/tiki/platforms/analytics/internal/cohort"
	"github.com/tikiclone/tiki/platforms/analytics/internal/config"
	"github.com/tikiclone/tiki/platforms/analytics/internal/dashboard"
	"github.com/tikiclone/tiki/platforms/analytics/internal/events"
	"github.com/tikiclone/tiki/platforms/analytics/internal/funnel"
	"github.com/tikiclone/tiki/platforms/analytics/internal/health"
	"github.com/tikiclone/tiki/platforms/analytics/internal/logging"
	"github.com/tikiclone/tiki/platforms/analytics/internal/metrics"
	"github.com/tikiclone/tiki/platforms/analytics/internal/report_scheduler"
	"github.com/tikiclone/tiki/platforms/analytics/internal/session"
	"github.com/tikiclone/tiki/platforms/analytics/internal/tracing"
	httptransport "github.com/tikiclone/tiki/platforms/analytics/internal/transport/http"
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

	shutdownTracer, err := tracing.Init(cfg.AppName, cfg.OpenTelemetry.Endpoint)
	if err != nil {
		logger.Warn("failed to init tracer", zap.Error(err))
	}
	if shutdownTracer != nil {
		defer shutdownTracer()
	}

	var pub events.Publisher
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		p := events.NewKafkaProducer(cfg.Kafka.Brokers, cfg.AppName)
		defer p.Close()
		pub = p
	} else {
		pub = events.NewNoOpPublisher()
	}
	_ = pub
	_ = metrics.EventsIngestedTotal
	_ = metrics.ReportsGeneratedTotal
	_ = metrics.DashboardViewsTotal
	_ = metrics.QueryLatency
	_ = metrics.ScheduledReportsTotal
	_ = metrics.ActiveDashboards
	_ = metrics.EventProcessingDuration
	_ = metrics.FunnelAnalysesTotal
	_ = metrics.CohortAnalysesTotal
	_ = metrics.ActiveSessions

	eventRepo := events.NewInMemoryRepository()
	eventSvc := events.NewService(eventRepo)

	analyticsRepo := analytics.NewInMemoryRepository()
	analyticsSvc := analytics.NewService(analyticsRepo, eventSvc)

	funnelRepo := funnel.NewInMemoryRepository()
	funnelSvc := funnel.NewService(funnelRepo, eventSvc)

	cohortRepo := cohort.NewInMemoryRepository()
	cohortSvc := cohort.NewService(cohortRepo, eventSvc)

	sessionRepo := session.NewInMemoryRepository()
	sessionSvc := session.NewService(sessionRepo, eventSvc, cfg.Analytics.SessionTimeoutMinutes)

	dashboardRepo := dashboard.NewInMemoryRepository()
	dashboardSvc := dashboard.NewService(dashboardRepo)

	scheduleRepo := report_scheduler.NewInMemoryRepository()
	scheduleSvc := report_scheduler.NewService(scheduleRepo)

	handler := httptransport.NewHandler(analyticsSvc, eventSvc, funnelSvc, cohortSvc, sessionSvc, dashboardSvc, scheduleSvc, pub)
	hc := health.NewChecker(cfg.AppName, "1.0.0")
	router := httptransport.NewRouter(handler, hc)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	router.Setup(engine)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      engine,
		ReadTimeout:       5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		logger.Info("starting analytics server", zap.Int("port", cfg.HTTPPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
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
