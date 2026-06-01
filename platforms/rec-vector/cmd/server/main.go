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

	"github.com/tikiclone/tiki/platforms/rec-vector/internal/collabvector"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/config"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/health"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/itemembedding"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/realtime"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/similarity"
	httptransport "github.com/tikiclone/tiki/platforms/rec-vector/internal/transport/http"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/userembedding"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/vectorstore"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"
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


	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vectorStore := vectorstore.NewInMemoryStore()

	userEmbRepo := userembedding.NewInMemoryRepository()
	userEmbSvc := userembedding.NewService(userEmbRepo)

	itemEmbRepo := itemembedding.NewInMemoryRepository()
	itemEmbSvc := itemembedding.NewService(itemEmbRepo)

	similaritySvc := similarity.NewService(vectorStore)

	collabRepo := collabvector.NewInMemoryRepository()
	collabSvc := collabvector.NewService(collabRepo)

	realtimeRepo := realtime.NewInMemoryRepository()
	realtimeSvc := realtime.NewService(realtimeRepo, itemEmbSvc, vectorStore)

	handler := httptransport.NewHandler(vectorStore, userEmbSvc, itemEmbSvc, similaritySvc, collabSvc, realtimeSvc)
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
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	srv.Shutdown(shutdownCtx)
	cancel()
}
