// internal/routes/auth_routes.go
package routes

import (
	"github.com/bimal009/Zovly/internal/handler"
	"github.com/bimal009/Zovly/pkg/utils"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, h *handler.AuthHandler, jwtUtil *utils.JWTUtil) {
	auth := rg.Group("/auth")
	{
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/verify-otp", h.VerifyOTP)
		auth.POST("/reset-password", h.ResetPassword)

	}
}
