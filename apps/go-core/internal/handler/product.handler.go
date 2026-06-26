// internal/handler/product_handler.go
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func businessIDFromCtx(c *gin.Context) (string, bool) {
	raw, exists := c.Get("businessID")
	if !exists {
		return "", false
	}
	id, ok := raw.(string)
	return id, ok
}

func paginationFromQuery(c *gin.Context) (limit, offset int) {
	limit = 20
	offset = 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	return
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (h *ProductHandler) Create(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	var req models.CreateProductInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.BadRequest("invalid request body"))
		return
	}
	req.BusinessID = businessID

	product, err := h.productService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, responses.Created("product created successfully", product))
}

// ─── GetByID ──────────────────────────────────────────────────────────────────

func (h *ProductHandler) GetByID(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	id := c.Param("id")

	product, err := h.productService.GetByID(c.Request.Context(), id, businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, responses.NotFound("product not found"))
		return
	}

	c.JSON(http.StatusOK, responses.Success("product fetched successfully", product))
}

// ─── List ─────────────────────────────────────────────────────────────────────

func (h *ProductHandler) List(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	limit, offset := paginationFromQuery(c)

	f := repository.ListProductsFilter{
		Limit:  limit,
		Offset: offset,
	}
	if s := c.Query("status"); s != "" {
		status := models.ProductStatus(s)
		f.Status = &status
	}
	if q := c.Query("search"); q != "" {
		f.Search = q
	}

	products, err := h.productService.List(c.Request.Context(), businessID, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Paginated("products fetched successfully", products, len(products), limit, offset))
}

func (h *ProductHandler) Count(c *gin.Context) {
	businessID := c.Query("businessID")
	if businessID == "" {
		c.JSON(http.StatusBadRequest, responses.BadRequest("businessID is required"))
		return
	}

	categorySlug := c.Query("categorySlug")

	count, err := h.productService.Count(c.Request.Context(), businessID, categorySlug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Success("product count fetched successfully", gin.H{"count": count}))
}

func (h *ProductHandler) Update(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	id := c.Param("id")

	var req models.UpdateProductInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.BadRequest("invalid request body"))
		return
	}

	product, err := h.productService.Update(c.Request.Context(), id, businessID, req)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, responses.NotFound("product not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Success("product updated successfully", product))
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (h *ProductHandler) Delete(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	id := c.Param("id")

	if err := h.productService.Delete(c.Request.Context(), id, businessID); err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Success[any]("product deleted successfully", nil))
}

// ─── LowStock ─────────────────────────────────────────────────────────────────

func (h *ProductHandler) LowStock(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	products, err := h.productService.LowStock(c.Request.Context(), businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Success("low stock products fetched successfully", products))
}
