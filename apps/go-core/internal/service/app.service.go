package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
)

type AppService interface {
	GetConnections(ctx context.Context, businessID string) (*models.AppConnections, error)
}

type appService struct {
	appRepo repository.AppRepo
	log     *slog.Logger
}

func NewAppService(appRepo repository.AppRepo, log *slog.Logger) AppService {
	return &appService{
		appRepo: appRepo,
		log:     log,
	}
}

func (s *appService) GetConnections(ctx context.Context, businessID string) (*models.AppConnections, error) {
	conn, err := s.appRepo.GetByBusinessID(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("get app connections: %w", err)
	}
	return conn, nil
}
