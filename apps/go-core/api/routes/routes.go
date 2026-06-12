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

	// ── products ──────────────────────────────────────────────────────────
	// authMiddleware  → valid session
	// businessMiddleware → sets businessID on ctx, verifies ownership
	products := api.Group("/products")
	products.Use(authMiddleware, businessMiddleware)
	{
		products.POST("", productHandler.Create)
		products.GET("", productHandler.List)
		products.GET("/low-stock", productHandler.LowStock) // before /:id or gin matches it
		products.GET("/:id", productHandler.GetByID)
		products.PATCH("/:id", productHandler.Update)
		products.DELETE("/:id", productHandler.Delete)
	}

	// ── services ──────────────────────────────────────────────────────────
	services := api.Group("/services")
	services.Use(authMiddleware, businessMiddleware)
	{
		services.POST("", serviceHandler.Create)
		services.GET("", serviceHandler.List)
		services.GET("/ai-context", serviceHandler.ListForAIContext) // before /:id
		services.GET("/:id", serviceHandler.GetByID)
		services.PATCH("/:id", serviceHandler.Update)
		services.DELETE("/:id", serviceHandler.Delete)
	}
}
