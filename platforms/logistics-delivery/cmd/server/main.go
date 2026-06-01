package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	sharedRedis "github.com/tikiclone/tiki/packages/go-shared/pkg/redis"
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"

	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/config"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/couriers"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/dispatch"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/estimations"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/events"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/fulfillment"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/health"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/idempotency"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/logging"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/pickups"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/replay"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/routing"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/shipments"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/synchronization"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/tracking"
	httpTransport "github.com/tikiclone/tiki/platforms/logistics-delivery/internal/transport/http"
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

	pgPool, err := connectPostgres(ctx, cfg)
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer pgPool.Close()

	producer := events.NewKafkaProducer(cfg.Kafka.Brokers, cfg.Kafka.ShipmentTopic)
	defer producer.Close()

	shipmentRepo := shipments.NewPostgresRepository(pgPool)
	trackingRepo := tracking.NewPostgresRepository(pgPool)
	routingRepo := routing.NewPostgresRepository(pgPool)
	dispatchRepo := dispatch.NewPostgresRepository(pgPool)
	courierRepo := couriers.NewPostgresRepository(pgPool)
	fulfillmentRepo := fulfillment.NewPostgresRepository(pgPool)
	pickupRepo := pickups.NewPostgresRepository(pgPool)
	estimationRepo := estimations.NewPostgresRepository(pgPool)

	var courierWebhookStore couriers.WebhookStore
	if redisClient != nil {
		courierWebhookStore = couriers.NewRedisWebhookStore(redisClient, 24*time.Hour)
	} else {
		courierWebhookStore = &noopWebhookStore{}
	}

	replaySvc := replay.NewService(30 * time.Minute)
	syncSvc := synchronization.NewService()

	idempotencyStore := idempotency.NewStore(redisClient, 24*time.Hour)

	shipmentSvc := shipments.NewService(shipmentRepo, producer)
	trackingSvc := tracking.NewService(trackingRepo, producer)
	routingSvc := routing.NewService(routingRepo)
	dispatchSvc := dispatch.NewService(dispatchRepo, producer)
	courierSvc := couriers.NewService(courierRepo, courierWebhookStore, producer)
	fulfillmentSvc := fulfillment.NewService(fulfillmentRepo, producer)
	pickupSvc := pickups.NewService(pickupRepo, producer)
	estimationSvc := estimations.NewService(estimationRepo, producer)

	handler := httpTransport.NewHandler(
		shipmentSvc, trackingSvc, routingSvc, dispatchSvc,
		courierSvc, fulfillmentSvc, pickupSvc, estimationSvc,
		replaySvc, syncSvc, idempotencyStore,
	)

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

func connectPostgres(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, cfg.Postgres.DSN)
}

type noopWebhookStore struct{}

func (n *noopWebhookStore) IsDuplicate(ctx context.Context, eventID string) (bool, error) { return false, nil }
func (n *noopWebhookStore) MarkProcessed(ctx context.Context, eventID string) error         { return nil }
