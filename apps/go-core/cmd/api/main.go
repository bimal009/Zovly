package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/config/database"
	"github.com/bimal009/Zovly/internal/config/redis"
	"github.com/bimal009/Zovly/internal/handler"
	"github.com/bimal009/Zovly/internal/middlewares"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/internal/routes"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/logger"
	"github.com/bimal009/Zovly/pkg/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// ── config & infra ─────────────────────────────────────────────────────────
	cfg := config.MustLoad()
	slog := logger.New(cfg.App.Env)

	db, err := database.NewDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	rdb, err := redis.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	// ── dependencies ───────────────────────────────────────────────────────────
	jwtUtil := utils.NewJWTUtil(*cfg)
	userRepo := repository.NewUserRepo(db)
	accountRepo := repository.NewAccountRepo(db)
	authService := service.NewAuthService(db, userRepo, accountRepo, rdb, slog, jwtUtil)
	authHandler := handler.NewAuthHandler(authService, *cfg)

	// ── gin setup ──────────────────────────────────────────────────────────────
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.App.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	limiter := middlewares.NewRateLimiter(6, time.Second)
	r.Use(limiter.LimitMiddleWare())

	api := r.Group("/api/v1")
	routes.RegisterAuthRoutes(api, authHandler, jwtUtil)

	httpServer := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.App.Port, "env", cfg.App.Env)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("server forced shutdown: %v", err)
	}

	slog.Info("server exited")
}
