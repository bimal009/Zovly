package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type MemberInviteRepo interface {
	HasPendingByEmail(ctx context.Context, email string) (bool, error)
}

type memberInviteRepo struct {
	db *sqlx.DB
}

func NewMemberInviteRepo(db *sqlx.DB) MemberInviteRepo {
	return &memberInviteRepo{db: db}
}

func (r *memberInviteRepo) HasPendingByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `
		SELECT EXISTS (
			SELECT 1 FROM member_invites
			WHERE invited_email = $1
			  AND status = 'pending'
			  AND expires_at > NOW()
		)`, email)
	if err != nil {
		return false, fmt.Errorf("check pending invite: %w", err)
	}
	return exists, nil
}
