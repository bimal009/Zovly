// internal/models/session.go
package models

import (
	"time"
)

type Session struct {
	ID        string    `db:"id"         json:"id"`
	UserID    string    `db:"user_id"    json:"user_id"`
	Token     string    `db:"token"      json:"-"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type SessionWithUser struct {
	ID        string    `db:"id"         json:"id"`
	UserID    string    `db:"user_id"    json:"user_id"`
	Token     string    `db:"token"      json:"-"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	UserName          string   `db:"user_name"           json:"name"`
	UserEmail         string   `db:"user_email"          json:"email"`
	UserRole          UserRole `db:"user_role"           json:"role"`
	UserOnboarded     bool     `db:"user_onboarded"      json:"onboarded"`
	UserEmailVerified bool     `db:"user_email_verified" json:"email_verified"`
	UserImage         *string  `db:"user_image"          json:"image,omitempty"`
}
