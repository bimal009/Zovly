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

	// peek at object field to route to the right parser
	var base struct {
		Object string `json:"object"`
	}
	if err := json.Unmarshal(body, &base); err != nil {
		h.log.Error("failed to parse webhook object", "error", err)
		c.Status(http.StatusBadRequest)
		return
	}

	switch base.Object {
	case "page":
		var payload models.FacebookWebhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			h.log.Error("failed to parse facebook payload", "error", err)
			c.Status(http.StatusBadRequest)
			return
		}
		for _, entry := range payload.Entry {
			for _, event := range entry.Messaging {
				if event.Message == nil || event.Message.IsEcho {
					continue
				}
				pageID := event.Recipient.ID
				if pageID == "" {
					pageID = entry.ID
				}
				if err := h.facebookService.HandleFacebookInboundMessage(
					c.Request.Context(), models.PlatformFacebook, pageID, event,
				); err != nil {
					h.log.Error("handle facebook message failed",
						"page_id", pageID, "error", err)
				}
			}
		}

	case "instagram":
		var payload models.InstagramWebhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			h.log.Error("failed to parse instagram payload", "error", err)
			c.Status(http.StatusBadRequest)
			return
		}
		for _, entry := range payload.Entry {
			for _, igEvent := range entry.Messaging {
				if igEvent.Message == nil || igEvent.Message.IsEcho {
					continue
				}
				accountID := igEvent.Recipient.ID
				if accountID == "" {
					accountID = entry.ID
				}
				if err := h.facebookService.HandleFacebookInboundMessage(
					c.Request.Context(), models.PlatformInstagram, accountID, igEventToFb(igEvent),
				); err != nil {
					h.log.Error("handle instagram message failed",
						"account_id", accountID, "error", err)
				}
			}
		}
	}

	c.Status(http.StatusOK)
}

// igEventToFb converts an Instagram messaging event to the Facebook equivalent
// so that ChatService.HandleInboundMessage can work with a single event type.
// Both platforms share the same Messenger API structure (sender/recipient/message).
func igEventToFb(e models.InstagramMessagingEvent) models.FacebookMessagingEvent {
	fb := models.FacebookMessagingEvent{
		Sender:    models.FacebookUser{ID: e.Sender.ID},
		Recipient: models.FacebookUser{ID: e.Recipient.ID},
		Timestamp: e.Timestamp,
	}
	if e.Message != nil {
		msg := &models.FacebookMessage{
			Mid:    e.Message.Mid,
			Text:   e.Message.Text,
			IsEcho: e.Message.IsEcho,
		}
		for _, a := range e.Message.Attachments {
			msg.Attachments = append(msg.Attachments, models.FacebookAttachment{
				Type:    models.FacebookAttachmentType(a.Type),
				Payload: models.FacebookAttachmentPayload{URL: a.Payload.URL},
			})
		}
		fb.Message = msg
	}
	return fb
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
