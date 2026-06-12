package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ServiceRepo interface {
	Create(ctx context.Context, input models.CreateServiceInput) (*models.Service, error)
	GetByID(ctx context.Context, id, businessID string) (*models.Service, error)
	List(ctx context.Context, businessID string, f ListServicesFilter) ([]models.Service, error)
	Update(ctx context.Context, id, businessID string, input models.UpdateServiceInput) (*models.Service, error)
	Delete(ctx context.Context, id, businessID string) error
	ListForAIContext(ctx context.Context, businessID string) ([]models.Service, error)
}

type serviceRepo struct {
	db *sqlx.DB
}

func NewServiceRepo(db *sqlx.DB) ServiceRepo {
	return &serviceRepo{db: db}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (r *serviceRepo) Create(ctx context.Context, input models.CreateServiceInput) (*models.Service, error) {
	const q = `
		INSERT INTO services (
			business_id, type, status, name, description,
			price, cost_price, mrp, currency,
			requires_deposit, deposit_amount,
			duration_min, buffer_min, max_advance_days, google_calendar_id,
			max_concurrent,
			billing_interval, trial_days,
			session_count, validity_days,
			features, images
		) VALUES (
			:business_id, :type, :status, :name, :description,
			:price, :cost_price, :mrp, :currency,
			:requires_deposit, :deposit_amount,
			:duration_min, :buffer_min, :max_advance_days, :google_calendar_id,
			:max_concurrent,
			:billing_interval, :trial_days,
			:session_count, :validity_days,
			:features, :images
		)
		RETURNING id`

	status := input.Status
	if status == "" {
		status = models.ServiceStatusActive
	}
	currency := input.Currency
	if currency == "" {
		currency = "NPR"
	}

	row, err := r.db.NamedQueryContext(ctx, q, map[string]any{
		"business_id":        input.BusinessID,
		"type":               string(input.Type),
		"status":             status,
		"name":               input.Name,
		"description":        input.Description,
		"price":              input.Price,
		"cost_price":         input.CostPrice,
		"mrp":                input.MRP,
		"currency":           currency,
		"requires_deposit":   input.RequiresDeposit,
		"deposit_amount":     input.DepositAmount,
		"duration_min":       input.DurationMin,
		"buffer_min":         input.BufferMin,
		"max_advance_days":   input.MaxAdvanceDays,
		"google_calendar_id": input.GoogleCalendarID,
		"max_concurrent":     input.MaxConcurrent,
		"billing_interval":   input.BillingInterval,
		"trial_days":         input.TrialDays,
		"session_count":      input.SessionCount,
		"validity_days":      input.ValidityDays,
		"features":           models.ServiceFeatures(input.Features),
		"images":             pq.Array(orEmptySlice(input.Images)),
	})
	if err != nil {
		return nil, fmt.Errorf("service create: %w", err)
	}
	defer row.Close()

	var id string
	if row.Next() {
		if err := row.Scan(&id); err != nil {
			return nil, fmt.Errorf("service create scan id: %w", err)
		}
	}

	return r.GetByID(ctx, id, input.BusinessID)
}

// ─── GetByID ──────────────────────────────────────────────────────────────────

func (r *serviceRepo) GetByID(ctx context.Context, id, businessID string) (*models.Service, error) {
	var s models.Service
	err := r.db.GetContext(ctx, &s, `
		SELECT * FROM services
		WHERE id = $1 AND business_id = $2
	`, id, businessID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("service get: %w", err)
	}
	return &s, nil
}

// ─── List ─────────────────────────────────────────────────────────────────────

type ListServicesFilter struct {
	Type   *models.ServiceType
	Status *models.ServiceStatus
	Search string
	Limit  int
	Offset int
}

func (r *serviceRepo) List(ctx context.Context, businessID string, f ListServicesFilter) ([]models.Service, error) {
	args := []any{businessID}
	conds := []string{"business_id = $1"}
	i := 2

	if f.Type != nil {
		conds = append(conds, fmt.Sprintf("type = $%d", i))
		args = append(args, string(*f.Type))
		i++
	}
	if f.Status != nil {
		conds = append(conds, fmt.Sprintf("status = $%d", i))
		args = append(args, string(*f.Status))
		i++
	}
	if f.Search != "" {
		conds = append(conds, fmt.Sprintf("name ILIKE $%d", i))
		args = append(args, "%"+f.Search+"%")
		i++
	}

	limit := 50
	if f.Limit > 0 && f.Limit <= 100 {
		limit = f.Limit
	}

	q := fmt.Sprintf(`
		SELECT * FROM services
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		strings.Join(conds, " AND "), i, i+1,
	)
	args = append(args, limit, f.Offset)

	var services []models.Service
	if err := r.db.SelectContext(ctx, &services, q, args...); err != nil {
		return nil, fmt.Errorf("service list: %w", err)
	}
	return services, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (r *serviceRepo) Update(ctx context.Context, id, businessID string, input models.UpdateServiceInput) (*models.Service, error) {
	fields := map[string]any{}

	if input.Status != nil {
		fields["status"] = string(*input.Status)
	}
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.Price != nil {
		fields["price"] = *input.Price
	}
	if input.CostPrice != nil {
		fields["cost_price"] = *input.CostPrice
	}
	if input.MRP != nil {
		fields["mrp"] = *input.MRP
	}
	if input.Currency != nil {
		fields["currency"] = *input.Currency
	}
	if input.RequiresDeposit != nil {
		fields["requires_deposit"] = *input.RequiresDeposit
	}
	if input.DepositAmount != nil {
		fields["deposit_amount"] = *input.DepositAmount
	}
	if input.DurationMin != nil {
		fields["duration_min"] = *input.DurationMin
	}
	if input.BufferMin != nil {
		fields["buffer_min"] = *input.BufferMin
	}
	if input.MaxAdvanceDays != nil {
		fields["max_advance_days"] = *input.MaxAdvanceDays
	}
	if input.GoogleCalendarID != nil {
		fields["google_calendar_id"] = *input.GoogleCalendarID
	}
	if input.MaxConcurrent != nil {
		fields["max_concurrent"] = *input.MaxConcurrent
	}
	if input.BillingInterval != nil {
		fields["billing_interval"] = string(*input.BillingInterval)
	}
	if input.TrialDays != nil {
		fields["trial_days"] = *input.TrialDays
	}
	if input.SessionCount != nil {
		fields["session_count"] = *input.SessionCount
	}
	if input.ValidityDays != nil {
		fields["validity_days"] = *input.ValidityDays
	}
	if input.Features != nil {
		fields["features"] = models.ServiceFeatures(input.Features)
	}
	if input.Images != nil {
		fields["images"] = pq.Array(orEmptySlice(input.Images))
	}

	if len(fields) == 0 {
		return r.GetByID(ctx, id, businessID)
	}

	fields["updated_at"] = time.Now()
	fields["id"] = id
	fields["business_id"] = businessID

	setClauses := make([]string, 0, len(fields))
	for col := range fields {
		if col == "id" || col == "business_id" {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", col, col))
	}

	q := fmt.Sprintf(`
		UPDATE services
		SET %s
		WHERE id = :id AND business_id = :business_id`,
		strings.Join(setClauses, ", "),
	)

	if _, err := r.db.NamedExecContext(ctx, q, fields); err != nil {
		return nil, fmt.Errorf("service update: %w", err)
	}

	return r.GetByID(ctx, id, businessID)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (r *serviceRepo) Delete(ctx context.Context, id, businessID string) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM services WHERE id = $1 AND business_id = $2
	`, id, businessID)
	if err != nil {
		return fmt.Errorf("service delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("service delete: not found or not owned")
	}
	return nil
}

// ─── ListForAIContext ─────────────────────────────────────────────────────────

func (r *serviceRepo) ListForAIContext(ctx context.Context, businessID string) ([]models.Service, error) {
	var services []models.Service
	err := r.db.SelectContext(ctx, &services, `
		SELECT id, type, name, description, price, mrp, currency,
		       duration_min, max_concurrent, billing_interval,
		       session_count, validity_days
		FROM services
		WHERE business_id = $1
		  AND status = 'active'
		ORDER BY type, name
	`, businessID)
	if err != nil {
		return nil, fmt.Errorf("service ai context: %w", err)
	}
	return services, nil
}
