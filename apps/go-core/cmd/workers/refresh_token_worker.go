package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/bimal009/Zovly/internal/service"
)

type RefreshInstagramWorker struct {
	instagramService service.InstagramService
	log              *slog.Logger
}

func NewRefreshInstagramWorker(instagramService service.InstagramService, log *slog.Logger) *RefreshInstagramWorker {
	return &RefreshInstagramWorker{
		instagramService: instagramService,
		log:              log,
	}
}

func (r *RefreshInstagramWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	r.refresh(ctx)

	for {
		select {
		case <-ctx.Done():
			r.log.Info("instagram token refresh worker stopped")
			return
		case <-ticker.C:
			r.refresh(ctx)
		}
	}
}

func (r *RefreshInstagramWorker) refresh(ctx context.Context) {
	r.log.Info("checking for expiring instagram tokens")

	if err := r.instagramService.RefreshExpiringTokens(ctx); err != nil {
		r.log.Error("instagram token refresh failed", "err", err)
	}
}
