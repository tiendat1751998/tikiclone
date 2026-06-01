package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	sharedRedis "github.com/tikiclone/tiki/packages/go-shared/pkg/redis"
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"

	"github.com/tikiclone/tiki/platforms/notification/internal/config"
	"github.com/tikiclone/tiki/platforms/notification/internal/dispatcher"
	"github.com/tikiclone/tiki/platforms/notification/internal/email"
	"github.com/tikiclone/tiki/platforms/notification/internal/events"
	"github.com/tikiclone/tiki/platforms/notification/internal/health"
	"github.com/tikiclone/tiki/platforms/notification/internal/inapp"
	"github.com/tikiclone/tiki/platforms/notification/internal/logging"
	"github.com/tikiclone/tiki/platforms/notification/internal/notifier"
	"github.com/tikiclone/tiki/platforms/notification/internal/preferences"
	"github.com/tikiclone/tiki/platforms/notification/internal/push"
	"github.com/tikiclone/tiki/platforms/notification/internal/sms"
	"github.com/tikiclone/tiki/platforms/notification/internal/template"
	"github.com/tikiclone/tiki/platforms/notification/internal/tracing"
	httpTransport "github.com/tikiclone/tiki/platforms/notification/internal/transport/http"
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


	logger := logging.NewLogger(cfg.Env)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	redisClient, err := sharedRedis.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Warn("redis not available", zap.Error(err))
		redisClient = nil
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	shutdownTracer, err := tracing.Init(cfg.AppName, cfg.OTEL.Endpoint)
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

	notifRepo := notifier.NewInMemoryRepository()
	pushRepo := push.NewInMemoryRepository()
	emailRepo := email.NewInMemoryRepository()
	smsRepo := sms.NewInMemoryRepository()
	inappRepo := inapp.NewInMemoryRepository()
	prefRepo := preferences.NewInMemoryRepository()
	tmplRepo := template.NewInMemoryRepository()
	dispatchRepo := dispatcher.NewInMemoryRepository()

	notifSvc := notifier.NewService(notifRepo, pub)
	pushSvc := push.NewService(pushRepo, pub)
	emailSvc := email.NewService(emailRepo, pub, cfg.SMTP.From)
	smsSvc := sms.NewService(smsRepo, pub, cfg.Twilio.FromNumber)
	inappSvc := inapp.NewService(inappRepo, pub)
	prefSvc := preferences.NewService(prefRepo)
	tmplSvc := template.NewService(tmplRepo)
	dispatchSvc := dispatcher.NewService(dispatchRepo, notifSvc, pushSvc, emailSvc, smsSvc, inappSvc, prefSvc)

	handler := httpTransport.NewHandler(notifSvc, pushSvc, emailSvc, smsSvc, inappSvc, prefSvc, tmplSvc, dispatchSvc)
	hc := health.NewChecker(cfg.AppName, cfg.Version)
	router := httpTransport.NewRouter(handler, hc)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	router.Setup(engine)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      engine,
		ReadTimeout:       5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		logger.Info("starting server", zap.String("port", cfg.HTTPPort))
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
