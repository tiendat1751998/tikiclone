package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/dispute"
	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/ledger"
	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/payment"
	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/payout"
	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/reconciliation"
	httptransport "github.com/tikiclone/tiki/platforms/payment-ledger/internal/transport/http"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"
)

func init() {
	// Tune GC for low-latency: more frequent GCs, less heap growth
	if gogc := os.Getenv("GOGC"); gogc == "" {
		os.Setenv("GOGC", "50")
	}
}

func main() {
	paymentRepo := payment.NewInMemoryRepository()
	// Auto-tune GOMAXPROCS for container environments
	_, _ = automaxprocs.Set()

	paymentService := payment.NewService(paymentRepo)

	accountRepo := ledger.NewInMemoryAccountRepo()
	txnRepo := ledger.NewInMemoryTransactionRepo()
	entryRepo := ledger.NewInMemoryEntryRepo()
	ledgerService := ledger.NewService(accountRepo, txnRepo, entryRepo)

	payoutRepo := payout.NewInMemoryPayoutRepo()
	batchRepo := payout.NewInMemoryBatchRepo()
	payoutService := payout.NewService(payoutRepo, batchRepo)

	reconRepo := reconciliation.NewInMemoryRepository()
	reconService := reconciliation.NewService(reconRepo, nil, nil)

	disputeRepo := dispute.NewInMemoryRepository()
	disputeService := dispute.NewService(disputeRepo)

	handler := httptransport.NewHandler(paymentService, ledgerService, payoutService, reconService, disputeService)
	router := httptransport.NewRouter(handler)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	router.Setup(engine)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", 8080),
		Handler:      engine,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Printf("starting payment-ledger server on :8080")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
