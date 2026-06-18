package models

import "time"

type Conversation struct {
	ID            string     `db:"id" json:"id"`
	BusinessID    string     `db:"business_id" json:"business_id"`
	Platform      string     `db:"platform" json:"platform"`
	ThreadID      string     `db:"thread_id" json:"thread_id"`
	ContactID     string     `db:"contact_id" json:"contact_id"`
	LastMessageAt *time.Time `db:"last_message_at" json:"last_message_at"`
	AIEnabled     bool       `db:"ai_enabled" json:"ai_enabled"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
}

type CreateConversation struct {
	BusinessID string `db:"business_id"`
	Platform   string `db:"platform"`
	ThreadID   string `db:"thread_id"`
	ContactID  string `db:"contact_id"`
}
