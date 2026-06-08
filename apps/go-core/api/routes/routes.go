package routes

import (
	"github.com/bimal009/Zovly/internal/handler"
	"github.com/bimal009/Zovly/pkg/utils"
	"github.com/gin-gonic/gin"
)

func RegisterAll(
	r *gin.Engine,
	jwtUtil *utils.JWTUtil,
	planHandler *handler.PlanHandler,
) {
	api := r.Group("/api/v1")
	plans:=api.Group("/plans")
	plans.GET("/all", planHandler.GetAll)
}