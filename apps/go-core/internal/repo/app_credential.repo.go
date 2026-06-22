package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type AppCredentialRepo interface {
	Upsert(ctx context.Context, tx *sqlx.Tx, cred models.CreateAppCredential) error
	TogglePageActive(ctx context.Context, tx *sqlx.Tx, businessID, platformAccountID string) (bool, error)
	SetActiveByApp(ctx context.Context, businessID, appName string, active bool) error
	HasActiveByApp(ctx context.Context, businessID, appName string) (bool, error)
	GetExpiringInstaTokens(ctx context.Context) ([]models.AppCredential, error)
	UpdateToken(ctx context.Context, id, encToken string, expiresAt time.Time) error
	MarkTokenError(ctx context.Context, id, errMsg string) error
	ListByApp(ctx context.Context, businessID, appName string) ([]models.AppCredential, error)
	UpdateWebhookSubscribed(ctx context.Context, businessID, platformAccountID, appName string) error
	GetByPlatformAccountID(ctx context.Context, platformID string) (models.AppCredential, error)
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

// SetActiveByApp flips is_active for every credential a business holds under the
// given app. Instagram has a single account per business, so a simple app-scoped
// update is enough — unlike Facebook, which toggles individual pages by id.
func (r *appCredentialRepo) SetActiveByApp(ctx context.Context, businessID, appName string, active bool) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE app_credentials
		SET is_active = $3, updated_at = now()
		WHERE business_id = $1 AND app_name = $2
	`, businessID, appName, active)
	if err != nil {
		return fmt.Errorf("set active by app: %w", err)
	}
	return nil
}

func (r *appCredentialRepo) GetExpiringInstaTokens(ctx context.Context) ([]models.AppCredential, error) {
	query := `
		SELECT * FROM app_credentials
		WHERE app_name = 'instagram'
		  AND token_expires_at IS NOT NULL
		  AND token_expires_at <= NOW() + INTERVAL '24 hours'
		  AND token_expires_at > NOW()
	`
	var creds []models.AppCredential
	if err := r.db.SelectContext(ctx, &creds, query); err != nil {
		return nil, fmt.Errorf("get expiring instagram tokens: %w", err)
	}
	return creds, nil
}

func (r *appCredentialRepo) UpdateToken(ctx context.Context, id, encToken string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE app_credentials
		SET access_token     = $1,
		    token_expires_at = $2,
		    error_message    = NULL,
		    updated_at       = now()
		WHERE id = $3
	`, encToken, expiresAt, id)
	if err != nil {
		return fmt.Errorf("update token: %w", err)
	}
	return nil
}

func (r *appCredentialRepo) HasActiveByApp(ctx context.Context, businessID, appName string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `
		SELECT EXISTS(
			SELECT 1 FROM app_credentials
			WHERE business_id = $1 AND app_name = $2 AND is_active = true
		)
	`, businessID, appName)
	if err != nil {
		return false, fmt.Errorf("has active by app: %w", err)
	}
	return exists, nil
}

func (r *appCredentialRepo) MarkTokenError(ctx context.Context, id, errMsg string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE app_credentials
		SET error_message = $1,
		    updated_at    = now()
		WHERE id = $2
	`, errMsg, id)
	if err != nil {
		return fmt.Errorf("mark token error: %w", err)
	}
	return nil
}

func (r *appCredentialRepo) UpdateWebhookSubscribed(ctx context.Context, businessID, platformAccountID, appName string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE app_credentials
		SET webhook_subscribed_at = now(), updated_at = now()
		WHERE business_id = $1 AND platform_account_id = $2 AND app_name = $3
	`, businessID, platformAccountID, appName)
	if err != nil {
		return fmt.Errorf("update webhook subscribed: %w", err)
	}
	return nil
}

func (r *appCredentialRepo) GetByPlatformAccountID(ctx context.Context, platformID string) (models.AppCredential, error) {
	var cred models.AppCredential
	err := r.db.GetContext(ctx, &cred,
		`SELECT * FROM app_credentials WHERE platform_account_id = $1 LIMIT 1`,
		platformID,
	)
	if err != nil {
		return models.AppCredential{}, fmt.Errorf("get by platform account id: %w", err)
	}
	return cred, nil
}
