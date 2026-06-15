package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
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
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

type igTokenResponse struct {
	AccessToken string `json:"access_token"`
	UserID      string `json:"user_id"`
	Permissions string `json:"permissions,omitempty"`
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
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("token exchange failed"))
		return
	}

	longLived, err := h.exchangeForLongLivedToken(ctx, shortLived.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("long-lived token exchange failed"))
		return
	}

	if err := h.instagramService.SaveConnection(ctx, businessID, shortLived.UserID, longLived); err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to save connection"))
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, h.cfg.App.FrontendURL+"/settings/connections?instagram=connected")
}
