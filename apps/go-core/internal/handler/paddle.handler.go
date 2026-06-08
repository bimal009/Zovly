package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	paddle "github.com/PaddleHQ/paddle-go-sdk/v5"
	"github.com/bimal009/Zovly/internal/config"
	"github.com/gin-gonic/gin"
)

type PaddleHandler struct {
	cfg config.Config
}

func NewPaddleHandler(cfg config.Config) *PaddleHandler {
	return &PaddleHandler{cfg: cfg}
}

func (h *PaddleHandler) Webhook(c *gin.Context) {
	verifier := paddle.NewWebhookVerifier(h.cfg.Paddle.WebhookSecret)

	handler := verifier.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the raw body (already verified at this point)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}

		// Decode into a generic map to inspect event_type
		var event map[string]interface{}
		if err := json.Unmarshal(body, &event); err != nil {
			http.Error(w, "failed to parse event", http.StatusBadRequest)
			return
		}

		switch event["event_type"] {
		case "subscription.created":
			fmt.Println("Subscription Created")
		
		case "subscription.updated":
			fmt.Println("Subscription Updated")
		
		case "transaction.completed":
			fmt.Println("Transaction Completed")
		
		default:
			fmt.Printf("Unhandled Event: %v\n", event["event_type"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	}))

	handler.ServeHTTP(c.Writer, c.Request)
}

