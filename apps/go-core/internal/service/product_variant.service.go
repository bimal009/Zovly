package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type ProductVariantService interface {
	Create(ctx context.Context, input models.CreateProductVariantInput) (*models.ProductVariant, error)
}

type productVariantService struct {
	db                 *sqlx.DB
	rdb                *redis.Client
	logger             *slog.Logger
	productVariantRepo repository.ProductVariantRepo
}

func NewProductVariantService(
	db *sqlx.DB,
	rdb *redis.Client,
	logger *slog.Logger,
	productVariantRepo repository.ProductVariantRepo,
) ProductVariantService {
	return &productVariantService{
		db:                 db,
		rdb:                rdb,
		logger:             logger,
		productVariantRepo: productVariantRepo,
	}
}

func (s *productVariantService) Create(ctx context.Context, input models.CreateProductVariantInput) (*models.ProductVariant, error) {
	s.logger.Info("product variant create", "product_id", input.ProductID, "business_id", input.BusinessID, "name", input.Name)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	variant, err := s.productVariantRepo.Create(ctx, tx, input)
	if err != nil {
		s.logger.Error("product variant create failed", "product_id", input.ProductID, "error", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	s.rdb.Del(ctx, productKey(input.ProductID, input.BusinessID), productListKey(input.BusinessID))

	s.logger.Info("product variant created", "id", variant.ID, "product_id", variant.ProductID)
	return variant, nil
}
