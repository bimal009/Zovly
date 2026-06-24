package models

import "time"

type Category struct {
	ID         string `db:"id"          json:"id"`
	BusinessID string `db:"business_id" json:"business_id"`

	Name        string  `db:"name"        json:"name"`
	Description *string `db:"description" json:"description,omitempty"`
	Slug        *string `db:"slug"        json:"slug,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CreateCategoryInput struct {
	BusinessID string `json:"business_id" validate:"required,uuid"`

	Name        string  `json:"name"        validate:"required,min=1,max=255"`
	Description *string `json:"description"`
	Slug        *string `json:"slug"        validate:"omitempty,min=1,max=255"`
}

type UpdateCategoryInput struct {
	Name        *string `json:"name"        validate:"omitempty,min=1,max=255"`
	Description *string `json:"description"`
	Slug        *string `json:"slug"        validate:"omitempty,min=1,max=255"`
}
