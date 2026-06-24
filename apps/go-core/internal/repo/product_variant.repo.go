package repository

import (
	"context"
	"fmt"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ProductVariantRepo interface {
	Create(ctx context.Context, tx *sqlx.Tx, input models.CreateProductVariantInput) (*models.ProductVariant, error)
}

type productVariantRepo struct {
	db *sqlx.DB
}

func NewProductVariantRepo(db *sqlx.DB) ProductVariantRepo {
	return &productVariantRepo{db: db}
}

func (r *productVariantRepo) Create(ctx context.Context, tx *sqlx.Tx, input models.CreateProductVariantInput) (*models.ProductVariant, error) {
	const q = `
		INSERT INTO product_variants (
			product_id, business_id, name, sku, attributes,
			price, discount, stock_qty, low_stock_threshold, images
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10
		)
		RETURNING *`

	var v models.ProductVariant
	err := tx.QueryRowxContext(ctx, q,
		input.ProductID,
		input.BusinessID,
		input.Name,
		input.SKU,
		input.Attributes,
		input.Price,
		input.Discount,
		input.StockQty,
		input.LowStockThreshold,
		pq.Array(orEmptySlice(input.Images)),
	).StructScan(&v)
	if err != nil {
		return nil, fmt.Errorf("product variant create: %w", err)
	}

	return &v, nil
}
