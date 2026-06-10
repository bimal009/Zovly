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
	authMiddleware gin.HandlerFunc,
) {
	plans := api.Group("/plans")
	plans.GET("/all", planHandler.GetAll)

	paddlePayment := api.Group("/payment")
	paddlePayment.POST("/paddle/webhook", paddleHandler.Webhook)

	imageKit := api.Group("/images")
	imageKit.Use(authMiddleware)
	imageKit.GET("/auth", imageHandler.CreateToken)
}
