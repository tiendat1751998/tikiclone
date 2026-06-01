package main
import ("context"; "fmt"; "net/http"; "os"; "os/signal"; "syscall"; "time"; "github.com/gin-gonic/gin"; sharedRedis "github.com/tikiclone/tiki/packages/go-shared/pkg/redis"; "github.com/tikiclone/tiki/platforms/user-behavior/internal/health"; "github.com/tikiclone/tiki/packages/go-shared/pkg/observability"; "github.com/tikiclone/tiki/platforms/user-behavior/internal/application"; "github.com/tikiclone/tiki/platforms/user-behavior/internal/config"; httptransport "github.com/tikiclone/tiki/platforms/user-behavior/internal/transport/http"; kafkatransport "github.com/tikiclone/tiki/platforms/user-behavior/internal/transport/kafka"; "go.uber.org/zap"
	automaxprocs "go.uber.org/automaxprocs/maxprocs")
var version = "1.0.0"

func init() {
	// Tune GC for low-latency: more frequent GCs, less heap growth
	if gogc := os.Getenv("GOGC"); gogc == "" {
		os.Setenv("GOGC", "50")
	}
}

func main() {
	cfg := config.Load(); logger := observability.InitLogger(cfg.AppName, cfg.LogLevel)

	// Auto-tune GOMAXPROCS for container environments
	if _, err := automaxprocs.Set(); err != nil {
		logger.Warn("failed to set automaxprocs", zap.Error(err))
	}

	shutdownTracer, _ := observability.InitTracer(cfg.OpenTelemetry.ServiceName, cfg.OpenTelemetry.Endpoint)
	defer shutdownTracer(); defer observability.Sync()
	redisClient, err := sharedRedis.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil { logger.Warn("redis not available", zap.Error(err)); redisClient = nil }
	var pub application.EventPublisher
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" { p := kafkatransport.NewProducer(cfg.Kafka.Brokers, cfg.AppName); defer p.Close(); pub = p }
	behaviorService := application.NewBehaviorService(pub)
	hc := health.NewChecker(cfg.AppName, version, redisClient)
	gin.SetMode(getGinMode(cfg.AppEnv)); engine := gin.New()
	handler := httptransport.NewHandler(behaviorService); router := httptransport.NewRouter(handler, hc.HealthChecker()); router.Setup(engine)
	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", cfg.HTTPPort), Handler: engine, ReadTimeout:       5 * time.Second, WriteTimeout:      10 * time.Second, IdleTimeout:       120 * time.Second}
	quit := make(chan os.Signal, 1); signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() { logger.Info("starting user behavior service", zap.Int("http_port", cfg.HTTPPort)); if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed { logger.Fatal("http server failed", zap.Error(err)) } }()
	<-quit; logger.Info("shutting down..."); ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second); defer cancel()
	httpServer.Shutdown(ctx); if redisClient != nil { redisClient.Close() }; logger.Info("stopped")
}
func getGinMode(env string) string { if env == "production" || env == "staging" { return gin.ReleaseMode }; return gin.DebugMode }
