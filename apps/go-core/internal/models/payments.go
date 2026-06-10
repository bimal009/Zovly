package models

import "time"

type PaymentStatus string


const (
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusFailed  PaymentStatus = "failed"
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type PaymentRecord struct {
	ID             string `db:"id"              json:"id"`
	BusinessID     string `db:"business_id"     json:"business_id"`
	SubscriptionID *string `db:"subscription_id" json:"subscription_id,omitempty"`
	PlanID         *string `db:"plan_id"         json:"plan_id,omitempty"`

	// ── Billing ───────────────────────────────────────────────────────────────
	BillingCycle BillingCycle `db:"billing_cycle" json:"billing_cycle"`

	// ── Paddle IDs ────────────────────────────────────────────────────────────
	PaddleTransactionID  string  `db:"paddle_transaction_id"  json:"paddle_transaction_id"`
	PaddleSubscriptionID *string `db:"paddle_subscription_id" json:"paddle_subscription_id,omitempty"`
	PaddleCustomerID     *string `db:"paddle_customer_id"     json:"paddle_customer_id,omitempty"`

	// ── Payment details ───────────────────────────────────────────────────────
	Amount   int    `db:"amount"   json:"amount"`
	Currency string `db:"currency" json:"currency"`

	// ── Period ────────────────────────────────────────────────────────────────
	PeriodStart time.Time `db:"period_start" json:"period_start"`
	PeriodEnd   time.Time `db:"period_end"   json:"period_end"`

	Status PaymentStatus `db:"status" json:"status"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}