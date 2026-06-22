package handler

import (
	"log/slog"
	"net/http"

	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/gin-gonic/gin"
)

type AppHandler struct {
	appService service.AppService
	log        *slog.Logger
}

func NewAppHandler(appService service.AppService, log *slog.Logger) *AppHandler {
	return &AppHandler{
		appService: appService,
		log:        log,
	}
}

func (h *AppHandler) GetConnections(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	conn, err := h.appService.GetConnections(c.Request.Context(), businessID)
	if err != nil {
		h.log.Error("failed to get app connections", "business_id", businessID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to get app connections"))
		return
	}

	c.JSON(http.StatusOK, responses.Success("business app connections", conn))
}
