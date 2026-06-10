package handler

import (
	"fmt"
	"net/http"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/models"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/gin-gonic/gin"
	imagekit "github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
)

type ImageHandler struct {
	cfg *config.Config
}

func NewImageHandler(cfg *config.Config) *ImageHandler {
	return &ImageHandler{cfg: cfg}
}

func (i *ImageHandler) CreateToken(c *gin.Context) {
	client := imagekit.NewClient(
		option.WithPrivateKey(i.cfg.ImageKit.PrivateKey),
	)

	params, err := client.Helper.GetAuthenticationParameters("", 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(fmt.Sprintf("failed to generate auth token: %s", err.Error())))
		return
	}

	data := models.ImageAuthTokenResponse{
		Signature: params["signature"].(string),
		Expire:    params["expire"].(int64),
		Token:     params["token"].(string),
	}

	c.JSON(http.StatusOK, responses.Success("auth token generated", data))
}
