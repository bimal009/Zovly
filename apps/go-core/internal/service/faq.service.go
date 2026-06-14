package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/jmoiron/sqlx"
)

type FaqService interface {
	Create(ctx context.Context, input models.CreateFaqRequest, businessID string) error
	GetAll(ctx context.Context, businessID string) ([]models.Faq, error)
}

type faqService struct {
	faqRepo       repository.FaqRepo
	knowledgeRepo repository.BusinessKnowledgeRepo
	log           *slog.Logger
	db            *sqlx.DB
	cfg           config.Config
	httpClient    *http.Client
}

func NewFaqService(
	faqRepo repository.FaqRepo,
	knowledgeRepo repository.BusinessKnowledgeRepo,
	log *slog.Logger,
	db *sqlx.DB,
	cfg config.Config,
) FaqService {
	return &faqService{
		faqRepo:       faqRepo,
		knowledgeRepo: knowledgeRepo,
		log:           log,
		db:            db,
		cfg:           cfg,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *faqService) embedFaq(ctx context.Context, input models.CreateFaqRequest) ([]models.FaqChunksResponse, error) {
	body, err := json.Marshal(map[string]string{
		"question": input.Question,
		"answer":   input.Answer,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	url := s.cfg.App.AIServiceURL + "/api/v1/ml/embed/faq"
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

	var result []models.FaqChunksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}
	return result, nil
}

func (s *faqService) Create(ctx context.Context, input models.CreateFaqRequest, businessID string) error {
	// 1. embed first — network call, kept outside the tx
	chunks, err := s.embedFaq(ctx, input)
	if err != nil {
		return fmt.Errorf("embed faq: %w", err)
	}

	// metadata stored on every chunk: { question } for citations
	meta, err := json.Marshal(map[string]string{"question": input.Question})
	if err != nil {
		return fmt.Errorf("marshal chunk metadata: %w", err)
	}

	// 2. tx wraps only the writes
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() // no-op once Commit succeeds

	// 3. create faq AND get its id (source_id for the chunks)
	faqID, err := s.faqRepo.Create(ctx, tx, input, businessID)
	if err != nil {
		return fmt.Errorf("create faq: %w", err)
	}

	// 4. stamp business/source/metadata onto chunks, bulk insert on the tx
	inserts := models.ToChunkInserts(chunks, businessID, faqID, models.SourceFaq, meta)
	if err := s.knowledgeRepo.Create(ctx, tx, inserts); err != nil {
		return fmt.Errorf("create knowledge chunks: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	s.log.InfoContext(ctx, "faq created with embeddings",
		"business_id", businessID, "faq_id", faqID, "chunks", len(inserts))
	return nil
}

func (s *faqService) GetAll(ctx context.Context, businessID string) ([]models.Faq, error) {
	faqs, err := s.faqRepo.GetAll(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("get faqs: %w", err)
	}
	return faqs, nil
}
