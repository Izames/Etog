package worker

import (
	"Etog/internal/domain/entity"
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCleaner_Run(t *testing.T) {
	t.Logf("test started.")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	err = db.AutoMigrate(&entity.RefreshToken{})
	if err != nil {
		t.Fatal(err)
	}

	expiredToken := entity.RefreshToken{
		UserId:   1,
		TokenId:  []byte("expired"),
		ExpireAt: time.Now().Add(-time.Hour), // уже просрочен
	}

	validToken := entity.RefreshToken{
		UserId:   2,
		TokenId:  []byte("valid"),
		ExpireAt: time.Now().Add(time.Hour), // еще валиден
	}

	db.Create(&expiredToken)
	db.Create(&validToken)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cleaner := NewCleaner(db, logger, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cleaner.Run(ctx)

	time.Sleep(30 * time.Second)

	var tokens []entity.RefreshToken
	db.Find(&tokens)

	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}

	if string(tokens[0].TokenId) != "valid" {
		t.Fatalf("wrong token left, got %s", tokens[0].TokenId)
	}
}
