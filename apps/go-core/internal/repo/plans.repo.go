package repository

import (
	"context"
	"database/sql"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type PlanRepo interface {
	GetPlans(ctx context.Context) ([]models.Plan, error)
	GetByID(ctx context.Context, id string) (*models.Plan, error)
	GetByPaddlePriceID(ctx context.Context, priceID string) (*models.Plan, error)
}

type planRepo struct {
	db *sqlx.DB
}

func NewPlansRepo(db *sqlx.DB) PlanRepo {
	return &planRepo{db: db}
}

func (r *planRepo) GetPlans(ctx context.Context) ([]models.Plan, error) {
	var plans []models.Plan
	err := r.db.SelectContext(ctx, &plans, `
		SELECT * FROM plans
		WHERE is_active = TRUE
		ORDER BY monthly_price ASC
	`)
	if err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *planRepo) GetByID(ctx context.Context, id string) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.GetContext(ctx, &plan, `
		SELECT * FROM plans WHERE id = $1
	`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &plan, err
}

func (r *planRepo) GetByPaddlePriceID(ctx context.Context, priceID string) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.GetContext(ctx, &plan, `
		SELECT * FROM plans
		WHERE paddle_price_id_monthly = $1
		   OR paddle_price_id_yearly  = $1
		LIMIT 1
	`, priceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &plan, err
}
