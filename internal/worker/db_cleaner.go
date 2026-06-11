package worker

import (
	"Etog/internal/domain/entity"
	"context"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

type Cleaner struct {
	db       *gorm.DB
	slog     *slog.Logger
	interval time.Duration
}

func NewCleaner(db *gorm.DB, slog *slog.Logger, interval time.Duration) *Cleaner {
	return &Cleaner{db: db, slog: slog, interval: interval}
}

func (cleaner *Cleaner) Run(ctx context.Context) {
	cleaner.slog.Info("Cleaner started successfully\n")
	ticker := time.NewTicker(cleaner.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				cleaner.slog.Info("Cleaner shutting down successfully\n")
				return
			case <-ticker.C:
				result := cleaner.db.Model(&entity.RefreshToken{}).
					Select("token_id").
					Where("expire_at < ?", time.Now()).
					Limit(10000).
					Delete(&entity.RefreshToken{})
				if result.Error != nil {
					if errors.Is(result.Error, gorm.ErrRecordNotFound) {
						cleaner.slog.Info("Cleaner did not find any refresh tokens\n")
					} else {
						cleaner.slog.Error("Failed to delete expired refresh token", "error", result.Error)
					}
				} else {
					cleaner.slog.Info("Deleted expired refresh token successfully\n")
				}
			}
		}
	}()

}
