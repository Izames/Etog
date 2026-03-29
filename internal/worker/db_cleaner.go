package worker

import (
	"Etog/internal/domain/entity"
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

type Cleaner struct {
	db   *gorm.DB
	slog *slog.Logger
}

func NewCleaner(db *gorm.DB, slog *slog.Logger) *Cleaner {
	return &Cleaner{db: db, slog: slog}
}

func (cleaner *Cleaner) Run(ctx context.Context) {
	cleaner.slog.Info("Cleaner started successfully\n")
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				cleaner.slog.Info("Cleaner shutting down successfully\n")
				return
			case <-ticker.C:
				result := cleaner.db.Where("id IN (?)",
					cleaner.db.
						Model(&entity.RefreshToken{}).
						Select("id").
						Where("expire_at < ?", time.Now()).
						Limit(10000),
				).
					Delete(&entity.RefreshToken{})
				if result.Error != nil {
					cleaner.slog.Error("Failed to delete expired refresh token")
				} else {
					cleaner.slog.Info("Deleted expired refresh token successfully\n")
				}
			}
		}
	}()

}
