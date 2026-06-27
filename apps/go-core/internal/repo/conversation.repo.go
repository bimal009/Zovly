package repository

import (
	"context"
	"fmt"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type ConversationRepo interface {
	FindOrCreate(ctx context.Context, tx *sqlx.Tx, conv models.CreateConversation) (*models.Conversation, error)
	GetByID(ctx context.Context, id, businessID string) (*models.Conversation, error)
	ListByBusiness(ctx context.Context, businessID string, limit, offset int) ([]models.Conversation, int, error)
	SetActiveProduct(ctx context.Context, conversationID, productID string) error
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
            (business_id, platform, thread_id, contact_id, contact_name, contact_username, contact_avatar_url, last_message_at)
        VALUES
            (:business_id, :platform, :thread_id, :contact_id, :contact_name, :contact_username, :contact_avatar_url, now())
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

func (r *conversationRepo) GetByID(ctx context.Context, id, businessID string) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.GetContext(ctx, &conv, `
        SELECT * FROM conversations WHERE id = $1 AND business_id = $2
    `, id, businessID)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *conversationRepo) SetActiveProduct(ctx context.Context, conversationID, productID string) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE conversations
        SET active_product_id = $1, active_product_at = now()
        WHERE id = $2
    `, productID, conversationID)
	if err != nil {
		return fmt.Errorf("set active product: %w", err)
	}
	return nil
}

func (r *conversationRepo) ListByBusiness(ctx context.Context, businessID string, limit, offset int) ([]models.Conversation, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `
        SELECT COUNT(*) FROM conversations WHERE business_id = $1
    `, businessID); err != nil {
		return nil, 0, fmt.Errorf("count conversations: %w", err)
	}

	var convs []models.Conversation
	err := r.db.SelectContext(ctx, &convs, `
        SELECT * FROM conversations
        WHERE business_id = $1
        ORDER BY last_message_at DESC NULLS LAST
        LIMIT $2 OFFSET $3
    `, businessID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list conversations: %w", err)
	}
	return convs, total, nil
}
