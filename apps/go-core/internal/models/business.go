package models

import "time"

type BusinessType string

const (
	BusinessTypeProduct BusinessType = "product"
	BusinessTypeService BusinessType = "service"
	BusinessTypeBoth    BusinessType = "both"

	
)

type Business struct {
	ID          string       `db:"id"          json:"id"`
	Name        string       `db:"name"        json:"name"`
	Slug        string       `db:"slug"        json:"slug"`
	Description *string      `db:"description" json:"description,omitempty"`
	Logo        *string      `db:"logo"        json:"logo,omitempty"`
	Website     *string      `db:"website"     json:"website,omitempty"`
	Phone       *string      `db:"phone"       json:"phone,omitempty"`
	Address     *string      `db:"address"     json:"address,omitempty"`
	City        *string      `db:"city"        json:"city,omitempty"`
	Country     string       `db:"country"     json:"country"`
	Type        BusinessType `db:"type"        json:"type"`
	Plan        Plan         `db:"plan"        json:"plan"`
	CreatedAt   time.Time    `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"  json:"updated_at"`
}

// BusinessUpdate holds mutable fields for PATCH-style updates.
type BusinessUpdate struct {
	Name        *string
	Description *string
	Logo        *string
	Website     *string
	Phone       *string
	Address     *string
	City        *string
	Country     *string
	Type        *BusinessType
	Plan        *Plan
}