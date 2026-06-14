package repository

import (
	"context"
	"fmt"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type BusinessKnowledgeRepo interface {
	Create(ctx context.Context, tx *sqlx.Tx, chunks []models.CreateKnowledgeChunk) error
}

type businessKnowledgeRepo struct {
	db *sqlx.DB
}

func NewBusinessKnowledgeRepo(db *sqlx.DB) BusinessKnowledgeRepo {
	return &businessKnowledgeRepo{
		db: db,
	}
}

func (r *businessKnowledgeRepo) Create(ctx context.Context, tx *sqlx.Tx, chunks []models.CreateKnowledgeChunk) error {
	if len(chunks) == 0 {
		return nil
	}

	query := `
		INSERT INTO knowledge_chunks
			(business_id, source_type, source_id, chunk_index, content, embedding, metadata)
		VALUES
			(:business_id, :source_type, :source_id, :chunk_index, :content, :embedding, :metadata)
	`

	if _, err := tx.NamedExecContext(ctx, query, chunks); err != nil {
		return fmt.Errorf("create knowledge chunks: %w", err)
	}
	return nil
}
