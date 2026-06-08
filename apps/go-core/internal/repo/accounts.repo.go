package repository

import (
	"context"
	"fmt"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type AccountRepo interface {
	DeleteByUserID(ctx context.Context, userID string) error
	Get(ctx context.Context) ([]*models.Account, error)
	GetByUserID(ctx context.Context, userID string) (*models.Account, error)
}

type accountRepo struct {
	db *sqlx.DB
}

func NewAccountRepo(db *sqlx.DB) AccountRepo {
	return &accountRepo{db: db}
}





func (r *accountRepo) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM accounts WHERE user_id=$1`, userID)
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}

func (r *accountRepo) Get(ctx context.Context) ([]*models.Account, error) {
	var accounts []*models.Account
	if err := r.db.SelectContext(ctx, &accounts, `SELECT * FROM accounts`); err != nil {
		return nil, fmt.Errorf("get accounts: %w", err)
	}
	return accounts, nil
}

func (r *accountRepo) GetByUserID(ctx context.Context, userID string) (*models.Account, error) {
	var account models.Account
	if err := r.db.GetContext(ctx, &account,
		`SELECT * FROM accounts WHERE user_id=$1`, userID,
	); err != nil {
		return nil, fmt.Errorf("get account by user_id: %w", err)
	}
	return &account, nil
}
