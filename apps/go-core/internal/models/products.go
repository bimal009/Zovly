// internal/models/product.go
package models

import (
	"time"

	"github.com/lib/pq"
)

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
	ProductStatusArchived ProductStatus = "archived"
)

// ─── Core model ──────────────────────────────────────────────────────────────

type Product struct {
	ID         string  `db:"id"          json:"id"`
	BusinessID string  `db:"business_id" json:"business_id"`
	CategoryID *string `db:"category_id" json:"category_id,omitempty"`

	// Core
	Name        string         `db:"name"        json:"name"`
	Description *string        `db:"description" json:"description,omitempty"`
	SKU         *string        `db:"sku"         json:"sku,omitempty"`
	Status      ProductStatus  `db:"status"      json:"status"`
	Tags        pq.StringArray `db:"tags"        json:"tags"`

	// Pricing (cents)
	Price     int    `db:"price"      json:"price"`
	CostPrice *int   `db:"cost_price" json:"cost_price,omitempty"` // never exposed to customer
	Discount  int    `db:"discount"   json:"discount"`             // percentage 0-100
	Currency  string `db:"currency"   json:"currency"`

	// Inventory
	StockQty          int  `db:"stock_qty"           json:"stock_qty"`
	LowStockThreshold *int `db:"low_stock_threshold" json:"low_stock_threshold,omitempty"`

	// Media
	Images   pq.StringArray   `db:"images" json:"images"`
	Variants []ProductVariant `db:"-" json:"variants,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ─── Create ───────────────────────────────────────────────────────────────────

type CreateProductInput struct {
	BusinessID string  `json:"business_id" validate:"required,uuid"`
	CategoryID *string `json:"category_id" validate:"omitempty,uuid"`

	Name        string        `json:"name"        validate:"required,min=1,max=255"`
	Description *string       `json:"description"`
	SKU         *string       `json:"sku"`
	Status      ProductStatus `json:"status"      validate:"omitempty,oneof=active inactive archived"`
	Tags        []string      `json:"tags"        validate:"omitempty,dive,min=1"`

	Price     int    `json:"price"      validate:"required,gt=0"`
	CostPrice *int   `json:"cost_price" validate:"omitempty,gt=0"`
	Discount  int    `json:"discount"   validate:"min=0,max=100"`
	Currency  string `json:"currency"   validate:"omitempty,len=3"` // ISO 4217

	StockQty          int  `json:"stock_qty"           validate:"min=0"`
	LowStockThreshold *int `json:"low_stock_threshold" validate:"omitempty,min=0"`

	Images []string `json:"images" validate:"omitempty,dive,url"`

	Variants []CreateProductVariantInput `json:"variants" validate:"omitempty,dive"`
}

// ─── Update ───────────────────────────────────────────────────────────────────
// All fields are pointers — only non-nil fields are written to DB

type UpdateProductInput struct {
	CategoryID  *string        `json:"category_id" validate:"omitempty,uuid"`
	Name        *string        `json:"name"        validate:"omitempty,min=1,max=255"`
	Description *string        `json:"description"`
	SKU         *string        `json:"sku"`
	Status      *ProductStatus `json:"status"      validate:"omitempty,oneof=active inactive archived"`
	Tags        []string       `json:"tags"        validate:"omitempty,dive,min=1"`

	Price     *int    `json:"price"      validate:"omitempty,gt=0"`
	CostPrice *int    `json:"cost_price" validate:"omitempty,gt=0"`
	Discount  *int    `json:"discount"   validate:"omitempty,min=0,max=100"`
	Currency  *string `json:"currency"   validate:"omitempty,len=3"`

	StockQty          *int `json:"stock_qty"           validate:"omitempty,min=0"`
	LowStockThreshold *int `json:"low_stock_threshold" validate:"omitempty,min=0"`

	Images []string `json:"images" validate:"omitempty,dive,url"`
}
