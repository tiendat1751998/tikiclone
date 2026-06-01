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

	"github.com/tikiclone/tiki/platforms/developer/internal/apikeys"
	"github.com/tikiclone/tiki/platforms/developer/internal/cicd"
	"github.com/tikiclone/tiki/platforms/developer/internal/docs"
	"github.com/tikiclone/tiki/platforms/developer/internal/onboarding"
	"github.com/tikiclone/tiki/platforms/developer/internal/sdk"
	httptransport "github.com/tikiclone/tiki/platforms/developer/internal/transport/http"
	"github.com/tikiclone/tiki/platforms/developer/internal/webhooks"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"
)

func init() {
	// Tune GC for low-latency: more frequent GCs, less heap growth
	if gogc := os.Getenv("GOGC"); gogc == "" {
		os.Setenv("GOGC", "50")
	}
}

func main() {
	apikeysRepo := apikeys.NewInMemoryRepository()
	// Auto-tune GOMAXPROCS for container environments
	_, _ = automaxprocs.Set()

	apikeysSvc := apikeys.NewService(apikeysRepo)

	docsRepo := docs.NewInMemoryRepository()
	docsSvc := docs.NewService(docsRepo)

	sdkRepo := sdk.NewInMemoryRepository()
	sdkSvc := sdk.NewService(sdkRepo)

	webhookRepo := webhooks.NewInMemoryWebhookRepository()
	deliveryRepo := webhooks.NewInMemoryDeliveryRepository()
	webhookSvc := webhooks.NewService(webhookRepo, deliveryRepo)

	cicdRepo := cicd.NewInMemoryRepository()
	cicdSvc := cicd.NewService(cicdRepo)

	onboardingRepo := onboarding.NewInMemoryRepository()
	onboardingSvc := onboarding.NewService(onboardingRepo)

	handler := httptransport.NewHandler(apikeysSvc, docsSvc, sdkSvc, webhookSvc, cicdSvc, onboardingSvc)
	router := httptransport.NewRouter(handler)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	router.Setup(engine)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      engine,
		ReadTimeout:       5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		fmt.Println("starting developer platform server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
