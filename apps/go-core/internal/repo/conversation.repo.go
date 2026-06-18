package repository

import (
	"context"
	"fmt"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type ConversationRepo interface {
	FindOrCreate(ctx context.Context, tx *sqlx.Tx, conv models.CreateConversation) (*models.Conversation, error)
}
type conversationRepo struct {
	db *sqlx.DB
}

func NewconversationRepo(db *sqlx.DB) ConversationRepo {
	return &conversationRepo{
		db: db,
	}
}

func (r *conversationRepo) FindOrCreate(ctx context.Context, tx *sqlx.Tx, conv models.CreateConversation) (*models.Conversation, error) {
	var existing models.Conversation
	err := tx.GetContext(ctx, &existing, `
        SELECT * FROM conversations
        WHERE business_id = $1 AND platform = $2 AND contact_id = $3
    `, conv.BusinessID, conv.Platform, conv.ContactID)

	if err == nil {
		tx.ExecContext(ctx, `
            UPDATE conversations SET last_message_at = now() WHERE id = $1
        `, existing.ID)
		return &existing, nil
	}

	// create new conversation
	query := `
        INSERT INTO conversations
            (business_id, platform, thread_id, contact_id, last_message_at, ai_enabled)
        VALUES
            (:business_id, :platform, :thread_id, :contact_id, now(), true)
        RETURNING *
    `
	rows, err := sqlx.NamedQueryContext(ctx, tx, query, conv)
	if err != nil {
		return nil, fmt.Errorf("create conversation: %w", err)
	}
	defer rows.Close()

	var created models.Conversation
	if rows.Next() {
		rows.StructScan(&created)
	}
	return &created, rows.Err()
}
