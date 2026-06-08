package handler

import (
	"net/http"

	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/gin-gonic/gin"
)


type PlanHandler struct{
	planService service.PlanService
}

func NewPlanHandler(planService service.PlanService)*PlanHandler{
	return  &PlanHandler{
		planService: planService,
	}
}

func (h *PlanHandler) GetAll(c *gin.Context) {
	plans, err := h.planService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.BadRequest("failed to get plans"))
		return
	}

	c.JSON(http.StatusOK, responses.Success("Plans fetched successfully", plans))
}