package service

import (
	"context"
	"log/slog"

	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
)

type SubscriptionService interface {
	Create(ctx context.Context, subs models.BusinessSubscription) (*models.BusinessSubscription, error)
}

type subscriptionService struct {
	subscriptionRepo repository.SubscriptionRepo
	log              *slog.Logger
}

func NewSubscriptionService(subscriptionRepo repository.SubscriptionRepo, log *slog.Logger) SubscriptionService {
	return &subscriptionService{
		subscriptionRepo: subscriptionRepo,
		log:              log,
	}
}

func (s *subscriptionService) Create(ctx context.Context, subs models.BusinessSubscription) (*models.BusinessSubscription, error) {
	id, err := s.subscriptionRepo.Create(ctx, nil, subs)
	if err != nil {
		return nil, err
	}
	subs.ID = id
	return &subs, nil
}
