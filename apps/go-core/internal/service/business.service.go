package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/jmoiron/sqlx"
)

type BusinessService interface {
	Create(ctx context.Context, input models.Business, userId string) (*models.Business, error)
	GetByUserId(ctx context.Context, userId string) (*models.BusinessWithMembers, error)
}

type businessService struct {
	db                 *sqlx.DB
	businessRepo       repository.BusinessRepo
	businessMemberRepo repository.BusinessMemberRepo
	userRepo           repository.UserRepo
	appRepo            repository.AppRepo
	log                *slog.Logger
}

func NewBusinessService(
	db *sqlx.DB,
	businessRepo repository.BusinessRepo,
	businessMemberRepo repository.BusinessMemberRepo,
	userRepo repository.UserRepo,
	log *slog.Logger,
	appRepo repository.AppRepo,
) BusinessService {
	return &businessService{
		db:                 db,
		businessRepo:       businessRepo,
		businessMemberRepo: businessMemberRepo,
		userRepo:           userRepo,
		log:                log,
		appRepo:            appRepo,
	}
}

func ptr[T any](v T) *T { return &v }

func (s *businessService) Create(ctx context.Context, input models.Business, userId string) (*models.Business, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	created, err := s.businessRepo.CreateTx(ctx, tx, input)
	if err != nil {
		return nil, fmt.Errorf("insert: %w", err)
	}

	member := models.BusinessMember{
		BusinessID: created.ID,
		UserID:     userId,
		Role:       models.MemberRoleOwner,

		CanManageContent:   true,
		CanViewAnalytics:   true,
		CanManageAds:       true,
		CanReadDMs:         true,
		CanReplyDMs:        true,
		CanReadComments:    true,
		CanReplyComments:   true,
		CanViewLeads:       true,
		CanManageLeads:     true,
		CanViewBookings:    true,
		CanManageBookings:  true,
		CanViewInventory:   true,
		CanManageInventory: true,
		CanViewOrders:      true,
		CanManageSettings:  true,
		CanManageMembers:   true,
		CanManageBilling:   true,
	}

	_, err = s.businessMemberRepo.CreateTx(ctx, tx, member)
	if err != nil {
		return nil, fmt.Errorf("create business member: %w", err)
	}

	if err = s.appRepo.Create(ctx, tx, created.ID); err != nil {
		return nil, fmt.Errorf("create app connections: %w", err)
	}

	_, err = s.userRepo.UpdateTx(ctx, tx, userId, models.UserUpdate{
		Role:      ptr(models.RoleVendor),
		Onboarded: ptr(true),
	})
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	s.log.InfoContext(ctx,
		"business created",
		"business_id", created.ID,
		"owner_user_id", userId,
	)

	return created, nil
}

func (s *businessService) GetByUserId(ctx context.Context, userId string) (*models.BusinessWithMembers, error) {
	result, err := s.businessRepo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("get business: %w", err)
	}
	return result, nil
}
