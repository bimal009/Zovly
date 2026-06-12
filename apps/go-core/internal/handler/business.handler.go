package handler

import (
	"net/http"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/gin-gonic/gin"
)

type BusinessHandler struct {
	businessService service.BusinessService
}

func NewBusinessHandler(
	businessService service.BusinessService,
) *BusinessHandler {
	return &BusinessHandler{
		businessService: businessService,
	}
}

func (h *BusinessHandler) Get(c *gin.Context) {
	raw, _ := c.Get("session")
	session := raw.(*models.SessionWithUser)

	result, err := h.businessService.GetByUserId(c.Request.Context(), session.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, responses.NotFound("business not found"))
		return
	}

	c.JSON(http.StatusOK, responses.Success("ok", result))
}

func (h *BusinessHandler) Create(c *gin.Context) {
	raw, _ := c.Get("session")
	session := raw.(*models.SessionWithUser)
	if session.UserOnboarded {
		c.JSON(http.StatusForbidden, responses.Forbidden("already onboarded"))
		return
	}

	var req models.Business

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusBadRequest,
			responses.BadRequest("invalid request body"),
		)
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(
			http.StatusUnauthorized,
			responses.Unauthorized("unauthorized"),
		)
		return
	}

	business, err := h.businessService.Create(
		c.Request.Context(),
		req,
		userID.(string),
	)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			responses.BadRequest(err.Error()),
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		responses.Success("Business created sucessfully", business),
	)
}
