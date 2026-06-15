package repository

import (
	"context"
	"fmt"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type AppRepo interface {
	Create(ctx context.Context, tx *sqlx.Tx, businessId string) error
	Update(ctx context.Context, tx *sqlx.Tx, businessID string, platform models.ConnectionType) error
	GetByBusinessID(ctx context.Context, businessID string) (*models.AppConnections, error)
}

type appRepo struct {
	db *sqlx.DB
}

func NewAppRepo(db *sqlx.DB) AppRepo {
	return &appRepo{db: db}
}

func (r *appRepo) Create(ctx context.Context, tx *sqlx.Tx, businessID string) error {
	query := `INSERT INTO app_connections (business_id) VALUES (:business_id)`

	_, err := tx.NamedExecContext(ctx, query, map[string]any{
		"business_id": businessID,
	})
	if err != nil {
		return fmt.Errorf("create app connections: %w", err)
	}
	return nil
}

func (r *appRepo) GetByBusinessID(ctx context.Context, businessID string) (*models.AppConnections, error) {
	var conn models.AppConnections
	err := r.db.GetContext(ctx, &conn, `SELECT * FROM app_connections WHERE business_id = $1`, businessID)
	if err != nil {
		return nil, fmt.Errorf("get app connections: %w", err)
	}
	return &conn, nil
}

func (r *appRepo) Update(ctx context.Context, tx *sqlx.Tx, businessID string, platform models.ConnectionType) error {
	column, ok := models.ConnectionColumns[platform]
	if !ok {
		return fmt.Errorf("unknown connection type: %s", platform)
	}

	query := fmt.Sprintf(`
		UPDATE app_connections
		SET %s = true, updated_at = now()
		WHERE business_id = $1
	`, column)

	_, err := tx.ExecContext(ctx, query, businessID)
	if err != nil {
		return fmt.Errorf("update app connections: %w", err)
	}
	return nil
}
