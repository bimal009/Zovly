package models

import "time"

type BillingCycle string
type PlanStatus   string

const (
	BillingCycleMonthly BillingCycle = "monthly"
	BillingCycleYearly  BillingCycle = "yearly"

	PlanStatusActive    PlanStatus = "active"
	PlanStatusTrialing  PlanStatus = "trialing"
	PlanStatusPastDue   PlanStatus = "past_due"
	PlanStatusPaused    PlanStatus = "paused"
	PlanStatusCancelled PlanStatus = "cancelled"
	PlanStatusExpired   PlanStatus = "expired"
)

type BusinessSubscription struct {
	ID           string       `db:"id"            json:"id"`
	BusinessID   string       `db:"business_id"   json:"business_id"`
	PlanID       string       `db:"plan_id"        json:"plan_id"`

	PaddleSubscriptionID *string `db:"paddle_subscription_id" json:"paddle_subscription_id,omitempty"`
	PaddleCustomerID     *string `db:"paddle_customer_id"     json:"paddle_customer_id,omitempty"`
	PaddlePriceID        *string `db:"paddle_price_id"        json:"paddle_price_id,omitempty"`

	BillingCycle BillingCycle `db:"billing_cycle" json:"billing_cycle"`
	Status       PlanStatus   `db:"status"        json:"status"`

	AiRepliesUsed int        `db:"ai_replies_used" json:"ai_replies_used"`
	PostsUsed     int        `db:"posts_used"      json:"posts_used"`
	UsageResetAt  *time.Time `db:"usage_reset_at"  json:"usage_reset_at,omitempty"`

	TrialStartedAt *time.Time `db:"trial_started_at" json:"trial_started_at,omitempty"`
	TrialEndsAt    *time.Time `db:"trial_ends_at"    json:"trial_ends_at,omitempty"`

	CurrentPeriodStart *time.Time `db:"current_period_start" json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time `db:"current_period_end"   json:"current_period_end,omitempty"`
	CancelAtPeriodEnd  bool       `db:"cancel_at_period_end" json:"cancel_at_period_end"`
	CancelledAt        *time.Time `db:"cancelled_at"         json:"cancelled_at,omitempty"`
	PausedAt           *time.Time `db:"paused_at"            json:"paused_at,omitempty"`

	Notes *string `db:"notes" json:"notes,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// SubscriptionUpdate — only non-nil fields get written to DB
type SubscriptionUpdate struct {
	PlanID               string
	PaddleSubscriptionID *string
	PaddleCustomerID     *string
	PaddlePriceID        *string
	Status               string
	CurrentPeriodStart   *time.Time
	CurrentPeriodEnd     *time.Time
	CancelAtPeriodEnd    bool
	CancelledAt          *time.Time
	PausedAt             *time.Time
	TrialStartedAt       *time.Time
	TrialEndsAt          *time.Time
}