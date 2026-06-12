// internal/service/service_service.go
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

type ServiceService interface {
	Create(ctx context.Context, input models.CreateServiceInput) (*models.Service, error)
	GetByID(ctx context.Context, id, businessID string) (*models.Service, error)
	List(ctx context.Context, businessID string, f repository.ListServicesFilter) ([]models.Service, error)
	Update(ctx context.Context, id, businessID string, input models.UpdateServiceInput) (*models.Service, error)
	Delete(ctx context.Context, id, businessID string) error
	ListForAIContext(ctx context.Context, businessID string) ([]models.Service, error)
}

type svcService struct {
	db          *sqlx.DB
	rdb         *redis.Client
	logger      *slog.Logger
	serviceRepo repository.ServiceRepo
}

func NewServiceService(
	db *sqlx.DB,
	rdb *redis.Client,
	logger *slog.Logger,
	serviceRepo repository.ServiceRepo,
) ServiceService {
	return &svcService{
		db:          db,
		rdb:         rdb,
		logger:      logger,
		serviceRepo: serviceRepo,
	}
}

// ─── cache helpers ────────────────────────────────────────────────────────────

func svcKey(id, businessID string) string {
	return fmt.Sprintf("%s%s:%s", constants.ServicesKeys, businessID, id)
}

func svcListKey(businessID string) string {
	return fmt.Sprintf("%s%s:list", constants.ServicesKeys, businessID)
}

func svcAIContextKey(businessID string) string {
	return fmt.Sprintf("%s%s:ai_context", constants.ServicesKeys, businessID)
}

func (s *svcService) invalidateBusinessCache(ctx context.Context, businessID string) {
	keys := []string{svcListKey(businessID), svcAIContextKey(businessID)}
	if err := s.rdb.Del(ctx, keys...).Err(); err != nil {
		s.logger.Warn("service cache invalidate failed", "business_id", businessID, "error", err)
	}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (s *svcService) Create(ctx context.Context, input models.CreateServiceInput) (*models.Service, error) {
	s.logger.Info("service create", "business_id", input.BusinessID, "name", input.Name, "type", input.Type)

	svc, err := s.serviceRepo.Create(ctx, input)
	if err != nil {
		s.logger.Error("service create failed", "business_id", input.BusinessID, "error", err)
		return nil, err
	}

	s.invalidateBusinessCache(ctx, input.BusinessID)

	s.logger.Info("service created", "id", svc.ID, "business_id", svc.BusinessID, "type", svc.Type)
	return svc, nil
}

// ─── GetByID ──────────────────────────────────────────────────────────────────

func (s *svcService) GetByID(ctx context.Context, id, businessID string) (*models.Service, error) {
	cacheKey := svcKey(id, businessID)

	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var svc models.Service
		if err := json.Unmarshal([]byte(cached), &svc); err == nil {
			return &svc, nil
		}
	}
	if err != nil && err != redis.Nil {
		s.logger.Warn("service cache get failed", "id", id, "error", err)
	}

	svc, err := s.serviceRepo.GetByID(ctx, id, businessID)
	if err != nil {
		s.logger.Error("service get failed", "id", id, "business_id", businessID, "error", err)
		return nil, err
	}
	if svc == nil {
		return nil, nil
	}

	if data, err := json.Marshal(svc); err == nil {
		if err := s.rdb.Set(ctx, cacheKey, data, constants.TTLMedium).Err(); err != nil {
			s.logger.Warn("service cache set failed", "id", id, "error", err)
		}
	}

	return svc, nil
}

// ─── List ─────────────────────────────────────────────────────────────────────

func (s *svcService) List(ctx context.Context, businessID string, f repository.ListServicesFilter) ([]models.Service, error) {
	cacheKey := svcListKey(businessID)

	// only cache unfiltered first page
	useCache := f.Type == nil && f.Status == nil && f.Offset == 0

	if useCache {
		cached, err := s.rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var svcs []models.Service
			if err := json.Unmarshal([]byte(cached), &svcs); err == nil {
				return svcs, nil
			}
		}
		if err != nil && err != redis.Nil {
			s.logger.Warn("service list cache get failed", "business_id", businessID, "error", err)
		}
	}

	svcs, err := s.serviceRepo.List(ctx, businessID, f)
	if err != nil {
		s.logger.Error("service list failed", "business_id", businessID, "error", err)
		return nil, err
	}

	if useCache {
		if data, err := json.Marshal(svcs); err == nil {
			if err := s.rdb.Set(ctx, cacheKey, data, constants.TTLShort).Err(); err != nil {
				s.logger.Warn("service list cache set failed", "business_id", businessID, "error", err)
			}
		}
	}

	return svcs, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (s *svcService) Update(ctx context.Context, id, businessID string, input models.UpdateServiceInput) (*models.Service, error) {
	s.logger.Info("service update", "id", id, "business_id", businessID)

	svc, err := s.serviceRepo.Update(ctx, id, businessID, input)
	if err != nil {
		s.logger.Error("service update failed", "id", id, "business_id", businessID, "error", err)
		return nil, err
	}
	if svc == nil {
		return nil, nil
	}

	s.rdb.Del(ctx, svcKey(id, businessID))
	s.invalidateBusinessCache(ctx, businessID)

	s.logger.Info("service updated", "id", id, "business_id", businessID)
	return svc, nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (s *svcService) Delete(ctx context.Context, id, businessID string) error {
	s.logger.Info("service delete", "id", id, "business_id", businessID)

	if err := s.serviceRepo.Delete(ctx, id, businessID); err != nil {
		s.logger.Error("service delete failed", "id", id, "business_id", businessID, "error", err)
		return err
	}

	s.rdb.Del(ctx, svcKey(id, businessID))
	s.invalidateBusinessCache(ctx, businessID)

	s.logger.Info("service deleted", "id", id, "business_id", businessID)
	return nil
}

// ─── ListForAIContext ─────────────────────────────────────────────────────────

func (s *svcService) ListForAIContext(ctx context.Context, businessID string) ([]models.Service, error) {
	cacheKey := svcAIContextKey(businessID)

	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var svcs []models.Service
		if err := json.Unmarshal([]byte(cached), &svcs); err == nil {
			return svcs, nil
		}
	}
	if err != nil && err != redis.Nil {
		s.logger.Warn("service ai context cache get failed", "business_id", businessID, "error", err)
	}

	svcs, err := s.serviceRepo.ListForAIContext(ctx, businessID)
	if err != nil {
		s.logger.Error("service ai context failed", "business_id", businessID, "error", err)
		return nil, err
	}

	// TTLShort — AI context should reflect recent changes quickly
	if data, err := json.Marshal(svcs); err == nil {
		if err := s.rdb.Set(ctx, cacheKey, data, constants.TTLShort).Err(); err != nil {
			s.logger.Warn("service ai context cache set failed", "business_id", businessID, "error", err)
		}
	}

	return svcs, nil
}
