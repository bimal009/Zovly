package handler

import (
	"net/http"

	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	imageKitService service.ImageKitService
}

func NewImageHandler(imageKitService service.ImageKitService) *ImageHandler {
	return &ImageHandler{
		imageKitService: imageKitService,
	}
}

func (h *ImageHandler) CreateToken(c *gin.Context) {
	data, err := h.imageKitService.GetToken(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			responses.InternalServerError("failed to generate auth token"),
		)
		return
	}

	c.JSON(
		http.StatusOK,
		responses.Success("auth token generated", data),
	)
}
