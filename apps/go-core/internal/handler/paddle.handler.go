package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	paddle "github.com/PaddleHQ/paddle-go-sdk/v5"
	paddlenotification "github.com/PaddleHQ/paddle-go-sdk/v5/pkg/paddlenotification"
	"github.com/bimal009/Zovly/internal/config"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/gin-gonic/gin"
)

type PaddleHandler struct {
	cfg       config.Config
	paddleSvc service.PaddleService
}

func NewPaddleHandler(
	cfg config.Config,
	subRepo repository.SubscriptionRepo,
	planRepo repository.PlanRepo,
	payRepo repository.PaymentRepo,
) *PaddleHandler {
	return &PaddleHandler{
		cfg:       cfg,
		paddleSvc: service.NewPaddleService(subRepo, planRepo, payRepo),
	}
}

func (h *PaddleHandler) Webhook(c *gin.Context) {
	verifier := paddle.NewWebhookVerifier(h.cfg.Paddle.WebhookSecret)

	handler := verifier.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("[paddle] ❌ failed to read body: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var generic paddlenotification.GenericNotificationEvent
		if err := json.Unmarshal(body, &generic); err != nil {
			fmt.Printf("[paddle] ❌ failed to parse event: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// ── Log every event ───────────────────────────────────────────────────
		dataJSON, _ := json.MarshalIndent(generic.Data, "", "  ")
		fmt.Printf("\n========================================\n")
		fmt.Printf("[paddle] 📦 %s\n", generic.EventType)
		fmt.Printf("[paddle] 🆔 %s\n", generic.EventID)
		fmt.Printf("[paddle] 🕐 %s\n", generic.OccurredAt)
		fmt.Printf("----------------------------------------\n")
		fmt.Printf("%s\n", string(dataJSON))
		fmt.Printf("========================================\n\n")

		// ── Dispatch ──────────────────────────────────────────────────────────
		if err := h.paddleSvc.HandleEvent(r.Context(), body, generic.EventType); err != nil {
			fmt.Printf("[paddle] ❌ handler error: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	}))

	handler.ServeHTTP(c.Writer, c.Request)
}

func (h *PaddleHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/webhook", h.Webhook)
}
