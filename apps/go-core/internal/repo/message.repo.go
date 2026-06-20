package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type MessageRepo interface {
	Create(ctx context.Context, tx *sqlx.Tx, msg models.CreateMessage) (*models.Message, error)
	UpdateVectorized(ctx context.Context, messageID string, vectorized bool) error
	UpdateStatus(ctx context.Context, messageID string, status models.MessageStatus, platformMsgID string, errMsg *string) error
	GetByConversation(ctx context.Context, conversationID string, after *time.Time, limit int) ([]models.Message, error)
	GetPendingOutbound(ctx context.Context) ([]models.Message, error)
	GetUnrepliedInbound(ctx context.Context, conversationID string) ([]models.Message, error)
}

type messageRepo struct {
	db *sqlx.DB
}

func NewMessageRepo(db *sqlx.DB) MessageRepo {
	return &messageRepo{db: db}
}

func (r *messageRepo) Create(ctx context.Context, tx *sqlx.Tx, msg models.CreateMessage) (*models.Message, error) {
	query := `
		INSERT INTO messages
			(conversation_id, business_id, direction, sent_by, content,
			 media_url, media_type, status)
		VALUES
			(:conversation_id, :business_id, :direction, :sent_by, :content,
			 :media_url, :media_type, :status)
		RETURNING *
	`

	rows, err := sqlx.NamedQueryContext(ctx, tx, query, msg)
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}
	defer rows.Close()

	var created models.Message
	if rows.Next() {
		if err := rows.StructScan(&created); err != nil {
			return nil, fmt.Errorf("scan created message: %w", err)
		}
	}
	return &created, rows.Err()
}

func (r *messageRepo) UpdateVectorized(ctx context.Context, messageID string, vectorized bool) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE messages
		SET is_vectorized = $1
		WHERE id = $2
	`, vectorized, messageID)
	if err != nil {
		return fmt.Errorf("update vectorized: %w", err)
	}
	return nil
}

func (r *messageRepo) UpdateStatus(ctx context.Context, messageID string, status models.MessageStatus, platformMsgID string, errMsg *string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE messages
		SET status              = $1,
		    platform_message_id = $2,
		    error_message       = $3,
		    sent_to_platform_at = CASE WHEN $1 = 'sent' THEN now() ELSE sent_to_platform_at END,
		    sent_at             = CASE WHEN $1 = 'sent' THEN now() ELSE sent_at END
		WHERE id = $4
	`, status, platformMsgID, errMsg, messageID)
	if err != nil {
		return fmt.Errorf("update message status: %w", err)
	}
	return nil
}

func (r *messageRepo) GetByConversation(ctx context.Context, conversationID string, after *time.Time, limit int) ([]models.Message, error) {
	var messages []models.Message

	if after != nil {
		err := r.db.SelectContext(ctx, &messages, `
			SELECT * FROM messages
			WHERE conversation_id = $1 AND sent_at > $2
			ORDER BY sent_at ASC
			LIMIT $3
		`, conversationID, *after, limit)
		return messages, err
	}

	err := r.db.SelectContext(ctx, &messages, `
		SELECT * FROM messages
		WHERE conversation_id = $1
		ORDER BY sent_at ASC
		LIMIT $2
	`, conversationID, limit)
	return messages, err
}
func (r *messageRepo) GetUnrepliedInbound(ctx context.Context, conversationID string) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.SelectContext(ctx, &messages, `
		SELECT * FROM messages
		WHERE conversation_id = $1
		  AND direction = 'in'
		  AND sent_at > COALESCE(
		      (SELECT MAX(sent_at) FROM messages
		       WHERE conversation_id = $1 AND direction = 'out'),
		      '1970-01-01'
		  )
		ORDER BY sent_at ASC
	`, conversationID)
	return messages, err
}

func (r *messageRepo) GetPendingOutbound(ctx context.Context) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.SelectContext(ctx, &messages, `
        SELECT * FROM messages
        WHERE direction = 'out'
          AND status = 'pending'
        ORDER BY sent_at ASC
        LIMIT 50
    `)
	return messages, err
}
