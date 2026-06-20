package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bimal009/Zovly/api/routes"
	"github.com/bimal009/Zovly/cmd/workers"
	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/config/database"
	"github.com/bimal009/Zovly/internal/config/redis"
	"github.com/bimal009/Zovly/internal/handler"
	"github.com/bimal009/Zovly/internal/middlewares"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	_ = godotenv.Load() // load .env if present; no-op in prod where vars are injected

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

	planRepo := repository.NewPlansRepo(db)
	subRepo := repository.NewSubscriptionRepo(db)
	payRepo := repository.NewPaymentRepo(db)
	sessionRepo := repository.NewSessionRepo(db)
	userRepo := repository.NewUserRepo(db)
	businessRepo := repository.NewBusinessRepo(db)
	businessMemberRepo := repository.NewBusinessMemberRepo(db)
	memberInviteRepo := repository.NewMemberInviteRepo(db)
	productRepo := repository.NewProductRepo(db)
	serviceRepo := repository.NewServiceRepo(db)
	faqRepo := repository.NewFaqRepo(db)
	knowledgeRepo := repository.NewBusinessKnowledgeRepo(db)
	appRepo := repository.NewAppRepo(db)
	appCredentialRepo := repository.NewAppCredentialRepo(db)
	messageRepo := repository.NewMessageRepo(db)
	messageEmbedRepo := repository.NewMessageEmbeddingRepo(db)
	conversationRepo := repository.NewconversationRepo(db)

	planService := service.NewPlanService(db, rdb, slog, planRepo)
	businessService := service.NewBusinessService(db, businessRepo, businessMemberRepo, userRepo, slog, appRepo)
	productService := service.NewProductService(db, rdb, slog, productRepo)
	serviceService := service.NewServiceService(db, rdb, slog, serviceRepo)
	faqService := service.NewFaqService(faqRepo, knowledgeRepo, slog, db, *cfg)
	facebookService := service.NewFacebookService(db, appCredentialRepo, appRepo, cfg, slog)
	instagramService := service.NewInstagramService(db, appCredentialRepo, appRepo, cfg, slog)
	chatService := service.NewChatService(db, messageRepo, appCredentialRepo, messageEmbedRepo, conversationRepo, *cfg, rdb, slog)

	planHandler := handler.NewPlanHandler(planService)
	paddleHandler := handler.NewPaddleHandler(*cfg, subRepo, planRepo, payRepo)
	businessHandler := handler.NewBusinessHandler(businessService)
	imagekitService := service.NewImageKitService(cfg)
	imageHandler := handler.NewImageHandler(imagekitService)
	productHandler := handler.NewProductHandler(productService)
	serviceHandler := handler.NewServiceHandler(serviceService)
	faqHandler := handler.NewFaqHandler(faqService)
	facebookHandler := handler.NewFacebookHandler(facebookService, chatService, rdb, cfg, slog)
	instagramHandler := handler.NewInstagramHandler(rdb, cfg, slog, instagramService)
	chatHandler := handler.NewChatHandler(facebookService, chatService, rdb, cfg, slog)
	inboxHandler := handler.NewInboxHandler(conversationRepo, messageRepo)

	authMiddleware := middlewares.RequireAuth(sessionRepo)
	businessMiddleware := middlewares.RequireBusiness(businessService, memberInviteRepo)
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
	routes.RegisterAll(api, planHandler, paddleHandler, imageHandler, businessHandler, productHandler, serviceHandler, faqHandler, facebookHandler, instagramHandler, chatHandler, inboxHandler, authMiddleware, businessMiddleware)
	httpServer := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	igWorker := workers.NewRefreshInstagramWorker(instagramService, slog)
	go igWorker.Run(ctx)

	aiWorker := workers.NewAIWorker(chatService, slog, rdb, messageRepo, messageEmbedRepo, appCredentialRepo, *cfg, db)
	go aiWorker.Run(ctx)

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
