package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/pkg/utils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type InstagramConnectionStatus struct {
	Connected bool                  `json:"connected"`
	Account   *models.AppCredential `json:"account,omitempty"`
}

type InstagramService interface {
	SaveConnection(ctx context.Context, businessID, igUserID, igUsername string, token *models.IgLongLivedResponse) error
	GetConnectionStatus(ctx context.Context, businessID string) (*InstagramConnectionStatus, error)
	RefreshToken(ctx context.Context, accessToken string) (*models.IgLongLivedResponse, error)
	RefreshExpiringTokens(ctx context.Context) error
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
) InstagramService {
	return &instagramService{
		db:                db,
		appCredentialRepo: appCredentialRepo,
		appRepo:           appRepo,
		cfg:               cfg,
		httpClient:        &http.Client{Timeout: 10 * time.Second},
		log:               log,
	}
}

func (s *instagramService) GetConnectionStatus(ctx context.Context, businessID string) (*InstagramConnectionStatus, error) {
	conn, err := s.appRepo.GetByBusinessID(ctx, businessID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("get connection status: %w", err)
	}
	if err == sql.ErrNoRows || !conn.Instagram {
		return &InstagramConnectionStatus{Connected: false}, nil
	}

	creds, err := s.appCredentialRepo.ListByApp(ctx, businessID, "instagram")
	if err != nil {
		return nil, fmt.Errorf("list instagram credentials: %w", err)
	}

	var account *models.AppCredential
	if len(creds) > 0 {
		account = &creds[0]
	}

	return &InstagramConnectionStatus{Connected: true, Account: account}, nil
}

func (s *instagramService) SaveConnection(ctx context.Context, businessID, igUserID, igUsername string, token *models.IgLongLivedResponse) error {
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
		BusinessID:          businessID,
		AppName:             "instagram",
		AccessToken:         &encToken,
		TokenExpiresAt:      &expiresAt,
		PlatformAccountID:   &igUserID,
		PlatformAccountName: &igUsername,
		Scopes:              pq.StringArray{"instagram_business_basic", "instagram_business_content_publish", "instagram_business_manage_comments", "instagram_business_manage_messages"},
		IsActive:            true,
		ConnectedAt:         &now,
	}

	if err := s.appCredentialRepo.Upsert(ctx, tx, cred); err != nil {
		return fmt.Errorf("upsert instagram credential: %w", err)
	}
	if err := s.appRepo.Update(ctx, tx, businessID, models.ConnectionInstagram); err != nil {
		return fmt.Errorf("flip instagram connection flag: %w", err)
	}
	return tx.Commit()
}

func (s *instagramService) RefreshToken(ctx context.Context, accessToken string) (*models.IgLongLivedResponse, error) {
	params := url.Values{
		"grant_type":   {"ig_refresh_token"},
		"access_token": {accessToken},
	}
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		"https://graph.instagram.com/refresh_access_token?"+params.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("build refresh request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call refresh endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("refresh failed (%d): %s", resp.StatusCode, body)
	}

	var result models.IgLongLivedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode refresh response: %w", err)
	}
	return &result, nil
}

func (s *instagramService) RefreshExpiringTokens(ctx context.Context) error {
	creds, err := s.appCredentialRepo.GetExpiringInstaTokens(ctx)
	if err != nil {
		return fmt.Errorf("get expiring tokens: %w", err)
	}

	if len(creds) == 0 {
		s.log.Info("no expiring instagram tokens found")
		return nil
	}

	s.log.Info("refreshing expiring instagram tokens", "count", len(creds))

	key, err := base64.StdEncoding.DecodeString(s.cfg.App.EncryptionKey)
	if err != nil || len(key) != 32 {
		return fmt.Errorf("invalid encryption key configuration")
	}

	for _, cred := range creds {
		if cred.AccessToken == nil {
			s.log.Warn("skipping credential with nil token", "id", cred.ID)
			continue
		}

		plainToken, err := utils.Decrypt(*cred.AccessToken, key)
		if err != nil {
			s.log.Error("failed to decrypt token", "id", cred.ID, "err", err)
			continue // don't stop — try the rest
		}

		newToken, err := s.RefreshToken(ctx, plainToken)
		if err != nil {
			s.log.Error("failed to refresh instagram token", "id", cred.ID, "err", err)
			// mark as error in db so the UI can prompt reconnect
			s.appCredentialRepo.MarkTokenError(ctx, cred.ID, err.Error())
			continue
		}

		newEncToken, err := utils.Encrypt(newToken.AccessToken, key)
		if err != nil {
			s.log.Error("failed to encrypt new token", "id", cred.ID, "err", err)
			continue
		}

		newExpiresAt := time.Now().Add(time.Duration(newToken.ExpiresIn) * time.Second)
		if err := s.appCredentialRepo.UpdateToken(ctx, cred.ID, newEncToken, newExpiresAt); err != nil {
			s.log.Error("failed to update token in db", "id", cred.ID, "err", err)
			continue
		}

		s.log.Info("instagram token refreshed", "id", cred.ID, "expires_at", newExpiresAt)
	}

	return nil
}
