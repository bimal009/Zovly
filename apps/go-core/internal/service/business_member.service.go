package service

import (
	"context"
	"log/slog"

	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/jmoiron/sqlx"
)

type BusinessMemberService interface {
	Create(ctx context.Context, input models.BusinessMember) (*models.BusinessMember, error)
}

type businessMemberService struct {
	db                 *sqlx.DB
	businessMemberRepo repository.BusinessMemberRepo
	log                *slog.Logger
}

func NewBusinessMemberService(
	db *sqlx.DB,
	businessMemberRepo repository.BusinessMemberRepo,
	log *slog.Logger,
) BusinessMemberService {
	return &businessMemberService{
		db:                 db,
		businessMemberRepo: businessMemberRepo,
		log:                log,
	}
}

func (s *businessMemberService) Create(ctx context.Context, input models.BusinessMember) (*models.BusinessMember, error) {
	s.log.Info("creating business member",
		"business_id", input.BusinessID,
		"user_id", input.UserID,
		"role", input.Role,
	)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		s.log.Error("failed to begin transaction", "error", err)
		return nil, err
	}
	defer tx.Rollback()

	member, err := s.businessMemberRepo.CreateTx(ctx, tx, input)
	if err != nil {
		s.log.Error("failed to create business member", "error", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.log.Error("failed to commit transaction", "error", err)
		return nil, err
	}

	s.log.Info("business member created",
		"id", member.ID,
		"business_id", member.BusinessID,
		"user_id", member.UserID,
	)

	return member, nil
}
