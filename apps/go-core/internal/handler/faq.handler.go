package handler

import (
	"net/http"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/gin-gonic/gin"
)

type FaqHandler struct {
	faqService service.FaqService
}

func NewFaqHandler(faqService service.FaqService) *FaqHandler {
	return &FaqHandler{faqService: faqService}
}

func (h *FaqHandler) Create(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	var req models.CreateFaqRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.BadRequest("invalid request body"))
		return
	}

	err := h.faqService.Create(c.Request.Context(), req, businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, responses.Created[any]("faq created successfully", nil))
}

func (h *FaqHandler) GetAll(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	faqs, err := h.faqService.GetAll(c.Request.Context(), businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Success("ok", faqs))
}
