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
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/billing/internal/config"
	"github.com/tikiclone/tiki/platforms/billing/internal/domain"
	"github.com/tikiclone/tiki/platforms/billing/internal/events"
	"github.com/tikiclone/tiki/platforms/billing/internal/health"
	"github.com/tikiclone/tiki/platforms/billing/internal/infrastructure/postgres"
	"github.com/tikiclone/tiki/platforms/billing/internal/ledger"
	"github.com/tikiclone/tiki/platforms/billing/internal/settlements"
	httptransport "github.com/tikiclone/tiki/platforms/billing/internal/transport/http"
	kafkatransport "github.com/tikiclone/tiki/platforms/billing/internal/transport/kafka"
	"github.com/tikiclone/tiki/platforms/billing/internal/wallets"
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"
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

	logger := observability.InitLogger(cfg.AppName, cfg.LogLevel)

	// Auto-tune GOMAXPROCS for container environments
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

	pgPool, err := postgres.NewPool(ctx, cfg.Postgres.DSN)
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer pgPool.Close()

	var publisher events.Publisher
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		p := kafkatransport.NewProducer(cfg.Kafka.Brokers, cfg.AppName)
		defer p.Close()
		publisher = p
	}

	accountRepo := postgres.NewAccountRepo(pgPool)
	txnRepo := postgres.NewTransactionRepo(pgPool)
	ledgerRepo := postgres.NewLedgerEntryRepo(pgPool)
	walletRepo := postgres.NewWalletRepo(pgPool)
	settlementRepo := postgres.NewSettlementRepo(pgPool)

	ledgerEngine := ledger.NewEngine(accountRepo, txnRepo, ledgerRepo, walletRepo, publisher)

	ledgerBridge := &wallets.LedgerBridge{
		PostTransaction: func(ctx context.Context, txnType domain.TransactionType, debit, credit string, amount int64, currency, description string) (*domain.Transaction, error) {
			return ledgerEngine.PostTransaction(ctx, txnType, debit, credit, amount, currency, description)
		},
	}

	walletService := wallets.NewService(walletRepo, accountRepo, ledgerBridge, publisher)
	settlementService := settlements.NewService(settlementRepo, publisher,
		func(ctx context.Context, txnType domain.TransactionType, debit, credit string, amount int64, currency, description string) (*domain.Transaction, error) {
			return ledgerEngine.PostTransaction(ctx, txnType, debit, credit, amount, currency, description)
		},
	)

	hc := health.NewChecker(cfg.AppName, version, nil)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	handler := httptransport.NewHandler(walletService, ledgerEngine, settlementService)
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
		logger.Info("starting billing service", zap.Int("http_port", cfg.HTTPPort))
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
	cancel()
	wg.Wait()
	logger.Info("stopped")
}
