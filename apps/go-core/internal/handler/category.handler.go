// internal/handler/category_handler.go
package handler

import (
	"net/http"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

func (h *CategoryHandler) Create(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	var req models.CreateCategoryInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.BadRequest("invalid request body"))
		return
	}
	req.BusinessID = businessID

	if err := h.categoryService.Create(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, responses.Created[any]("category created successfully", nil))
}

func (h *CategoryHandler) Get(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	id := c.Param("id")

	category, err := h.categoryService.Get(c.Request.Context(), id, businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}
	if category == nil {
		c.JSON(http.StatusNotFound, responses.NotFound("category not found"))
		return
	}

	c.JSON(http.StatusOK, responses.Success("category fetched successfully", category))
}

func (h *CategoryHandler) GetAll(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	categories, err := h.categoryService.GetAll(c.Request.Context(), businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Success("categories fetched successfully", categories))
}
