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

	"github.com/tikiclone/tiki/platforms/aiml/internal/config"
	"github.com/tikiclone/tiki/platforms/aiml/internal/embeddings"
	"github.com/tikiclone/tiki/platforms/aiml/internal/experiments"
	"github.com/tikiclone/tiki/platforms/aiml/internal/featurestore"
	"github.com/tikiclone/tiki/platforms/aiml/internal/health"
	"github.com/tikiclone/tiki/platforms/aiml/internal/inference"
	"github.com/tikiclone/tiki/platforms/aiml/internal/modelregistry"
	"github.com/tikiclone/tiki/platforms/aiml/internal/training"
	httptransport "github.com/tikiclone/tiki/platforms/aiml/internal/transport/http"
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

	featureRepo := featurestore.NewInMemoryRepository()
	featureSvc := featurestore.NewService(featureRepo)

	modelRepo := modelregistry.NewInMemoryRepository()
	modelSvc := modelregistry.NewService(modelRepo)

	trainingRepo := training.NewInMemoryRepository()
	trainingSvc := training.NewService(trainingRepo)

	predictor := inference.NewMockPredictor()
	inferenceSvc := inference.NewService(predictor)

	embedGen := embeddings.NewEmbeddingGenerator()
	embedStore := embeddings.NewInMemoryVectorStore()
	embedSvc := embeddings.NewService(embedGen, embedStore)

	experimentRepo := experiments.NewInMemoryRepository()
	experimentSvc := experiments.NewService(experimentRepo)

	handler := httptransport.NewHandler(featureSvc, modelSvc, trainingSvc, inferenceSvc, embedSvc, experimentSvc)
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
