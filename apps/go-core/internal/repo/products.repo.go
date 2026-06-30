package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

type ProductRepo interface {
	Create(ctx context.Context, tx *sqlx.Tx, input models.CreateProductInput) (*models.Product, error)
	GetByID(ctx context.Context, id, businessID string) (*models.Product, error)
	GetByCategorySlug(ctx context.Context, slug, businessID string, limit, offset int) ([]models.Product, error)
	List(ctx context.Context, businessID string, f ListProductsFilter) (ListProductsResult, error)
	Count(ctx context.Context, businessID, categorySlug string) (int, error)
	Update(ctx context.Context, tx *sqlx.Tx, id, businessID string, input models.UpdateProductInput) error
	Delete(ctx context.Context, id, businessID string) error
	AdjustStock(ctx context.Context, tx sqlx.ExtContext, id, businessID string, delta int) (*models.Product, error)
	LowStock(ctx context.Context, businessID string) ([]models.Product, error)
	HybridSearch(ctx context.Context, businessID, query string, vec pgvector.Vector) ([]models.ProductSearchCandidate, error)
}

type productRepo struct {
	db *sqlx.DB
}

func NewProductRepo(db *sqlx.DB) ProductRepo {
	return &productRepo{db: db}
}

func (r *productRepo) Create(ctx context.Context, tx *sqlx.Tx, input models.CreateProductInput) (*models.Product, error) {
	const q = `
		INSERT INTO products (
			business_id, category_id, name, description, sku, status, tags, attributes,
			price, cost_price, discount, currency,
			stock_qty, low_stock_threshold,
			images
		) VALUES (
			:business_id, :category_id, :name, :description, :sku, :status, :tags, :attributes,
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
		"category_id":         input.CategoryID,
		"name":                input.Name,
		"description":         input.Description,
		"sku":                 input.SKU,
		"status":              status,
		"tags":                pq.Array(orEmptySlice(input.Tags)),
		"attributes":          orNullJSON(input.Attributes),
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
		SELECT p.*, COALESCE(jsonb_agg(
			(to_jsonb(v) - 'created_at' - 'updated_at') || jsonb_build_object(
				'created_at', to_char(v.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
				'updated_at', to_char(v.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
			) ORDER BY v.created_at
		) FILTER (WHERE v.id IS NOT NULL), '[]'::jsonb) AS variants
		FROM products p
		LEFT JOIN product_variants v ON v.product_id = p.id
		WHERE p.id = $1 AND p.business_id = $2
		GROUP BY p.id
	`, id, businessID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("product get: %w", err)
	}
	return &p, nil
}

func (r *productRepo) GetByCategorySlug(ctx context.Context, slug, businessID string, limit, offset int) ([]models.Product, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var products []models.Product
	err := r.db.SelectContext(ctx, &products, `
		SELECT p.*, COALESCE(jsonb_agg(
			(to_jsonb(v) - 'created_at' - 'updated_at') || jsonb_build_object(
				'created_at', to_char(v.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
				'updated_at', to_char(v.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
			) ORDER BY v.created_at
		) FILTER (WHERE v.id IS NOT NULL), '[]'::jsonb) AS variants
		FROM products p
		JOIN categories c ON c.id = p.category_id
		LEFT JOIN product_variants v ON v.product_id = p.id
		WHERE c.slug = $1 AND p.business_id = $2
		GROUP BY p.id
		ORDER BY p.created_at DESC
		LIMIT $3 OFFSET $4
	`, slug, businessID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("product get by category slug: %w", err)
	}
	return products, nil
}

type ListProductsFilter struct {
	Status       *models.ProductStatus
	Search       string
	CategorySlug string
	Limit        int
	Offset       int
	page         int
}

// Hybrid product search: fuses three signals over Reciprocal Rank Fusion.
const (
	hybridListLimit = 30 // candidates pulled per signal before fusion
	hybridReturn    = 20 // fused candidates returned to the reranker
	rrfK            = 60 // RRF damping constant
)

func (r *productRepo) HybridSearch(ctx context.Context, businessID, query string, vec pgvector.Vector) ([]models.ProductSearchCandidate, error) {
	if strings.TrimSpace(query) == "" {
		return []models.ProductSearchCandidate{}, nil
	}

	semantic, err := r.rankSemantic(ctx, businessID, vec)
	if err != nil {
		return nil, fmt.Errorf("hybrid semantic: %w", err)
	}
	fts, err := r.rankFullText(ctx, businessID, query)
	if err != nil {
		return nil, fmt.Errorf("hybrid fulltext: %w", err)
	}
	trgm, err := r.rankTrigram(ctx, businessID, query)
	if err != nil {
		return nil, fmt.Errorf("hybrid trigram: %w", err)
	}

	scores := make(map[string]float64)
	fuse := func(ids []string) {
		for i, id := range ids {
			scores[id] += 1.0 / float64(rrfK+i+1)
		}
	}
	fuse(semantic)
	fuse(fts)
	fuse(trgm)

	if len(scores) == 0 {
		return []models.ProductSearchCandidate{}, nil
	}

	ids := make([]string, 0, len(scores))
	for id := range scores {
		ids = append(ids, id)
	}
	sort.SliceStable(ids, func(i, j int) bool { return scores[ids[i]] > scores[ids[j]] })
	if len(ids) > hybridReturn {
		ids = ids[:hybridReturn]
	}

	candidates, err := r.loadCandidates(ctx, businessID, ids)
	if err != nil {
		return nil, fmt.Errorf("hybrid load: %w", err)
	}

	// Order by fused score and attach it.
	byID := make(map[string]models.ProductSearchCandidate, len(candidates))
	for _, c := range candidates {
		byID[c.SourceID] = c
	}
	out := make([]models.ProductSearchCandidate, 0, len(ids))
	for _, id := range ids {
		if c, ok := byID[id]; ok {
			c.Score = scores[id]
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *productRepo) rankSemantic(ctx context.Context, businessID string, vec pgvector.Vector) ([]string, error) {
	const q = `
		SELECT id FROM (
			SELECT DISTINCT ON (p.id) p.id, kc.embedding <=> $1 AS dist
			FROM products p
			JOIN knowledge_chunks kc ON kc.source_id = p.id AND kc.source_type = 'product'
			WHERE p.business_id = $2 AND p.status = 'active'
			ORDER BY p.id, dist
		) s
		ORDER BY s.dist
		LIMIT $3`
	var ids []string
	if err := r.db.SelectContext(ctx, &ids, q, vec, businessID, hybridListLimit); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *productRepo) rankFullText(ctx context.Context, businessID, query string) ([]string, error) {
	const q = `
		SELECT p.id
		FROM products p, websearch_to_tsquery('simple', $1) AS qq
		WHERE p.business_id = $2 AND p.status = 'active' AND p.search_tsv @@ qq
		ORDER BY ts_rank(p.search_tsv, qq) DESC
		LIMIT $3`
	var ids []string
	if err := r.db.SelectContext(ctx, &ids, q, query, businessID, hybridListLimit); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *productRepo) rankTrigram(ctx context.Context, businessID, query string) ([]string, error) {
	const q = `
		SELECT p.id
		FROM products p
		WHERE p.business_id = $2 AND p.status = 'active' AND p.name % $1
		ORDER BY similarity(p.name, $1) DESC
		LIMIT $3`
	var ids []string
	if err := r.db.SelectContext(ctx, &ids, q, query, businessID, hybridListLimit); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *productRepo) loadCandidates(ctx context.Context, businessID string, ids []string) ([]models.ProductSearchCandidate, error) {
	if len(ids) == 0 {
		return []models.ProductSearchCandidate{}, nil
	}
	const q = `
		SELECT
			p.id,
			p.name,
			COALESCE(
				(SELECT kc.content FROM knowledge_chunks kc
				 WHERE kc.source_id = p.id AND kc.source_type = 'product'
				 ORDER BY kc.chunk_index LIMIT 1),
				p.name || '. ' || COALESCE(p.description, '')
			) AS passage
		FROM products p
		WHERE p.business_id = $1 AND p.id = ANY($2)`
	var out []models.ProductSearchCandidate
	if err := r.db.SelectContext(ctx, &out, q, businessID, pq.Array(ids)); err != nil {
		return nil, err
	}
	return out, nil
}

type ListProductsResult struct {
	Products []models.Product
	Total    int
}

func (r *productRepo) List(ctx context.Context, businessID string, f ListProductsFilter) (ListProductsResult, error) {
	args := []any{businessID}
	conds := []string{"p.business_id = $1", "p.status != 'archived'"}
	join := ""
	i := 2

	if f.Status != nil {
		conds = append(conds, fmt.Sprintf("p.status = $%d", i))
		args = append(args, string(*f.Status))
		i++
	}
	if f.Search != "" {
		pattern := "%" + strings.ToLower(f.Search) + "%"
		conds = append(conds, fmt.Sprintf(
			`(p.name ILIKE $%d OR p.sku ILIKE $%d OR p.description ILIKE $%d OR EXISTS (
				SELECT 1 FROM unnest(p.tags) AS t(tag) WHERE t.tag ILIKE $%d
			))`, i, i, i, i,
		))
		args = append(args, pattern)
		i++
	}
	if f.CategorySlug != "" {
		join = "JOIN categories c ON c.id = p.category_id"
		conds = append(conds, fmt.Sprintf("c.slug = $%d", i))
		args = append(args, f.CategorySlug)
		i++
	}

	limit := 20
	if f.Limit > 0 && f.Limit <= 100 {
		limit = f.Limit
	}

	q := fmt.Sprintf(`
		SELECT
			p.*,
			COUNT(*) OVER() AS total_count
		FROM products p
		%s
		WHERE %s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d`,
		join, strings.Join(conds, " AND "), i, i+1,
	)
	args = append(args, limit, f.Offset)

	var rows []struct {
		models.Product
		TotalCount int `db:"total_count"`
	}
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return ListProductsResult{}, fmt.Errorf("product list: %w", err)
	}

	result := ListProductsResult{}
	if len(rows) > 0 {
		result.Total = rows[0].TotalCount
	}
	result.Products = make([]models.Product, len(rows))
	for i, row := range rows {
		result.Products[i] = row.Product
	}
	return result, nil
}

func (r *productRepo) Count(ctx context.Context, businessID, categorySlug string) (int, error) {
	args := []any{businessID}
	join := ""
	conds := "p.business_id = $1"

	if categorySlug != "" {
		join = "JOIN categories c ON c.id = p.category_id"
		conds += " AND c.slug = $2"
		args = append(args, categorySlug)
	}

	var count int
	q := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM products p
		%s
		WHERE %s
	`, join, conds)
	if err := r.db.GetContext(ctx, &count, q, args...); err != nil {
		return 0, fmt.Errorf("product count: %w", err)
	}
	return count, nil
}

func (r *productRepo) Update(ctx context.Context, tx *sqlx.Tx, id, businessID string, input models.UpdateProductInput) error {
	fields := map[string]any{}

	if input.Name != nil {
		fields["name"] = *input.Name
	}

	if input.CategoryID != nil && *input.CategoryID != "" {
		fields["category_id"] = *input.CategoryID
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
	if input.Tags != nil {
		fields["tags"] = pq.Array(orEmptySlice(input.Tags))
	}
	if input.Attributes != nil {
		fields["attributes"] = orNullJSON(input.Attributes)
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

	// nothing to change on the products row itself (e.g. only variants changed)
	if len(fields) == 0 {
		return nil
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

	if _, err := tx.NamedExecContext(ctx, q, fields); err != nil {
		return fmt.Errorf("product update: %w", err)
	}
	return nil
}

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
