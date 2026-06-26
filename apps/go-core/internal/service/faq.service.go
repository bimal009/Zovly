package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/embed"
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
	embedder      *embed.Client
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
		embedder:      embed.New(cfg.App.AIServiceURL),
	}
}

func (s *faqService) Create(ctx context.Context, input models.CreateFaqRequest, businessID string) error {
	chunks, err := s.embedder.EmbedFaq(ctx, input.Question, input.Answer)
	if err != nil {
		return fmt.Errorf("embed faq: %w", err)
	}

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
