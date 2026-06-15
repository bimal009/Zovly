package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/pkg/utils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type InstagramService interface {
	SaveConnection(ctx context.Context, businessID, igUserID string, token *models.IgLongLivedResponse) error
}

type instagramService struct {
	db                *sqlx.DB
	appCredentialRepo repository.AppCredentialRepo
	appRepo           repository.AppRepo
	cfg               *config.Config
	httpClient        *http.Client
	log               *slog.Logger
}

func NewInstagramService(
	db *sqlx.DB,
	appCredentialRepo repository.AppCredentialRepo,
	appRepo repository.AppRepo,
	cfg *config.Config,
	log *slog.Logger,
) FacebookService {
	return &facebookService{
		db:                db,
		appCredentialRepo: appCredentialRepo,
		appRepo:           appRepo,
		cfg:               cfg,
		httpClient:        &http.Client{Timeout: 10 * time.Second},
		log:               log,
	}
}

func (s *instagramService) SaveConnection(ctx context.Context, businessID, igUserID string, token *models.IgLongLivedResponse) error {
	key, err := base64.StdEncoding.DecodeString(s.cfg.App.EncryptionKey)
	if err != nil || len(key) != 32 {
		return fmt.Errorf("invalid encryption key configuration")
	}

	encToken, err := utils.Encrypt(token.AccessToken, key)
	if err != nil {
		return fmt.Errorf("encrypt ig token: %w", err)
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	expiresAt := now.Add(time.Duration(token.ExpiresIn) * time.Second)

	cred := models.CreateAppCredential{
		BusinessID:        businessID,
		AppName:           "instagram",
		AccessToken:       &encToken,
		TokenExpiresAt:    &expiresAt,
		PlatformAccountID: &igUserID,
		Scopes:            pq.StringArray{"instagram_business_basic", "instagram_business_content_publish", "instagram_business_manage_comments", "instagram_business_manage_messages"},
		IsActive:          true,
		ConnectedAt:       &now,
	}

	if err := s.appCredentialRepo.Upsert(ctx, tx, cred); err != nil {
		return fmt.Errorf("upsert instagram credential: %w", err)
	}
	if err := s.appRepo.Update(ctx, tx, businessID, models.ConnectionInstagram); err != nil {
		return fmt.Errorf("flip instagram connection flag: %w", err)
	}
	return tx.Commit()
}
