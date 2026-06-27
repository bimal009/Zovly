package models

import "time"

type Conversation struct {
	ID         string `db:"id" json:"id"`
	BusinessID string `db:"business_id" json:"business_id"`
	Platform   string `db:"platform" json:"platform"`
	ThreadID   string `db:"thread_id" json:"thread_id"`
	ContactID  string `db:"contact_id" json:"contact_id"`

	ContactName      *string `db:"contact_name" json:"contact_name"`
	ContactUsername  *string `db:"contact_username" json:"contact_username"`
	ContactAvatarURL *string `db:"contact_avatar_url" json:"contact_avatar_url"`

	LastMessageAt *time.Time `db:"last_message_at" json:"last_message_at"`

	ActiveProductID *string    `db:"active_product_id" json:"active_product_id"`
	ActiveProductAt *time.Time `db:"active_product_at" json:"active_product_at"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type CreateConversation struct {
	BusinessID string `db:"business_id"`
	Platform   string `db:"platform"`
	ThreadID   string `db:"thread_id"`
	ContactID  string `db:"contact_id"`

	ContactName      *string `db:"contact_name"`
	ContactUsername  *string `db:"contact_username"`
	ContactAvatarURL *string `db:"contact_avatar_url"`
}
