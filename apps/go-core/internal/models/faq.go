package models

import "time"

type Faq struct {
	ID         string    `db:"id"          json:"id"`
	BusinessID string    `db:"business_id" json:"business_id"`
	Question   string    `db:"question"    json:"question"`
	Answer     string    `db:"answer"      json:"answer"`
	IsActive   bool      `db:"is_active"   json:"is_active"`
	SortOrder  int       `db:"sort_order"  json:"sort_order"`
	CreatedAt  time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"  json:"updated_at"`
}

type CreateFaqRequest struct {
	Question string `json:"question"   validate:"required"`
	Answer   string `json:"answer"     validate:"required"`
}
