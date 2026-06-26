package middlewares

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/gin-gonic/gin"
)

func RequireInternal(internalToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if internalToken == "" {
			c.JSON(http.StatusUnauthorized, responses.Unauthorized("internal calls disabled"))
			c.Abort()
			return
		}

		authz := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(authz, "Bearer ")
		if !ok || token == "" {
			c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
			c.Abort()
			return
		}

		if subtle.ConstantTimeCompare([]byte(token), []byte(internalToken)) != 1 {
			c.JSON(http.StatusUnauthorized, responses.Unauthorized("invalid internal token"))
			c.Abort()
			return
		}

		c.Next()
	}
}
