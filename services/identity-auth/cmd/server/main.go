package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/identity-auth/internal/application"
	"github.com/tikiclone/tiki/services/identity-auth/internal/config"
	"github.com/tikiclone/tiki/services/identity-auth/internal/infrastructure/jwt"
	"github.com/tikiclone/tiki/services/identity-auth/internal/infrastructure/mysql"
	rediss "github.com/tikiclone/tiki/services/identity-auth/internal/infrastructure/redis"
	httptransport "github.com/tikiclone/tiki/services/identity-auth/internal/transport/http"
	"go.uber.org/automaxprocs/maxprocs"
)

func main() {
	if _, err := maxprocs.Set(); err != nil {
		log.Printf("Warning: failed to set automaxprocs: %v", err)
	}

	cfg := config.Load()

	db, err := mysql.NewDB(&cfg.MySQL)
	if err != nil {
		log.Fatalf("failed to connect to mysql: %v", err)
	}
	defer db.Close()

	var redisStore *rediss.Store
	if cfg.Redis.Addr != "" {
		redisStore = rediss.NewStore(cfg.Redis)
		defer redisStore.Close()
	}

	userRepo := mysql.NewUserRepository(db)
	tokenRepo := mysql.NewRefreshTokenRepository(db)
	failedLoginRepo := mysql.NewFailedLoginAttemptRepository(db)
	outboxRepo := mysql.NewOutboxEventRepository(db)
	roleRepo := mysql.NewRoleRepository(db)
	userRoleRepo := mysql.NewUserRoleRepository(db)

	jwtService := jwt.NewService(cfg.JWT)

	authService := application.NewAuthService(
		cfg,
		userRepo,
		tokenRepo,
		failedLoginRepo,
		outboxRepo,
		roleRepo,
		userRoleRepo,
		jwtService,
		redisStore,
	)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	handler := httptransport.NewHandler(authService)
	httptransport.SetupRouter(engine, handler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      engine,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("starting identity-auth service on port %d", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("service stopped")
}