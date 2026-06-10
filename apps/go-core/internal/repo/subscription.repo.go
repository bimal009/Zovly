package repository

import (
	"context"
	"database/sql"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type SubscriptionRepo interface {
	Create(ctx context.Context, tx *sqlx.Tx, subs models.BusinessSubscription) (string, error)
	GetByBusinessID(ctx context.Context, businessID string) (*models.BusinessSubscription, error)
	GetByPaddleSubscriptionID(ctx context.Context, paddleSubID string) (*models.BusinessSubscription, error)
	GetByPaddleCustomerID(ctx context.Context, customerID string) (*models.BusinessSubscription, error)
	Update(ctx context.Context, tx *sqlx.Tx, id string, update models.SubscriptionUpdate) error
}

type subscriptionRepo struct {
	db *sqlx.DB
}

func NewSubscriptionRepo(db *sqlx.DB) SubscriptionRepo {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) Create(ctx context.Context, tx *sqlx.Tx, subs models.BusinessSubscription) (string, error) {
	var id string
	err := tx.QueryRowContext(ctx, `
		INSERT INTO business_subscriptions (
			business_id, plan_id,
			paddle_subscription_id, paddle_customer_id, paddle_price_id,
			billing_cycle, status,
			ai_replies_used, posts_used, usage_reset_at,
			trial_started_at, trial_ends_at,
			current_period_start, current_period_end,
			cancel_at_period_end, cancelled_at, paused_at,
			notes
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18
		)
		RETURNING id`,
		subs.BusinessID, subs.PlanID,
		subs.PaddleSubscriptionID, subs.PaddleCustomerID, subs.PaddlePriceID,
		subs.BillingCycle, subs.Status,
		subs.AiRepliesUsed, subs.PostsUsed, subs.UsageResetAt,
		subs.TrialStartedAt, subs.TrialEndsAt,
		subs.CurrentPeriodStart, subs.CurrentPeriodEnd,
		subs.CancelAtPeriodEnd, subs.CancelledAt, subs.PausedAt,
		subs.Notes,
	).Scan(&id)
	return id, err
}

func (r *subscriptionRepo) GetByBusinessID(ctx context.Context, businessID string) (*models.BusinessSubscription, error) {
	var sub models.BusinessSubscription
	err := r.db.GetContext(ctx, &sub, `
		SELECT * FROM business_subscriptions
		WHERE business_id = $1
	`, businessID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &sub, err
}

func (r *subscriptionRepo) GetByPaddleSubscriptionID(ctx context.Context, paddleSubID string) (*models.BusinessSubscription, error) {
	var sub models.BusinessSubscription
	err := r.db.GetContext(ctx, &sub, `
		SELECT * FROM business_subscriptions
		WHERE paddle_subscription_id = $1
	`, paddleSubID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &sub, err
}

func (r *subscriptionRepo) GetByPaddleCustomerID(ctx context.Context, customerID string) (*models.BusinessSubscription, error) {
	var sub models.BusinessSubscription
	err := r.db.GetContext(ctx, &sub, `
		SELECT * FROM business_subscriptions
		WHERE paddle_customer_id = $1
	`, customerID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &sub, err
}

func (r *subscriptionRepo) Update(ctx context.Context, tx *sqlx.Tx, id string, u models.SubscriptionUpdate) error {
	q := `
		UPDATE business_subscriptions SET
			plan_id                = COALESCE($2, plan_id),
			paddle_subscription_id = COALESCE($3, paddle_subscription_id),
			paddle_customer_id     = COALESCE($4, paddle_customer_id),
			paddle_price_id        = COALESCE($5, paddle_price_id),
			status                 = COALESCE($6, status),
			current_period_start   = COALESCE($7, current_period_start),
			current_period_end     = COALESCE($8, current_period_end),
			cancel_at_period_end   = $9,
			cancelled_at           = COALESCE($10, cancelled_at),
			paused_at              = COALESCE($11, paused_at),
			trial_started_at       = COALESCE($12, trial_started_at),
			trial_ends_at          = COALESCE($13, trial_ends_at),
			updated_at             = NOW()
		WHERE id = $1
	`

	var planID *string
	if u.PlanID != "" {
		planID = &u.PlanID
	}

	var status *string
	if u.Status != "" {
		status = &u.Status
	}

	// Use tx if provided, otherwise use db directly
	if tx != nil {
		_, err := tx.ExecContext(ctx, q,
			id, planID,
			u.PaddleSubscriptionID, u.PaddleCustomerID, u.PaddlePriceID,
			status,
			u.CurrentPeriodStart, u.CurrentPeriodEnd,
			u.CancelAtPeriodEnd,
			u.CancelledAt, u.PausedAt,
			u.TrialStartedAt, u.TrialEndsAt,
		)
		return err
	}

	_, err := r.db.ExecContext(ctx, q,
		id, planID,
		u.PaddleSubscriptionID, u.PaddleCustomerID, u.PaddlePriceID,
		status,
		u.CurrentPeriodStart, u.CurrentPeriodEnd,
		u.CancelAtPeriodEnd,
		u.CancelledAt, u.PausedAt,
		u.TrialStartedAt, u.TrialEndsAt,
	)
	return err
}
