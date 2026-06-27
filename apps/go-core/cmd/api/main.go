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
	"github.com/bimal009/Zovly/internal/embed"
	"github.com/bimal009/Zovly/internal/handler"
	"github.com/bimal009/Zovly/internal/middlewares"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/internal/task"
	"github.com/bimal009/Zovly/pkg/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
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

	rdb, err := redis.NewRedis(cfg.Redis.Url)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	// Asynq connection. Parsed from the same Redis URL the app uses. If you want
	// the queue isolated from your cache/locks (recommended), point it at a
	// dedicated DB number — e.g. set the URL's path to /1.
	asynqOpt, err := asynq.ParseRedisURI(cfg.Redis.Url)
	if err != nil {
		log.Fatalf("failed to parse redis uri for asynq: %v", err)
	}

	queueClient := task.NewClient(asynqOpt, slog)
	defer queueClient.Close()

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
	productVariantRepo := repository.NewProductVariantRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)

	planService := service.NewPlanService(db, rdb, slog, planRepo)
	businessService := service.NewBusinessService(db, businessRepo, businessMemberRepo, userRepo, slog, appRepo)
	appService := service.NewAppService(appRepo, slog)
	embedClient := embed.New(cfg.App.AIServiceURL)
	productService := service.NewProductService(db, rdb, slog, productRepo, productVariantRepo, knowledgeRepo, categoryRepo, conversationRepo, embedClient)
	categoryService := service.NewCategoryService(rdb, slog, categoryRepo)
	serviceService := service.NewServiceService(db, rdb, slog, serviceRepo)
	faqService := service.NewFaqService(faqRepo, knowledgeRepo, slog, db, *cfg)

	chatService := service.NewChatService(db, messageRepo, appCredentialRepo, messageEmbedRepo, conversationRepo, *cfg, rdb, slog, queueClient)
	facebookService := service.NewFacebookService(db, appCredentialRepo, appRepo, messageRepo, cfg, chatService, slog)
	instagramService := service.NewInstagramService(db, appCredentialRepo, appRepo, messageRepo, cfg, slog, chatService)

	planHandler := handler.NewPlanHandler(planService)
	paddleHandler := handler.NewPaddleHandler(*cfg, subRepo, planRepo, payRepo)
	businessHandler := handler.NewBusinessHandler(businessService)
	imagekitService := service.NewImageKitService(cfg)
	imageHandler := handler.NewImageHandler(imagekitService)
	productHandler := handler.NewProductHandler(productService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	serviceHandler := handler.NewServiceHandler(serviceService)
	faqHandler := handler.NewFaqHandler(faqService)
	facebookHandler := handler.NewFacebookHandler(facebookService, chatService, rdb, cfg, slog)
	instagramHandler := handler.NewInstagramHandler(rdb, cfg, slog, instagramService)
	inboxHandler := handler.NewInboxHandler(conversationRepo, messageRepo)
	appHandler := handler.NewAppHandler(appService, slog)

	authMiddleware := middlewares.RequireAuth(sessionRepo)
	businessMiddleware := middlewares.RequireBusiness(businessService, memberInviteRepo)
	internalMiddleware := middlewares.RequireInternal(cfg.App.InternalCallsToken)
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
	routes.RegisterAll(api, planHandler, paddleHandler, imageHandler, businessHandler, productHandler, categoryHandler, serviceHandler, faqHandler, facebookHandler, instagramHandler, inboxHandler, appHandler, authMiddleware, businessMiddleware, internalMiddleware)
	httpServer := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	igWorker := workers.NewRefreshInstagramWorker(instagramService, slog)
	go igWorker.Run(ctx)

	aiWorker := workers.NewAIWorker(chatService, productService, slog, rdb, messageRepo, messageEmbedRepo, appCredentialRepo, *cfg, db, queueClient)

	queueServer := task.NewServer(asynqOpt, slog)
	aiWorker.Register(queueServer)
	if err := queueServer.Start(); err != nil {
		log.Fatalf("failed to start asynq server: %v", err)
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

	// Stop accepting new webhooks first, then drain in-flight queue tasks.
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("server forced shutdown: %v", err)
	}
	queueServer.Shutdown()

	slog.Info("server exited")
}
