package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/models"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type ChatHandler struct {
	facebookService service.FacebookService
	chatService     service.ChatService
	rdb             *redis.Client
	cfg             *config.Config
	httpClient      *http.Client
	log             *slog.Logger
}

func NewChatHandler(
	facebookService service.FacebookService,
	chatService service.ChatService,
	rdb *redis.Client,
	cfg *config.Config,
	log *slog.Logger,
) *ChatHandler {
	return &ChatHandler{
		facebookService: facebookService,
		chatService:     chatService,
		rdb:             rdb,
		cfg:             cfg,
		httpClient:      &http.Client{Timeout: 10 * time.Second},
		log:             log,
	}
}

func (h *ChatHandler) HandleChallenge(c *gin.Context) {
	challenge := c.Query("hub.challenge")
	token := c.Query("hub.verify_token")

	if token == h.cfg.Meta.WebhookVerifyToken {
		h.log.Info("facebook webhook verified")
		c.String(http.StatusOK, challenge)
		return
	}

	h.log.Warn("facebook webhook verification failed")
	c.Status(http.StatusForbidden)
}

func (h *ChatHandler) MetaWebhook(c *gin.Context) {

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.log.Error("failed to read webhook body", "error", err)
		c.Status(http.StatusBadRequest)
		return
	}

	sig := c.GetHeader("X-Hub-Signature-256")
	if !verifySignature(body, sig, h.cfg.Meta.AppSecret) {
		h.log.Warn("facebook webhook signature mismatch")
		c.Status(http.StatusUnauthorized)
		return
	}

	var payload models.MetaWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.log.Error("failed to parse webhook payload", "error", err)
		c.Status(http.StatusBadRequest)
		return
	}

	switch payload.Object {
	case "page":
		for _, entry := range payload.Entry {
			for _, event := range entry.Messaging {
				if event.Message != nil && !event.Message.IsEcho {
					if err := h.chatService.HandleInboundMessage(
						c.Request.Context(),
						"facebook",
						entry.ID,
						event,
					); err != nil {
						h.log.Error(
							"handle facebook message failed",
							"page_id",
							entry.ID,
							"error",
							err,
						)
					}
				}
			}
		}

	case "instagram":
		for _, entry := range payload.Entry {
			for _, event := range entry.Messaging {
				if event.Message != nil && !event.Message.IsEcho {
					if err := h.chatService.HandleInboundMessage(
						c.Request.Context(),
						"instagram",
						entry.ID,
						event,
					); err != nil {
						h.log.Error(
							"handle instagram message failed",
							"page_id",
							entry.ID,
							"error",
							err,
						)
					}
				}
			}
		}
	}

	c.Status(http.StatusOK)
}

func verifySignature(body []byte, header, appSecret string) bool {
	const prefix = "sha256="

	if !strings.HasPrefix(header, prefix) {
		return false
	}

	got, err := hex.DecodeString(strings.TrimPrefix(header, prefix))
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(body)

	return hmac.Equal(mac.Sum(nil), got)
}
