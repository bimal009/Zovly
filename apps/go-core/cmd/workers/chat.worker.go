package workers

import (
	"bytes"
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

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/utils"
	"github.com/jmoiron/sqlx"
	"github.com/pgvector/pgvector-go"
	"github.com/redis/go-redis/v9"
)

type AIWorker struct {
	chatService          service.ChatService
	messageRepo          repository.MessageRepo
	messageEmbeddingRepo repository.MessageEmbeddingRepo
	credentialRepo       repository.AppCredentialRepo
	db                   *sqlx.DB
	log                  *slog.Logger
	rdb                  *redis.Client
	httpClient           *http.Client
	cfg                  config.Config
}

func NewAIWorker(
	chatService service.ChatService,
	log *slog.Logger,
	rdb *redis.Client,
	messageRepo repository.MessageRepo,
	messageEmbeddingRepo repository.MessageEmbeddingRepo,
	credentialRepo repository.AppCredentialRepo,
	cfg config.Config,
	db *sqlx.DB,
) *AIWorker {
	return &AIWorker{
		chatService:          chatService,
		log:                  log,
		rdb:                  rdb,
		messageRepo:          messageRepo,
		messageEmbeddingRepo: messageEmbeddingRepo,
		credentialRepo:       credentialRepo,
		cfg:                  cfg,
		db:                   db,
		httpClient:           &http.Client{Timeout: 60 * time.Second},
	}
}

func (w *AIWorker) Run(ctx context.Context) {
	if err := w.rdb.XGroupCreateMkStream(ctx, "chat:messages", "ai-workers", "0").Err(); err != nil {
		w.log.Warn("consumer group setup", "info", err.Error())
	}

	w.log.Info("ai worker started", "stream", "chat:messages", "group", "ai-workers")

	for {
		select {
		case <-ctx.Done():
			w.log.Info("ai worker stopped")
			return
		default:
		}

		results, err := w.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    "ai-workers",
			Consumer: "worker-1",
			Streams:  []string{"chat:messages", ">"},
			Count:    10,
			Block:    5 * time.Second,
		}).Result()

		if err != nil {
			if err.Error() != "redis: nil" {
				w.log.Error("stream read failed", "err", err)
			}
			continue
		}

		if len(results) == 0 || len(results[0].Messages) == 0 {
			continue
		}

		w.log.Info("messages received from stream", "count", len(results[0].Messages))

		for _, msg := range results[0].Messages {
			if err := w.processAndReply(ctx, msg); err != nil {
				w.log.Error("process failed", "id", msg.ID, "err", err)
				continue
			}
			w.rdb.XAck(ctx, "chat:messages", "ai-workers", msg.ID)
		}
	}
}

func (w *AIWorker) processAndReply(ctx context.Context, msg redis.XMessage) error {
	conversationID, _ := msg.Values["conversation_id"].(string)
	businessID, _ := msg.Values["business_id"].(string)
	messageID, _ := msg.Values["message_id"].(string)

	if conversationID == "" || businessID == "" {
		return fmt.Errorf("malformed stream entry: missing ids")
	}

	w.log.Info("processing message", "conversation_id", conversationID, "message_id", messageID)

	lockKey := fmt.Sprintf("processing:%s", conversationID)
	locked, err := w.rdb.SetNX(ctx, lockKey, "1", 30*time.Second).Result()
	if err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	if !locked {
		w.log.Info("conversation already being processed, skipping", "conversation_id", conversationID)
		return nil
	}
	defer w.rdb.Del(ctx, lockKey)

	time.Sleep(12 * time.Second)

	unreplied, err := w.messageRepo.GetUnrepliedInbound(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("get unreplied inbound: %w", err)
	}
	if len(unreplied) == 0 {
		w.log.Info("no unreplied messages, skipping", "conversation_id", conversationID)
		return nil
	}
	w.log.Info("unreplied messages found", "conversation_id", conversationID, "count", len(unreplied))

	var parts []string
	for _, m := range unreplied {
		if m.Content != nil && *m.Content != "" {
			parts = append(parts, *m.Content)
		}
	}
	if len(parts) == 0 {
		w.log.Info("media-only messages, nothing to embed", "conversation_id", conversationID)
		return nil
	}
	combined := strings.Join(parts, "\n")

	const maxQueryChars = 2000
	searchText := combined
	if len(searchText) > maxQueryChars {
		searchText = searchText[:maxQueryChars]
	}

	w.log.Info("embedding message", "conversation_id", conversationID, "chars", len(searchText))
	queryVec, err := w.embedChat(ctx, searchText, "query")
	if err != nil {
		return fmt.Errorf("embed query: %w", err)
	}

	passageVec, err := w.embedChat(ctx, searchText, "passage")
	if err != nil {
		w.log.Error("embed passage failed", "conversation_id", conversationID, "err", err)
	} else {
		emb := models.CreateMessageEmbedding{
			MessageID:      messageID,
			BusinessID:     businessID,
			ConversationID: conversationID,
			Content:        combined,
			Embedding:      passageVec,
		}
		if err := w.messageEmbeddingRepo.CreateAndMarkVectorized(ctx, emb); err != nil {
			w.log.Error("store embedding failed", "conversation_id", conversationID, "err", err)
		} else {
			w.log.Info("embedding stored", "message_id", messageID)
		}
	}

	w.log.Info("fetching context", "conversation_id", conversationID, "business_id", businessID)

	knowledge, err := w.searchKnowledge(ctx, businessID, queryVec, 5)
	if err != nil {
		w.log.Error("knowledge search failed", "err", err)
	} else {
		w.log.Info("knowledge chunks", "count", len(knowledge))
	}

	pastChats, err := w.searchPastChats(ctx, businessID, conversationID, queryVec, 3)
	if err != nil {
		w.log.Error("past chat search failed", "err", err)
	} else {
		w.log.Info("past chat chunks", "count", len(pastChats))
	}

	business, err := w.getBusiness(ctx, businessID)
	if err != nil {
		w.log.Error("business fetch failed", "err", err)
	}

	history, err := w.getConversationHistory(ctx, conversationID, 10)
	if err != nil {
		w.log.Error("conversation history fetch failed", "err", err)
	} else {
		w.log.Info("history loaded", "conversation_id", conversationID, "messages", len(history))
	}

	customer, err := w.getCustomerProfile(ctx, conversationID)
	if err != nil {
		w.log.Error("customer profile fetch failed", "err", err)
	} else {
		w.log.Info("customer loaded", "contact_id", customer.ContactID, "platform", customer.Platform)
	}

	w.log.Info("handing off to trySend", "conversation_id", conversationID, "platform", customer.Platform)
	return w.trySend(ctx, customer.Platform, businessID, conversationID, customer.ContactID, combined, knowledge, pastChats, business, history, customer)
}

func (w *AIWorker) callChatReply(
	ctx context.Context,
	businessID, message string,
	knowledge []models.KnowledgeChunk,
	pastChats []models.PastChatChunk,
	business *models.Business,
	history []models.Message,
	customer *models.Conversation,
) (string, error) {
	payload := models.ChatReplyPayload{
		BusinessID: businessID,
		Message:    message,
		Context: models.ChatContext{
			Knowledge:        knowledge,
			PastChats:        pastChats,
			Business:         business,
			PastConversation: history,
			Customer:         customer,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal chat reply request: %w", err)
	}

	reqURL := w.cfg.App.AIServiceURL + "/api/v1/ml/chat/reply"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build chat reply request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call chat reply service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("chat reply service returned status %d", resp.StatusCode)
	}

	var result models.ChatReplyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode chat reply response: %w", err)
	}
	return result.Reply, nil
}

func (w *AIWorker) embedChat(ctx context.Context, message, kind string) (pgvector.Vector, error) {
	body, err := json.Marshal(map[string]string{"message": message, "kind": kind})
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("marshal embed request: %w", err)
	}

	reqURL := w.cfg.App.AIServiceURL + "/api/v1/ml/chat/embed"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("build embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("call embed service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return pgvector.Vector{}, fmt.Errorf("embed service returned status %d", resp.StatusCode)
	}

	var result models.ChatEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return pgvector.Vector{}, fmt.Errorf("decode embed response: %w", err)
	}
	if len(result.Embeddings) == 0 {
		return pgvector.Vector{}, fmt.Errorf("embed service returned no embeddings")
	}
	return pgvector.NewVector(result.Embeddings[0].Embedding), nil
}

func (w *AIWorker) searchKnowledge(ctx context.Context, businessID string, vec pgvector.Vector, k int) ([]models.KnowledgeChunk, error) {
	query := `
		SELECT content, source_type, 1 - (embedding <=> $1) AS score
		FROM knowledge_chunks
		WHERE business_id = $2
		ORDER BY embedding <=> $1
		LIMIT $3`

	var results []models.KnowledgeChunk
	if err := w.db.SelectContext(ctx, &results, query, vec, businessID, k); err != nil {
		return nil, err
	}

	filtered := make([]models.KnowledgeChunk, 0, len(results))
	for _, r := range results {
		if r.Score >= 0.80 {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

func (w *AIWorker) searchPastChats(ctx context.Context, businessID, conversationID string, vec pgvector.Vector, k int) ([]models.PastChatChunk, error) {
	// exclude the current conversation — those messages belong in history, not "similar past chats"
	query := `
		SELECT content, conversation_id, 1 - (embedding <=> $1) AS score
		FROM message_embeddings
		WHERE business_id = $2 AND conversation_id != $3
		ORDER BY embedding <=> $1
		LIMIT $4`

	var results []models.PastChatChunk
	if err := w.db.SelectContext(ctx, &results, query, vec, businessID, conversationID, k); err != nil {
		return nil, err
	}

	filtered := make([]models.PastChatChunk, 0, len(results))
	for _, r := range results {
		if r.Score >= 0.85 { // higher bar for past chats than knowledge
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

func (w *AIWorker) getBusiness(ctx context.Context, businessID string) (*models.Business, error) {
	var b models.Business
	if err := w.db.GetContext(ctx, &b, `SELECT * FROM business WHERE id = $1`, businessID); err != nil {
		return nil, err
	}
	return &b, nil
}

func (w *AIWorker) getCustomerProfile(ctx context.Context, conversationID string) (*models.Conversation, error) {
	var c models.Conversation
	if err := w.db.GetContext(ctx, &c, `SELECT * FROM conversations WHERE id = $1`, conversationID); err != nil {
		return nil, err
	}
	return &c, nil
}

func (w *AIWorker) getConversationHistory(ctx context.Context, conversationID string, n int) ([]models.Message, error) {
	// newest-first from DB, then reverse to chronological for the LLM
	query := `
		SELECT * FROM messages
		WHERE conversation_id = $1 AND content IS NOT NULL
		ORDER BY sent_at DESC
		LIMIT $2`

	var results []models.Message
	if err := w.db.SelectContext(ctx, &results, query, conversationID, n); err != nil {
		return nil, err
	}

	// reverse to chronological (oldest first)
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}
	return results, nil
}
func (w *AIWorker) trySend(
	ctx context.Context,
	platform, businessID, conversationID, recipientID, message string,
	knowledge []models.KnowledgeChunk,
	pastChats []models.PastChatChunk,
	business *models.Business,
	history []models.Message,
	customer *models.Conversation,
) error {
	rateLimitKey := fmt.Sprintf("rate:dm:%s", businessID)
	count, _ := w.rdb.Incr(ctx, rateLimitKey).Result()
	if count == 1 {
		w.rdb.Expire(ctx, rateLimitKey, time.Hour)
	}
	if count > 200 {
		w.rdb.Decr(ctx, rateLimitKey)
		w.log.Warn("rate limited", "conversation_id", conversationID, "count", count)
		return nil
	}

	w.log.Info("calling AI service", "conversation_id", conversationID, "business_id", businessID)
	reply, err := w.callChatReply(ctx, businessID, message, knowledge, pastChats, business, history, customer)
	if err != nil {
		w.rdb.Decr(ctx, rateLimitKey)
		// return error so the stream message is NOT acked and will be retried
		return fmt.Errorf("chat reply: %w", err)
	}
	w.log.Info("AI reply received", "conversation_id", conversationID, "reply_len", len(reply))

	pending := models.MessageStatusPending
	aiSender := models.MessageSenderAI
	tx, err := w.db.BeginTxx(ctx, nil)
	if err != nil {
		w.log.Error("begin tx failed", "conversation_id", conversationID, "err", err)
		return fmt.Errorf("begin tx: %w", err)
	}
	outbound, err := w.messageRepo.Create(ctx, tx, models.CreateMessage{
		ConversationID: conversationID,
		BusinessID:     businessID,
		Direction:      models.MessageDirectionOut,
		SentBy:         &aiSender,
		Content:        &reply,
		Status:         &pending,
	})
	if err != nil {
		tx.Rollback()
		w.log.Error("create outbound message failed", "conversation_id", conversationID, "err", err)
		return fmt.Errorf("create outbound: %w", err)
	}
	if err := tx.Commit(); err != nil {
		w.log.Error("commit outbound message failed", "conversation_id", conversationID, "err", err)
		return fmt.Errorf("commit outbound: %w", err)
	}
	w.log.Info("outbound message saved", "message_id", outbound.ID, "conversation_id", conversationID)

	// from here on: outbound row exists, so mark failed rather than retrying
	w.log.Info("fetching credentials", "business_id", businessID, "platform", platform)
	creds, err := w.credentialRepo.ListByApp(ctx, businessID, platform)
	if err != nil || len(creds) == 0 {
		w.log.Error("no credential found", "business_id", businessID, "platform", platform)
		errMsg := "no credential found"
		w.messageRepo.UpdateStatus(ctx, outbound.ID, models.MessageStatusFailed, "", &errMsg)
		return nil
	}
	encKey, err := base64.StdEncoding.DecodeString(w.cfg.App.EncryptionKey)
	if err != nil {
		w.log.Error("decode encryption key failed", "err", err)
		errMsg := "decrypt failed"
		w.messageRepo.UpdateStatus(ctx, outbound.ID, models.MessageStatusFailed, "", &errMsg)
		return nil
	}
	token, err := utils.Decrypt(*creds[0].AccessToken, encKey)
	if err != nil {
		w.log.Error("decrypt token failed", "err", err)
		errMsg := "decrypt failed"
		w.messageRepo.UpdateStatus(ctx, outbound.ID, models.MessageStatusFailed, "", &errMsg)
		return nil
	}

	w.log.Info("sending reply via platform", "platform", platform, "recipient_id", recipientID)
	platformMsgID, err := w.sendReply(ctx, platform, token, recipientID, reply)
	if err != nil {
		w.log.Error("platform send failed", "platform", platform, "conversation_id", conversationID, "err", err)
		errMsg := err.Error()
		w.messageRepo.UpdateStatus(ctx, outbound.ID, models.MessageStatusFailed, "", &errMsg)
		w.rdb.Decr(ctx, rateLimitKey)
		return nil
	}

	w.messageRepo.UpdateStatus(ctx, outbound.ID, models.MessageStatusSent, platformMsgID, nil)
	w.log.Info("ai reply sent", "conversation_id", conversationID, "platform_message_id", platformMsgID)
	return nil
}

type graphSendResponse struct {
	RecipientID string `json:"recipient_id"`
	MessageID   string `json:"message_id"`
	Error       *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

func (w *AIWorker) sendMessengerReply(ctx context.Context, pageToken, recipientID, text string) (string, error) {
	payload := map[string]interface{}{
		"recipient":      map[string]string{"id": recipientID},
		"message":        map[string]string{"text": text},
		"messaging_type": "RESPONSE", // replying within the 24h window
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal send payload: %w", err)
	}

	reqURL := "https://graph.facebook.com/v25.0/me/messages?access_token=" + url.QueryEscape(pageToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build send request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call send endpoint: %w", err)
	}
	defer resp.Body.Close()

	var result graphSendResponse
	rawBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(rawBody, &result)

	if resp.StatusCode != http.StatusOK {
		if result.Error != nil {
			return "", fmt.Errorf("messenger send failed (code %d): %s", result.Error.Code, result.Error.Message)
		}
		return "", fmt.Errorf("messenger send failed (%d): %s", resp.StatusCode, rawBody)
	}

	return result.MessageID, nil
}

// sendInstagramReply sends a text reply via Instagram DM.
func (w *AIWorker) sendInstagramReply(ctx context.Context, igToken, recipientID, text string) (string, error) {
	payload := map[string]interface{}{
		"recipient": map[string]string{"id": recipientID},
		"message":   map[string]string{"text": text},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal send payload: %w", err)
	}

	reqURL := "https://graph.instagram.com/v25.0/me/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build send request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+igToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call send endpoint: %w", err)
	}
	defer resp.Body.Close()

	var result graphSendResponse
	rawBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(rawBody, &result)

	if resp.StatusCode != http.StatusOK {
		if result.Error != nil {
			return "", fmt.Errorf("instagram send failed (code %d): %s", result.Error.Code, result.Error.Message)
		}
		return "", fmt.Errorf("instagram send failed (%d): %s", resp.StatusCode, rawBody)
	}

	return result.MessageID, nil
}

func (w *AIWorker) sendReply(ctx context.Context, platform, token, recipientID, text string) (string, error) {
	switch platform {
	case "facebook":
		return w.sendMessengerReply(ctx, token, recipientID, text)
	case "instagram":
		return w.sendInstagramReply(ctx, token, recipientID, text)
	default:
		return "", fmt.Errorf("unsupported platform: %s", platform)
	}
}
