// internal/service/category.service.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bimal009/Zovly/internal/constants"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/redis/go-redis/v9"
)

type CategoryService interface {
	Create(ctx context.Context, input models.CreateCategoryInput) error
	Get(ctx context.Context, id, businessID string) (*models.Category, error)
	GetAll(ctx context.Context, businessID string) ([]models.Category, error)
}

type categoryService struct {
	rdb          *redis.Client
	logger       *slog.Logger
	categoryRepo repository.CategoryRepo
}

func NewCategoryService(
	rdb *redis.Client,
	logger *slog.Logger,
	categoryRepo repository.CategoryRepo,
) CategoryService {
	return &categoryService{
		rdb:          rdb,
		logger:       logger,
		categoryRepo: categoryRepo,
	}
}

func categoryKey(id, businessID string) string {
	return fmt.Sprintf("%s%s:%s", constants.CategoriesKeys, businessID, id)
}

func categoryListKey(businessID string) string {
	return fmt.Sprintf("%s%s:list", constants.CategoriesKeys, businessID)
}

func (s *categoryService) invalidateBusinessCache(ctx context.Context, businessID string) {
	if err := s.rdb.Del(ctx, categoryListKey(businessID)).Err(); err != nil {
		s.logger.Warn("category cache invalidate failed", "business_id", businessID, "error", err)
	}
}

func (s *categoryService) Create(ctx context.Context, input models.CreateCategoryInput) error {
	input.Name = strings.ToUpper(strings.TrimSpace(input.Name))
	if input.Slug != nil {
		slug := strings.ToLower(strings.TrimSpace(*input.Slug))
		input.Slug = &slug
	}

	s.logger.Info("category create", "business_id", input.BusinessID, "name", input.Name)

	exists, err := s.categoryRepo.ExistsByName(ctx, input.BusinessID, input.Name)
	if err != nil {
		s.logger.Error("category exists check failed", "business_id", input.BusinessID, "error", err)
		return err
	}
	if exists {
		s.logger.Warn("category already exists", "business_id", input.BusinessID, "name", input.Name)
		return fmt.Errorf("category %q already exists", input.Name)
	}

	if err := s.categoryRepo.Create(ctx, input); err != nil {
		s.logger.Error("category create failed", "business_id", input.BusinessID, "error", err)
		return err
	}

	s.invalidateBusinessCache(ctx, input.BusinessID)

	s.logger.Info("category created", "business_id", input.BusinessID, "name", input.Name)
	return nil
}

func (s *categoryService) Get(ctx context.Context, id, businessID string) (*models.Category, error) {
	cacheKey := categoryKey(id, businessID)

	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var category models.Category
		if err := json.Unmarshal([]byte(cached), &category); err == nil {
			return &category, nil
		}
	}
	if err != nil && err != redis.Nil {
		s.logger.Warn("category cache get failed", "id", id, "error", err)
	}

	category, err := s.categoryRepo.Get(ctx, id, businessID)
	if err != nil {
		s.logger.Error("category get failed", "id", id, "business_id", businessID, "error", err)
		return nil, err
	}
	if category == nil {
		return nil, nil
	}

	if data, err := json.Marshal(category); err == nil {
		if err := s.rdb.Set(ctx, cacheKey, data, constants.TTLMedium).Err(); err != nil {
			s.logger.Warn("category cache set failed", "id", id, "error", err)
		}
	}

	return category, nil
}

func (s *categoryService) GetAll(ctx context.Context, businessID string) ([]models.Category, error) {
	cacheKey := categoryListKey(businessID)

	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var categories []models.Category
		if err := json.Unmarshal([]byte(cached), &categories); err == nil {
			return categories, nil
		}
	}
	if err != nil && err != redis.Nil {
		s.logger.Warn("category list cache get failed", "business_id", businessID, "error", err)
	}

	categories, err := s.categoryRepo.GetAll(ctx, businessID)
	if err != nil {
		s.logger.Error("category list failed", "business_id", businessID, "error", err)
		return nil, err
	}

	if data, err := json.Marshal(categories); err == nil {
		if err := s.rdb.Set(ctx, cacheKey, data, constants.TTLShort).Err(); err != nil {
			s.logger.Warn("category list cache set failed", "business_id", businessID, "error", err)
		}
	}

	return categories, nil
}
