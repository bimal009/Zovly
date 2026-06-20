package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/pkg/utils"
	imagekit "github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type ChatService interface {
	FindOrCreate(ctx context.Context, tx *sqlx.Tx, conv models.CreateConversation) (*models.Conversation, error)
	HandleFacebookInboundMessage(ctx context.Context, platform models.Platform, pageID string, event models.FacebookMessagingEvent) error
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
		httpClient:        &http.Client{Timeout: 30 * time.Second},
		rdb:               rdb,
		log:               log,
	}
}

func (s *chatService) HandleFacebookInboundMessage(ctx context.Context, platform models.Platform, pageID string, event models.FacebookMessagingEvent) error {
	s.log.Info("inbound message received", "platform", platform, "page_id", pageID, "sender", event.Sender.ID)

	cred, err := s.appCredentialRepo.GetByPlatformAccountID(ctx, pageID)
	if err != nil {
		s.log.Error("credential lookup failed", "platform_id", pageID, "err", err)
		return fmt.Errorf("get credential for page %s: %w", pageID, err)
	}

	encKey, err := base64.StdEncoding.DecodeString(s.cfg.App.EncryptionKey)
	if err != nil {
		s.log.Error("decode encryption key failed", "err", err)
		return fmt.Errorf("decode encryption key: %w", err)
	}
	pageToken, err := utils.Decrypt(*cred.AccessToken, encKey)
	if err != nil {
		s.log.Error("token decrypt failed", "err", err)
		return fmt.Errorf("decrypt access token: %w", err)
	}

	user, err := s.fetchUserProfile(ctx, event.Sender.ID, pageToken)
	if err != nil {
		s.log.Error("fetch user profile failed", "sender_id", event.Sender.ID, "err", err)
		return fmt.Errorf("fetch user profile: %w", err)
	}
	s.log.Info("user profile fetched", "name", user.FirstName+" "+user.LastName)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	fullName := user.FirstName + " " + user.LastName
	conv, err := s.conversationRepo.FindOrCreate(ctx, tx, models.CreateConversation{
		BusinessID:       cred.BusinessID,
		Platform:         string(platform),
		ThreadID:         event.Sender.ID,
		ContactID:        event.Sender.ID,
		ContactName:      &fullName,
		ContactAvatarURL: &user.ProfilePic,
	})
	if err != nil {
		s.log.Error("find or create conversation failed", "err", err)
		return fmt.Errorf("find or create conversation: %w", err)
	}
	s.log.Info("conversation ready", "conversation_id", conv.ID, "business_id", cred.BusinessID)

	// 1) User text. Meta can send text and attachments in the SAME event,
	//    so this runs independently of the attachment loop below.
	if event.Message.Text != "" {
		text := event.Message.Text

		insertedMsg, err := s.messageRepo.Create(ctx, tx, models.CreateMessage{
			ConversationID: conv.ID,
			BusinessID:     cred.BusinessID,
			Direction:      models.MessageDirectionIn,
			SentBy:         nil,
			Content:        &text,
			Status:         nil,
		})
		if err != nil {
			s.log.Error("create message failed", "err", err)
			return fmt.Errorf("create message: %w", err)
		}

		if err := s.StreamMessage(ctx, platform, insertedMsg.ID, cred.BusinessID, conv.ID, false); err != nil {
			s.log.Error("stream message failed", "err", err)
			return fmt.Errorf("stream message: %w", err)
		}
	}

	// 2) Attachments. Message creation happens INSIDE each case so every media
	//    type stores what's relevant to it (e.g. image -> AI description).
	for _, attachment := range event.Message.Attachments {
		mediaURL := attachment.Payload.URL
		uploadedURL, err := s.UploadFileFromURL(ctx, mediaURL, pageID+"_"+attachment.Payload.URL[len(attachment.Payload.URL)-16:])
		if err != nil {
			s.log.Warn("imagekit upload failed, using original url", "err", err)
			uploadedURL = mediaURL
		}

		switch attachment.Type {

		case models.FacebookAttachmentTypeImage:
			mediaType := models.MediaTypeImage

			// Describe the image now and store it as Content, so the row has
			// real text for embedding + the AI reply. Failure is non-fatal:
			// we still persist the image, just without a description.
			var content *string
			if desc, derr := s.GetImageDetails(ctx, uploadedURL); derr != nil {
				s.log.Warn("get image details failed", "err", derr)
			} else if desc != "" {
				content = &desc
			}

			insertedMsg, err := s.messageRepo.Create(ctx, tx, models.CreateMessage{
				ConversationID: conv.ID,
				BusinessID:     cred.BusinessID,
				Direction:      models.MessageDirectionIn,
				SentBy:         nil,
				Content:        content,
				MediaUrl:       &uploadedURL,
				MediaType:      &mediaType,
				Status:         nil,
			})
			if err != nil {
				s.log.Error("create image message failed", "err", err)
				return fmt.Errorf("create image message: %w", err)
			}

			if err := s.StreamMessage(ctx, platform, insertedMsg.ID, cred.BusinessID, conv.ID, true); err != nil {
				s.log.Error("stream image message failed", "err", err)
				return fmt.Errorf("stream image message: %w", err)
			}

		case models.FacebookAttachmentTypeAudio:
			mediaType := models.MediaTypeAudio

			var transcribe *string

			if trans, terr := s.GetAudioDetails(ctx, uploadedURL); terr != nil {
				s.log.Warn("get audio details failed", "err", terr)
			} else if trans != "" {
				transcribe = &trans
			}

			insertedMsg, err := s.messageRepo.Create(ctx, tx, models.CreateMessage{
				ConversationID: conv.ID,
				BusinessID:     cred.BusinessID,
				Direction:      models.MessageDirectionIn,
				SentBy:         nil,
				Content:        transcribe,
				MediaUrl:       &uploadedURL,
				MediaType:      &mediaType,
				Status:         nil,
			})
			if err != nil {
				s.log.Error("create audio message failed", "err", err)
				return fmt.Errorf("create audio message: %w", err)
			}

			if err := s.StreamMessage(ctx, platform, insertedMsg.ID, cred.BusinessID, conv.ID, true); err != nil {
				s.log.Error("stream audio message failed", "err", err)
				return fmt.Errorf("stream audio message: %w", err)
			}

		case models.FacebookAttachmentTypeVideo, models.FacebookAttachmentTypeFile:
			// AI can't read video or arbitrary files — keep the message but
			// hand the conversation off to a human instead of replying.
			mediaType := models.MediaTypeVideo
			if attachment.Type == models.FacebookAttachmentTypeFile {
				mediaType = models.MediaTypeDocument
			}

			insertedMsg, err := s.messageRepo.Create(ctx, tx, models.CreateMessage{
				ConversationID: conv.ID,
				BusinessID:     cred.BusinessID,
				Direction:      models.MessageDirectionIn,
				SentBy:         nil,
				MediaUrl:       &uploadedURL,
				MediaType:      &mediaType,
				Status:         nil,
			})
			if err != nil {
				s.log.Error("create attachment message failed", "type", attachment.Type, "err", err)
				return fmt.Errorf("create attachment message: %w", err)
			}

			// TODO: wire your real handoff here, e.g.
			//   if err := s.conversationRepo.MarkNeedsHuman(ctx, tx, conv.ID); err != nil {
			//       s.log.Error("mark needs-human failed", "err", err)
			//       return fmt.Errorf("mark needs-human: %w", err)
			//   }
			s.log.Info("attachment needs human handoff",
				"type", attachment.Type,
				"conversation_id", conv.ID,
				"message_id", insertedMsg.ID,
			)

		default:
			s.log.Warn("unknown attachment type", "type", attachment.Type)
			continue
		}
	}

	return tx.Commit()
}

func (s *chatService) FindOrCreate(ctx context.Context, tx *sqlx.Tx, conv models.CreateConversation) (*models.Conversation, error) {
	return s.conversationRepo.FindOrCreate(ctx, tx, conv)
}

func (s *chatService) StreamMessage(
	ctx context.Context,
	platform models.Platform,
	messageId string,
	businessId string,
	conversationId string,
	attachments bool,

) error {

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

	s.log.Info(
		"message published to stream",
		"message_id", messageId,
		"conversation_id", conversationId,
	)

	return nil
}

func (s *chatService) fetchUserProfile(ctx context.Context, psid, pageToken string) (*models.MessengerProfile, error) {
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

func (s *chatService) UploadFileFromURL(ctx context.Context, fileURL string, filename string) (string, error) {
	client := imagekit.NewClient(
		option.WithPrivateKey(s.cfg.ImageKit.PrivateKey),
	)

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

	result, err := client.Files.Upload(ctx, imagekit.FileUploadParams{
		File:     httpResp.Body,
		FileName: filename,
	})
	if err != nil {
		return "", fmt.Errorf("imagekit upload: %w", err)
	}

	return result.URL, nil
}

func (s *chatService) GetImageDetails(ctx context.Context, fileURL string) (string, error) {
	payload := map[string]string{"url": fileURL}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal image details request: %w", err)
	}

	endpoint := s.cfg.App.AIServiceURL + "/api/v1/ml/chat/images"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("build image details request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call image details endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("image details endpoint returned %d: %s", resp.StatusCode, raw)
	}

	var result struct {
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode image details response: %w", err)
	}

	return result.Description, nil
}
func (s *chatService) GetAudioDetails(ctx context.Context, fileURL string) (string, error) {
	payload := map[string]string{"url": fileURL}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal audio request: %w", err)
	}

	endpoint := s.cfg.App.AIServiceURL + "/api/v1/ml/chat/audio"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("build audio request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call audio endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("audio endpoint returned %d: %s", resp.StatusCode, raw)
	}

	var result struct {
		Transcript string `json:"transcript"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode audio response: %w", err)
	}
	s.log.Info("audio transcript", "transcript", result.Transcript)
	return result.Transcript, nil
}
