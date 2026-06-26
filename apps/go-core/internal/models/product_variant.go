// internal/models/product_variant.go
package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// ProductVariants is a slice of variants that can be scanned directly from a
// JSON/JSONB column — lets the products query LEFT JOIN + jsonb_agg the
// variants in a single round trip instead of a second query.
type ProductVariants []ProductVariant

func (vs *ProductVariants) Scan(src any) error {
	if src == nil {
		*vs = nil
		return nil
	}

	var data []byte
	switch s := src.(type) {
	case []byte:
		data = s
	case string:
		data = []byte(s)
	default:
		return fmt.Errorf("ProductVariants.Scan: unsupported type %T", src)
	}

	if len(data) == 0 {
		*vs = nil
		return nil
	}
	return json.Unmarshal(data, (*[]ProductVariant)(vs))
}

// ─── Core model ──────────────────────────────────────────────────────────────

type ProductVariant struct {
	ID         string `db:"id"          json:"id"`
	ProductID  string `db:"product_id"  json:"product_id"`
	BusinessID string `db:"business_id" json:"business_id"`

	// Identity
	Name string  `db:"name" json:"name"` // "Red / Medium", "500ml", "Large"
	SKU  *string `db:"sku"  json:"sku,omitempty"`

	// structured option values — { "color": "red", "size": "M" }
	Attributes json.RawMessage `db:"attributes" json:"attributes,omitempty"`

	// Pricing (cents) — null means inherit the parent product's value
	Price     *int `db:"price"      json:"price,omitempty"`
	CostPrice *int `db:"cost_price" json:"cost_price,omitempty"` // never exposed to customer
	Discount  *int `db:"discount"   json:"discount,omitempty"`

	// Inventory
	StockQty          int  `db:"stock_qty"           json:"stock_qty"`
	LowStockThreshold *int `db:"low_stock_threshold" json:"low_stock_threshold,omitempty"`

	// Media — variant-specific images
	Images pq.StringArray `db:"images" json:"images"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ─── Create ───────────────────────────────────────────────────────────────────

type CreateProductVariantInput struct {
	ProductID  string `json:"product_id"  validate:"omitempty,uuid"`
	BusinessID string `json:"business_id" validate:"omitempty,uuid"`

	Name string  `json:"name" validate:"required,min=1,max=255"`
	SKU  *string `json:"sku"`

	Attributes json.RawMessage `json:"attributes"`

	Price     *int `json:"price"      validate:"omitempty,gt=0"`
	CostPrice *int `json:"cost_price" validate:"omitempty,gt=0"`
	Discount  *int `json:"discount"   validate:"omitempty,min=0,max=100"`

	StockQty          int  `json:"stock_qty"           validate:"min=0"`
	LowStockThreshold *int `json:"low_stock_threshold" validate:"omitempty,min=0"`

	Images []string `json:"images" validate:"omitempty,max=4,dive,url"`
}

// ─── Update ───────────────────────────────────────────────────────────────────
// All fields are pointers — only non-nil fields are written to DB

type UpdateProductVariantInput struct {
	Name *string `json:"name" validate:"omitempty,min=1,max=255"`
	SKU  *string `json:"sku"`

	Attributes json.RawMessage `json:"attributes"`

	Price     *int `json:"price"      validate:"omitempty,gt=0"`
	CostPrice *int `json:"cost_price" validate:"omitempty,gt=0"`
	Discount  *int `json:"discount"   validate:"omitempty,min=0,max=100"`

	StockQty          *int `json:"stock_qty"           validate:"omitempty,min=0"`
	LowStockThreshold *int `json:"low_stock_threshold" validate:"omitempty,min=0"`

	Images []string `json:"images" validate:"omitempty,max=4,dive,url"`
}
