package repository

import (
	"context"
	"fmt"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type BusinessKnowledgeRepo interface {
	Create(ctx context.Context, tx *sqlx.Tx, chunks []models.CreateKnowledgeChunk) error
	DeleteBySource(ctx context.Context, tx *sqlx.Tx, businessID, sourceID string, sourceType models.KnowledgeSourceType) error
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

func (r *businessKnowledgeRepo) DeleteBySource(ctx context.Context, tx *sqlx.Tx, businessID, sourceID string, sourceType models.KnowledgeSourceType) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM knowledge_chunks
		WHERE business_id = $1 AND source_id = $2 AND source_type = $3
	`, businessID, sourceID, sourceType)
	if err != nil {
		return fmt.Errorf("delete knowledge chunks by source: %w", err)
	}
	return nil
}
