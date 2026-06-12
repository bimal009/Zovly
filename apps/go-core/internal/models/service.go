// internal/models/service.go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// ─── ServiceFeature ───────────────────────────────────────────────────────────

type ServiceFeature struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type ServiceFeatures []ServiceFeature

func (f ServiceFeatures) Value() (driver.Value, error) {
	if f == nil {
		return "[]", nil
	}
	b, err := json.Marshal(f)
	return string(b), err
}

func (f *ServiceFeatures) Scan(src any) error {
	if src == nil {
		*f = ServiceFeatures{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("ServiceFeatures.Scan: unsupported type %T", src)
	}
	if len(b) == 0 || string(b) == "null" {
		*f = ServiceFeatures{}
		return nil
	}
	return json.Unmarshal(b, f)
}

type ServiceType string
type ServiceStatus string
type BillingInterval string

const (
	ServiceTypeAppointment ServiceType = "appointment"
	ServiceTypeMembership  ServiceType = "membership"
	ServiceTypeClass       ServiceType = "class"
	ServiceTypePackage     ServiceType = "package"
)

const (
	ServiceStatusActive   ServiceStatus = "active"
	ServiceStatusInactive ServiceStatus = "inactive"
	ServiceStatusArchived ServiceStatus = "archived"
)

const (
	BillingIntervalWeekly    BillingInterval = "weekly"
	BillingIntervalMonthly   BillingInterval = "monthly"
	BillingIntervalQuarterly BillingInterval = "quarterly"
	BillingIntervalYearly    BillingInterval = "yearly"
)

// ─── Core model ──────────────────────────────────────────────────────────────

type Service struct {
	ID         string `db:"id"          json:"id"`
	BusinessID string `db:"business_id" json:"business_id"`

	// Core
	Type        ServiceType   `db:"type"        json:"type"`
	Status      ServiceStatus `db:"status"      json:"status"`
	Name        string        `db:"name"        json:"name"`
	Description *string       `db:"description" json:"description,omitempty"`

	// Pricing (cents)
	Price     int    `db:"price"      json:"price"`
	CostPrice *int   `db:"cost_price" json:"cost_price,omitempty"`
	MRP       *int   `db:"mrp"        json:"mrp,omitempty"`
	Currency  string `db:"currency"   json:"currency"`

	// Payment
	RequiresDeposit bool `db:"requires_deposit" json:"requires_deposit"`
	DepositAmount   *int `db:"deposit_amount"   json:"deposit_amount,omitempty"`

	// appointment + class
	DurationMin      *int    `db:"duration_min"     json:"duration_min,omitempty"`
	BufferMin        *int    `db:"buffer_min"       json:"buffer_min,omitempty"`
	MaxAdvanceDays   *int    `db:"max_advance_days" json:"max_advance_days,omitempty"`
	GoogleCalendarID *string `db:"google_calendar_id" json:"google_calendar_id,omitempty"`

	// class only
	MaxConcurrent *int `db:"max_concurrent" json:"max_concurrent,omitempty"`

	// membership only
	BillingInterval *BillingInterval `db:"billing_interval" json:"billing_interval,omitempty"`
	TrialDays       *int             `db:"trial_days"       json:"trial_days,omitempty"`

	// package only
	SessionCount *int `db:"session_count" json:"session_count,omitempty"`
	ValidityDays *int `db:"validity_days" json:"validity_days,omitempty"`

	// Features / inclusions (shown to customers)
	Features ServiceFeatures `db:"features" json:"features"`

	// Media
	Images pq.StringArray `db:"images" json:"images"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ─── Create ───────────────────────────────────────────────────────────────────

type CreateServiceInput struct {
	BusinessID string `json:"business_id" validate:"required,uuid"`

	Type        ServiceType   `json:"type"        validate:"required,oneof=appointment membership class package"`
	Status      ServiceStatus `json:"status"      validate:"omitempty,oneof=active inactive archived"`
	Name        string        `json:"name"        validate:"required,min=1,max=255"`
	Description *string       `json:"description"`

	Price     int    `json:"price"      validate:"required,gt=0"`
	CostPrice *int   `json:"cost_price" validate:"omitempty,gt=0"`
	MRP       *int   `json:"mrp"        validate:"omitempty,gt=0"`
	Currency  string `json:"currency"   validate:"omitempty,len=3"`

	RequiresDeposit bool `json:"requires_deposit"`
	DepositAmount   *int `json:"deposit_amount" validate:"omitempty,gt=0"`

	// appointment + class
	DurationMin      *int    `json:"duration_min"      validate:"omitempty,gt=0"`
	BufferMin        *int    `json:"buffer_min"        validate:"omitempty,min=0"`
	MaxAdvanceDays   *int    `json:"max_advance_days"  validate:"omitempty,gt=0"`
	GoogleCalendarID *string `json:"google_calendar_id"`

	// class only
	MaxConcurrent *int `json:"max_concurrent" validate:"omitempty,gt=0"`

	// membership only
	BillingInterval *BillingInterval `json:"billing_interval" validate:"omitempty,oneof=weekly monthly quarterly yearly"`
	TrialDays       *int             `json:"trial_days"       validate:"omitempty,min=0"`

	// package only
	SessionCount *int `json:"session_count" validate:"omitempty,gt=0"`
	ValidityDays *int `json:"validity_days" validate:"omitempty,gt=0"`

	Images    []string         `json:"images"   validate:"omitempty,dive,url"`
	Features  []ServiceFeature `json:"features"`
}

// ─── Update ───────────────────────────────────────────────────────────────────
// All fields are pointers — only non-nil fields are written to DB

type UpdateServiceInput struct {
	Status      *ServiceStatus `json:"status"      validate:"omitempty,oneof=active inactive archived"`
	Name        *string        `json:"name"        validate:"omitempty,min=1,max=255"`
	Description *string        `json:"description"`

	Price     *int    `json:"price"      validate:"omitempty,gt=0"`
	CostPrice *int    `json:"cost_price" validate:"omitempty,gt=0"`
	MRP       *int    `json:"mrp"        validate:"omitempty,gt=0"`
	Currency  *string `json:"currency"   validate:"omitempty,len=3"`

	RequiresDeposit *bool `json:"requires_deposit"`
	DepositAmount   *int  `json:"deposit_amount" validate:"omitempty,gt=0"`

	// appointment + class
	DurationMin      *int    `json:"duration_min"      validate:"omitempty,gt=0"`
	BufferMin        *int    `json:"buffer_min"        validate:"omitempty,min=0"`
	MaxAdvanceDays   *int    `json:"max_advance_days"  validate:"omitempty,gt=0"`
	GoogleCalendarID *string `json:"google_calendar_id"`

	// class only
	MaxConcurrent *int `json:"max_concurrent" validate:"omitempty,gt=0"`

	// membership only
	BillingInterval *BillingInterval `json:"billing_interval" validate:"omitempty,oneof=weekly monthly quarterly yearly"`
	TrialDays       *int             `json:"trial_days"       validate:"omitempty,min=0"`

	// package only
	SessionCount *int `json:"session_count" validate:"omitempty,gt=0"`
	ValidityDays *int `json:"validity_days" validate:"omitempty,gt=0"`

	Images   []string         `json:"images"   validate:"omitempty,dive,url"`
	Features []ServiceFeature `json:"features"`
}
