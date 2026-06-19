package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
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
}

func NewChatService(
	db *sqlx.DB,
	messageRepo repository.MessageRepo,
	appCredentialRepo repository.AppCredentialRepo,
	messageEmbedRepo repository.MessageEmbeddingRepo,
	conversationRepo repository.ConversationRepo,
	cfg config.Config,
	rdb *redis.Client,
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
	}
}

func (s *chatService) HandleInboundMessage(ctx context.Context, platform string, platformID string, event models.MessagingEvent) error {
	cred, err := s.appCredentialRepo.GetByPlatformAccountID(ctx, platformID)
	if err != nil {
		return fmt.Errorf("get credential for page %s: %w", platformID, err)
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	conv, err := s.conversationRepo.FindOrCreate(ctx, tx, models.CreateConversation{
		BusinessID: cred.BusinessID,
		Platform:   "facebook",
		ThreadID:   event.Sender.ID,
		ContactID:  event.Sender.ID,
	})
	if err != nil {
		return fmt.Errorf("find or create conversation: %w", err)
	}

	text := event.Message.Text
	newMessage := models.CreateMessage{
		ConversationID: conv.ID,
		BusinessID:     cred.BusinessID,
		Direction:      models.MessageDirectionIn,
		SentBy:         nil,
		Content:        &text,
		Status:         nil,
	}

	insertedMsg, err := s.messageRepo.Create(ctx, tx, newMessage)
	if err != nil {
		return fmt.Errorf("create message: %w", err)
	}

	if _, err := s.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "chat:messages",
		Values: map[string]interface{}{
			"message_id":      insertedMsg.ID,
			"business_id":     insertedMsg.BusinessID,
			"conversation_id": insertedMsg.ConversationID,
		},
	}).Result(); err != nil {
		return fmt.Errorf("publish message to stream: %w", err)
	}

	return tx.Commit()
}

func (s *chatService) FindOrCreate(ctx context.Context, tx *sqlx.Tx, conv models.CreateConversation) (*models.Conversation, error) {
	return s.conversationRepo.FindOrCreate(ctx, tx, conv)
}

type chatEmbedResponse struct {
	Embeddings []models.FaqChunksResponse `json:"embeddings"`
}

func (s *chatService) embedChat(ctx context.Context, message string) ([]models.FaqChunksResponse, error) {
	body, err := json.Marshal(map[string]string{"message": message})
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	url := s.cfg.App.AIServiceURL + "/api/v1/ml/chat/embed"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call embed service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed service returned status %d", resp.StatusCode)
	}

	var result chatEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}
	return result.Embeddings, nil
}

// func (w *chatService) processAndReply(ctx context.Context, msg redis.XMessage) error {
// 	conversationID := msg.Values["conversation_id"].(string)
// 	businessID := msg.Values["business_id"].(string)

// 	lockKey := fmt.Sprintf("processing:%s", conversationID)
// 	locked, _ := w.rdb.SetNX(ctx, lockKey, "1", 30*time.Second).Result()
// 	if !locked {
// 		return nil
// 	}
// 	defer w.rdb.Del(ctx, lockKey)

// 	time.Sleep(12 * time.Second)

// 	unreplied, err := w.messageRepo.GetPendingOutbound(ctx)
// 	if err != nil || len(unreplied) == 0 {
// 		return err
// 	}

// 	var parts []string
// 	for _, m := range unreplied {
// 		if m.Content != nil {
// 			parts = append(parts, *m.Content)
// 		}
// 	}
// 	combined := strings.Join(parts, "\n")

// 	// // generate reply — always do this regardless of rate limit
// 	// reply, err := w.pyml.Chat(ctx, businessID, combined)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	// write outbound message as pending
// 	pending := models.MessageStatusPending
// 	outbound, err := w.messageRepo.Create(ctx, nil, models.CreateMessage{
// 		ConversationID: conversationID,
// 		BusinessID:     businessID,
// 		Direction:      models.MessageDirectionOut,
// 		SentBy:         ptr(models.MessageSenderAI),
// 		Content:        &reply,
// 		Status:         &pending,
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	// NOW check rate limit before sending
// 	w.trySend(ctx, outbound, businessID)
// 	return nil
// }

// func (w *AIWorker) trySend(ctx context.Context, msg *models.Message, businessID string) {
// 	rateLimitKey := fmt.Sprintf("rate:dm:%s", businessID)

// 	count, _ := w.rdb.Incr(ctx, rateLimitKey).Result()
// 	if count == 1 {
// 		w.rdb.Expire(ctx, rateLimitKey, time.Hour)
// 	}

// 	if count > 200 {
// 		// over limit — decrement back, message stays pending
// 		w.rdb.Decr(ctx, rateLimitKey)
// 		w.log.Warn("rate limited, reply queued", "message_id", msg.ID)
// 		return
// 	}

// 	// under limit — send via Graph API
// 	platformMsgID, err := w.sendViaGraphAPI(ctx, msg)
// 	if err != nil {
// 		w.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusFailed, "", ptr(err.Error()))
// 		w.rdb.Decr(ctx, rateLimitKey)
// 		return
// 	}

// 	w.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusSent, platformMsgID, nil)
// }
