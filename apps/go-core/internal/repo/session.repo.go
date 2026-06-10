package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type SessionRepo interface {
	GetByToken(ctx context.Context, token string) (*models.SessionWithUser, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Session, error)
}

type sessionRepo struct {
	db *sqlx.DB
}

func NewSessionRepo(db *sqlx.DB) SessionRepo {
	return &sessionRepo{db: db}
}



func (r *sessionRepo) GetByToken(ctx context.Context, token string) (*models.SessionWithUser, error) {
	var session models.SessionWithUser
	err := r.db.GetContext(ctx, &session, `
		SELECT
			s.id, s.user_id, s.token, s.expires_at, s.created_at, s.updated_at,
			u.name  AS user_name,
			u.email AS user_email,
			u.role  AS user_role,
			u.is_onboarded  AS user_onboarded,
			u.email_verified AS user_email_verified,
			u.image AS user_image
		FROM session s
		JOIN "user" u ON u.id = s.user_id
		WHERE s.token=$1 AND s.expires_at > NOW()`, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get session by token: %w", err)
	}
	return &session, nil
}

func (r *sessionRepo) GetByUserID(ctx context.Context, userID string) ([]*models.Session, error) {
	var sessions []*models.Session
	err := r.db.SelectContext(ctx, &sessions, `
		SELECT * FROM sessions 
		WHERE user_id=$1 AND expires_at > NOW()`, userID)
	if err != nil {
		return nil, fmt.Errorf("get sessions by user_id: %w", err)
	}
	return sessions, nil
}
