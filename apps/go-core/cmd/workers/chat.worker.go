package workers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/embed"
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
	productService       service.ProductService
	messageRepo          repository.MessageRepo
	messageEmbeddingRepo repository.MessageEmbeddingRepo
	credentialRepo       repository.AppCredentialRepo
	db                   *sqlx.DB
	log                  *slog.Logger
	rdb                  *redis.Client
	httpClient           *http.Client
	cfg                  config.Config
	embedder             *embed.Client
}

func NewAIWorker(
	chatService service.ChatService,
	productService service.ProductService,
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
		productService:       productService,
		log:                  log,
		rdb:                  rdb,
		messageRepo:          messageRepo,
		messageEmbeddingRepo: messageEmbeddingRepo,
		credentialRepo:       credentialRepo,
		cfg:                  cfg,
		db:                   db,
		httpClient:           &http.Client{Timeout: 60 * time.Second},
		embedder:             embed.New(cfg.App.AIServiceURL),
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
			if !errors.Is(err, redis.Nil) {
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

	// debounce lock — TTL generously exceeds worst-case processing
	// (15s sleep + media resolution + embed + search + Claude)
	lockKey := fmt.Sprintf("processing:%s", conversationID)
	locked, err := w.rdb.SetNX(ctx, lockKey, "1", 90*time.Second).Result()
	if err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	if !locked {
		w.log.Info("conversation already being processed, skipping", "conversation_id", conversationID)
		return nil
	}
	defer w.rdb.Del(ctx, lockKey)

	// debounce — wait for more messages in the burst (same window for all types,
	// since media content is resolved below rather than waited-for)
	time.Sleep(15 * time.Second)

	unreplied, err := w.messageRepo.GetUnrepliedInbound(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("get unreplied inbound: %w", err)
	}
	if len(unreplied) == 0 {
		w.log.Info("no unreplied messages, skipping", "conversation_id", conversationID)
		return nil
	}
	w.log.Info("unreplied messages found", "conversation_id", conversationID, "count", len(unreplied))

	// resolve media → text for rows the webhook stored without content.
	// This is where image-describe / audio-transcribe happens (async, off the
	// webhook path), and the result is persisted so retries don't reprocess.
	for i := range unreplied {
		m := &unreplied[i]
		if (m.Content != nil && *m.Content != "") || m.MediaUrl == nil || m.MediaType == nil {
			continue
		}
		var text string
		switch *m.MediaType {
		case models.MediaTypeImage:
			if d, derr := w.chatService.GetImageDetails(ctx, *m.MediaUrl); derr == nil {
				text = d
			} else {
				w.log.Warn("describe image failed", "message_id", m.ID, "err", derr)
			}
		case models.MediaTypeAudio:
			if t, terr := w.chatService.GetAudioDetails(ctx, *m.MediaUrl); terr == nil {
				text = t
			} else {
				w.log.Warn("transcribe audio failed", "message_id", m.ID, "err", terr)
			}
		}
		if text != "" {
			m.Content = &text
			if err := w.messageRepo.UpdateContent(ctx, m.ID, text); err != nil {
				w.log.Error("persist media content failed", "message_id", m.ID, "err", err)
			}
		}
	}

	var parts []string
	for _, m := range unreplied {
		if m.Content != nil && *m.Content != "" {
			parts = append(parts, *m.Content)
		}
	}
	if len(parts) == 0 {
		w.log.Info("no resolvable content, skipping", "conversation_id", conversationID)
		return nil
	}
	combined := strings.Join(parts, "\n")

	const maxQueryChars = 2000
	searchText := combined
	if len(searchText) > maxQueryChars {
		searchText = searchText[:maxQueryChars]
	}

	w.log.Info("embedding message", "conversation_id", conversationID, "chars", len(searchText))
	queryVec, err := w.embedder.EmbedChat(ctx, searchText, "query")
	if err != nil {
		return fmt.Errorf("embed query: %w", err)
	}

	passageVec, err := w.embedder.EmbedChat(ctx, searchText, "passage")
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
		// can't send a reply without the platform + recipient id — fail (retryable)
		w.log.Error("customer profile fetch failed", "err", err)
		return fmt.Errorf("customer profile required: %w", err)
	}
	w.log.Info("customer loaded", "contact_id", customer.ContactID, "platform", customer.Platform)

	// the product currently under discussion (if any, cached in Redis with 24h
	// TTL) — lets the AI resolve follow-ups like "how much?" without re-searching
	var activeProductID string
	if activeProduct, aerr := w.productService.GetActiveProduct(ctx, businessID, conversationID); aerr != nil {
		w.log.Error("active product fetch failed", "conversation_id", conversationID, "err", aerr)
	} else if activeProduct != nil {
		activeProductID = activeProduct.ID
		w.log.Info("active product loaded", "conversation_id", conversationID, "active_product_id", activeProductID)
	}

	w.log.Info("handing off to trySend", "conversation_id", conversationID, "platform", customer.Platform)
	return w.trySend(ctx, customer.Platform, businessID, conversationID, customer.ContactID, combined, activeProductID, knowledge, pastChats, business, history, customer)
}

func (w *AIWorker) callChatReply(
	ctx context.Context,
	businessID, conversationID, message, activeProductID string,
	knowledge []models.KnowledgeChunk,
	pastChats []models.PastChatChunk,
	business *models.Business,
	history []models.Message,
	customer *models.Conversation,
) (string, []string, error) {
	payload := models.ChatReplyPayload{
		BusinessID:      businessID,
		ConversationID:  conversationID,
		Message:         message,
		ActiveProductID: activeProductID,
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
		return "", nil, fmt.Errorf("marshal chat reply request: %w", err)
	}

	reqURL := w.cfg.App.AIServiceURL + "/api/v1/ml/chat/reply"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return "", nil, fmt.Errorf("build chat reply request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("call chat reply service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("chat reply service returned status %d", resp.StatusCode)
	}

	var result models.ChatReplyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", nil, fmt.Errorf("decode chat reply response: %w", err)
	}
	return result.Reply, result.Images, nil
}

func (w *AIWorker) searchKnowledge(ctx context.Context, businessID string, vec pgvector.Vector, k int) ([]models.KnowledgeChunk, error) {
	query := `
		SELECT content, source_type, source_id, 1 - (embedding <=> $1) AS score
		FROM knowledge_chunks
		WHERE business_id = $2
		ORDER BY embedding <=> $1
		LIMIT $3`

	var results []models.KnowledgeChunk
	if err := w.db.SelectContext(ctx, &results, query, vec, businessID, k); err != nil {
		return nil, err
	}

	for _, r := range results {
		w.log.Info("knowledge candidate", "score", r.Score, "type", r.SourceType,
			"content", r.Content[:min(50, len(r.Content))])
	}

	filtered := make([]models.KnowledgeChunk, 0, len(results))
	for _, r := range results {
		if r.Score >= 0.74 {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

func (w *AIWorker) searchPastChats(ctx context.Context, businessID, conversationID string, vec pgvector.Vector, k int) ([]models.PastChatChunk, error) {
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
		if r.Score >= 0.85 {
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
	query := `
		SELECT * FROM (
			SELECT * FROM messages
			WHERE conversation_id = $1 AND content IS NOT NULL
			ORDER BY sent_at DESC
			LIMIT $2
		) sub
		ORDER BY sent_at ASC`

	var results []models.Message
	if err := w.db.SelectContext(ctx, &results, query, conversationID, n); err != nil {
		return nil, err
	}
	return results, nil
}

func (w *AIWorker) trySend(
	ctx context.Context,
	platform, businessID, conversationID, recipientID, message, activeProductID string,
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
	reply, images, err := w.callChatReply(ctx, businessID, conversationID, message, activeProductID, knowledge, pastChats, business, history, customer)
	if err != nil {
		w.rdb.Decr(ctx, rateLimitKey)
		return fmt.Errorf("chat reply: %w", err) // not acked → retried
	}
	w.log.Info("AI reply received", "conversation_id", conversationID, "reply_len", len(reply), "images", len(images), "msg", reply)

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

	// send any product images the AI included, each as its own message
	for _, imgURL := range images {
		w.sendImageMessage(ctx, platform, businessID, conversationID, recipientID, token, imgURL)
	}

	return nil
}

// sendImageMessage persists and delivers a single image attachment. Failures are
// logged and recorded on the message row but never abort the reply — the text
// has already been sent successfully.
func (w *AIWorker) sendImageMessage(ctx context.Context, platform, businessID, conversationID, recipientID, token, imgURL string) {
	imageType := models.MediaTypeImage
	sent := models.MessageStatusSent
	aiSender := models.MessageSenderAI

	tx, err := w.db.BeginTxx(ctx, nil)
	if err != nil {
		w.log.Error("begin tx for image failed", "conversation_id", conversationID, "err", err)
		return
	}
	pending := models.MessageStatusPending
	imgMsg, err := w.messageRepo.Create(ctx, tx, models.CreateMessage{
		ConversationID: conversationID,
		BusinessID:     businessID,
		Direction:      models.MessageDirectionOut,
		SentBy:         &aiSender,
		MediaUrl:       &imgURL,
		MediaType:      &imageType,
		Status:         &pending,
	})
	if err != nil {
		tx.Rollback()
		w.log.Error("create outbound image message failed", "conversation_id", conversationID, "err", err)
		return
	}
	if err := tx.Commit(); err != nil {
		w.log.Error("commit outbound image message failed", "conversation_id", conversationID, "err", err)
		return
	}

	w.log.Info("sending image via platform", "platform", platform, "recipient_id", recipientID, "url", imgURL)
	platformMsgID, err := w.sendImage(ctx, platform, token, recipientID, imgURL)
	if err != nil {
		w.log.Error("platform image send failed", "platform", platform, "conversation_id", conversationID, "err", err)
		errMsg := err.Error()
		w.messageRepo.UpdateStatus(ctx, imgMsg.ID, models.MessageStatusFailed, "", &errMsg)
		return
	}
	w.messageRepo.UpdateStatus(ctx, imgMsg.ID, sent, platformMsgID, nil)
	w.log.Info("ai image sent", "conversation_id", conversationID, "platform_message_id", platformMsgID)
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
		"messaging_type": "RESPONSE",
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

func imageAttachmentMessage(imageURL string) map[string]interface{} {
	return map[string]interface{}{
		"attachment": map[string]interface{}{
			"type": "image",
			"payload": map[string]interface{}{
				"url":         imageURL,
				"is_reusable": true,
			},
		},
	}
}

func (w *AIWorker) sendMessengerImage(ctx context.Context, pageToken, recipientID, imageURL string) (string, error) {
	payload := map[string]interface{}{
		"recipient":      map[string]string{"id": recipientID},
		"message":        imageAttachmentMessage(imageURL),
		"messaging_type": "RESPONSE",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal image payload: %w", err)
	}

	reqURL := "https://graph.facebook.com/v25.0/me/messages?access_token=" + url.QueryEscape(pageToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build image request: %w", err)
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
			return "", fmt.Errorf("messenger image send failed (code %d): %s", result.Error.Code, result.Error.Message)
		}
		return "", fmt.Errorf("messenger image send failed (%d): %s", resp.StatusCode, rawBody)
	}
	return result.MessageID, nil
}

func (w *AIWorker) sendInstagramImage(ctx context.Context, igToken, recipientID, imageURL string) (string, error) {
	payload := map[string]interface{}{
		"recipient": map[string]string{"id": recipientID},
		"message":   imageAttachmentMessage(imageURL),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal image payload: %w", err)
	}

	reqURL := "https://graph.instagram.com/v25.0/me/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build image request: %w", err)
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
			return "", fmt.Errorf("instagram image send failed (code %d): %s", result.Error.Code, result.Error.Message)
		}
		return "", fmt.Errorf("instagram image send failed (%d): %s", resp.StatusCode, rawBody)
	}
	return result.MessageID, nil
}

func (w *AIWorker) sendImage(ctx context.Context, platform, token, recipientID, imageURL string) (string, error) {
	switch platform {
	case "facebook":
		return w.sendMessengerImage(ctx, token, recipientID, imageURL)
	case "instagram":
		return w.sendInstagramImage(ctx, token, recipientID, imageURL)
	default:
		return "", fmt.Errorf("unsupported platform: %s", platform)
	}
}
