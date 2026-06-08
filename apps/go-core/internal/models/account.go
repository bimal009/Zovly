// internal/models/account.go
package models

import "time"

type AccountType string
type TokenType string

const (
	AccountTypeCredential AccountType = "credential"
	AccountTypeGoogle     AccountType = "google"
	TokenTypeBearer       TokenType   = "Bearer"
)

type Account struct {
	ID                string      `db:"id"                  json:"id"`
	UserID            string      `db:"user_id"             json:"user_id"`
	Type              AccountType `db:"type"                json:"type"`
	Provider          string      `db:"provider"            json:"provider"`
	ProviderAccountID string      `db:"provider_account_id" json:"provider_account_id"`
	PasswordHash      *string     `db:"password_hash"       json:"-"`
	AccessToken       *string     `db:"access_token"        json:"-"`
	RefreshToken      *string     `db:"refresh_token"       json:"-"`
	ExpiresAt         *int64      `db:"expires_at"          json:"-"`
	TokenType         *TokenType  `db:"token_type"          json:"-"`
	Scope             *string     `db:"scope"               json:"-"`
	IDToken           *string     `db:"id_token"            json:"-"`
	CreatedAt         time.Time   `db:"created_at"          json:"created_at"`
	UpdatedAt         time.Time   `db:"updated_at"          json:"updated_at"`
}
type AccountTokenUpdate struct {
	AccessToken  *string
	RefreshToken *string
	ExpiresAt    *int64 // matches db:"expires_at" int64 on Account
}
