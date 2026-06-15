package models

import (
	"time"

	"github.com/lib/pq"
)

type AppCredential struct {
	ID         string `db:"id" json:"id"`
	BusinessID string `db:"business_id" json:"business_id"`
	AppName    string `db:"app_name" json:"app_name"` // "facebook" | "instagram" | "tiktok" | ...

	AccessToken    *string        `db:"access_token" json:"-"`  // encrypted, never serialized
	RefreshToken   *string        `db:"refresh_token" json:"-"` // encrypted, never serialized
	TokenExpiresAt *time.Time     `db:"token_expires_at" json:"token_expires_at"`
	Scopes         pq.StringArray `db:"scopes" json:"scopes"`

	PublicKey  *string `db:"public_key" json:"-"` // encrypted
	SecretKey  *string `db:"secret_key" json:"-"` // encrypted
	MerchantID *string `db:"merchant_id" json:"merchant_id"`

	PlatformAccountID   *string `db:"platform_account_id" json:"platform_account_id"`
	PlatformAccountName *string `db:"platform_account_name" json:"platform_account_name"`

	WebhookVerifyToken  *string    `db:"webhook_verify_token" json:"-"`
	WebhookSubscribedAt *time.Time `db:"webhook_subscribed_at" json:"webhook_subscribed_at"`
	IsActive            bool       `db:"is_active" json:"is_active"`
	ConnectedAt         *time.Time `db:"connected_at" json:"connected_at"`
	DisconnectedAt      *time.Time `db:"disconnected_at" json:"disconnected_at"`
	LastSyncAt          *time.Time `db:"last_sync_at" json:"last_sync_at"`
	ErrorMessage        *string    `db:"error_message" json:"error_message"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CreateAppCredential struct {
	BusinessID string `db:"business_id"`
	AppName    string `db:"app_name"`

	AccessToken    *string        `db:"access_token"`
	RefreshToken   *string        `db:"refresh_token"`
	TokenExpiresAt *time.Time     `db:"token_expires_at"`
	Scopes         pq.StringArray `db:"scopes"`

	PlatformAccountID   *string `db:"platform_account_id"`
	PlatformAccountName *string `db:"platform_account_name"`

	ConnectedAt *time.Time `db:"connected_at"`

	IsActive bool `db:"is_active"`
}
