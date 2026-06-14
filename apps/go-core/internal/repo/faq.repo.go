package repository

import (
	"context"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type FaqRepo interface {
	Create(ctx context.Context, tx *sqlx.Tx, faq models.CreateFaqRequest, businessID string) (string, error)
	GetAll(ctx context.Context, businessID string) ([]models.Faq, error)
}

type faqRepo struct {
	db *sqlx.DB
}

func NewFaqRepo(db *sqlx.DB) FaqRepo {
	return &faqRepo{db: db}
}
func (r *faqRepo) Create(ctx context.Context, tx *sqlx.Tx, faq models.CreateFaqRequest, businessID string) (string, error) {
	query := `
		INSERT INTO faqs (business_id, question, answer, sort_order)
		VALUES (:business_id, :question, :answer,
			(SELECT COALESCE(MAX(sort_order), 0) + 1 FROM faqs WHERE business_id = :business_id))
		RETURNING id
	`
	rows, err := sqlx.NamedQueryContext(ctx, tx, query, map[string]any{
		"business_id": businessID,
		"question":    faq.Question,
		"answer":      faq.Answer,
	})
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var id string
	if rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
	}
	return id, rows.Err()
}
func (r *faqRepo) GetAll(ctx context.Context, businessID string) ([]models.Faq, error) {
	query := `SELECT * FROM faqs WHERE business_id=$1 ORDER BY sort_order ASC`

	var faqs []models.Faq
	err := r.db.SelectContext(ctx, &faqs, query, businessID)
	return faqs, err
}
