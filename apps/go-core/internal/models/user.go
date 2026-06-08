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
	Onboarded     bool      `db:"onboarded"      json:"onboarded"`
	CreatedAt     time.Time `db:"created_at"     json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"     json:"updated_at"`
}

type UserProfile struct {
	User
	AccountType string `json:"account_type"`
}
