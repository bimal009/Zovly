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
	"strings"
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
	FetchUserProfile(ctx context.Context, accessToken string) (*models.InstagramUser, error)
	SubscribeWebhook(ctx context.Context, businessID string) error
	HandleInstagramInboundMessage(ctx context.Context, platform models.Platform, instagramUserID string, event models.InstagramMessagingEvent) error
}

type instagramService struct {
	db                *sqlx.DB
	appCredentialRepo repository.AppCredentialRepo
	appRepo           repository.AppRepo
	cfg               *config.Config
	httpClient        *http.Client
	log               *slog.Logger
	chatService       ChatService
	messageRepo       repository.MessageRepo
}

func NewInstagramService(
	db *sqlx.DB,
	appCredentialRepo repository.AppCredentialRepo,
	appRepo repository.AppRepo,
	messageRepo repository.MessageRepo,
	cfg *config.Config,
	log *slog.Logger,
	chatService ChatService,
) InstagramService {
	return &instagramService{
		db:                db,
		appCredentialRepo: appCredentialRepo,
		appRepo:           appRepo,
		cfg:               cfg,
		httpClient:        &http.Client{Timeout: 10 * time.Second},
		log:               log,
		chatService:       chatService,
		messageRepo:       messageRepo,
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

// SubscribeWebhook subscribes the business's connected Instagram account to the
// messaging webhook fields and stamps webhook_subscribed_at on the credential.
// Mirrors the Facebook SubscribeMessengerPage flow, but Instagram has a single
// account per business so no page id is needed.
func (s *instagramService) SubscribeWebhook(ctx context.Context, businessID string) error {
	creds, err := s.appCredentialRepo.ListByApp(ctx, businessID, "instagram")
	if err != nil {
		return fmt.Errorf("list instagram credentials: %w", err)
	}
	if len(creds) == 0 {
		return fmt.Errorf("no instagram account connected for business %s", businessID)
	}

	cred := creds[0]
	if cred.AccessToken == nil {
		return fmt.Errorf("instagram account has no access token")
	}
	if cred.PlatformAccountID == nil {
		return fmt.Errorf("instagram account has no platform account id")
	}

	key, err := base64.StdEncoding.DecodeString(s.cfg.App.EncryptionKey)
	if err != nil || len(key) != 32 {
		return fmt.Errorf("invalid encryption key configuration")
	}

	token, err := utils.Decrypt(*cred.AccessToken, key)
	if err != nil {
		return fmt.Errorf("decrypt instagram token: %w", err)
	}

	if err := s.subscribeWebhooks(ctx, token); err != nil {
		return fmt.Errorf("subscribe instagram webhooks: %w", err)
	}

	if err := s.appCredentialRepo.UpdateWebhookSubscribed(ctx, businessID, *cred.PlatformAccountID, "instagram"); err != nil {
		return fmt.Errorf("mark webhook subscribed: %w", err)
	}
	return nil
}

func (s *instagramService) subscribeWebhooks(ctx context.Context, accessToken string) error {
	params := url.Values{
		"subscribed_fields": {"messages,messaging_postbacks,messaging_seen,message_reactions"},
		"access_token":      {accessToken},
	}
	reqURL := "https://graph.instagram.com/v25.0/me/subscribed_apps"

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
	s.log.Info("instagram webhooks subscribed")
	return nil
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

func (s *instagramService) HandleInstagramInboundMessage(ctx context.Context, platform models.Platform, instagramUserID string, event models.InstagramMessagingEvent) error {
	s.log.Info("inbound message received", "platform", platform, "instagram_user_id", instagramUserID, "sender", event.Sender.ID)

	cred, err := s.appCredentialRepo.GetByPlatformAccountID(ctx, instagramUserID)
	if err != nil {
		s.log.Error("credential lookup failed", "platform_id", instagramUserID, "err", err)
		return fmt.Errorf("get credential for instagram user %s: %w", instagramUserID, err)
	}

	encKey, err := base64.StdEncoding.DecodeString(s.cfg.App.EncryptionKey)
	if err != nil {
		s.log.Error("decode encryption key failed", "err", err)
		return fmt.Errorf("decode encryption key: %w", err)
	}
	instagramToken, err := utils.Decrypt(*cred.AccessToken, encKey)
	if err != nil {
		s.log.Error("token decrypt failed", "err", err)
		return fmt.Errorf("decrypt access token: %w", err)
	}

	senderProfile, err := s.fetchSenderProfile(ctx, event.Sender.ID, instagramToken)
	if err != nil {
		s.log.Warn("fetch sender profile failed, using fallback", "sender_id", event.Sender.ID, "err", err)
		senderProfile = &models.InstagramUser{}
	}
	s.log.Info("sender profile fetched", "username", senderProfile.Username)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	conv, err := s.chatService.FindOrCreate(ctx, tx, models.CreateConversation{
		BusinessID:       cred.BusinessID,
		Platform:         string(platform),
		ThreadID:         event.Sender.ID,
		ContactID:        event.Sender.ID,
		ContactName:      &senderProfile.Name,
		ContactUsername:  &senderProfile.Username,
		ContactAvatarURL: &senderProfile.ProfilePic,
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
	// 2) Attachments — size-gated, type-aware.
	for _, attachment := range event.Message.Attachments {
		switch attachment.Type {

		case models.InstagramAttachmentTypeImage:
			s.chatService.HandleImageAttachment(ctx, tx, platform, conv, &cred, instagramUserID, attachment.Payload.URL)

		case models.InstagramAttachmentTypeAudio:
			s.chatService.HandleAudioAttachment(ctx, tx, platform, conv, &cred, instagramUserID, attachment.Payload.URL)

		case models.InstagramAttachmentTypeVideo:
			s.chatService.HandleUnprocessableAttachment(ctx, tx, platform, conv, &cred, instagramUserID,
				attachment.Payload.URL, models.MediaTypeVideo, "video",
				"[Customer sent a video that can't be processed automatically]")

		case models.InstagramAttachmentTypeFile:
			s.chatService.HandleUnprocessableAttachment(ctx, tx, platform, conv, &cred, instagramUserID,
				attachment.Payload.URL, models.MediaTypeDocument, "file",
				"[Customer sent a file that can't be processed automatically]")

		case models.InstagramAttachmentTypeReel:
			s.chatService.HandleSharedContent(ctx, tx, platform, conv, &cred, attachment.Payload.URL, "reel")

		case models.InstagramAttachmentTypeStory:
			s.chatService.HandleSharedContent(ctx, tx, platform, conv, &cred, attachment.Payload.URL, "story")

		case models.InstagramAttachmentTypeLink, models.InstagramAttachmentTypeURL:
			s.chatService.HandleSharedContent(ctx, tx, platform, conv, &cred, attachment.Payload.URL, "link")

		default:
			s.log.Warn("unknown attachment type", "type", attachment.Type)
			continue
		}
	}

	return tx.Commit()
}

func (s *instagramService) fetchSenderProfile(ctx context.Context, senderID, accessToken string) (*models.InstagramUser, error) {
	params := url.Values{
		"fields":       {"name,username,profile_picture_url"},
		"access_token": {accessToken},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://graph.instagram.com/v25.0/%s?%s", senderID, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("build sender profile request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call sender profile endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sender profile fetch failed (%d): %s", resp.StatusCode, body)
	}

	var result models.InstagramUser
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode sender profile: %w", err)
	}
	return &result, nil
}

func (s *instagramService) FetchUserProfile(ctx context.Context, accessToken string) (*models.InstagramUser, error) {
	params := url.Values{
		"fields":       {"id,user_id,username,profile_picture_url,name,followers_count,follows_count,media_count"},
		"access_token": {accessToken},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://graph.instagram.com/me?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build instagram user request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call instagram user endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("instagram user fetch failed (%d): %s", resp.StatusCode, body)
	}

	var result models.InstagramUser
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode instagram user response: %w", err)
	}
	return &result, nil
}
