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
)

type ChatService interface {
	Chat(ctx context.Context, message string, convoId string, businessId string) error
}

type chatService struct {
	db               *sqlx.DB
	messageEmbedRepo repository.MessageEmbeddingRepo
	messageRepo      repository.MessageRepo
	cfg              config.Config
	httpClient       *http.Client
}

func NewChatService(
	db *sqlx.DB,
	messageRepo repository.MessageRepo,
	messageEmbedRepo repository.MessageEmbeddingRepo,
	cfg config.Config,
) ChatService {
	return &chatService{
		db:               db,
		messageEmbedRepo: messageEmbedRepo,
		messageRepo:      messageRepo,
		cfg:              cfg,
		httpClient:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *chatService) Chat(ctx context.Context, message string, convoId string, businessId string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	sender := models.MessageSenderAI
	status := models.MessageStatusPending
	newMessage := models.CreateMessage{
		ConversationID: convoId,
		BusinessID:     businessId,
		Direction:      models.MessageDirectionIn,
		SentBy:         &sender,
		Content:        &message,
		MediaUrl:       nil,
		MediaType:      nil,
		Status:         &status,
	}

	if _, err := s.messageRepo.Create(ctx, tx, newMessage); err != nil {
		return fmt.Errorf("create message: %w", err)
	}

	return tx.Commit()
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
