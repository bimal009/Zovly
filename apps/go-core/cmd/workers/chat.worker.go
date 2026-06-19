package workers

// import (
// 	"context"
// 	"fmt"
// 	"log/slog"
// 	"strings"
// 	"time"

// 	"github.com/bimal009/Zovly/internal/models"
// 	repository "github.com/bimal009/Zovly/internal/repo"
// 	"github.com/bimal009/Zovly/internal/service"
// 	"github.com/redis/go-redis/v9"
// )

// type AIWorker struct {
// 	chatService service.ChatService
// 	messageRepo repository.MessageRepo
// 	log         *slog.Logger
// 	rdb         *redis.Client
// }

// func NewAIWorker(chatService service.ChatService, log *slog.Logger, rdb *redis.Client) *AIWorker {
// 	return &AIWorker{
// 		chatService: chatService,
// 		log:         log,
// 		rdb:         rdb,
// 	}
// }

// func (w *AIWorker) Run(ctx context.Context) {
// 	w.rdb.XGroupCreateMkStream(ctx, "chat:messages", "ai-workers", "0")

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			w.log.Info("ai worker stopped")
// 			return
// 		default:
// 		}

// 		results, err := w.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
// 			Group:    "ai-workers",
// 			Consumer: "worker-1",
// 			Streams:  []string{"chat:messages", ">"},
// 			Count:    10,
// 			Block:    5 * time.Second,
// 		}).Result()
// 		if err != nil || len(results) == 0 {
// 			continue
// 		}

// 		for _, msg := range results[0].Messages {
// 			if err := w.processAndReply(ctx, msg); err != nil {
// 				w.log.Error("process failed", "id", msg.ID, "err", err)
// 				continue
// 			}
// 			w.rdb.XAck(ctx, "chat:messages", "ai-workers", msg.ID)
// 		}
// 	}
// }

// func (w *AIWorker) processAndReply(ctx context.Context, msg redis.XMessage) error {
// 	conversationID := msg.Values["conversation_id"].(string)
// 	businessID := msg.Values["business_id"].(string)

// 	lockKey := fmt.Sprintf("processing:%s", conversationID)
// 	locked, _ := w.rdb.SetNX(ctx, lockKey, "1", 30*time.Second).Result()
// 	if !locked {
// 		return nil
// 	}
// 	defer w.rdb.Del(ctx, lockKey)

// 	time.Sleep(12 * time.Second)

// 	unreplied, err := w.messageRepo.GetUnrepliedInbound(ctx, conversationID)
// 	if err != nil || len(unreplied) == 0 {
// 		return err
// 	}

// 	var parts []string
// 	for _, m := range unreplied {
// 		if m.Content != nil {
// 			parts = append(parts, *m.Content)
// 		}
// 	}
// 	combined := strings.Join(parts, "\n")

// 	// generate reply — always do this regardless of rate limit
// 	reply, err := w.pyml.Chat(ctx, businessID, combined)
// 	if err != nil {
// 		return err
// 	}

// 	// write outbound message as pending
// 	pending := models.MessageStatusPending
// 	outbound, err := w.messageRepo.Create(ctx, nil, models.CreateMessage{
// 		ConversationID: conversationID,
// 		BusinessID:     businessID,
// 		Direction:      models.MessageDirectionOut,
// 		SentBy:         ptr(models.MessageSenderAI),
// 		Content:        &reply,
// 		Status:         &pending,
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	// NOW check rate limit before sending
// 	w.trySend(ctx, outbound, businessID)
// 	return nil
// }

// func (w *AIWorker) trySend(ctx context.Context, msg *models.Message, businessID string) {
// 	rateLimitKey := fmt.Sprintf("rate:dm:%s", businessID)

// 	count, _ := w.rdb.Incr(ctx, rateLimitKey).Result()
// 	if count == 1 {
// 		w.rdb.Expire(ctx, rateLimitKey, time.Hour)
// 	}

// 	if count > 200 {
// 		// over limit — decrement back, message stays pending
// 		w.rdb.Decr(ctx, rateLimitKey)
// 		w.log.Warn("rate limited, reply queued", "message_id", msg.ID)
// 		return
// 	}

// 	// under limit — send via Graph API
// 	platformMsgID, err := w.sendViaGraphAPI(ctx, msg)
// 	if err != nil {
// 		w.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusFailed, "", ptr(err.Error()))
// 		w.rdb.Decr(ctx, rateLimitKey)
// 		return
// 	}

// 	w.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusSent, platformMsgID, nil)
// }
