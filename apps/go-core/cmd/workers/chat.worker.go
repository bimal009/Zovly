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
	"unicode"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/embed"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/internal/task"
	"github.com/bimal009/Zovly/pkg/utils"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"github.com/pgvector/pgvector-go"
	"github.com/redis/go-redis/v9"
)

const (
	knowledgeScoreThreshold = 0.78
	pastChatScoreThreshold  = 0.86
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
	queue                *task.Client // for re-scheduling the debounce-gap drain
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
	q *task.Client,
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
		queue:                q,
	}
}

// Register wires this handler onto an asynq server. Call once at startup.
func (w *AIWorker) Register(srv *task.Server) {
	srv.Register(task.TypeChatReply, w.HandleChatReply)
}

// HandleChatReply is the asynq handler. asynq calls it once per debounced task,
// on its own goroutine, with a context already bounded by the task timeout.
//
// Error contract:
//   - return nil               → task done, removed.
//   - return err               → asynq retries with backoff; after MaxRetry the
//     task is archived (the dead-letter equivalent, visible in asynqmon).
//   - return ...asynq.SkipRetry → permanent failure, archived immediately.
func (w *AIWorker) HandleChatReply(ctx context.Context, t *asynq.Task) error {
	p, err := task.ParseChatReplyPayload(t)
	if err != nil {
		// Unusable payload — never going to succeed, so skip retries.
		return fmt.Errorf("%v: %w", err, asynq.SkipRetry)
	}
	if p.BusinessID == "" || p.ConversationID == "" {
		return fmt.Errorf("missing ids in payload: %w", asynq.SkipRetry)
	}

	conversationID := p.ConversationID
	businessID := p.BusinessID
	w.log.Info("processing chat reply", "conversation_id", conversationID)

	unreplied, err := w.messageRepo.GetUnrepliedInbound(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("get unreplied inbound: %w", err)
	}
	if len(unreplied) == 0 {
		w.log.Debug("no unreplied messages, nothing to do", "conversation_id", conversationID)
		return nil
	}

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
			if tr, terr := w.chatService.GetAudioDetails(ctx, *m.MediaUrl); terr == nil {
				text = tr
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
		w.log.Debug("no resolvable content, nothing to do", "conversation_id", conversationID)
		return nil
	}
	combined := strings.Join(parts, "\n")

	// Newest message + its id: the cutoff detects inbound that lands mid-process,
	// and the id keys the stored passage embedding.
	var cutoff time.Time
	var newestMessageID string
	for _, m := range unreplied {
		if m.SentAt.After(cutoff) {
			cutoff = m.SentAt
			newestMessageID = m.ID
		}
	}

	const maxQueryChars = 2000
	searchText := combined
	if len(searchText) > maxQueryChars {
		searchText = searchText[:maxQueryChars]
	}

	retrievalText := lastSubstantiveMessage(unreplied)

	// Store the passage embedding for the cross-conversation past-chat corpus —
	// only for substantive bursts, so greeting/emoji-only turns don't pollute it.
	if retrievalText != "" {
		if passageVec, perr := w.embedder.EmbedChat(ctx, searchText, "passage"); perr != nil {
			w.log.Error("embed passage failed", "conversation_id", conversationID, "err", perr)
		} else {
			emb := models.CreateMessageEmbedding{
				MessageID:      newestMessageID,
				BusinessID:     businessID,
				ConversationID: conversationID,
				Content:        combined,
				Embedding:      passageVec,
			}
			if err := w.messageEmbeddingRepo.CreateAndMarkVectorized(ctx, emb); err != nil {
				w.log.Error("store embedding failed", "conversation_id", conversationID, "err", err)
			}
		}
	}

	var activeProductID string
	if activeProduct, aerr := w.productService.GetActiveProduct(ctx, businessID, conversationID); aerr != nil {
		w.log.Error("active product fetch failed", "conversation_id", conversationID, "err", aerr)
	} else if activeProduct != nil {
		activeProductID = activeProduct.ID
	}

	if retrievalText == "" {
		retrievalText = searchText
	}
	if len(retrievalText) > maxQueryChars {
		retrievalText = retrievalText[:maxQueryChars]
	}

	queryVec, err := w.embedder.EmbedChat(ctx, retrievalText, "query")
	if err != nil {
		return fmt.Errorf("embed query: %w", err)
	}

	knowledge, err := w.searchKnowledge(ctx, businessID, queryVec, 5)
	if err != nil {
		w.log.Error("knowledge search failed", "err", err)
	}

	if products, perr := w.searchProducts(ctx, businessID, retrievalText, queryVec, 10); perr != nil {
		w.log.Error("product search failed", "err", perr)
	} else {
		knowledge = append(knowledge, products...)
	}

	pastChats, err := w.searchPastChats(ctx, businessID, conversationID, queryVec, 3)
	if err != nil {
		w.log.Error("past chat search failed", "err", err)
	}

	business, err := w.getBusiness(ctx, businessID)
	if err != nil {
		w.log.Error("business fetch failed", "err", err)
	}

	history, err := w.getConversationHistory(ctx, conversationID, 10)
	if err != nil {
		w.log.Error("conversation history fetch failed", "err", err)
	}

	customer, err := w.getCustomerProfile(ctx, conversationID)
	if err != nil {
		// Can't send without platform + recipient id — transient, so retry.
		return fmt.Errorf("customer profile required: %w", err)
	}

	if err := w.trySend(ctx, customer.Platform, businessID, conversationID, customer.ContactID, combined, activeProductID, knowledge, pastChats, business, history, customer); err != nil {
		return err
	}

	// Debounce gap: a message can arrive after we read unreplied but before this
	// task finishes; its webhook enqueue hit the still-active TaskID and was
	// coalesced away. Re-check; if anything newer than `cutoff` landed, schedule a
	// drain pass (separate task ID, so it enqueues even though we're still active).
	if remaining, rerr := w.messageRepo.GetUnrepliedInbound(ctx, conversationID); rerr != nil {
		w.log.Error("debounce re-check failed", "conversation_id", conversationID, "err", rerr)
	} else {
		for _, m := range remaining {
			if m.SentAt.After(cutoff) {
				w.log.Info("new inbound during processing — scheduling drain", "conversation_id", conversationID)
				if eerr := w.queue.EnqueueReplyDrain(ctx, businessID, conversationID); eerr != nil {
					w.log.Error("drain enqueue failed", "conversation_id", conversationID, "err", eerr)
				}
				break
			}
		}
	}
	return nil
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
	// Over-fetch so that after de-duping by source_id we still have k distinct
	// items. Products are excluded — they come from the dedicated hybrid path in
	// searchProducts so the two routes can't disagree. This covers FAQ/policy/post.
	const preDedupLimit = 15
	query := `
		SELECT content, source_type, source_id, 1 - (embedding <=> $1) AS score
		FROM knowledge_chunks
		WHERE business_id = $2 AND source_type != 'product'
		ORDER BY embedding <=> $1
		LIMIT $3`

	var results []models.KnowledgeChunk
	if err := w.db.SelectContext(ctx, &results, query, vec, businessID, preDedupLimit); err != nil {
		return nil, err
	}

	if w.log.Enabled(ctx, slog.LevelDebug) {
		for _, r := range results {
			w.log.Debug("knowledge candidate", "score", r.Score, "type", r.SourceType,
				"content", r.Content[:min(50, len(r.Content))])
		}
	}

	seen := make(map[string]struct{}, len(results))
	filtered := make([]models.KnowledgeChunk, 0, k)
	for _, r := range results {
		if r.Score < knowledgeScoreThreshold {
			continue
		}
		key := r.SourceType + ":" + r.SourceID
		if r.SourceID != "" {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
		}
		filtered = append(filtered, r)
		if len(filtered) >= k {
			break
		}
	}
	return filtered, nil
}

func isEmptyOrEmoji(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

func lastSubstantiveMessage(msgs []models.Message) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Content == nil {
			continue
		}
		c := strings.TrimSpace(*msgs[i].Content)
		if c == "" || isEmptyOrEmoji(c) {
			continue
		}
		return c
	}
	return ""
}

func (w *AIWorker) searchProducts(ctx context.Context, businessID, queryText string, queryVec pgvector.Vector, k int) ([]models.KnowledgeChunk, error) {
	candidates, err := w.productService.HybridSearch(ctx, businessID, queryText, queryVec)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	ordered := w.rerankProducts(ctx, queryText, candidates)

	byID := make(map[string]models.ProductSearchCandidate, len(candidates))
	for _, c := range candidates {
		byID[c.SourceID] = c
	}

	out := make([]models.KnowledgeChunk, 0, k)
	for i, id := range ordered {
		if len(out) >= k {
			break
		}
		c, ok := byID[id]
		if !ok {
			continue
		}
		out = append(out, models.KnowledgeChunk{
			Content:    c.Passage,
			SourceType: string(models.SourceProduct),
			SourceID:   c.SourceID,
			Score:      1.0 - float64(i)*0.001,
		})
	}
	return out, nil
}

func (w *AIWorker) rerankProducts(ctx context.Context, query string, candidates []models.ProductSearchCandidate) []string {
	fallback := func() []string {
		ids := make([]string, len(candidates))
		for i, c := range candidates {
			ids[i] = c.SourceID
		}
		return ids
	}

	type rerankCand struct {
		SourceID string `json:"source_id"`
		Passage  string `json:"passage"`
	}
	payloadCands := make([]rerankCand, len(candidates))
	for i, c := range candidates {
		payloadCands[i] = rerankCand{SourceID: c.SourceID, Passage: c.Passage}
	}

	body, err := json.Marshal(map[string]any{"query": query, "candidates": payloadCands})
	if err != nil {
		w.log.Error("marshal rerank request failed", "err", err)
		return fallback()
	}

	reqURL := w.cfg.App.AIServiceURL + "/api/v1/ml/chat/rerank"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		w.log.Error("build rerank request failed", "err", err)
		return fallback()
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		w.log.Error("rerank call failed", "err", err)
		return fallback()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.log.Error("rerank returned non-200", "status", resp.StatusCode)
		return fallback()
	}

	var result struct {
		SourceIDs []string `json:"source_ids"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		w.log.Error("decode rerank response failed", "err", err)
		return fallback()
	}
	if len(result.SourceIDs) == 0 {
		return fallback()
	}
	return result.SourceIDs
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
		if r.Score >= pastChatScoreThreshold {
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

	reply, images, err := w.callChatReply(ctx, businessID, conversationID, message, activeProductID, knowledge, pastChats, business, history, customer)
	if err != nil {
		w.rdb.Decr(ctx, rateLimitKey)
		return fmt.Errorf("chat reply: %w", err)
	}
	w.log.Info("AI reply received", "conversation_id", conversationID, "reply_len", len(reply), "images", len(images))

	pending := models.MessageStatusPending
	aiSender := models.MessageSenderAI
	tx, err := w.db.BeginTxx(ctx, nil)
	if err != nil {
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
		return fmt.Errorf("create outbound: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit outbound: %w", err)
	}

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

	for _, imgURL := range images {
		w.sendImageMessage(ctx, platform, businessID, conversationID, recipientID, token, imgURL)
	}
	return nil
}

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
