package repository

import (
	"context"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type PlanRepo interface {
	GetPlans(ctx context.Context) ([]models.Plan, error)
}

type planRepo struct {
	db *sqlx.DB
}

func NewPlansRepo(db *sqlx.DB) PlanRepo {
	return &planRepo{
		db: db,
	}
}

func (r *planRepo) GetPlans(ctx context.Context) ([]models.Plan, error) {
	var plans []models.Plan

	query := `
		SELECT *
		FROM plans
		WHERE is_active = TRUE
		ORDER BY monthly_price ASC
	`

	err := r.db.SelectContext(ctx, &plans, query)
	if err != nil {
		return nil, err
	}

	return plans, nil
}