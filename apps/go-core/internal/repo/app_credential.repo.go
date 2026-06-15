package repository

import (
	"context"
	"fmt"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type AppCredentialRepo interface {
	Upsert(ctx context.Context, tx *sqlx.Tx, cred models.CreateAppCredential) error
	TogglePageActive(ctx context.Context, tx *sqlx.Tx, businessID, platformAccountID string) (bool, error)
	ListByApp(ctx context.Context, businessID, appName string) ([]models.AppCredential, error)
}

type appCredentialRepo struct {
	db *sqlx.DB
}

func NewAppCredentialRepo(db *sqlx.DB) AppCredentialRepo {
	return &appCredentialRepo{db: db}
}

func (r *appCredentialRepo) ListByApp(ctx context.Context, businessID, appName string) ([]models.AppCredential, error) {
	var creds []models.AppCredential
	err := r.db.SelectContext(ctx, &creds,
		`SELECT * FROM app_credentials WHERE business_id = $1 AND app_name = $2`,
		businessID, appName,
	)
	if err != nil {
		return nil, fmt.Errorf("list app credentials: %w", err)
	}
	return creds, nil
}

func (r *appCredentialRepo) Upsert(ctx context.Context, tx *sqlx.Tx, cred models.CreateAppCredential) error {
	query := `
INSERT INTO app_credentials
	(business_id, app_name, access_token, refresh_token, token_expires_at,
	 scopes, platform_account_id, platform_account_name, is_active, connected_at)
VALUES
	(:business_id, :app_name, :access_token, :refresh_token, :token_expires_at,
	 :scopes, :platform_account_id, :platform_account_name, :is_active, :connected_at)
ON CONFLICT (business_id, app_name, platform_account_id)
DO UPDATE SET
	access_token          = EXCLUDED.access_token,
	refresh_token         = EXCLUDED.refresh_token,
	token_expires_at      = EXCLUDED.token_expires_at,
	scopes                = EXCLUDED.scopes,
	platform_account_name = EXCLUDED.platform_account_name,
	connected_at          = EXCLUDED.connected_at,
	disconnected_at       = NULL,
	error_message         = NULL,
	updated_at            = now()
	`

	if _, err := tx.NamedExecContext(ctx, query, cred); err != nil {
		return fmt.Errorf("upsert app credential: %w", err)
	}
	return nil
}

func (r *appCredentialRepo) TogglePageActive(ctx context.Context, tx *sqlx.Tx, businessID, platformAccountID string) (bool, error) {
	query := `
		UPDATE app_credentials
		SET is_active = NOT is_active, updated_at = now()
		WHERE business_id = $1 AND app_name = 'facebook' AND platform_account_id = $2
		RETURNING is_active
	`

	var newState bool
	if err := tx.GetContext(ctx, &newState, query, businessID, platformAccountID); err != nil {
		return false, fmt.Errorf("toggle page active: %w", err)
	}
	return newState, nil
}
