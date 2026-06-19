package workers

// import (
// 	"context"
// 	"log/slog"
// 	"time"

// 	repository "github.com/bimal009/Zovly/internal/repo"
// 	"github.com/bimal009/Zovly/internal/service"
// 	"github.com/redis/go-redis/v9"
// )

// type ChatWorker struct {
// 	chatService service.ChatService
// 	messageRepo repository.MessageRepo
// 	log         *slog.Logger
// 	rdb         *redis.Client
// }

// func NewChatWorker(chatService service.ChatService, log *slog.Logger, rdb *redis.Client) *ChatWorker {
// 	return &ChatWorker{
// 		chatService: chatService,
// 		log:         log,
// 		rdb:         rdb,
// 	}
// }

// func (c *ChatWorker) Run(ctx context.Context) {
// 	ticker := time.NewTicker(30 * time.Second)
// 	defer ticker.Stop()

// 	c.refresh(ctx)

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			c.log.Info("instagram token refresh worker stopped")
// 			return
// 		case <-ticker.C:
// 			c.refresh(ctx)
// 		}
// 	}
// }

// func (w *ChatWorker) retryPending(ctx context.Context) {
// 	pending, err := w.messageRepo.GetPendingOutbound(ctx)
// 	if err != nil || len(pending) == 0 {
// 		return
// 	}

// 	for _, msg := range pending {
// 		w.aiWorker.trySend(ctx, &msg, msg.BusinessID)
// 	}
// }
