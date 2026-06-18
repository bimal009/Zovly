package repository

import (
	"context"
	"fmt"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type MessageEmbeddingRepo interface {
	Create(ctx context.Context, tx *sqlx.Tx, emb models.CreateMessageEmbedding) error
	CreateAndMarkVectorized(ctx context.Context, emb models.CreateMessageEmbedding) error
}

type messageEmbeddingRepo struct {
	db *sqlx.DB
}

func NewMessageEmbeddingRepo(db *sqlx.DB) MessageEmbeddingRepo {
	return &messageEmbeddingRepo{db: db}
}

func (r *messageEmbeddingRepo) Create(ctx context.Context, tx *sqlx.Tx, emb models.CreateMessageEmbedding) error {
	query := `
		INSERT INTO message_embeddings
			(message_id, business_id, conversation_id, content, embedding)
		VALUES
			(:message_id, :business_id, :conversation_id, :content, :embedding)
	`
	if _, err := tx.NamedExecContext(ctx, query, emb); err != nil {
		return fmt.Errorf("create message embedding: %w", err)
	}
	return nil
}

func (r *messageEmbeddingRepo) CreateAndMarkVectorized(ctx context.Context, emb models.CreateMessageEmbedding) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO message_embeddings
			(message_id, business_id, conversation_id, content, embedding)
		VALUES
			(:message_id, :business_id, :conversation_id, :content, :embedding)
		ON CONFLICT (message_id) DO NOTHING
	`
	if _, err := tx.NamedExecContext(ctx, insertQuery, emb); err != nil {
		return fmt.Errorf("insert message embedding: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE messages SET is_vectorized = true WHERE id = $1
	`, emb.MessageID); err != nil {
		return fmt.Errorf("mark vectorized: %w", err)
	}

	return tx.Commit()
}
