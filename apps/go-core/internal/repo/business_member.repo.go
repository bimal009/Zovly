package repository

import (
	"context"
	"fmt"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type BusinessMemberRepo interface {
	CreateTx(ctx context.Context, tx *sqlx.Tx, business models.BusinessMember) (*models.BusinessMember, error)
}
type businessMemberRepo struct {
	db *sqlx.DB
}

func NewBusinessMemberRepo(db *sqlx.DB) BusinessMemberRepo {
	return &businessMemberRepo{
		db: db,
	}
}

func (r *businessMemberRepo) CreateTx(ctx context.Context, tx *sqlx.Tx, member models.BusinessMember) (*models.BusinessMember, error) {
	rows, err := sqlx.NamedQueryContext(ctx, tx, `
		INSERT INTO business_members (
			business_id,
			user_id,
			role,
			can_manage_content,
			can_view_analytics,
			can_manage_ads,
			can_read_dms,
			can_reply_dms,
			can_read_comments,
			can_reply_comments,
			can_view_leads,
			can_manage_leads,
			can_view_bookings,
			can_manage_bookings,
			can_view_inventory,
			can_manage_inventory,
			can_view_orders,
			can_manage_settings,
			can_manage_members,
			can_manage_billing
		)
		VALUES (
			:business_id,
			:user_id,
			:role,
			:can_manage_content,
			:can_view_analytics,
			:can_manage_ads,
			:can_read_dms,
			:can_reply_dms,
			:can_read_comments,
			:can_reply_comments,
			:can_view_leads,
			:can_manage_leads,
			:can_view_bookings,
			:can_manage_bookings,
			:can_view_inventory,
			:can_manage_inventory,
			:can_view_orders,
			:can_manage_settings,
			:can_manage_members,
			:can_manage_billing
		)
		RETURNING *
	`, member)
	if err != nil {
		return nil, fmt.Errorf("createTx business member: %w", err)
	}
	defer rows.Close()
	var created models.BusinessMember
	if rows.Next() {
		if err := rows.StructScan(&created); err != nil {
			return nil, fmt.Errorf("createTx business member: scan: %w", err)
		}
	}
	return &created, nil
}
