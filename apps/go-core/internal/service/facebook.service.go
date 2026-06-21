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
	"strings"
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
	SubscribeMessengerPage(ctx context.Context, businessID, pageID string) error
	HandleFacebookInboundMessage(ctx context.Context, platform models.Platform, pageID string, event models.FacebookMessagingEvent) error
}

type facebookService struct {
	db                *sqlx.DB
	appCredentialRepo repository.AppCredentialRepo
	appRepo           repository.AppRepo
	messageRepo       repository.MessageRepo
	cfg               *config.Config
	httpClient        *http.Client
	chatService       ChatService
	log               *slog.Logger
}

func NewFacebookService(
	db *sqlx.DB,
	appCredentialRepo repository.AppCredentialRepo,
	appRepo repository.AppRepo,
	messageRepo repository.MessageRepo,
	cfg *config.Config,
	chatService ChatService,
	log *slog.Logger,
) FacebookService {
	return &facebookService{
		db:                db,
		appCredentialRepo: appCredentialRepo,
		appRepo:           appRepo,
		messageRepo:       messageRepo,
		cfg:               cfg,
		httpClient:        &http.Client{Timeout: 10 * time.Second},
		chatService:       chatService,
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

func (s *facebookService) SubscribeMessengerPage(ctx context.Context, businessID, pageID string) error {
	creds, err := s.appCredentialRepo.ListByApp(ctx, businessID, "facebook")
	if err != nil {
		return fmt.Errorf("list facebook credentials: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(s.cfg.App.EncryptionKey)
	if err != nil || len(key) != 32 {
		return fmt.Errorf("invalid encryption key configuration")
	}

	var pageToken string
	var found bool
	for _, cred := range creds {
		if cred.PlatformAccountID != nil && *cred.PlatformAccountID == pageID {
			if cred.AccessToken == nil {
				return fmt.Errorf("page %s has no access token", pageID)
			}
			token, err := utils.Decrypt(*cred.AccessToken, key)
			if err != nil {
				return fmt.Errorf("decrypt page token: %w", err)
			}
			pageToken = token
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("page %s not found for business %s", pageID, businessID)
	}

	if err := s.subscribePageWebhooks(ctx, pageID, pageToken); err != nil {
		return fmt.Errorf("subscribe page webhooks: %w", err)
	}

	if err := s.appCredentialRepo.UpdateWebhookSubscribed(ctx, businessID, pageID, "facebook"); err != nil {
		return fmt.Errorf("mark webhook subscribed: %w", err)
	}
	return nil
}

func (s *facebookService) subscribePageWebhooks(ctx context.Context, pageID, pageToken string) error {
	params := url.Values{
		"subscribed_fields": {"messages,messaging_postbacks,message_deliveries,message_reads"},
		"access_token":      {pageToken},
	}
	reqURL := fmt.Sprintf("https://graph.facebook.com/v25.0/%s/subscribed_apps", pageID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("build subscribe request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call subscribe endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("subscribe failed (%d): %s", resp.StatusCode, body)
	}
	return nil
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

// ── inbound message handling ─────────────────────────────────────────
func (s *facebookService) HandleFacebookInboundMessage(ctx context.Context, platform models.Platform, pageID string, event models.FacebookMessagingEvent) error {
	s.log.Info("inbound message received", "platform", platform, "page_id", pageID, "sender", event.Sender.ID)

	cred, err := s.appCredentialRepo.GetByPlatformAccountID(ctx, pageID)
	if err != nil {
		s.log.Error("credential lookup failed", "platform_id", pageID, "err", err)
		return fmt.Errorf("get credential for page %s: %w", pageID, err)
	}

	encKey, err := base64.StdEncoding.DecodeString(s.cfg.App.EncryptionKey)
	if err != nil {
		s.log.Error("decode encryption key failed", "err", err)
		return fmt.Errorf("decode encryption key: %w", err)
	}
	pageToken, err := utils.Decrypt(*cred.AccessToken, encKey)
	if err != nil {
		s.log.Error("token decrypt failed", "err", err)
		return fmt.Errorf("decrypt access token: %w", err)
	}

	user, err := s.chatService.FetchUserProfile(ctx, event.Sender.ID, pageToken)
	if err != nil {
		s.log.Warn("fetch user profile failed, using fallback", "sender_id", event.Sender.ID, "err", err)
		user = &models.MessengerProfile{}
	}
	s.log.Info("user profile fetched", "name", user.FirstName+" "+user.LastName)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	fullName := user.FirstName + " " + user.LastName
	conv, err := s.chatService.FindOrCreate(ctx, tx, models.CreateConversation{
		BusinessID:       cred.BusinessID,
		Platform:         string(platform),
		ThreadID:         event.Sender.ID,
		ContactID:        event.Sender.ID,
		ContactName:      &fullName,
		ContactAvatarURL: &user.ProfilePic,
	})
	if err != nil {
		s.log.Error("find or create conversation failed", "err", err)
		return fmt.Errorf("find or create conversation: %w", err)
	}
	s.log.Info("conversation ready", "conversation_id", conv.ID, "business_id", cred.BusinessID)

	// 1) Text — Meta can send text and attachments in the SAME event,
	//    so this runs independently of the attachment loop below.
	if event.Message.Text != "" {
		text := event.Message.Text
		insertedMsg, err := s.messageRepo.Create(ctx, tx, models.CreateMessage{
			ConversationID: conv.ID,
			BusinessID:     cred.BusinessID,
			Direction:      models.MessageDirectionIn,
			SentBy:         nil,
			Content:        &text,
			Status:         nil,
		})
		if err != nil {
			s.log.Error("create message failed", "err", err)
			return fmt.Errorf("create message: %w", err)
		}
		if err := s.chatService.StreamMessage(ctx, platform, insertedMsg.ID, cred.BusinessID, conv.ID, false); err != nil {
			s.log.Error("stream message failed", "err", err)
			return fmt.Errorf("stream message: %w", err)
		}
	}

	// 2) Attachments — size-gated, type-aware.
	for _, attachment := range event.Message.Attachments {
		switch attachment.Type {

		case models.FacebookAttachmentTypeImage:
			s.chatService.HandleImageAttachment(ctx, tx, platform, conv, &cred, pageID, attachment.Payload.URL)

		case models.FacebookAttachmentTypeAudio:
			s.chatService.HandleAudioAttachment(ctx, tx, platform, conv, &cred, pageID, attachment.Payload.URL)

		case models.FacebookAttachmentTypeVideo:
			s.chatService.HandleUnprocessableAttachment(ctx, tx, platform, conv, &cred, pageID,
				attachment.Payload.URL, models.MediaTypeVideo, "video",
				"[Customer sent a video that can't be processed automatically]")

		case models.FacebookAttachmentTypeFile:
			s.chatService.HandleUnprocessableAttachment(ctx, tx, platform, conv, &cred, pageID,
				attachment.Payload.URL, models.MediaTypeDocument, "file",
				"[Customer sent a file that can't be processed automatically]")

		case models.FacebookAttachmentTypeFallback:
			s.chatService.HandleSharedLink(ctx, tx, platform, conv, &cred, attachment)

		default:
			s.log.Warn("unknown attachment type", "type", attachment.Type)
			continue
		}
	}

	return tx.Commit()
}
