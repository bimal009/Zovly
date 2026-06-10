package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	paddlenotification "github.com/PaddleHQ/paddle-go-sdk/v5/pkg/paddlenotification"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
)

type PaddleService interface {
	HandleEvent(ctx context.Context, body []byte, eventType paddlenotification.EventTypeName) error
}

type paddleService struct {
	subRepo  repository.SubscriptionRepo
	planRepo repository.PlanRepo
	payRepo  repository.PaymentRepo
}

func NewPaddleService(
	subRepo repository.SubscriptionRepo,
	planRepo repository.PlanRepo,
	payRepo repository.PaymentRepo,
) PaddleService {
	return &paddleService{subRepo: subRepo, planRepo: planRepo, payRepo: payRepo}
}

func (s *paddleService) HandleEvent(ctx context.Context, body []byte, eventType paddlenotification.EventTypeName) error {
	switch eventType {
	case paddlenotification.EventTypeNameSubscriptionCreated:
		var e paddlenotification.SubscriptionCreated
		if err := json.Unmarshal(body, &e); err != nil {
			return err
		}
		return s.onSubscriptionCreated(ctx, &e)

	case paddlenotification.EventTypeNameSubscriptionUpdated:
		var e paddlenotification.SubscriptionUpdated
		if err := json.Unmarshal(body, &e); err != nil {
			return err
		}
		return s.onSubscriptionUpdated(ctx, &e)

	case paddlenotification.EventTypeNameSubscriptionCanceled:
		var e paddlenotification.SubscriptionCanceled
		if err := json.Unmarshal(body, &e); err != nil {
			return err
		}
		return s.onSubscriptionCancelled(ctx, &e)

	case paddlenotification.EventTypeNameSubscriptionPaused:
		var e paddlenotification.SubscriptionPaused
		if err := json.Unmarshal(body, &e); err != nil {
			return err
		}
		return s.onSubscriptionPaused(ctx, &e)

	case paddlenotification.EventTypeNameSubscriptionResumed:
		var e paddlenotification.SubscriptionResumed
		if err := json.Unmarshal(body, &e); err != nil {
			return err
		}
		return s.onSubscriptionResumed(ctx, &e)

	case paddlenotification.EventTypeNameTransactionCompleted:
		var e paddlenotification.TransactionCompleted
		if err := json.Unmarshal(body, &e); err != nil {
			return err
		}
		return s.onTransactionCompleted(ctx, &e)

	case paddlenotification.EventTypeNameTransactionPaymentFailed:
		var e paddlenotification.TransactionPaymentFailed
		if err := json.Unmarshal(body, &e); err != nil {
			return err
		}
		return s.onTransactionFailed(ctx, &e)

	default:
		fmt.Printf("[paddle] unhandled event: %s\n", eventType)
		return nil
	}
}

func (s *paddleService) onSubscriptionCreated(ctx context.Context, e *paddlenotification.SubscriptionCreated) error {
	d := e.Data

	// business_id must be passed in custom_data when creating the Paddle checkout session
	businessID, _ := d.CustomData["business_id"].(string)
	if businessID == "" {
		return fmt.Errorf("onSubscriptionCreated: missing business_id in custom_data")
	}

	plan, err := s.planRepo.GetByPaddlePriceID(ctx, d.Items[0].Price.ID)
	if err != nil || plan == nil {
		return fmt.Errorf("onSubscriptionCreated: plan not found for price %s", d.Items[0].Price.ID)
	}

	paddleSubID := d.ID
	customerID := d.CustomerID
	priceID := d.Items[0].Price.ID
	status := mapStatus(string(d.Status))
	billingCycle := mapBillingCycle(d.BillingCycle)

	var periodStart, periodEnd *time.Time
	if d.CurrentBillingPeriod != nil {
		periodStart = parseRFC3339(d.CurrentBillingPeriod.StartsAt)
		periodEnd = parseRFC3339(d.CurrentBillingPeriod.EndsAt)
	}

	// Try to find an existing subscription for this business (e.g. re-subscribing after cancel)
	existing, err := s.subRepo.GetByBusinessID(ctx, businessID)
	if err != nil {
		return fmt.Errorf("onSubscriptionCreated: lookup failed: %w", err)
	}

	if existing != nil {
		return s.subRepo.Update(ctx, nil, existing.ID, models.SubscriptionUpdate{
			PlanID:               plan.ID,
			PaddleSubscriptionID: &paddleSubID,
			PaddleCustomerID:     &customerID,
			PaddlePriceID:        &priceID,
			Status:               status,
			CurrentPeriodStart:   periodStart,
			CurrentPeriodEnd:     periodEnd,
			CancelAtPeriodEnd:    false,
		})
	}

	// No existing record — create a fresh subscription
	_, err = s.subRepo.Create(ctx, nil, models.BusinessSubscription{
		BusinessID:           businessID,
		PlanID:               plan.ID,
		PaddleSubscriptionID: &paddleSubID,
		PaddleCustomerID:     &customerID,
		PaddlePriceID:        &priceID,
		BillingCycle:         billingCycle,
		Status:               models.PlanStatus(status),
		CurrentPeriodStart:   periodStart,
		CurrentPeriodEnd:     periodEnd,
		CancelAtPeriodEnd:    false,
	})
	return err
}

func (s *paddleService) onSubscriptionUpdated(ctx context.Context, e *paddlenotification.SubscriptionUpdated) error {
	d := e.Data

	sub, err := s.subRepo.GetByPaddleSubscriptionID(ctx, d.ID)
	if err != nil || sub == nil {
		return fmt.Errorf("onSubscriptionUpdated: not found %s", d.ID)
	}

	plan, err := s.planRepo.GetByPaddlePriceID(ctx, d.Items[0].Price.ID)
	if err != nil || plan == nil {
		return fmt.Errorf("onSubscriptionUpdated: plan not found")
	}

	priceID := d.Items[0].Price.ID
	status := mapStatus(string(d.Status))

	var periodStart, periodEnd *time.Time
	if d.CurrentBillingPeriod != nil {
		periodStart = parseRFC3339(d.CurrentBillingPeriod.StartsAt)
		periodEnd = parseRFC3339(d.CurrentBillingPeriod.EndsAt)
	}

	return s.subRepo.Update(ctx, nil, sub.ID, models.SubscriptionUpdate{
		PlanID:             plan.ID,
		PaddlePriceID:      &priceID,
		Status:             status,
		CurrentPeriodStart: periodStart,
		CurrentPeriodEnd:   periodEnd,
		CancelAtPeriodEnd:  d.ScheduledChange != nil,
	})
}

func (s *paddleService) onSubscriptionCancelled(ctx context.Context, e *paddlenotification.SubscriptionCanceled) error {
	sub, err := s.subRepo.GetByPaddleSubscriptionID(ctx, e.Data.ID)
	if err != nil || sub == nil {
		return fmt.Errorf("onSubscriptionCancelled: not found %s", e.Data.ID)
	}
	now := time.Now()
	return s.subRepo.Update(ctx, nil, sub.ID, models.SubscriptionUpdate{
		Status:      "cancelled",
		CancelledAt: &now,
	})
}

func (s *paddleService) onSubscriptionPaused(ctx context.Context, e *paddlenotification.SubscriptionPaused) error {
	sub, err := s.subRepo.GetByPaddleSubscriptionID(ctx, e.Data.ID)
	if err != nil || sub == nil {
		return fmt.Errorf("onSubscriptionPaused: not found %s", e.Data.ID)
	}
	now := time.Now()
	return s.subRepo.Update(ctx, nil, sub.ID, models.SubscriptionUpdate{
		Status:   "paused",
		PausedAt: &now,
	})
}

func (s *paddleService) onSubscriptionResumed(ctx context.Context, e *paddlenotification.SubscriptionResumed) error {
	sub, err := s.subRepo.GetByPaddleSubscriptionID(ctx, e.Data.ID)
	if err != nil || sub == nil {
		return fmt.Errorf("onSubscriptionResumed: not found %s", e.Data.ID)
	}
	return s.subRepo.Update(ctx, nil, sub.ID, models.SubscriptionUpdate{
		Status: "active",
	})
}

func (s *paddleService) onTransactionCompleted(ctx context.Context, e *paddlenotification.TransactionCompleted) error {
	d := e.Data

	if d.SubscriptionID == nil {
		return nil // one-time transaction, skip
	}

	sub, err := s.subRepo.GetByPaddleSubscriptionID(ctx, *d.SubscriptionID)
	if err != nil || sub == nil {
		return fmt.Errorf("onTransactionCompleted: sub not found")
	}

	plan, err := s.planRepo.GetByPaddlePriceID(ctx, d.Items[0].Price.ID)
	if err != nil || plan == nil {
		return fmt.Errorf("onTransactionCompleted: plan not found")
	}

	var periodStart, periodEnd time.Time
	if d.BillingPeriod != nil {
		if t := parseRFC3339(d.BillingPeriod.StartsAt); t != nil {
			periodStart = *t
		}
		if t := parseRFC3339(d.BillingPeriod.EndsAt); t != nil {
			periodEnd = *t
		}
	}

	_, err = s.payRepo.Create(ctx, nil, models.PaymentRecord{
		BusinessID:           sub.BusinessID,
		SubscriptionID:       &sub.ID,
		PlanID:               &plan.ID,
		BillingCycle:         sub.BillingCycle,
		PaddleTransactionID:  d.ID,
		PaddleSubscriptionID: d.SubscriptionID,
		PaddleCustomerID:     d.CustomerID,
		Amount:               parseAmount(d.Details.Totals.Total),
		Currency:             string(d.CurrencyCode),
		PeriodStart:          periodStart,
		PeriodEnd:            periodEnd,
		Status:               "paid",
	})
	return err
}

func (s *paddleService) onTransactionFailed(ctx context.Context, e *paddlenotification.TransactionPaymentFailed) error {
	if e.Data.SubscriptionID == nil {
		return nil
	}
	sub, err := s.subRepo.GetByPaddleSubscriptionID(ctx, *e.Data.SubscriptionID)
	if err != nil || sub == nil {
		return fmt.Errorf("onTransactionFailed: sub not found")
	}
	return s.subRepo.Update(ctx, nil, sub.ID, models.SubscriptionUpdate{
		Status: "past_due",
	})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func mapStatus(paddleStatus string) string {
	switch paddleStatus {
	case "active":
		return "active"
	case "trialing":
		return "trialing"
	case "past_due":
		return "past_due"
	case "paused":
		return "paused"
	case "canceled":
		return "cancelled"
	default:
		return "active"
	}
}

func mapBillingCycle(d paddlenotification.Duration) models.BillingCycle {
	if d.Interval == paddlenotification.IntervalYear {
		return models.BillingCycleYearly
	}
	return models.BillingCycleMonthly
}

func parseRFC3339(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

func parseAmount(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
