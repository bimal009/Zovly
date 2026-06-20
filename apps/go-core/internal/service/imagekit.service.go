package service

import (
	"context"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/models"
	imagekit "github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
)

type ImageKitService interface {
	GetToken(ctx context.Context) (models.ImageAuthTokenResponse, error)
}

type imageKitService struct {
	cfg *config.Config
}

func NewImageKitService(cfg *config.Config) ImageKitService {
	return &imageKitService{
		cfg: cfg,
	}
}

func (s *imageKitService) GetToken(ctx context.Context) (models.ImageAuthTokenResponse, error) {
	client := imagekit.NewClient(
		option.WithPrivateKey(s.cfg.ImageKit.PrivateKey),
	)

	params, err := client.Helper.GetAuthenticationParameters("", 0)
	if err != nil {
		return models.ImageAuthTokenResponse{}, err
	}

	return models.ImageAuthTokenResponse{
		Signature: params["signature"].(string),
		Expire:    params["expire"].(int64),
		Token:     params["token"].(string),
	}, nil
}
