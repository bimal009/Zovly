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
	Connected bool `json:"connected"`
	// FacebookLinked reports whether a connected Facebook Page has this
	// Instagram professional account linked to it (verified via the Graph API).
	// Instagram only delivers DMs through that link, so the UI prompts the user
	// to link a Page when this is false.
	FacebookLinked bool                  `json:"facebook_linked"`
	Account        *models.AppCredential `json:"account,omitempty"`
}

type InstagramService interface {
	SaveConnection(ctx context.Context, businessID, igUserID, igUsername string, token *models.IgLongLivedResponse) error
	GetConnectionStatus(ctx context.Context, businessID string) (*InstagramConnectionStatus, error)
	VeirfyFacebookAndInstaConnection(ctx context.Context, businessID string) (bool, error)
	ActivateConnection(ctx context.Context, businessID string) error
	RefreshToken(ctx context.Context, accessToken string) (*models.IgLongLivedResponse, error)
	RefreshExpiringTokens(ctx context.Context) error
	FetchUserProfile(ctx context.Context, accessToken string) (*models.InstagramUser, error)
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

func (s *instagramService) VeirfyFacebookAndInstaConnection(ctx context.Context, businessID string) (bool, error) {
	facebook, err := s.appCredentialRepo.ListByApp(ctx, businessID, string(models.PlatformFacebook))
	if err != nil {
		return false, fmt.Errorf("list facebook credentials: %w", err)
	}
	if len(facebook) == 0 {
		return false, nil
	}

	key, err := base64.StdEncoding.DecodeString(s.cfg.App.EncryptionKey)
	if err != nil || len(key) != 32 {
		return false, fmt.Errorf("invalid encryption key configuration")
	}

	for i := range facebook {
		fb := facebook[i]
		if fb.AccessToken == nil || fb.PlatformAccountID == nil {
			continue
		}

		pageToken, err := utils.Decrypt(*fb.AccessToken, key)
		if err != nil {
			s.log.Error("failed to decrypt facebook page token", "page_id", *fb.PlatformAccountID, "err", err)
			continue
		}

		igAccountID, err := s.fetchLinkedInstagramAccount(ctx, *fb.PlatformAccountID, pageToken)
		if err != nil {
			s.log.Warn("failed to check instagram link for page", "page_id", *fb.PlatformAccountID, "err", err)
			continue
		}
		if igAccountID != "" {
			s.log.Info("facebook page linked to instagram business account",
				"page_id", *fb.PlatformAccountID, "ig_account_id", igAccountID)
			return true, nil
		}
	}

	return false, nil
}

func (s *instagramService) fetchLinkedInstagramAccount(ctx context.Context, pageID, pageToken string) (string, error) {
	params := url.Values{
		"fields":       {"instagram_business_account"},
		"access_token": {pageToken},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://graph.facebook.com/v25.0/%s?%s", pageID, params.Encode()), nil)
	if err != nil {
		return "", fmt.Errorf("build instagram link request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call instagram link endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read instagram link response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("instagram link fetch failed (%d): %s", resp.StatusCode, body)
	}

	var result struct {
		InstagramBusinessAccount *struct {
			ID string `json:"id"`
		} `json:"instagram_business_account"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("decode instagram link response: %w", err)
	}
	if result.InstagramBusinessAccount == nil {
		return "", nil
	}
	return result.InstagramBusinessAccount.ID, nil
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

	facebookLinked, err := s.VeirfyFacebookAndInstaConnection(ctx, businessID)
	if err != nil {
		s.log.Warn("failed to verify facebook-instagram link", "business_id", businessID, "err", err)
		facebookLinked = false
	}

	return &InstagramConnectionStatus{
		Connected:      true,
		FacebookLinked: facebookLinked,
		Account:        account,
	}, nil
}

func (s *instagramService) ActivateConnection(ctx context.Context, businessID string) error {
	if err := s.appCredentialRepo.SetActiveByApp(ctx, businessID, "instagram", true); err != nil {
		return fmt.Errorf("activate instagram connection: %w", err)
	}
	return nil
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

		IsActive:    false,
		ConnectedAt: &now,
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

func (s *instagramService) HandleInstagramInboundMessage(ctx context.Context, platform models.Platform, instagramUserID string, event models.InstagramMessagingEvent) error {
	s.log.Info("inbound message received", "platform", platform, "instagram_user_id", instagramUserID, "sender", event.Sender.ID)

	cred, err := s.appCredentialRepo.GetByPlatformAccountID(ctx, instagramUserID)
	if err != nil {
		s.log.Error("credential lookup failed", "platform_id", instagramUserID, "err", err)
		return fmt.Errorf("get credential for instagram user %s: %w", instagramUserID, err)
	}

	senderProfile := &models.InstagramUser{}
	if pageToken, perr := s.getFacebookPageToken(ctx, cred.BusinessID); perr != nil {
		s.log.Warn("no facebook page token for ig profile lookup", "business_id", cred.BusinessID, "err", perr)
	} else if profile, gerr := s.fetchSenderProfileViaGraph(ctx, event.Sender.ID, pageToken); gerr != nil {
		s.log.Warn("graph sender profile fetch failed", "sender_id", event.Sender.ID, "err", gerr)
	} else {
		senderProfile = profile
	}

	s.log.Info("sender profile fetched",
		"sender_id", event.Sender.ID,
		"name", senderProfile.Name,
		"username", senderProfile.Username,
		"has_avatar", senderProfile.ProfilePic != "",
	)

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

		// Shared post — Instagram hands us the media as a direct image asset on
		// the lookaside CDN, so run it through the AI image route.
		case models.InstagramAttachmentTypePost:
			s.chatService.HandleImageAttachment(ctx, tx, platform, conv, &cred, instagramUserID, attachment.Payload.URL)

		// Reel/story/link — no media asset, but Instagram includes the caption in
		// the payload, so hand that text to the AI alongside the link.
		case models.InstagramAttachmentTypeReel:
			s.chatService.HandleSharedContent(ctx, tx, platform, conv, &cred, attachment.Payload.URL, "reel", attachment.Payload.Title)

		case models.InstagramAttachmentTypeStory:
			s.chatService.HandleSharedContent(ctx, tx, platform, conv, &cred, attachment.Payload.URL, "story", attachment.Payload.Title)

		case models.InstagramAttachmentTypeLink, models.InstagramAttachmentTypeURL:
			s.chatService.HandleSharedContent(ctx, tx, platform, conv, &cred, attachment.Payload.URL, "link", attachment.Payload.Title)

		default:
			s.log.Warn("unknown attachment type", "type", attachment.Type)
			continue
		}
	}

	return tx.Commit()
}

func (s *instagramService) getFacebookPageToken(ctx context.Context, businessID string) (string, error) {
	creds, err := s.appCredentialRepo.ListByApp(ctx, businessID, "facebook")
	if err != nil {
		return "", fmt.Errorf("list facebook credentials: %w", err)
	}

	var chosen *models.AppCredential
	for i := range creds {
		if creds[i].AccessToken == nil {
			continue
		}
		if creds[i].IsActive {
			chosen = &creds[i]
			break
		}
		if chosen == nil {
			chosen = &creds[i]
		}
	}
	if chosen == nil {
		return "", fmt.Errorf("no facebook page with access token for business %s", businessID)
	}

	key, err := base64.StdEncoding.DecodeString(s.cfg.App.EncryptionKey)
	if err != nil || len(key) != 32 {
		return "", fmt.Errorf("invalid encryption key configuration")
	}
	token, err := utils.Decrypt(*chosen.AccessToken, key)
	if err != nil {
		return "", fmt.Errorf("decrypt facebook page token: %w", err)
	}
	return token, nil
}

func (s *instagramService) fetchSenderProfileViaGraph(ctx context.Context, senderID, pageToken string) (*models.InstagramUser, error) {
	params := url.Values{
		"fields":       {"name,username,profile_pic,follower_count"},
		"access_token": {pageToken},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://graph.facebook.com/v25.0/%s?%s", senderID, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("build graph sender profile request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call graph sender profile endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read graph sender profile response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graph sender profile fetch failed (%d): %s", resp.StatusCode, body)
	}

	var result struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		Username      string `json:"username"`
		ProfilePic    string `json:"profile_pic"`
		FollowerCount int    `json:"follower_count"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode graph sender profile: %w", err)
	}
	return &models.InstagramUser{
		ID:         result.ID,
		Name:       result.Name,
		Username:   result.Username,
		ProfilePic: result.ProfilePic,
	}, nil
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
