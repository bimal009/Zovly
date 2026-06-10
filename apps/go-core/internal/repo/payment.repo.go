package repository

import (
	"context"
	"database/sql"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
)

type PaymentRepo interface {
	Create(ctx context.Context, tx *sqlx.Tx, payment models.PaymentRecord) (*models.PaymentRecord, error)
	GetByID(ctx context.Context, id string) (*models.PaymentRecord, error)
	GetByBusinessID(ctx context.Context, businessID string) ([]models.PaymentRecord, error)
	GetByPaddleTransactionID(ctx context.Context, txnID string) (*models.PaymentRecord, error)
	GetBySubscriptionID(ctx context.Context, subscriptionID string) ([]models.PaymentRecord, error)
	UpdateStatus(ctx context.Context, paddleTxnID string, status string) error
}

type paymentRepo struct {
	db *sqlx.DB // ← pointer, was value before
}

func NewPaymentRepo(db *sqlx.DB) PaymentRepo {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(ctx context.Context, tx *sqlx.Tx, payment models.PaymentRecord) (*models.PaymentRecord, error) {
	rows, err := tx.QueryContext(ctx, `
		INSERT INTO payment_records (
			business_id, subscription_id, plan_id, billing_cycle,
			paddle_transaction_id, paddle_subscription_id, paddle_customer_id,
			amount, currency, period_start, period_end, status
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12
		)
		RETURNING *`,
		payment.BusinessID, payment.SubscriptionID, payment.PlanID, payment.BillingCycle,
		payment.PaddleTransactionID, payment.PaddleSubscriptionID, payment.PaddleCustomerID,
		payment.Amount, payment.Currency, payment.PeriodStart, payment.PeriodEnd, payment.Status,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPayment(rows)
}

func (r *paymentRepo) GetByID(ctx context.Context, id string) (*models.PaymentRecord, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM payment_records WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPayment(rows)
}

func (r *paymentRepo) GetByBusinessID(ctx context.Context, businessID string) ([]models.PaymentRecord, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM payment_records WHERE business_id = $1 ORDER BY created_at DESC`, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPayments(rows)
}

func (r *paymentRepo) GetByPaddleTransactionID(ctx context.Context, txnID string) (*models.PaymentRecord, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM payment_records WHERE paddle_transaction_id = $1`, txnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPayment(rows)
}

func (r *paymentRepo) GetBySubscriptionID(ctx context.Context, subscriptionID string) ([]models.PaymentRecord, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM payment_records WHERE subscription_id = $1 ORDER BY created_at DESC`, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPayments(rows)
}

func (r *paymentRepo) UpdateStatus(ctx context.Context, paddleTxnID string, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE payment_records
		SET status = $2, updated_at = NOW()
		WHERE paddle_transaction_id = $1
	`, paddleTxnID, status)
	return err
}

// ── Scanners ──────────────────────────────────────────────────────────────────

func scanPayment(rows *sql.Rows) (*models.PaymentRecord, error) {
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	var p models.PaymentRecord
	err := rows.Scan(
		&p.ID, &p.BusinessID, &p.SubscriptionID, &p.PlanID, &p.BillingCycle,
		&p.PaddleTransactionID, &p.PaddleSubscriptionID, &p.PaddleCustomerID,
		&p.Amount, &p.Currency, &p.PeriodStart, &p.PeriodEnd,
		&p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	return &p, err
}

func scanPayments(rows *sql.Rows) ([]models.PaymentRecord, error) {
	var payments []models.PaymentRecord
	for rows.Next() {
		var p models.PaymentRecord
		if err := rows.Scan(
			&p.ID, &p.BusinessID, &p.SubscriptionID, &p.PlanID, &p.BillingCycle,
			&p.PaddleTransactionID, &p.PaddleSubscriptionID, &p.PaddleCustomerID,
			&p.Amount, &p.Currency, &p.PeriodStart, &p.PeriodEnd,
			&p.Status, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}
