// internal/service/product_service.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/bimal009/Zovly/internal/constants"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

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
	db          *sqlx.DB
	rdb         *redis.Client
	logger      *slog.Logger
	productRepo repository.ProductRepo
}

func NewProductService(
	db *sqlx.DB,
	rdb *redis.Client,
	logger *slog.Logger,
	productRepo repository.ProductRepo,
) ProductService {
	return &productService{
		db:          db,
		rdb:         rdb,
		logger:      logger,
		productRepo: productRepo,
	}
}

// ─── cache helpers ────────────────────────────────────────────────────────────

func productKey(id, businessID string) string {
	return fmt.Sprintf("%s%s:%s", constants.ProductsKeys, businessID, id)
}

func productListKey(businessID string) string {
	return fmt.Sprintf("%s%s:list", constants.ProductsKeys, businessID)
}

func lowStockKey(businessID string) string {
	return fmt.Sprintf("%s%s:low_stock", constants.ProductsKeys, businessID)
}

func (s *productService) invalidateBusinessCache(ctx context.Context, businessID string) {
	keys := []string{productListKey(businessID), lowStockKey(businessID)}
	if err := s.rdb.Del(ctx, keys...).Err(); err != nil {
		s.logger.Warn("product cache invalidate failed", "business_id", businessID, "error", err)
	}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (s *productService) Create(ctx context.Context, input models.CreateProductInput) (*models.Product, error) {
	s.logger.Info("product create", "business_id", input.BusinessID, "name", input.Name)

	product, err := s.productRepo.Create(ctx, input)
	if err != nil {
		s.logger.Error("product create failed", "business_id", input.BusinessID, "error", err)
		return nil, err
	}

	s.invalidateBusinessCache(ctx, input.BusinessID)

	s.logger.Info("product created", "id", product.ID, "business_id", product.BusinessID)
	return product, nil
}

// ─── GetByID ──────────────────────────────────────────────────────────────────

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

// ─── List ─────────────────────────────────────────────────────────────────────

func (s *productService) List(ctx context.Context, businessID string, f repository.ListProductsFilter) ([]models.Product, error) {
	cacheKey := productListKey(businessID)

	// only cache unfiltered first page
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

// ─── Update ───────────────────────────────────────────────────────────────────

func (s *productService) Update(ctx context.Context, id, businessID string, input models.UpdateProductInput) (*models.Product, error) {
	s.logger.Info("product update", "id", id, "business_id", businessID)

	product, err := s.productRepo.Update(ctx, id, businessID, input)
	if err != nil {
		s.logger.Error("product update failed", "id", id, "business_id", businessID, "error", err)
		return nil, err
	}
	if product == nil {
		return nil, nil
	}

	// bust single + list cache
	s.rdb.Del(ctx, productKey(id, businessID))
	s.invalidateBusinessCache(ctx, businessID)

	s.logger.Info("product updated", "id", id, "business_id", businessID)
	return product, nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

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

// ─── AdjustStock ─────────────────────────────────────────────────────────────

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
