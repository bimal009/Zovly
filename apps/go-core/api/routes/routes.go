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
	serviceHandler *handler.ServiceHandler,
	faqHandler *handler.FaqHandler,
	facebookHandler *handler.FacebookHandler,
	instagramHandler *handler.InstagramHandler,
	authMiddleware gin.HandlerFunc,
	businessMiddleware gin.HandlerFunc,

) {
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

	connections := api.Group("/connections")
	connections.Use(authMiddleware, businessMiddleware)
	{
		connections.GET("/facebook", facebookHandler.GetConnectionStatus)
		connections.GET("/instagram", instagramHandler.GetConnectionStatus)

		fbPages := connections.Group("/facebook/pages")
		{
			fbPages.PATCH("/:pageId/toggle", facebookHandler.ToggleFacebookPage)
		}
	}
}
