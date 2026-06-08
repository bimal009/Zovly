package routes

import (
	"github.com/bimal009/Zovly/internal/handler"
	"github.com/bimal009/Zovly/internal/routes"
	"github.com/bimal009/Zovly/pkg/utils"
	"github.com/gin-gonic/gin"
)

func RegisterAll(
	r *gin.Engine,
	authHandler *handler.AuthHandler,

	jwtUtil *utils.JWTUtil,
) {
	api := r.Group("/api/v1")


}
