package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/models"
	"github.com/bimal009/Zovly/internal/service"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/bimal009/Zovly/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type FacebookHandler struct {
	facebookService service.FacebookService
	chatServive     service.ChatService
	rdb             *redis.Client
	cfg             *config.Config
	httpClient      *http.Client
	log             *slog.Logger
}

func NewFacebookHandler(facebookService service.FacebookService, chatService service.ChatService, rdb *redis.Client, cfg *config.Config, log *slog.Logger) *FacebookHandler {
	return &FacebookHandler{
		facebookService: facebookService,
		chatServive:     chatService,
		rdb:             rdb,
		cfg:             cfg,
		httpClient:      &http.Client{Timeout: 10 * time.Second},
		log:             log,
	}
}

func (h *FacebookHandler) ConnectFacebook(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	state, err := utils.GenerateSecureToken()
	if err != nil {
		h.log.Error("failed to generate oauth state token", "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to generate state"))
		return
	}

	if err := h.rdb.Set(c.Request.Context(), "oauth:facebook:state:"+state, businessID, 10*time.Minute).Err(); err != nil {
		h.log.Error("failed to store oauth state in redis", "business_id", businessID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to start oauth"))
		return
	}

	h.log.Info("redirecting to facebook oauth", "business_id", businessID)

	params := url.Values{
		"client_id":     {h.cfg.Meta.AppID},
		"redirect_uri":  {h.cfg.Meta.RedirectURI},
		"state":         {state},
		"scope":         {"pages_show_list,pages_manage_posts,pages_read_engagement,pages_manage_metadata,pages_read_user_content,pages_messaging,business_management"},
		"response_type": {"code"},
	}
	authURL := "https://www.facebook.com/v25.0/dialog/oauth?" + params.Encode()
	c.JSON(http.StatusOK, responses.Success("facebook oauth url", gin.H{"url": authURL}))
}

func (h *FacebookHandler) FacebookCallback(c *gin.Context) {
	ctx := c.Request.Context()

	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, responses.BadRequest("missing code or state"))
		return
	}

	if errParam := c.Query("error"); errParam != "" {
		c.JSON(http.StatusBadRequest, responses.BadRequest("facebook authorization denied: "+errParam))
		return
	}

	key := "oauth:facebook:state:" + state
	businessID, err := h.rdb.Get(ctx, key).Result()
	if err != nil {
		h.log.Warn("invalid or expired oauth state", "error", err)
		c.JSON(http.StatusBadRequest, responses.BadRequest("invalid or expired state"))
		return
	}
	h.rdb.Del(ctx, key)

	h.log.Info("facebook oauth callback received", "business_id", businessID)

	// 1. code → short-lived token
	shortLived, err := h.exchangeCodeForToken(ctx, code)
	if err != nil {
		h.log.Error("facebook token exchange failed", "business_id", businessID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("token exchange failed"))
		return
	}

	// 2. short-lived → long-lived token
	longLived, err := h.exchangeForLongLivedToken(ctx, shortLived.AccessToken)
	if err != nil {
		h.log.Error("facebook long-lived token exchange failed", "business_id", businessID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("long-lived token exchange failed"))
		return
	}

	// 3. long-lived user token → list of Pages + their Page tokens
	pages, err := h.fetchPages(ctx, longLived.AccessToken)
	if err != nil {
		h.log.Error("failed to fetch facebook pages", "business_id", businessID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to fetch pages"))
		return
	}
	if len(pages) == 0 {
		h.log.Warn("no facebook pages found for account", "business_id", businessID)
		c.JSON(http.StatusBadRequest, responses.BadRequest("no facebook pages found for this account"))
		return
	}

	if err := h.facebookService.SaveConnections(ctx, businessID, pages); err != nil {
		h.log.Error("failed to save facebook connections", "business_id", businessID, "page_count", len(pages), "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to save connections"))
		return
	}

	h.log.Info("facebook pages connected", "business_id", businessID, "page_count", len(pages))
	c.Redirect(http.StatusTemporaryRedirect, h.cfg.App.FrontendURL+"/"+businessID+"/connections/facebook")
}

func (h *FacebookHandler) GetConnectionStatus(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	status, err := h.facebookService.GetConnectionStatus(c, businessID)
	if err != nil {
		h.log.Error("failed to get facebook connection status", "business_id", businessID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to get connection status"))
		return
	}

	c.JSON(http.StatusOK, responses.Success("facebook connection status", status))
}

func (h *FacebookHandler) ToggleFacebookPage(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	pageID := c.Param("pageId")
	if pageID == "" {
		c.JSON(http.StatusBadRequest, responses.BadRequest("missing page id"))
		return
	}

	isActive, err := h.facebookService.TogglePage(c.Request.Context(), businessID, pageID)
	if err != nil {
		h.log.Error("failed to toggle facebook page", "business_id", businessID, "page_id", pageID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to toggle page"))
		return
	}

	h.log.Info("facebook page toggled", "business_id", businessID, "page_id", pageID, "is_active", isActive)
	c.JSON(http.StatusOK, responses.Success("page toggled", gin.H{"is_active": isActive}))
}

func (h *FacebookHandler) exchangeCodeForToken(ctx context.Context, code string) (*models.FbTokenResponse, error) {
	params := url.Values{
		"client_id":     {h.cfg.Meta.AppID},
		"client_secret": {h.cfg.Meta.AppSecret},
		"redirect_uri":  {h.cfg.Meta.RedirectURI},
		"code":          {code},
	}

	reqURL := "https://graph.facebook.com/v25.0/oauth/access_token?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call token endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("facebook token exchange failed (%d): %s", resp.StatusCode, body)
	}

	var result models.FbTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	return &result, nil
}

func (h *FacebookHandler) exchangeForLongLivedToken(ctx context.Context, shortLivedToken string) (*models.FbTokenResponse, error) {
	params := url.Values{
		"grant_type":        {"fb_exchange_token"},
		"client_id":         {h.cfg.Meta.AppID},
		"client_secret":     {h.cfg.Meta.AppSecret},
		"fb_exchange_token": {shortLivedToken},
	}

	reqURL := "https://graph.facebook.com/v25.0/oauth/access_token?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build long-lived token request: %w", err)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call long-lived token endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("facebook long-lived exchange failed (%d): %s", resp.StatusCode, body)
	}

	var result models.FbTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode long-lived token response: %w", err)
	}
	return &result, nil
}

func (h *FacebookHandler) fetchPages(ctx context.Context, userToken string) ([]models.FbPage, error) {
	params := url.Values{"access_token": {userToken}}
	reqURL := "https://graph.facebook.com/v25.0/me/accounts?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build pages request: %w", err)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call pages endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch pages failed (%d): %s", resp.StatusCode, body)
	}

	var result models.FbPagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode pages response: %w", err)
	}
	return result.Data, nil
}

func (h *FacebookHandler) TogglePage(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	var req struct {
		PageID string `json:"page_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.BadRequest("invalid request"))
		return
	}

	isActive, err := h.facebookService.TogglePage(c.Request.Context(), businessID, req.PageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to update page"))
		return
	}

	c.JSON(http.StatusOK, responses.Success("Facebook status updated successfully", gin.H{"is_active": isActive}))
}
func (h *FacebookHandler) SubscribeMessengerWebhook(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	pageID := c.Param("pageId")
	if pageID == "" {
		c.JSON(http.StatusBadRequest, responses.BadRequest("missing page id"))
		return
	}

	if err := h.facebookService.SubscribeMessengerPage(c.Request.Context(), businessID, pageID); err != nil {
		h.log.Error("failed to subscribe messenger webhook", "business_id", businessID, "page_id", pageID, "error", err)
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to subscribe page"))
		return
	}

	h.log.Info("messenger page subscribed", "business_id", businessID, "page_id", pageID)
	c.JSON(http.StatusOK, responses.Success[any]("page subscribed to messenger", nil))
}

func (h *FacebookHandler) handleComment(ctx context.Context, pageID string, change models.FacebookChangeEvent) {
	h.log.Info("page comment received",
		"page_id", pageID,
		"comment_id", change.Value.CommentID,
		"post_id", change.Value.PostID,
	)
	// TODO: enqueue to comment worker
}
