package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"fmt"
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"

	"github.com/tikiclone/tiki/platforms/fraud/internal/blacklist"
	fraudcase "github.com/tikiclone/tiki/platforms/fraud/internal/case"
	"github.com/tikiclone/tiki/platforms/fraud/internal/config"
	"github.com/tikiclone/tiki/platforms/fraud/internal/detection"
	"github.com/tikiclone/tiki/platforms/fraud/internal/events"
	"github.com/tikiclone/tiki/platforms/fraud/internal/health"
	"github.com/tikiclone/tiki/platforms/fraud/internal/logging"
	"github.com/tikiclone/tiki/platforms/fraud/internal/rules"
	"github.com/tikiclone/tiki/platforms/fraud/internal/scoring"
	"github.com/tikiclone/tiki/platforms/fraud/internal/streaming"
	"github.com/tikiclone/tiki/platforms/fraud/internal/tracing"
	httpTransport "github.com/tikiclone/tiki/platforms/fraud/internal/transport/http"
	"github.com/tikiclone/tiki/platforms/fraud/internal/verification"
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

	detectRepo := detection.NewInMemoryRepository()
	ruleRepo := rules.NewInMemoryRepository()
	scoreRepo := scoring.NewInMemoryRepository()
	streamRepo := streaming.NewInMemoryRepository()
	blacklistRepo := blacklist.NewInMemoryRepository()
	caseRepo := fraudcase.NewInMemoryRepository()
	verifyRepo := verification.NewInMemoryRepository()

	ruleSvc := rules.NewService(ruleRepo)
	scoreSvc := scoring.NewService(scoreRepo)
	streamSvc := streaming.NewService(streamRepo)
	blacklistSvc := blacklist.NewService(blacklistRepo)
	caseSvc := fraudcase.NewService(caseRepo)
	verifySvc := verification.NewService(verifyRepo, cfg.Verification.CodeExpiryMinutes)
	detectSvc := detection.NewService(detectRepo, ruleSvc, scoreSvc, streamSvc, blacklistSvc, float64(cfg.Fraud.DefaultThreshold))

	handler := httpTransport.NewHandler(detectSvc, ruleSvc, blacklistSvc, caseSvc, verifySvc)
	hc := health.NewChecker(cfg.AppName, "1.0.0")
	router := httpTransport.NewRouter(handler, hc)

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
		logger.Info("starting fraud detection server", zap.Int("port", cfg.HTTPPort))
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
