package models

import "time"

type UserRole string

const (
	RoleUser   UserRole = "user"
	RoleVendor UserRole = "vendor"
	RoleAdmin  UserRole = "admin"
)

type User struct {
	ID            string    `db:"id"             json:"id"`
	Name          string    `db:"name"           json:"name"`
	Email         string    `db:"email"          json:"email"`
	EmailVerified bool      `db:"email_verified" json:"email_verified"`
	Image         *string   `db:"image"          json:"image,omitempty"`
	Phone         *string   `db:"phone"          json:"phone,omitempty"`
	Role          UserRole  `db:"role"           json:"role"`
	BusinessID    *string   `db:"business_id"    json:"business_id,omitempty"`
	Onboarded     bool      `db:"is_onboarded"   json:"is_onboarded"`
	CreatedAt     time.Time `db:"created_at"     json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"     json:"updated_at"`
}

type UserUpdate struct {
	ID            string    `db:"id"`
	Name          *string   `db:"name"`
	Email         *string   `db:"email"`
	Image         *string   `db:"image"`
	Role          *UserRole `db:"role"`
	EmailVerified *bool     `db:"email_verified"`
	Onboarded     *bool     `db:"is_onboarded"`
}

type UserProfile struct {
	User
	AccountType string `json:"account_type"`
}
