// internal/handler/product_variant_handler.go
package handler

import (
	"net/http"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/gin-gonic/gin"
)

type ProductVariantHandler struct {
	productVariantService service.ProductVariantService
}

func NewProductVariantHandler(productVariantService service.ProductVariantService) *ProductVariantHandler {
	return &ProductVariantHandler{productVariantService: productVariantService}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (h *ProductVariantHandler) Create(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	var req models.CreateProductVariantInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.BadRequest("invalid request body"))
		return
	}
	req.BusinessID = businessID

	variant, err := h.productVariantService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, responses.Created("product variant created successfully", variant))
}
