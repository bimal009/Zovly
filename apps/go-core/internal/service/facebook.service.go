package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"database/sql"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/pkg/utils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type FacebookConnectionStatus struct {
	Connected bool                      `json:"connected"`
	Pages     []FacebookPageWithDetails `json:"pages"`
}

type FacebookService interface {
	SaveConnections(ctx context.Context, businessID string, pages []models.FbPage) error
	GetConnectionStatus(ctx context.Context, businessID string) (*FacebookConnectionStatus, error)
	TogglePage(ctx context.Context, businessID, pageID string) (bool, error)
}

type facebookService struct {
	db                *sqlx.DB
	appCredentialRepo repository.AppCredentialRepo
	appRepo           repository.AppRepo
	cfg               *config.Config
	httpClient        *http.Client
	log               *slog.Logger
}

func NewFacebookService(
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

type FacebookPageWithDetails struct {
	models.AppCredential
	Details *PageDetails `json:"details,omitempty"`
}

func (s *facebookService) GetConnectionStatus(ctx context.Context, businessID string) (*FacebookConnectionStatus, error) {
	conn, err := s.appRepo.GetByBusinessID(ctx, businessID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("get connection status: %w", err)
	}
	if err == sql.ErrNoRows {
		return &FacebookConnectionStatus{Connected: false, Pages: []FacebookPageWithDetails{}}, nil
	}

	creds, err := s.appCredentialRepo.ListByApp(ctx, businessID, "facebook")
	if err != nil {
		return nil, fmt.Errorf("list facebook pages: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(s.cfg.App.EncryptionKey)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("invalid encryption key configuration")
	}

	pages := make([]FacebookPageWithDetails, 0, len(creds))
	for _, cred := range creds {
		page := FacebookPageWithDetails{AppCredential: cred}

		if cred.AccessToken == nil || cred.PlatformAccountID == nil {
			pages = append(pages, page) // no token/id — skip enrichment, still show the row
			continue
		}

		token, err := utils.Decrypt(*cred.AccessToken, key)
		if err != nil {
			s.log.Warn("failed to decrypt page token", "page_id", *cred.PlatformAccountID, "err", err)
			pages = append(pages, page)
			continue
		}

		details, err := s.fetchPageDetails(ctx, *cred.PlatformAccountID, token)
		if err != nil {
			s.log.Warn("failed to fetch page details", "page_id", *cred.PlatformAccountID, "err", err)
			pages = append(pages, page) // token may be revoked — show row without details
			continue
		}

		page.Details = details
		pages = append(pages, page)
	}

	return &FacebookConnectionStatus{
		Connected: conn.Facebook,
		Pages:     pages,
	}, nil
}

func (s *facebookService) SaveConnections(ctx context.Context, businessID string, pages []models.FbPage) error {
	key, err := base64.StdEncoding.DecodeString(s.cfg.App.EncryptionKey)
	if err != nil || len(key) != 32 {
		return fmt.Errorf("invalid encryption key configuration")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	for _, page := range pages {
		encToken, err := utils.Encrypt(page.AccessToken, key)
		if err != nil {
			return fmt.Errorf("encrypt page token: %w", err)
		}

		fbCred := models.CreateAppCredential{
			BusinessID:          businessID,
			AppName:             "facebook",
			AccessToken:         &encToken,
			PlatformAccountID:   &page.ID,
			PlatformAccountName: &page.Name,
			Scopes:              pq.StringArray{"pages_show_list", "pages_read_engagement", "pages_messaging"},
			IsActive:            false,
			ConnectedAt:         &now,
		}

		if err := s.appCredentialRepo.Upsert(ctx, tx, fbCred); err != nil {
			return fmt.Errorf("upsert credential for page %s: %w", page.ID, err)
		}
	}

	if err := s.appRepo.Update(ctx, tx, businessID, models.ConnectionFacebook); err != nil {
		return fmt.Errorf("flip facebook connection flag: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
func (s *facebookService) TogglePage(ctx context.Context, businessID, pageID string) (bool, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	newState, err := s.appCredentialRepo.TogglePageActive(ctx, tx, businessID, pageID)
	if err != nil {
		return false, fmt.Errorf("toggle page: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit: %w", err)
	}
	return newState, nil
}

type PageDetails struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	About          string `json:"about,omitempty"`
	FanCount       int    `json:"fan_count"`
	FollowersCount int    `json:"followers_count"`
	Picture        struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
	Link     string `json:"link"`
	Category string `json:"category"`
}

func (s *facebookService) fetchPageDetails(ctx context.Context, pageID, pageToken string) (*PageDetails, error) {
	params := url.Values{
		"fields":       {"id,name,about,fan_count,followers_count,picture,link,category"},
		"access_token": {pageToken},
	}
	reqURL := fmt.Sprintf("https://graph.facebook.com/v25.0/%s?%s", pageID, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build page details request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call page details endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("page details failed (%d): %s", resp.StatusCode, body)
	}

	var details PageDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("decode page details: %w", err)
	}
	return &details, nil
}
