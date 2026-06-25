package routes

import (
	"github.com/bimal009/Zovly/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterAll(
	api *gin.RouterGroup,
	planHandler *handler.PlanHandler,
	paddleHandler *handler.PaddleHandler,
	imageHandler *handler.ImageHandler,
	businessHandler *handler.BusinessHandler,
	productHandler *handler.ProductHandler,
	categoryHandler *handler.CategoryHandler,
	serviceHandler *handler.ServiceHandler,
	faqHandler *handler.FaqHandler,
	facebookHandler *handler.FacebookHandler,
	instagramHandler *handler.InstagramHandler,
	inboxHandler *handler.InboxHandler,
	appHandler *handler.AppHandler,
	authMiddleware gin.HandlerFunc,
	businessMiddleware gin.HandlerFunc,

) {
	// Public webhook endpoints — no auth middleware (called by Meta servers).
	// Facebook and Instagram are separate routes, each verified with its own
	// app secret (META_APP_SECRET vs IG_APP_SECRET).
	webhook := api.Group("/webhook/meta")
	{
		webhook.GET("/facebook", facebookHandler.VerifyWebhook)
		webhook.POST("/facebook", facebookHandler.Webhook)
		webhook.GET("/instagram", instagramHandler.VerifyWebhook)
		webhook.POST("/instagram", instagramHandler.Webhook)
	}

	plans := api.Group("/plans")
	plans.GET("/all", planHandler.GetAll)

	paddlePayment := api.Group("/payment")
	paddlePayment.POST("/paddle/webhook", paddleHandler.Webhook)

	imageKit := api.Group("/images")
	imageKit.Use(authMiddleware)
	imageKit.GET("/auth", imageHandler.CreateToken)

	business := api.Group("/business")
	business.Use(authMiddleware)
	business.GET("/", businessHandler.Get)
	business.POST("/", businessHandler.Create)
	auth := api.Group("/auth")
	{
		auth.GET("/facebook/connect", authMiddleware, businessMiddleware, facebookHandler.ConnectFacebook)
		auth.GET("/facebook/callback", facebookHandler.FacebookCallback)
		auth.GET("/instagram/connect", authMiddleware, businessMiddleware, instagramHandler.ConnectInstagram)
		auth.GET("/instagram/callback", instagramHandler.InstagramCallback)
	}
	products := api.Group("/products")
	products.Use(authMiddleware, businessMiddleware)
	{
		products.POST("", productHandler.Create)
		products.GET("", productHandler.List)
		products.GET("/low-stock", productHandler.LowStock)
		products.GET("/:id", productHandler.GetByID)
		products.PATCH("/:id", productHandler.Update)
		products.DELETE("/:id", productHandler.Delete)
	}

	categories := api.Group("/categories")
	categories.Use(authMiddleware, businessMiddleware)
	{
		categories.POST("", categoryHandler.Create)
		categories.GET("", categoryHandler.GetAll)
		categories.GET("/:id", categoryHandler.Get)
	}

	services := api.Group("/services")
	services.Use(authMiddleware, businessMiddleware)
	{
		services.POST("", serviceHandler.Create)
		services.GET("", serviceHandler.List)
		services.GET("/ai-context", serviceHandler.ListForAIContext)
		services.GET("/:id", serviceHandler.GetByID)
		services.PATCH("/:id", serviceHandler.Update)
		services.DELETE("/:id", serviceHandler.Delete)
	}

	// ── faqs ──────────────────────────────────────────────────────────────
	faqs := api.Group("/faqs")
	faqs.Use(authMiddleware, businessMiddleware)
	{
		faqs.POST("/create", faqHandler.Create)
		faqs.GET("/all", faqHandler.GetAll)
	}

	inbox := api.Group("/inbox")
	inbox.Use(authMiddleware, businessMiddleware)
	{
		inbox.GET("/conversations", inboxHandler.ListConversations)
		inbox.GET("/conversations/:id/messages", inboxHandler.GetMessages)
	}

	connections := api.Group("/connections")
	connections.Use(authMiddleware, businessMiddleware)
	{
		connections.GET("/apps", appHandler.GetConnections)
		connections.GET("/facebook", facebookHandler.GetConnectionStatus)
		connections.GET("/instagram", instagramHandler.GetConnectionStatus)

		fbPages := connections.Group("/facebook/pages")
		{
			fbPages.PATCH("/:pageId/toggle", facebookHandler.ToggleFacebookPage)
		}

		messengerPages := connections.Group("/messenger/pages")
		{
			messengerPages.POST("/:pageId/subscribe", facebookHandler.SubscribeMessengerWebhook)
		}

		connections.POST("/instagram/activate", instagramHandler.ActivateInstagram)
	}
}
