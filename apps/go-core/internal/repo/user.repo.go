package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type UserRepo interface {
	Delete(ctx context.Context, userID string) error
	Get(ctx context.Context) ([]*models.User, error)
	GetByID(ctx context.Context, userID string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	CreateTx(ctx context.Context, tx *sqlx.Tx, user models.User) (*models.User, error)
	UpdateTx(ctx context.Context, tx *sqlx.Tx, user models.User, userID string) (*models.User, error)
}

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) UserRepo {
	return &userRepo{db: db}
}
func (r *userRepo) Delete(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (r *userRepo) Get(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.SelectContext(ctx, &users, `SELECT * FROM users`); err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}
	return users, nil
}

func (r *userRepo) GetByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	if err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE id=$1`, userID); err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE email=$1`, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepo) CreateTx(ctx context.Context, tx *sqlx.Tx, user models.User) (*models.User, error) {
	query := `
		INSERT INTO users (name, email, role)
		VALUES (:name, :email, :role)
		RETURNING *`

	rows, err := sqlx.NamedQueryContext(ctx, tx, query, user) // ← tx not r.db
	if err != nil {
		return nil, fmt.Errorf("create user tx: %w", err)
	}
	defer rows.Close()

	var created models.User
	if rows.Next() {
		if err := rows.StructScan(&created); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
	}
	return &created, nil
}

func (r *userRepo) UpdateTx(ctx context.Context, tx *sqlx.Tx, user models.User, userID string) (*models.User, error) {
	query := `
		UPDATE users
		SET name=:name, email=:email, image=:image,
		    email_verified=:email_verified, role=:role,
		    onboarded=:onboarded, updated_at=NOW()
		WHERE id=:id
		RETURNING *`

	user.ID = userID
	rows, err := sqlx.NamedQueryContext(ctx, tx, query, user)
	if err != nil {
		return nil, fmt.Errorf("update user tx: %w", err)
	}
	defer rows.Close()

	var updated models.User
	if rows.Next() {
		if err := rows.StructScan(&updated); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
	}
	return &updated, nil
}
