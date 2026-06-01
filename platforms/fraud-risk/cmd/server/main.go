package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/behavior"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/decisionlog"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/devicefp"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/riskscoring"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/ruleengine"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/transactionmon"
	httpTransport "github.com/tikiclone/tiki/platforms/fraud-risk/internal/transport/http"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"
)

func init() {
	// Tune GC for low-latency: more frequent GCs, less heap growth
	if gogc := os.Getenv("GOGC"); gogc == "" {
		os.Setenv("GOGC", "50")
	}
}

func main() {
	ruleRepo := ruleengine.NewInMemoryRuleRepository()
	// Auto-tune GOMAXPROCS for container environments
	_, _ = automaxprocs.Set()

	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	riskRepo := riskscoring.NewInMemoryRepository()
	deviceRepo := devicefp.NewInMemoryRepository()
	txnRepo := transactionmon.NewInMemoryRepository()
	behavProfileRepo := behavior.NewInMemoryProfileRepository()
	behavRuleRepo := behavior.NewInMemoryRuleRepository()
	decRepo := decisionlog.NewInMemoryRepository()

	ruleEng := ruleengine.NewEngine(ruleRepo, rulesetRepo)
	riskCalc := riskscoring.NewCalculator(riskRepo)
	deviceSvc := devicefp.NewService(deviceRepo)
	txnMon := transactionmon.NewMonitor(txnRepo, deviceSvc)
	behavAnalyzer := behavior.NewAnalyzer(behavProfileRepo, behavRuleRepo)
	decLogger := decisionlog.NewLogger(decRepo)

	handler := httpTransport.NewHandler(ruleEng, riskCalc, deviceSvc, txnMon, behavAnalyzer, decLogger)
	router := httpTransport.SetupRouter(handler)

	gin.SetMode(gin.ReleaseMode)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:       5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
