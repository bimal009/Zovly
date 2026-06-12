package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

// flat scan target for the business+member JOIN — only the 3 shared column names need aliases
type bwmScan struct {
	models.Business
	MemberID           string            `db:"member_id"`
	UserID             string            `db:"user_id"`
	MemberRole         models.MemberRole `db:"member_role"`
	CanManageContent   bool              `db:"can_manage_content"`
	CanViewAnalytics   bool              `db:"can_view_analytics"`
	CanManageAds       bool              `db:"can_manage_ads"`
	CanReadDMs         bool              `db:"can_read_dms"`
	CanReplyDMs        bool              `db:"can_reply_dms"`
	CanReadComments    bool              `db:"can_read_comments"`
	CanReplyComments   bool              `db:"can_reply_comments"`
	CanViewLeads       bool              `db:"can_view_leads"`
	CanManageLeads     bool              `db:"can_manage_leads"`
	CanViewBookings    bool              `db:"can_view_bookings"`
	CanManageBookings  bool              `db:"can_manage_bookings"`
	CanViewInventory   bool              `db:"can_view_inventory"`
	CanManageInventory bool              `db:"can_manage_inventory"`
	CanViewOrders      bool              `db:"can_view_orders"`
	CanManageSettings  bool              `db:"can_manage_settings"`
	CanManageMembers   bool              `db:"can_manage_members"`
	CanManageBilling   bool              `db:"can_manage_billing"`
	JoinedAt           time.Time         `db:"joined_at"`
	LastSeenAt         *time.Time        `db:"last_seen_at"`
	MemberCreatedAt    time.Time         `db:"member_created_at"`
	MemberUpdatedAt    time.Time         `db:"member_updated_at"`
}

type BusinessRepo interface {
	CreateTx(ctx context.Context, tx *sqlx.Tx, business models.Business) (*models.Business, error)
	GetByUserId(ctx context.Context, userID string) (*models.BusinessWithMembers, error)
}
type businessRepo struct {
	db *sqlx.DB
}

func NewBusinessRepo(db *sqlx.DB) BusinessRepo {
	return &businessRepo{
		db: db,
	}
}

func (r *businessRepo) CreateTx(ctx context.Context, tx *sqlx.Tx, business models.Business) (*models.Business, error) {
	rows, err := sqlx.NamedQueryContext(ctx, tx, `
		INSERT INTO business (name, description, logo, website, phone, address, city, country, type)
		VALUES (:name, :description, :logo, :website, :phone, :address, :city, :country, :type)
		RETURNING *`,
		business,
	)
	if err != nil {
		return nil, fmt.Errorf("createTx business: %w", err)
	}
	defer rows.Close()
	var created models.Business
	if rows.Next() {
		if err := rows.StructScan(&created); err != nil {
			return nil, fmt.Errorf("createTx business: scan: %w", err)
		}
	}
	return &created, nil
}

func (r *businessRepo) GetByUserId(ctx context.Context, userID string) (*models.BusinessWithMembers, error) {
	var row bwmScan
	err := r.db.GetContext(ctx, &row, `
		SELECT
			b.*,
			bm.id           AS member_id,
			bm.user_id,
			bm.role         AS member_role,
			bm.can_manage_content, bm.can_view_analytics, bm.can_manage_ads,
			bm.can_read_dms, bm.can_reply_dms, bm.can_read_comments, bm.can_reply_comments,
			bm.can_view_leads, bm.can_manage_leads,
			bm.can_view_bookings, bm.can_manage_bookings,
			bm.can_view_inventory, bm.can_manage_inventory, bm.can_view_orders,
			bm.can_manage_settings, bm.can_manage_members, bm.can_manage_billing,
			bm.joined_at, bm.last_seen_at,
			bm.created_at AS member_created_at,
			bm.updated_at AS member_updated_at
		FROM business_members bm
		JOIN business b ON bm.business_id = b.id
		WHERE bm.user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("get business by user: %w", err)
	}

	return &models.BusinessWithMembers{
		Business: row.Business,
		Member: models.BusinessMember{
			ID:                 row.MemberID,
			BusinessID:         row.Business.ID,
			UserID:             row.UserID,
			Role:               row.MemberRole,
			CanManageContent:   row.CanManageContent,
			CanViewAnalytics:   row.CanViewAnalytics,
			CanManageAds:       row.CanManageAds,
			CanReadDMs:         row.CanReadDMs,
			CanReplyDMs:        row.CanReplyDMs,
			CanReadComments:    row.CanReadComments,
			CanReplyComments:   row.CanReplyComments,
			CanViewLeads:       row.CanViewLeads,
			CanManageLeads:     row.CanManageLeads,
			CanViewBookings:    row.CanViewBookings,
			CanManageBookings:  row.CanManageBookings,
			CanViewInventory:   row.CanViewInventory,
			CanManageInventory: row.CanManageInventory,
			CanViewOrders:      row.CanViewOrders,
			CanManageSettings:  row.CanManageSettings,
			CanManageMembers:   row.CanManageMembers,
			CanManageBilling:   row.CanManageBilling,
			JoinedAt:           row.JoinedAt,
			LastSeenAt:         row.LastSeenAt,
			CreatedAt:          row.MemberCreatedAt,
			UpdatedAt:          row.MemberUpdatedAt,
		},
	}, nil
}
