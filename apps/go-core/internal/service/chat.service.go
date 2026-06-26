package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	imagekit "github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

const maxMediaBytes = 10 * 1024 * 1024 // 10 MB

type ChatService interface {
	FindOrCreate(ctx context.Context, tx *sqlx.Tx, conv models.CreateConversation) (*models.Conversation, error)
	StreamMessage(ctx context.Context, platform models.Platform, messageId, businessId, conversationId string, attachments bool) error
	FetchUserProfile(ctx context.Context, psid, pageToken string) (*models.MessengerProfile, error)

	HandleImageAttachment(ctx context.Context, tx *sqlx.Tx, platform models.Platform, conv *models.Conversation, cred *models.AppCredential, pageID, mediaURL string)
	HandleAudioAttachment(ctx context.Context, tx *sqlx.Tx, platform models.Platform, conv *models.Conversation, cred *models.AppCredential, pageID, mediaURL string)
	HandleUnprocessableAttachment(ctx context.Context, tx *sqlx.Tx, platform models.Platform, conv *models.Conversation, cred *models.AppCredential, pageID, mediaURL string, mediaType models.MessageMediaType, label, placeholder string)
	HandleSharedLink(ctx context.Context, tx *sqlx.Tx, platform models.Platform, conv *models.Conversation, cred *models.AppCredential, pageToken string, attachment models.FacebookAttachment)

	// called by the worker (async) to resolve media → text
	GetImageDetails(ctx context.Context, fileURL string) (string, error)
	GetAudioDetails(ctx context.Context, fileURL string) (string, error)

	HandleSharedContent(
		ctx context.Context, tx *sqlx.Tx, platform models.Platform,
		conv *models.Conversation, cred *models.AppCredential,
		contentURL, label, caption string,
	)
}

type chatService struct {
	db                *sqlx.DB
	messageEmbedRepo  repository.MessageEmbeddingRepo
	messageRepo       repository.MessageRepo
	conversationRepo  repository.ConversationRepo
	appCredentialRepo repository.AppCredentialRepo
	cfg               config.Config
	httpClient        *http.Client
	rdb               *redis.Client
	log               *slog.Logger
}

func NewChatService(
	db *sqlx.DB,
	messageRepo repository.MessageRepo,
	appCredentialRepo repository.AppCredentialRepo,
	messageEmbedRepo repository.MessageEmbeddingRepo,
	conversationRepo repository.ConversationRepo,
	cfg config.Config,
	rdb *redis.Client,
	log *slog.Logger,
) ChatService {
	return &chatService{
		db:                db,
		messageEmbedRepo:  messageEmbedRepo,
		messageRepo:       messageRepo,
		conversationRepo:  conversationRepo,
		appCredentialRepo: appCredentialRepo,
		cfg:               cfg,
		httpClient:        &http.Client{Timeout: 120 * time.Second},
		rdb:               rdb,
		log:               log,
	}
}

func (s *chatService) FindOrCreate(ctx context.Context, tx *sqlx.Tx, conv models.CreateConversation) (*models.Conversation, error) {
	return s.conversationRepo.FindOrCreate(ctx, tx, conv)
}

func (s *chatService) StreamMessage(ctx context.Context, platform models.Platform, messageId, businessId, conversationId string, attachments bool) error {
	if _, err := s.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "chat:messages",
		Values: map[string]interface{}{
			"message_id":      messageId,
			"business_id":     businessId,
			"conversation_id": conversationId,
			"platform":        string(platform),
			"attachments":     attachments,
		},
	}).Result(); err != nil {
		s.log.Error("publish to stream failed", "err", err)
		return fmt.Errorf("publish message to stream: %w", err)
	}
	s.log.Info("message published to stream", "message_id", messageId, "conversation_id", conversationId)
	return nil
}

func (s *chatService) HandleImageAttachment(ctx context.Context, tx *sqlx.Tx, platform models.Platform, conv *models.Conversation, cred *models.AppCredential, pageID, mediaURL string) {
	mediaType := models.MediaTypeImage

	if size, err := s.checkContentLength(ctx, mediaURL); err == nil && size > maxMediaBytes {
		s.log.Warn("image exceeds size limit", "size", size)
		content := "[Customer sent an image that's too large to process]"
		s.storeMediaMessage(ctx, tx, platform, conv, cred, &content, &mediaURL, &mediaType, false)
		return
	}

	uploadedURL, err := s.UploadFileFromURL(ctx, mediaURL, safeFilename(pageID, mediaURL))
	if err != nil {
		s.log.Warn("imagekit upload failed, using original url", "err", err)
		uploadedURL = mediaURL
	}

	s.storeMediaMessage(ctx, tx, platform, conv, cred, nil, &uploadedURL, &mediaType, true)
}

func (s *chatService) HandleAudioAttachment(ctx context.Context, tx *sqlx.Tx, platform models.Platform, conv *models.Conversation, cred *models.AppCredential, pageID, mediaURL string) {
	mediaType := models.MediaTypeAudio

	if size, err := s.checkContentLength(ctx, mediaURL); err == nil && size > maxMediaBytes {
		s.log.Warn("audio exceeds size limit", "size", size)
		content := "[Customer sent a voice message that's too large to process]"
		s.storeMediaMessage(ctx, tx, platform, conv, cred, &content, &mediaURL, &mediaType, false)
		return
	}

	uploadedURL, err := s.UploadFileFromURL(ctx, mediaURL, safeFilename(pageID, mediaURL))
	if err != nil {
		s.log.Warn("imagekit upload failed, using original url", "err", err)
		uploadedURL = mediaURL
	}

	s.storeMediaMessage(ctx, tx, platform, conv, cred, nil, &uploadedURL, &mediaType, true)
}

func (s *chatService) HandleUnprocessableAttachment(ctx context.Context, tx *sqlx.Tx, platform models.Platform, conv *models.Conversation, cred *models.AppCredential, pageID, mediaURL string, mediaType models.MessageMediaType, label, placeholder string) {
	content := placeholder
	storedURL := mediaURL

	if size, err := s.checkContentLength(ctx, mediaURL); err == nil && size > maxMediaBytes {
		s.log.Warn(label+" exceeds size limit", "size", size)
		content = fmt.Sprintf("[Customer sent a %s that's too large to process]", label)
	} else if uploaded, uerr := s.UploadFileFromURL(ctx, mediaURL, safeFilename(pageID, mediaURL)); uerr == nil {
		storedURL = uploaded
	} else {
		s.log.Warn("imagekit upload failed, using original url", "err", uerr)
	}

	// AI can't process video/file — store with placeholder, do NOT stream
	s.storeMediaMessage(ctx, tx, platform, conv, cred, &content, &storedURL, &mediaType, false)
	s.log.Info(label+" needs human handoff", "conversation_id", conv.ID)
}

// HandleSharedLink stores shared content the customer sent in a DM — a post,
// reel, story, or plain link. When we can resolve the object via the Graph API
// we feed the real caption/message to the AI; otherwise we fall back to the raw
// link so the AI at least knows something was shared.
func (s *chatService) HandleSharedLink(ctx context.Context, tx *sqlx.Tx, platform models.Platform, conv *models.Conversation, cred *models.AppCredential, pageToken string, attachment models.FacebookAttachment) {
	mediaType := models.MediaTypeLink
	link := attachment.Payload.URL
	label := facebookAttachmentLabel(attachment.Type)

	pageID := ""
	if cred.PlatformAccountID != nil {
		pageID = *cred.PlatformAccountID
	}

	details, err := s.fetchFacebookObjectDetails(ctx, link, pageID, pageToken)
	if err != nil {
		s.log.Warn("fetch shared content details failed, storing link only",
			"type", attachment.Type, "url", link, "err", err)
		content := fmt.Sprintf("[Customer shared a %s: %s]", label, link)
		s.storeMediaMessage(ctx, tx, platform, conv, cred, &content, &link, &mediaType, true)
		return
	}

	// The shared object carries an image — run it through the AI image route so
	// the model can read what's actually in the post, not just the link.
	if details.FullPicture != "" {
		s.HandleImageAttachment(ctx, tx, platform, conv, cred, pageID, details.FullPicture)
		return
	}

	content := fmt.Sprintf("[Customer shared a %s: %s]", label, link)
	if details.Message != "" {
		content = fmt.Sprintf("[Customer shared a %s: %s]\nContent: %s", label, link, details.Message)
	}
	s.storeMediaMessage(ctx, tx, platform, conv, cred, &content, &link, &mediaType, true)
}

// facebookAttachmentLabel maps a Graph attachment type to a human-readable noun
// used in the message we hand to the AI.
func facebookAttachmentLabel(t models.FacebookAttachmentType) string {
	switch t {
	case models.FacebookAttachmentTypeReel:
		return "reel"
	case models.FacebookAttachmentTypePost:
		return "post"
	case models.FacebookAttachmentTypeShare:
		return "post"
	case models.FacebookAttachmentTypeStoryMention:
		return "story"
	default:
		return "link"
	}
}

var fbObjectIDPattern = regexp.MustCompile(`(?:reel/|posts/|permalink/|videos/|story_fbid=|fbid=|/)(\d{6,})`)

var fbPfbidPattern = regexp.MustCompile(`(pfbid[0-9A-Za-z]+)`)

func (s *chatService) fetchFacebookObjectDetails(ctx context.Context, sharedURL, pageID, pageToken string) (*models.FacebookObjectDetails, error) {
	if pageToken == "" {
		return nil, fmt.Errorf("missing page token")
	}

	var objectID string
	if m := fbObjectIDPattern.FindStringSubmatch(sharedURL); len(m) >= 2 {
		objectID = m[1]
	} else if p := fbPfbidPattern.FindString(sharedURL); p != "" && pageID != "" {
		objectID = pageID + "_" + p
	}
	if objectID == "" {
		return nil, fmt.Errorf("no resolvable object id in url %q", sharedURL)
	}

	params := url.Values{
		"fields":       {"id,message,created_time,permalink_url,full_picture"},
		"access_token": {pageToken},
	}
	reqURL := fmt.Sprintf("https://graph.facebook.com/v25.0/%s?%s", objectID, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build object request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call object endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("object fetch failed (%d): %s", resp.StatusCode, body)
	}

	var details models.FacebookObjectDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("decode object details: %w", err)
	}
	return &details, nil
}

func (s *chatService) storeMediaMessage(ctx context.Context, tx *sqlx.Tx, platform models.Platform, conv *models.Conversation, cred *models.AppCredential, content *string, mediaURL *string, mediaType *models.MessageMediaType, stream bool) {
	insertedMsg, err := s.messageRepo.Create(ctx, tx, models.CreateMessage{
		ConversationID: conv.ID,
		BusinessID:     cred.BusinessID,
		Direction:      models.MessageDirectionIn,
		SentBy:         nil,
		Content:        content,
		MediaUrl:       mediaURL,
		MediaType:      mediaType,
		Status:         nil,
	})
	if err != nil {
		s.log.Error("create media message failed", "err", err)
		return
	}
	if stream {
		if err := s.StreamMessage(ctx, platform, insertedMsg.ID, cred.BusinessID, conv.ID, true); err != nil {
			s.log.Error("stream media message failed", "err", err)
		}
	}
}

func (s *chatService) checkContentLength(ctx context.Context, fileURL string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, fileURL, nil)
	if err != nil {
		return 0, fmt.Errorf("build head request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("head request: %w", err)
	}
	defer resp.Body.Close()
	return resp.ContentLength, nil
}

func safeFilename(pageID, rawURL string) string {
	suffix := rawURL
	if len(rawURL) > 16 {
		suffix = rawURL[len(rawURL)-16:]
	}
	return pageID + "_" + suffix
}

func (s *chatService) FetchUserProfile(ctx context.Context, psid, pageToken string) (*models.MessengerProfile, error) {
	params := url.Values{
		"fields":       {"first_name,last_name,profile_pic"},
		"access_token": {pageToken},
	}
	reqURL := fmt.Sprintf("https://graph.facebook.com/v25.0/%s?%s", psid, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build profile request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call profile endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("profile fetch failed (%d): %s", resp.StatusCode, body)
	}
	var profile models.MessengerProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("decode profile: %w", err)
	}
	return &profile, nil
}

func (s *chatService) UploadFileFromURL(ctx context.Context, fileURL, filename string) (string, error) {
	client := imagekit.NewClient(option.WithPrivateKey(s.cfg.ImageKit.PrivateKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return "", fmt.Errorf("build download request: %w", err)
	}
	httpResp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download file: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download file: unexpected status %d", httpResp.StatusCode)
	}

	limited := io.LimitReader(httpResp.Body, maxMediaBytes)
	result, err := client.Files.Upload(ctx, imagekit.FileUploadParams{File: limited, FileName: filename})
	if err != nil {
		return "", fmt.Errorf("imagekit upload: %w", err)
	}
	return result.URL, nil
}

func (s *chatService) GetImageDetails(ctx context.Context, fileURL string) (string, error) {
	return s.callMediaEndpoint(ctx, "/api/v1/ml/chat/images", fileURL, "description")
}

func (s *chatService) GetAudioDetails(ctx context.Context, fileURL string) (string, error) {
	return s.callMediaEndpoint(ctx, "/api/v1/ml/chat/audio", fileURL, "transcript")
}

func (s *chatService) callMediaEndpoint(ctx context.Context, path, fileURL, field string) (string, error) {
	body, err := json.Marshal(map[string]string{"url": fileURL})
	if err != nil {
		return "", fmt.Errorf("marshal media request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.App.AIServiceURL+path, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build media request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call media endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("media endpoint %s returned %d: %s", path, resp.StatusCode, raw)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode media response: %w", err)
	}
	return result[field], nil
}

func (s *chatService) HandleSharedContent(
	ctx context.Context, tx *sqlx.Tx, platform models.Platform,
	conv *models.Conversation, cred *models.AppCredential,
	contentURL, label, caption string,
) {
	mediaType := models.MediaTypeLink
	content := fmt.Sprintf("[Customer shared a %s: %s]", label, contentURL)
	if caption != "" {
		content = fmt.Sprintf("[Customer shared a %s: %s]\nCaption: %s", label, contentURL, caption)
	}
	s.storeMediaMessage(ctx, tx, platform, conv, cred, &content, &contentURL, &mediaType, true)
}
