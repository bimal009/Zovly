package handler

import (
	"net/http"

	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/gin-gonic/gin"
)

type ServiceHandler struct {
	serviceService service.ServiceService
}

func NewServiceHandler(serviceService service.ServiceService) *ServiceHandler {
	return &ServiceHandler{serviceService: serviceService}
}

func (h *ServiceHandler) Create(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	var req models.CreateServiceInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.BadRequest("invalid request body"))
		return
	}
	req.BusinessID = businessID

	svc, err := h.serviceService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, responses.Created("service created successfully", svc))
}

// ─── GetByID ──────────────────────────────────────────────────────────────────

func (h *ServiceHandler) GetByID(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	id := c.Param("id")

	svc, err := h.serviceService.GetByID(c.Request.Context(), id, businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}
	if svc == nil {
		c.JSON(http.StatusNotFound, responses.NotFound("service not found"))
		return
	}

	c.JSON(http.StatusOK, responses.Success("service fetched successfully", svc))
}

// ─── List ─────────────────────────────────────────────────────────────────────

func (h *ServiceHandler) List(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	limit, offset := paginationFromQuery(c)

	f := repository.ListServicesFilter{
		Limit:  limit,
		Offset: offset,
	}
	if t := c.Query("type"); t != "" {
		serviceType := models.ServiceType(t)
		f.Type = &serviceType
	}
	if s := c.Query("status"); s != "" {
		status := models.ServiceStatus(s)
		f.Status = &status
	}
	if q := c.Query("search"); q != "" {
		f.Search = q
	}

	svcs, err := h.serviceService.List(c.Request.Context(), businessID, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Paginated("services fetched successfully", svcs, len(svcs), limit, offset))
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (h *ServiceHandler) Update(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	id := c.Param("id")

	var req models.UpdateServiceInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.BadRequest("invalid request body"))
		return
	}

	svc, err := h.serviceService.Update(c.Request.Context(), id, businessID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}
	if svc == nil {
		c.JSON(http.StatusNotFound, responses.NotFound("service not found"))
		return
	}

	c.JSON(http.StatusOK, responses.Success("service updated successfully", svc))
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (h *ServiceHandler) Delete(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	id := c.Param("id")

	if err := h.serviceService.Delete(c.Request.Context(), id, businessID); err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Success[any]("service deleted successfully", nil))
}

// ─── ListForAIContext ─────────────────────────────────────────────────────────
// Internal endpoint — used by the AI worker to build RAG context

func (h *ServiceHandler) ListForAIContext(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	svcs, err := h.serviceService.ListForAIContext(c.Request.Context(), businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.Success("services fetched successfully", svcs))
}
