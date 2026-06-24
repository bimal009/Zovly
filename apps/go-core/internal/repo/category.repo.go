package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type CategoryRepo interface {
	Create(ctx context.Context, category models.CreateCategoryInput) error
	Get(ctx context.Context, id, businessID string) (*models.Category, error)
	GetAll(ctx context.Context, businessID string) ([]models.Category, error)
	ExistsByName(ctx context.Context, businessID, name string) (bool, error)
}

type categoryRepo struct {
	db *sqlx.DB
}

func NewCategoryRepo(db *sqlx.DB) CategoryRepo {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) Create(ctx context.Context, category models.CreateCategoryInput) error {
	const q = `
		INSERT INTO categories (business_id, name, description, slug)
		VALUES ($1, $2, $3, $4)`

	if _, err := r.db.ExecContext(ctx, q,
		category.BusinessID,
		category.Name,
		category.Description,
		category.Slug,
	); err != nil {
		return fmt.Errorf("category create: %w", err)
	}
	return nil
}

func (r *categoryRepo) Get(ctx context.Context, id, businessID string) (*models.Category, error) {
	var c models.Category
	err := r.db.GetContext(ctx, &c, `
		SELECT * FROM categories
		WHERE id = $1 AND business_id = $2
	`, id, businessID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("category get: %w", err)
	}
	return &c, nil
}

func (r *categoryRepo) GetAll(ctx context.Context, businessID string) ([]models.Category, error) {
	var categories []models.Category
	if err := r.db.SelectContext(ctx, &categories, `
		SELECT * FROM categories
		WHERE business_id = $1
		ORDER BY created_at DESC
	`, businessID); err != nil {
		return nil, fmt.Errorf("category list: %w", err)
	}
	return categories, nil
}

func (r *categoryRepo) ExistsByName(ctx context.Context, businessID, name string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `
		SELECT EXISTS (
			SELECT 1 FROM categories
			WHERE business_id = $1 AND name = $2
		)
	`, businessID, name)
	if err != nil {
		return false, fmt.Errorf("category exists by name: %w", err)
	}
	return exists, nil
}
