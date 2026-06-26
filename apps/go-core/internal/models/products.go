// internal/models/product.go
package models

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
	ProductStatusArchived ProductStatus = "archived"
)

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

	Attributes json.RawMessage `db:"attributes" json:"attributes,omitempty"`

	Price     float64  `db:"price"      json:"price"`
	CostPrice *float64 `db:"cost_price" json:"cost_price,omitempty"` // never exposed to customer
	Discount  int      `db:"discount"   json:"discount"`             // percentage 0-100
	Currency  string   `db:"currency"   json:"currency"`

	StockQty          int  `db:"stock_qty"           json:"stock_qty"`
	LowStockThreshold *int `db:"low_stock_threshold" json:"low_stock_threshold,omitempty"`

	Images   pq.StringArray  `db:"images" json:"images"`
	Variants ProductVariants `db:"variants" json:"variants,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CreateProductInput struct {
	BusinessID string  `json:"business_id" validate:"required,uuid"`
	CategoryID *string `json:"category_id" validate:"required,uuid"`

	Name        string        `json:"name"        validate:"required,min=1,max=255"`
	Description *string       `json:"description" validate:"omitempty,max=200"`
	SKU         *string       `json:"sku"`
	Status      ProductStatus `json:"status"      validate:"omitempty,oneof=active inactive archived"`
	Tags        []string      `json:"tags"        validate:"omitempty,max=5,dive,min=1"`

	Attributes json.RawMessage `json:"attributes"`

	Price     float64  `json:"price"      validate:"required,gt=0"`
	CostPrice *float64 `json:"cost_price" validate:"omitempty,gt=0"`
	Discount  int      `json:"discount"   validate:"min=0,max=100"`
	Currency  string   `json:"currency"   validate:"omitempty,len=3"` // ISO 4217

	StockQty          int  `json:"stock_qty"           validate:"min=0"`
	LowStockThreshold *int `json:"low_stock_threshold" validate:"omitempty,min=0"`

	Images []string `json:"images" validate:"omitempty,max=4,dive,url"`

	Variants []CreateProductVariantInput `json:"variants" validate:"omitempty,dive"`
}

type UpdateProductInput struct {
	CategoryID  *string        `json:"category_id" validate:"omitempty,uuid"`
	Name        *string        `json:"name"        validate:"omitempty,min=1,max=255"`
	Description *string        `json:"description" validate:"omitempty,max=200"`
	SKU         *string        `json:"sku"`
	Status      *ProductStatus `json:"status"      validate:"omitempty,oneof=active inactive archived"`
	Tags        []string       `json:"tags"        validate:"omitempty,max=5,dive,min=1"`

	Attributes json.RawMessage `json:"attributes"`

	Price     *float64 `json:"price"      validate:"omitempty,gt=0"`
	CostPrice *float64 `json:"cost_price" validate:"omitempty,gt=0"`
	Discount  *int     `json:"discount"   validate:"omitempty,min=0,max=100"`
	Currency  *string  `json:"currency"   validate:"omitempty,len=3"`

	StockQty          *int `json:"stock_qty"           validate:"omitempty,min=0"`
	LowStockThreshold *int `json:"low_stock_threshold" validate:"omitempty,min=0"`

	Images []string `json:"images" validate:"omitempty,max=4,dive,url"`

	Variants []CreateProductVariantInput `json:"variants" validate:"omitempty,dive"`
}
