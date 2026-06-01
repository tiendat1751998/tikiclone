package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tikiclone/tiki/platforms/service-mesh/internal/discovery"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/loadbalancer"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/mtls"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/resilience"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/telemetry"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/traffic"
	httpTransport "github.com/tikiclone/tiki/platforms/service-mesh/internal/transport/http"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"
)

func init() {
	// Tune GC for low-latency: more frequent GCs, less heap growth
	if gogc := os.Getenv("GOGC"); gogc == "" {
		os.Setenv("GOGC", "50")
	}
}

func main() {
	discRepo := discovery.NewInMemoryRepository()
	// Auto-tune GOMAXPROCS for container environments
	_, _ = automaxprocs.Set()

	discSvc := discovery.NewService(discRepo)

	ca, err := mtls.NewCertificateAuthority("TikiClone", "Service Mesh CA", 3650)
	if err != nil {
		panic(err)
	}
	certMgr := mtls.NewCertManager(ca)

	trafficRepo := traffic.NewInMemoryRepository()
	trafficEng := traffic.NewEngine(trafficRepo)

	lb := loadbalancer.NewLoadBalancer(loadbalancer.RoundRobin)

	executor := resilience.NewExecutor()

	telRepo := telemetry.NewInMemoryRepository()

	handler := httpTransport.NewHandler(discSvc, certMgr, trafficEng, lb, executor, telRepo)
	router := httpTransport.NewRouter(handler)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	router.Setup(engine)

	srv := &http.Server{
		Addr:         ":8095",
		Handler:      engine,
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
