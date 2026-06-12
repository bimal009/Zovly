package middlewares

import (
	"net/http"

	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/gin-gonic/gin"
)

func RequireBusiness(
	businessService service.BusinessService,
	inviteRepo repository.MemberInviteRepo,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		userEmail := c.GetString("userEmail")

		hasPending, err := inviteRepo.HasPendingByEmail(c.Request.Context(), userEmail)
		if err == nil && hasPending {
			c.JSON(http.StatusUnauthorized, responses.Unauthorized("accept your pending invite first"))
			c.Abort()
			return
		}

		result, err := businessService.GetByUserId(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.Unauthorized("no business found for this user"))
			c.Abort()
			return
		}

		c.Set("business", result.Business)
		c.Set("member", result.Member)
		c.Set("businessID", result.Business.ID)
		c.Next()
	}
}

// MemberFromCtx is a helper handlers can use to pull the member out of context.
func MemberFromCtx(c *gin.Context) (*models.BusinessMember, bool) {
	raw, exists := c.Get("member")
	if !exists {
		return nil, false
	}
	m, ok := raw.(models.BusinessMember)
	return &m, ok
}
