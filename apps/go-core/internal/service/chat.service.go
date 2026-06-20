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

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/pkg/utils"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type ChatService interface {
	FindOrCreate(ctx context.Context, tx *sqlx.Tx, conv models.CreateConversation) (*models.Conversation, error)
	HandleInboundMessage(ctx context.Context, platform string, platformID string, event models.MessagingEvent) error
}

type chatService struct {
	db                *sqlx.DB
	messageEmbedRepo  repository.MessageEmbeddingRepo
	messageRepo       repository.MessageRepo
	conversationRepo  repository.ConversationRepo
	appCredentialRepo repository.AppCredentialRepo
	cfg               config.Config
	httpClient        *http.Client
	rdb               *redis.Client
	log               *slog.Logger
}

func NewChatService(
	db *sqlx.DB,
	messageRepo repository.MessageRepo,
	appCredentialRepo repository.AppCredentialRepo,
	messageEmbedRepo repository.MessageEmbeddingRepo,
	conversationRepo repository.ConversationRepo,
	cfg config.Config,
	rdb *redis.Client,
	log *slog.Logger,
) ChatService {
	return &chatService{
		db:                db,
		messageEmbedRepo:  messageEmbedRepo,
		messageRepo:       messageRepo,
		conversationRepo:  conversationRepo,
		appCredentialRepo: appCredentialRepo,
		cfg:               cfg,
		httpClient:        &http.Client{Timeout: 30 * time.Second},
		rdb:               rdb,
		log:               log,
	}
}

func (s *chatService) HandleInboundMessage(ctx context.Context, platform string, platformID string, event models.MessagingEvent) error {
	s.log.Info("inbound message received", "platform", platform, "page_id", platformID, "sender", event.Sender.ID)

	cred, err := s.appCredentialRepo.GetByPlatformAccountID(ctx, platformID)
	if err != nil {
		s.log.Error("credential lookup failed", "platform_id", platformID, "err", err)
		return fmt.Errorf("get credential for page %s: %w", platformID, err)
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

	user, err := s.fetchUserProfile(ctx, event.Sender.ID, pageToken)
	if err != nil {
		s.log.Error("fetch user profile failed", "sender_id", event.Sender.ID, "err", err)
		return fmt.Errorf("fetch user profile: %w", err)
	}
	s.log.Info("user profile fetched", "name", user.FirstName+" "+user.LastName)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	fullName := user.FirstName + " " + user.LastName
	conv, err := s.conversationRepo.FindOrCreate(ctx, tx, models.CreateConversation{
		BusinessID:       cred.BusinessID,
		Platform:         platform,
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
	s.log.Info("inbound message saved", "message_id", insertedMsg.ID)

	if _, err := s.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "chat:messages",
		Values: map[string]interface{}{
			"message_id":      insertedMsg.ID,
			"business_id":     insertedMsg.BusinessID,
			"conversation_id": insertedMsg.ConversationID,
			"platform":        platform,
		},
	}).Result(); err != nil {
		s.log.Error("publish to stream failed", "err", err)
		return fmt.Errorf("publish message to stream: %w", err)
	}
	s.log.Info("message published to stream", "message_id", insertedMsg.ID, "conversation_id", conv.ID)

	return tx.Commit()
}

func (s *chatService) FindOrCreate(ctx context.Context, tx *sqlx.Tx, conv models.CreateConversation) (*models.Conversation, error) {
	return s.conversationRepo.FindOrCreate(ctx, tx, conv)
}

type MessengerProfile struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	ProfilePic string `json:"profile_pic"`
	ID         string `json:"id"`
}

func (s *chatService) fetchUserProfile(ctx context.Context, psid, pageToken string) (*MessengerProfile, error) {
	params := url.Values{
		"fields":       {"first_name,last_name,profile_pic"},
		"access_token": {pageToken},
	}
	reqURL := fmt.Sprintf("https://graph.facebook.com/v25.0/%s?%s", psid, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build profile request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call profile endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("profile fetch failed (%d): %s", resp.StatusCode, body)
	}

	var profile MessengerProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("decode profile: %w", err)
	}
	return &profile, nil
}
