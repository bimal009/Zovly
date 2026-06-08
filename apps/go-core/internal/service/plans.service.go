package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/bimal009/Zovly/internal/constants"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type PlanService interface {
	GetAll(ctx context.Context) ([]models.Plan, error)
}

type planService struct {
	db       *sqlx.DB
	rdb      *redis.Client
	logger   *slog.Logger
	planRepo repository.PlanRepo
}

func NewPlanService(
	db *sqlx.DB,
	rdb *redis.Client,
	logger *slog.Logger,
	planRepo repository.PlanRepo,
) PlanService {
	return &planService{
		db:       db,
		rdb:      rdb,
		logger:   logger,
		planRepo: planRepo,
	}
}

func (s *planService) GetAll(ctx context.Context) ([]models.Plan, error) {
	var plans []models.Plan

	cached, err := s.rdb.Get(ctx, constants.PlansCacheKey).Result()
	if err == nil {
		err = json.Unmarshal([]byte(cached), &plans)
		if err == nil {
			return plans, nil
		}
	}

	if err != nil && err != redis.Nil {
		s.logger.Error("redis get failed", "error", err)
	}

	plans, err = s.planRepo.GetPlans(ctx)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(plans)
	if err == nil {
		_ = s.rdb.Set(ctx, constants.PlansCacheKey, data, constants.TTLLong).Err()
	}

	return plans, nil
}