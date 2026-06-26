// internal/service/product_service.go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bimal009/Zovly/internal/constants"
	"github.com/bimal009/Zovly/internal/embed"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// ErrProductNotFound is returned when a product does not exist for the business.
// Handlers map it to a 404 "product not found" response.
var ErrProductNotFound = errors.New("product not found")

type ProductService interface {
	Create(ctx context.Context, input models.CreateProductInput) (*models.Product, error)
	GetByID(ctx context.Context, id, businessID string) (*models.Product, error)
	List(ctx context.Context, businessID string, f repository.ListProductsFilter) ([]models.Product, error)
	Update(ctx context.Context, id, businessID string, input models.UpdateProductInput) (*models.Product, error)
	Delete(ctx context.Context, id, businessID string) error
	AdjustStock(ctx context.Context, tx sqlx.ExtContext, id, businessID string, delta int) (*models.Product, error)
	LowStock(ctx context.Context, businessID string) ([]models.Product, error)
}

type productService struct {
	db                 *sqlx.DB
	rdb                *redis.Client
	logger             *slog.Logger
	productRepo        repository.ProductRepo
	productVariantRepo repository.ProductVariantRepo
	knowledgeRepo      repository.BusinessKnowledgeRepo
	embedder           *embed.Client
}

func NewProductService(
	db *sqlx.DB,
	rdb *redis.Client,
	logger *slog.Logger,
	productRepo repository.ProductRepo,
	productVariantRepo repository.ProductVariantRepo,
	knowledgeRepo repository.BusinessKnowledgeRepo,
	embedder *embed.Client,
) ProductService {
	return &productService{
		db:                 db,
		rdb:                rdb,
		logger:             logger,
		productRepo:        productRepo,
		productVariantRepo: productVariantRepo,
		knowledgeRepo:      knowledgeRepo,
		embedder:           embedder,
	}
}

func productKey(id, businessID string) string {
	return fmt.Sprintf("%s%s:%s", constants.ProductsKeys, businessID, id)
}

func productListKey(businessID string) string {
	return fmt.Sprintf("%s%s:list", constants.ProductsKeys, businessID)
}

func lowStockKey(businessID string) string {
	return fmt.Sprintf("%s%s:low_stock", constants.ProductsKeys, businessID)
}

func buildProductKnowledgePassage(name string, description *string, tags []string, attributes json.RawMessage, variantNames []string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Product: %s.", name)
	if description != nil && *description != "" {
		fmt.Fprintf(&b, " %s", strings.TrimRight(*description, "."))
		b.WriteString(".")
	}
	if len(tags) > 0 {
		fmt.Fprintf(&b, " Tags: %s.", strings.Join(tags, ", "))
	}
	if a := formatAttributes(attributes); a != "" {
		fmt.Fprintf(&b, " Attributes: %s.", a)
	}
	if len(variantNames) > 0 {
		fmt.Fprintf(&b, " Variants: %s.", strings.Join(variantNames, ", "))
	}

	return strings.TrimSpace(b.String())
}

func variantDisplayName(name string, attrs json.RawMessage) string {
	if a := formatAttributes(attrs); a != "" {
		return fmt.Sprintf("%s (%s)", name, a)
	}
	return name
}
func formatAttributes(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	parts := make([]string, 0, len(m))
	for k, val := range m {
		parts = append(parts, fmt.Sprintf("%s: %v", k, val))
	}
	return strings.Join(parts, ", ")
}

func (s *productService) invalidateBusinessCache(ctx context.Context, businessID string) {
	keys := []string{productListKey(businessID), lowStockKey(businessID)}
	if err := s.rdb.Del(ctx, keys...).Err(); err != nil {
		s.logger.Warn("product cache invalidate failed", "business_id", businessID, "error", err)
	}
}

func (s *productService) Create(ctx context.Context, input models.CreateProductInput) (*models.Product, error) {
	s.logger.Info("product create", "business_id", input.BusinessID, "name", input.Name)

	variantNames := make([]string, 0, len(input.Variants))
	for i := range input.Variants {
		variantNames = append(variantNames, variantDisplayName(input.Variants[i].Name, input.Variants[i].Attributes))
	}
	knowledgePassage := buildProductKnowledgePassage(input.Name, input.Description, input.Tags, input.Attributes, variantNames)

	embedChunks, err := s.embedder.EmbedFaq(ctx, knowledgePassage, "passage")
	if err != nil {
		s.logger.Error("product embed failed", "name", input.Name, "error", err)
		return nil, fmt.Errorf("embed product: %w", err)
	}
	s.logger.Info("product knowledge passage embedded", "name", input.Name, "chunks", len(embedChunks))

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	product, err := s.productRepo.Create(ctx, tx, input)
	if err != nil {
		s.logger.Error("product create failed", "business_id", input.BusinessID, "error", err)
		return nil, err
	}

	for i := range input.Variants {
		input.Variants[i].ProductID = product.ID
		input.Variants[i].BusinessID = product.BusinessID

		variant, err := s.productVariantRepo.Create(ctx, tx, input.Variants[i])
		if err != nil {
			s.logger.Error("product variant create failed",
				"product_id", product.ID,
				"name", input.Variants[i].Name,
				"error", err,
			)
			return nil, err
		}
		product.Variants = append(product.Variants, *variant)
	}

	metadata, err := json.Marshal(map[string]string{"name": product.Name})
	if err != nil {
		return nil, fmt.Errorf("marshal chunk metadata: %w", err)
	}

	chunks := models.ToChunkInserts(embedChunks, product.BusinessID, product.ID, models.SourceProduct, metadata)
	if err := s.knowledgeRepo.Create(ctx, tx, chunks); err != nil {
		s.logger.Error("product knowledge chunk create failed", "product_id", product.ID, "error", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	s.invalidateBusinessCache(ctx, input.BusinessID)

	s.logger.Info("product created", "id", product.ID, "business_id", product.BusinessID)
	return product, nil
}

func (s *productService) GetByID(ctx context.Context, id, businessID string) (*models.Product, error) {
	cacheKey := productKey(id, businessID)

	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var product models.Product
		if err := json.Unmarshal([]byte(cached), &product); err == nil {
			return &product, nil
		}
	}
	if err != nil && err != redis.Nil {
		s.logger.Warn("product cache get failed", "id", id, "error", err)
	}

	product, err := s.productRepo.GetByID(ctx, id, businessID)
	if err != nil {
		s.logger.Error("product get failed", "id", id, "business_id", businessID, "error", err)
		return nil, err
	}
	if product == nil {
		return nil, nil
	}

	if data, err := json.Marshal(product); err == nil {
		if err := s.rdb.Set(ctx, cacheKey, data, constants.TTLMedium).Err(); err != nil {
			s.logger.Warn("product cache set failed", "id", id, "error", err)
		}
	}

	return product, nil
}

func (s *productService) List(ctx context.Context, businessID string, f repository.ListProductsFilter) ([]models.Product, error) {
	cacheKey := productListKey(businessID)

	useCache := f.Status == nil && f.Offset == 0

	if useCache {
		cached, err := s.rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var products []models.Product
			if err := json.Unmarshal([]byte(cached), &products); err == nil {
				return products, nil
			}
		}
		if err != nil && err != redis.Nil {
			s.logger.Warn("product list cache get failed", "business_id", businessID, "error", err)
		}
	}

	products, err := s.productRepo.List(ctx, businessID, f)
	if err != nil {
		s.logger.Error("product list failed", "business_id", businessID, "error", err)
		return nil, err
	}

	if useCache {
		if data, err := json.Marshal(products); err == nil {
			if err := s.rdb.Set(ctx, cacheKey, data, constants.TTLShort).Err(); err != nil {
				s.logger.Warn("product list cache set failed", "business_id", businessID, "error", err)
			}
		}
	}

	return products, nil
}

func (s *productService) Update(ctx context.Context, id, businessID string, input models.UpdateProductInput) (*models.Product, error) {
	s.logger.Info("product update", "id", id, "business_id", businessID)

	// load current state — needed to 404 early and to build the new passage
	// from the *merged* (current + input) values, since input is partial.
	current, err := s.productRepo.GetByID(ctx, id, businessID)
	if err != nil {
		s.logger.Error("product update load failed", "id", id, "business_id", businessID, "error", err)
		return nil, err
	}
	if current == nil {
		return nil, ErrProductNotFound
	}

	name := current.Name
	if input.Name != nil {
		name = *input.Name
	}
	description := current.Description
	if input.Description != nil {
		description = input.Description
	}
	tags := []string(current.Tags)
	if input.Tags != nil {
		tags = input.Tags
	}
	attributes := current.Attributes
	if input.Attributes != nil {
		attributes = input.Attributes
	}
	variantNames := make([]string, 0, len(input.Variants))
	for i := range input.Variants {
		variantNames = append(variantNames, variantDisplayName(input.Variants[i].Name, input.Variants[i].Attributes))
	}

	knowledgePassage := buildProductKnowledgePassage(name, description, tags, attributes, variantNames)

	embedChunks, err := s.embedder.EmbedFaq(ctx, knowledgePassage, "passage")
	if err != nil {
		s.logger.Error("product embed failed", "id", id, "error", err)
		return nil, fmt.Errorf("embed product: %w", err)
	}

	metadata, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		return nil, fmt.Errorf("marshal chunk metadata: %w", err)
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. product row
	if err := s.productRepo.Update(ctx, tx, id, businessID, input); err != nil {
		s.logger.Error("product update failed", "id", id, "business_id", businessID, "error", err)
		return nil, err
	}

	// 2. replace variants wholesale
	if err := s.productVariantRepo.DeleteByProduct(ctx, tx, id, businessID); err != nil {
		s.logger.Error("product variant delete failed", "id", id, "error", err)
		return nil, err
	}
	for i := range input.Variants {
		input.Variants[i].ProductID = id
		input.Variants[i].BusinessID = businessID
		if _, err := s.productVariantRepo.Create(ctx, tx, input.Variants[i]); err != nil {
			s.logger.Error("product variant recreate failed", "product_id", id, "name", input.Variants[i].Name, "error", err)
			return nil, err
		}
	}

	if err := s.knowledgeRepo.DeleteBySource(ctx, tx, businessID, id, models.SourceProduct); err != nil {
		s.logger.Error("product knowledge chunk delete failed", "id", id, "error", err)
		return nil, err
	}
	chunks := models.ToChunkInserts(embedChunks, businessID, id, models.SourceProduct, metadata)
	if err := s.knowledgeRepo.Create(ctx, tx, chunks); err != nil {
		s.logger.Error("product knowledge chunk recreate failed", "id", id, "error", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	s.rdb.Del(ctx, productKey(id, businessID))
	s.invalidateBusinessCache(ctx, businessID)

	product, err := s.productRepo.GetByID(ctx, id, businessID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("product updated", "id", id, "business_id", businessID)
	return product, nil
}

func (s *productService) Delete(ctx context.Context, id, businessID string) error {
	s.logger.Info("product delete", "id", id, "business_id", businessID)

	if err := s.productRepo.Delete(ctx, id, businessID); err != nil {
		s.logger.Error("product delete failed", "id", id, "business_id", businessID, "error", err)
		return err
	}

	s.rdb.Del(ctx, productKey(id, businessID))
	s.invalidateBusinessCache(ctx, businessID)

	s.logger.Info("product deleted", "id", id, "business_id", businessID)
	return nil
}

func (s *productService) AdjustStock(ctx context.Context, tx sqlx.ExtContext, id, businessID string, delta int) (*models.Product, error) {
	s.logger.Info("product adjust stock", "id", id, "business_id", businessID, "delta", delta)

	product, err := s.productRepo.AdjustStock(ctx, tx, id, businessID, delta)
	if err != nil {
		s.logger.Error("product adjust stock failed", "id", id, "delta", delta, "error", err)
		return nil, err
	}

	// bust cache — stock is live data, don't serve stale
	s.rdb.Del(ctx, productKey(id, businessID))
	s.invalidateBusinessCache(ctx, businessID)

	s.logger.Info("product stock adjusted", "id", id, "new_qty", product.StockQty)
	return product, nil
}

// ─── LowStock ─────────────────────────────────────────────────────────────────

func (s *productService) LowStock(ctx context.Context, businessID string) ([]models.Product, error) {
	cacheKey := lowStockKey(businessID)

	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var products []models.Product
		if err := json.Unmarshal([]byte(cached), &products); err == nil {
			return products, nil
		}
	}
	if err != nil && err != redis.Nil {
		s.logger.Warn("low stock cache get failed", "business_id", businessID, "error", err)
	}

	products, err := s.productRepo.LowStock(ctx, businessID)
	if err != nil {
		s.logger.Error("low stock query failed", "business_id", businessID, "error", err)
		return nil, err
	}

	if data, err := json.Marshal(products); err == nil {
		if err := s.rdb.Set(ctx, cacheKey, data, constants.TTLShort).Err(); err != nil {
			s.logger.Warn("low stock cache set failed", "business_id", businessID, "error", err)
		}
	}

	return products, nil
}
