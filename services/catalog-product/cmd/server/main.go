package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	sharedConfig "github.com/tikiclone/tiki/packages/go-shared/pkg/config"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/kafka"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	sharedRedis "github.com/tikiclone/tiki/packages/go-shared/pkg/redis"
	catalogGrpc "github.com/tikiclone/tiki/services/catalog-product/internal/delivery/grpc"
	catalogHttp "github.com/tikiclone/tiki/services/catalog-product/internal/delivery/http"
	"github.com/tikiclone/tiki/services/catalog-product/internal/repository"
	"github.com/tikiclone/tiki/services/catalog-product/internal/usecase"
	"go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var version = "0.1.0"

func init() {
	// Tune GC for low-latency: more frequent GCs, less heap growth
	if gogc := os.Getenv("GOGC"); gogc == "" {
		os.Setenv("GOGC", "50")
	}
}

func main() {
	cfg := sharedConfig.Load()

	logger := observability.InitLogger("catalog-product", cfg.LogLevel)

	// Auto-tune GOMAXPROCS for container environments
	if _, err := automaxprocs.Set(); err != nil {
		logger.Warn("failed to set automaxprocs", zap.Error(err))
	}

	shutdownTracer, err := observability.InitTracer("catalog-product", cfg.OpenTelemetry.Endpoint)
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer shutdownTracer()
	defer observability.Sync()

	mongoClient, err := repository.NewMongoClient(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		logger.Fatal("failed to connect to mongodb", zap.Error(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		mongoClient.Disconnect(ctx)
	}()

	redisClient, err := sharedRedis.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Warn("redis not available, continuing without cache", zap.Error(err))
		redisClient = nil
	}

	kafkaProducer := kafka.NewProducer(cfg.Kafka.Brokers, "catalog-product")
	defer kafkaProducer.Close()

	productRepo := repository.NewProductRepository(mongoClient, cfg.MongoDB.Database)
	categoryRepo := repository.NewCategoryRepository(mongoClient, cfg.MongoDB.Database)
	productCache := repository.NewProductCache(redisClient)
	categoryCache := repository.NewCategoryCache(redisClient)
	productUseCase := usecase.NewProductUseCase(productRepo, productCache, kafkaProducer)
	categoryUseCase := usecase.NewCategoryUseCase(categoryRepo, categoryCache, kafkaProducer)

	healthChecker := health.NewChecker("catalog-product", version)
	healthChecker.AddCheck("mongodb", func(ctx context.Context) error {
		return mongoClient.Ping(ctx, nil)
	})
	if redisClient != nil {
		healthChecker.AddCheck("redis", func(ctx context.Context) error {
			return redisClient.Ping(ctx).Err()
		})
	}

	gin.SetMode(getGinMode(cfg.AppEnv))
	router := gin.New()
	router.Use(middleware.Recovery())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RequestID())
	router.Use(middleware.CORS())
	router.Use(middleware.OTelMiddleware("catalog-product"))
	router.Use(observability.ObserveHTTPMetrics("catalog-product"))

	router.GET("/health", healthChecker.LivenessHandler())
	router.GET("/ready", healthChecker.ReadinessHandler())
	router.GET("/metrics", observability.MetricsHandler())

	httpHandler := catalogHttp.NewHandler(productUseCase, categoryUseCase)
	api := router.Group("/api/v1")
	// TODO: Add auth middleware — GinJWTAuth requires secret from config
	// api.Use(auth.GinJWTAuth(cfg.JWTConfig.AccessSecret))
	{
		products := api.Group("/products")
		{
			products.GET("", httpHandler.ListProducts)
			products.GET("/search", httpHandler.SearchProducts)
			products.GET("/featured", httpHandler.GetFeaturedProducts)
			products.GET("/deals", httpHandler.GetDealsProducts)
			products.GET("/flash-sale", httpHandler.GetFlashSale)
			products.GET("/:spu_id", httpHandler.GetProduct)
			products.POST("", httpHandler.CreateProduct)
			products.PUT("/:spu_id", httpHandler.UpdateProduct)
			products.DELETE("/:spu_id", httpHandler.DeleteProduct)
		}
		categories := api.Group("/categories")
		{
			categories.GET("", httpHandler.ListCategories)
			categories.GET("/:slug", httpHandler.GetCategoryBySlug)
		}
	}

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadTimeout:       2 * time.Second,
			WriteTimeout:      2 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(catalogGrpc.OTelUnaryServerInterceptor()),
	)
	grpc_health_v1.RegisterHealthServer(grpcServer, &catalogGrpc.HealthServer{})
	reflection.Register(grpcServer)

	grpcPort := cfg.Port + 1
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		logger.Fatal("failed to listen grpc", zap.Int("port", grpcPort), zap.Error(err))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("starting grpc server", zap.Int("port", grpcPort))
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Fatal("grpc server failed", zap.Error(err))
		}
	}()

	go func() {
		logger.Info("starting http server", zap.Int("port", cfg.Port))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server failed", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("shutting down servers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	grpcServer.GracefulStop()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown error", zap.Error(err))
	}

	logger.Info("server stopped")
}

func getGinMode(env string) string {
	if env == "production" || env == "staging" {
		return gin.ReleaseMode
	}
	return gin.DebugMode
}
