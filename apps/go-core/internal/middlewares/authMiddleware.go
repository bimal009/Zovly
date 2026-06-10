package middlewares

import (
	"net/http"
	"strings"

	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/gin-gonic/gin"
)

func RequireAuth(sessionRepo repository.SessionRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		authz := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(authz, "Bearer ")
		if !ok || token == "" {
			c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
			c.Abort()
			return
		}

		session, err := sessionRepo.GetByToken(c.Request.Context(), token)
		if err != nil || session == nil {
			c.JSON(http.StatusUnauthorized, responses.Unauthorized("invalid or expired session"))
			c.Abort()
			return
		}

		c.Set("userID", session.UserID)
		c.Set("userRole", session.UserRole)
		c.Set("userEmail", session.UserEmail)
		c.Set("session", session)

		c.Next()
	}
}

func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
			c.Abort()
			return
		}

		role, ok := raw.(models.UserRole)
		if !ok {
			c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
			c.Abort()
			return
		}

		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, responses.Forbidden("insufficient permissions"))
		c.Abort()
	}
}
