package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ProductRepo interface {
	Create(ctx context.Context, tx *sqlx.Tx, input models.CreateProductInput) (*models.Product, error)
	GetByID(ctx context.Context, id, businessID string) (*models.Product, error)
	List(ctx context.Context, businessID string, f ListProductsFilter) ([]models.Product, error)
	Update(ctx context.Context, id, businessID string, input models.UpdateProductInput) (*models.Product, error)
	Delete(ctx context.Context, id, businessID string) error
	AdjustStock(ctx context.Context, tx sqlx.ExtContext, id, businessID string, delta int) (*models.Product, error)
	LowStock(ctx context.Context, businessID string) ([]models.Product, error)
}

type productRepo struct {
	db *sqlx.DB
}

func NewProductRepo(db *sqlx.DB) ProductRepo {
	return &productRepo{db: db}
}

// productVariantsAgg aggregates each product's variants into a single JSON
// column, so the product + its variants come back in one query (no N+1, no
// app-side grouping loop). Timestamps are reformatted to RFC3339 so they
// unmarshal cleanly into time.Time. Expects the products table aliased as "p".
const productVariantsAgg = `
		COALESCE(
			(
				SELECT jsonb_agg(
					(to_jsonb(v) - 'created_at' - 'updated_at') || jsonb_build_object(
						'created_at', to_char(v.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
						'updated_at', to_char(v.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
					)
					ORDER BY v.created_at ASC
				)
				FROM product_variants v
				WHERE v.product_id = p.id
			),
			'[]'::jsonb
		) AS variants`

// ─── Create ───────────────────────────────────────────────────────────────────

func (r *productRepo) Create(ctx context.Context, tx *sqlx.Tx, input models.CreateProductInput) (*models.Product, error) {
	const q = `
		INSERT INTO products (
			business_id, name, description, sku, status,
			price, cost_price, discount, currency,
			stock_qty, low_stock_threshold,
			images
		) VALUES (
			:business_id, :name, :description, :sku, :status,
			:price, :cost_price, :discount, :currency,
			:stock_qty, :low_stock_threshold,
			:images
		)
		RETURNING *`

	status := input.Status
	if status == "" {
		status = models.ProductStatusActive
	}
	currency := input.Currency
	if currency == "" {
		currency = "NPR"
	}

	rows, err := sqlx.NamedQueryContext(ctx, tx, q, map[string]any{
		"business_id":         input.BusinessID,
		"name":                input.Name,
		"description":         input.Description,
		"sku":                 input.SKU,
		"status":              status,
		"price":               input.Price,
		"cost_price":          input.CostPrice,
		"discount":            input.Discount,
		"currency":            currency,
		"stock_qty":           input.StockQty,
		"low_stock_threshold": input.LowStockThreshold,
		"images":              pq.Array(orEmptySlice(input.Images)),
	})
	if err != nil {
		return nil, fmt.Errorf("product create: %w", err)
	}
	defer rows.Close()

	var p models.Product
	if rows.Next() {
		if err := rows.StructScan(&p); err != nil {
			return nil, fmt.Errorf("product create scan: %w", err)
		}
	}

	return &p, nil
}

func (r *productRepo) GetByID(ctx context.Context, id, businessID string) (*models.Product, error) {
	var p models.Product
	err := r.db.GetContext(ctx, &p, `
		SELECT p.*, `+productVariantsAgg+`
		FROM products p
		WHERE p.id = $1 AND p.business_id = $2
	`, id, businessID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("product get: %w", err)
	}
	return &p, nil
}

// ─── List ─────────────────────────────────────────────────────────────────────

type ListProductsFilter struct {
	Status *models.ProductStatus
	Search string
	Limit  int
	Offset int
}

func (r *productRepo) List(ctx context.Context, businessID string, f ListProductsFilter) ([]models.Product, error) {
	args := []any{businessID}
	conds := []string{"business_id = $1"}
	i := 2

	if f.Status != nil {
		conds = append(conds, fmt.Sprintf("status = $%d", i))
		args = append(args, string(*f.Status))
		i++
	}
	if f.Search != "" {
		conds = append(conds, fmt.Sprintf("(name ILIKE $%d OR sku ILIKE $%d)", i, i))
		args = append(args, "%"+f.Search+"%")
		i++
	}

	limit := 50
	if f.Limit > 0 && f.Limit <= 100 {
		limit = f.Limit
	}

	q := fmt.Sprintf(`
		SELECT p.*, %s
		FROM products p
		WHERE %s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d`,
		productVariantsAgg, strings.Join(conds, " AND "), i, i+1,
	)
	args = append(args, limit, f.Offset)

	var products []models.Product
	if err := r.db.SelectContext(ctx, &products, q, args...); err != nil {
		return nil, fmt.Errorf("product list: %w", err)
	}
	return products, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (r *productRepo) Update(ctx context.Context, id, businessID string, input models.UpdateProductInput) (*models.Product, error) {
	fields := map[string]any{}

	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.SKU != nil {
		fields["sku"] = *input.SKU
	}
	if input.Status != nil {
		fields["status"] = string(*input.Status)
	}
	if input.Price != nil {
		fields["price"] = *input.Price
	}
	if input.CostPrice != nil {
		fields["cost_price"] = *input.CostPrice
	}
	if input.Discount != nil {
		fields["discount"] = *input.Discount
	}
	if input.Currency != nil {
		fields["currency"] = *input.Currency
	}
	if input.StockQty != nil {
		fields["stock_qty"] = *input.StockQty
	}
	if input.LowStockThreshold != nil {
		fields["low_stock_threshold"] = *input.LowStockThreshold
	}
	if input.Images != nil {
		fields["images"] = pq.Array(orEmptySlice(input.Images))
	}

	if len(fields) == 0 {
		return r.GetByID(ctx, id, businessID)
	}

	fields["updated_at"] = time.Now()
	fields["id"] = id
	fields["business_id"] = businessID

	setClauses := make([]string, 0, len(fields))
	for col := range fields {
		if col == "id" || col == "business_id" {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", col, col))
	}

	q := fmt.Sprintf(`
		UPDATE products
		SET %s
		WHERE id = :id AND business_id = :business_id`,
		strings.Join(setClauses, ", "),
	)

	if _, err := r.db.NamedExecContext(ctx, q, fields); err != nil {
		return nil, fmt.Errorf("product update: %w", err)
	}

	return r.GetByID(ctx, id, businessID)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (r *productRepo) Delete(ctx context.Context, id, businessID string) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM products WHERE id = $1 AND business_id = $2
	`, id, businessID)
	if err != nil {
		return fmt.Errorf("product delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("product delete: not found or not owned")
	}
	return nil
}

// ─── AdjustStock ─────────────────────────────────────────────────────────────

func (r *productRepo) AdjustStock(ctx context.Context, tx sqlx.ExtContext, id, businessID string, delta int) (*models.Product, error) {
	_, err := tx.ExecContext(ctx, `
		UPDATE products
		SET stock_qty  = stock_qty + $1,
		    updated_at = now()
		WHERE id = $2
		  AND business_id = $3
		  AND stock_qty + $1 >= 0
	`, delta, id, businessID)
	if err != nil {
		return nil, fmt.Errorf("product adjust stock: %w", err)
	}

	return r.GetByID(ctx, id, businessID)
}

// ─── LowStock ─────────────────────────────────────────────────────────────────

func (r *productRepo) LowStock(ctx context.Context, businessID string) ([]models.Product, error) {
	var products []models.Product
	err := r.db.SelectContext(ctx, &products, `
		SELECT * FROM products
		WHERE business_id = $1
		  AND status = 'active'
		  AND stock_qty <= COALESCE(low_stock_threshold, 5)
		ORDER BY stock_qty ASC
	`, businessID)
	if err != nil {
		return nil, fmt.Errorf("product low stock: %w", err)
	}
	return products, nil
}
