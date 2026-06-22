package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/models"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/bimal009/Zovly/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type InstagramHandler struct {
	instagramService service.InstagramService
	rdb              *redis.Client
	cfg              *config.Config
	httpClient       *http.Client
	log              *slog.Logger
}

func NewInstagramHandler(rdb *redis.Client, cfg *config.Config, log *slog.Logger, instagramService service.InstagramService) *InstagramHandler {
	return &InstagramHandler{
		instagramService: instagramService,
		rdb:              rdb,
		cfg:              cfg,
		httpClient:       &http.Client{Timeout: 10 * time.Second},
		log:              log,
	}
}

func (h *InstagramHandler) ConnectInstagram(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	state, err := utils.GenerateSecureToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to generate state"))
		return
	}

	if err := h.rdb.Set(c.Request.Context(), "oauth:instagram:state:"+state, businessID, 10*time.Minute).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to start oauth"))
		return
	}

	params := url.Values{
		"client_id":     {h.cfg.Instagram.AppID},
		"redirect_uri":  {h.cfg.Instagram.RedirectURI},
		"state":         {state},
		"scope":         {"instagram_business_basic,instagram_business_content_publish,instagram_business_manage_comments,instagram_business_manage_messages"},
		"response_type": {"code"},
	}
	authURL := "https://www.instagram.com/oauth/authorize?" + params.Encode()
	c.JSON(http.StatusOK, responses.Success("instagram oauth url", gin.H{"url": authURL}))
}

type igTokenResponse struct {
	AccessToken string   `json:"access_token"`
	UserID      int64    `json:"user_id"`
	Permissions []string `json:"permissions,omitempty"`
}

func (h *InstagramHandler) exchangeCodeForToken(ctx context.Context, code string) (*igTokenResponse, error) {
	form := url.Values{
		"client_id":     {h.cfg.Instagram.AppID},
		"client_secret": {h.cfg.Instagram.AppSecret},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {h.cfg.Instagram.RedirectURI},
		"code":          {code},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.instagram.com/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build ig token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call ig token endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ig token exchange failed (%d): %s", resp.StatusCode, body)
	}

	var result igTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode ig token response: %w", err)
	}
	return &result, nil
}

func (h *InstagramHandler) exchangeForLongLivedToken(ctx context.Context, shortLivedToken string) (*models.IgLongLivedResponse, error) {
	params := url.Values{
		"grant_type":    {"ig_exchange_token"},
		"client_secret": {h.cfg.Instagram.AppSecret},
		"access_token":  {shortLivedToken},
	}
	reqURL := "https://graph.instagram.com/access_token?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build ig long-lived request: %w", err)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call ig long-lived endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ig long-lived exchange failed (%d): %s", resp.StatusCode, body)
	}

	var result models.IgLongLivedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode ig long-lived response: %w", err)
	}
	return &result, nil
}
func (h *InstagramHandler) InstagramCallback(c *gin.Context) {
	ctx := c.Request.Context()

	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, responses.BadRequest("missing code or state"))
		return
	}
	if errParam := c.Query("error"); errParam != "" {
		c.JSON(http.StatusBadRequest, responses.BadRequest("instagram authorization denied: "+errParam))
		return
	}

	key := "oauth:instagram:state:" + state
	businessID, err := h.rdb.Get(ctx, key).Result()
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.BadRequest("invalid or expired state"))
		return
	}
	h.rdb.Del(ctx, key)

	shortLived, err := h.exchangeCodeForToken(ctx, code)
	if err != nil {
		h.log.Error("instagram token exchange failed", "business_id", businessID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	longLived, err := h.exchangeForLongLivedToken(ctx, shortLived.AccessToken)
	if err != nil {
		h.log.Error("instagram long-lived token exchange failed", "business_id", businessID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError(err.Error()))
		return
	}

	profile, err := h.instagramService.FetchUserProfile(ctx, longLived.AccessToken)
	if err != nil {
		h.log.Error("failed to fetch instagram profile", "business_id", businessID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to fetch instagram profile"))
		return
	}

	// Webhooks identify the account by its Instagram professional account ID
	// (the /me user_id field), NOT the OAuth token's app-scoped user_id. Storing
	// the wrong one makes inbound-message credential lookups miss.
	igUserID := profile.UserID
	if igUserID == "" {
		igUserID = strconv.FormatInt(shortLived.UserID, 10) // fallback
	}
	if err := h.instagramService.SaveConnection(ctx, businessID, igUserID, profile.Username, longLived); err != nil {
		h.log.Error("failed to save instagram connection", "business_id", businessID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to save connection"))
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, h.cfg.App.FrontendURL+"/"+businessID+"/connections/instagram")
}

func (h *InstagramHandler) GetConnectionStatus(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	status, err := h.instagramService.GetConnectionStatus(c.Request.Context(), businessID)
	if err != nil {
		h.log.Error("failed to get instagram connection status", "business_id", businessID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to get connection status"))
		return
	}

	c.JSON(http.StatusOK, responses.Success("instagram connection status", status))
}

// ActivateInstagram flips the connected (but inactive) Instagram credential to
// active. The OAuth callback stores it inactive on purpose, so the user must
// explicitly click "Connect with app" in the UI before messaging can be enabled.
func (h *InstagramHandler) ActivateInstagram(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	if err := h.instagramService.ActivateConnection(c.Request.Context(), businessID); err != nil {
		h.log.Error("failed to activate instagram connection", "business_id", businessID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to activate instagram connection"))
		return
	}

	h.log.Info("instagram connection activated", "business_id", businessID)
	c.JSON(http.StatusOK, responses.Success("instagram activated", gin.H{"is_active": true}))
}

// SubscribeWebhook subscribes the connected Instagram account to messaging
// webhook fields and stamps webhook_subscribed_at. Mirrors the Facebook
// messenger subscribe endpoint.
func (h *InstagramHandler) SubscribeWebhook(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	if err := h.instagramService.SubscribeWebhook(c.Request.Context(), businessID); err != nil {
		h.log.Error("failed to subscribe instagram webhook", "business_id", businessID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to subscribe instagram webhook"))
		return
	}

	h.log.Info("instagram webhook subscribed", "business_id", businessID)
	c.JSON(http.StatusOK, responses.Success[any]("instagram subscribed to webhooks", nil))
}

// VerifyWebhook handles Meta's GET verification challenge for the Instagram webhook.
func (h *InstagramHandler) VerifyWebhook(c *gin.Context) {
	challenge := c.Query("hub.challenge")
	token := c.Query("hub.verify_token")

	if token == h.cfg.Meta.WebhookVerifyToken {
		h.log.Info("instagram webhook verified")
		c.String(http.StatusOK, challenge)
		return
	}

	h.log.Warn("instagram webhook verification failed")
	c.Status(http.StatusForbidden)
}

// Webhook receives Instagram messaging events. Signature is verified with
// IG_APP_SECRET (separate from the Facebook app secret).
func (h *InstagramHandler) Webhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.log.Error("failed to read instagram webhook body", "error", err)
		c.Status(http.StatusBadRequest)
		return
	}

	// Instagram is connected via the Instagram Login app (IG_APP_ID), so Meta
	// signs its webhooks with IG_APP_SECRET — distinct from the Facebook app's
	// META_APP_SECRET.
	sig := c.GetHeader("X-Hub-Signature-256")
	if !verifyMetaSignature(body, sig, h.cfg.Meta.AppSecret) {
		h.log.Warn("instagram webhook signature mismatch")
		c.Status(http.StatusUnauthorized)
		return
	}

	var payload models.InstagramWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.log.Error("failed to parse instagram payload", "error", err)
		c.Status(http.StatusBadRequest)
		return
	}

	// Guard against misrouted events — only handle Instagram objects here.
	if payload.Object != "instagram" {
		h.log.Warn("instagram webhook received unexpected object — check dashboard callback URL", "object", payload.Object)
		c.Status(http.StatusOK)
		return
	}

	h.log.Info("instagram webhook received", "body", string(body))

	for _, entry := range payload.Entry {
		for _, igEvent := range entry.Messaging {
			if igEvent.Message == nil {
				// non-message events (message_edit/seen/reaction) — ignore
				continue
			}
			if igEvent.Message.IsEcho {
				// our own outbound message echoed back — ignore
				continue
			}
			accountID := igEvent.Recipient.ID
			if accountID == "" {
				accountID = entry.ID
			}
			if err := h.instagramService.HandleInstagramInboundMessage(
				c.Request.Context(), models.PlatformInstagram, accountID, igEvent,
			); err != nil {
				h.log.Error("handle instagram message failed", "account_id", accountID, "error", err)
			}
		}
	}

	c.Status(http.StatusOK)
}
