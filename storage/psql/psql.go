package psql

import (
	"Etog/internal/domain/entity"
	"Etog/storage"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage struct {
	db  *gorm.DB
	log *slog.Logger
}

func New(storagePath string, log *slog.Logger) *Storage {
	if log == nil {
		log = slog.Default()
	}
	db, err := gorm.Open(postgres.Open(storagePath), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	log.Info("PostgreSQL run successfully\n")
	return &Storage{db: db, log: log}
}

func (s *Storage) ReturnDb() *gorm.DB {
	return s.db
}

func (s *Storage) CreateMockEvent(mockEvent entity.MockEvent) error {
	result := s.db.Create(&mockEvent)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *Storage) GetMockEvent(id int) (*entity.MockEvent, error) {
	var mockEvent entity.MockEvent
	result := s.db.Where("id = ?", id).First(&mockEvent)
	if result.Error != nil {
		return nil, result.Error
	}
	return &mockEvent, nil
}

func (s *Storage) GetMockEvents() (*[]entity.MockEvent, error) {
	var mockEvents []entity.MockEvent
	result := s.db.Where("Deleted = false").Find(&mockEvents)
	if result.Error != nil {
		return nil, result.Error
	}
	return &mockEvents, nil
}

func (s *Storage) UpdateMockEvent(id int, newMockEvent map[string]interface{}) (*entity.MockEvent, error) {
	var mockEvent entity.MockEvent
	result := s.db.Model(&mockEvent).Where("id = ?", id).Updates(newMockEvent).Scan(&mockEvent)
	if result.Error != nil {
		return nil, result.Error
	}
	return &mockEvent, nil
}

func (s *Storage) DeleteMockEvent(id int) error {
	result := s.db.Delete(&entity.MockEvent{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *Storage) CreateAccount(ctx context.Context, account *entity.AccountDb) (int, error) {
	const op = "psql.CreateAccount"
	log := s.log.With(slog.String("op", op))

	result := s.db.WithContext(ctx).Create(account)
	if result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) && pgErr.Code == "23505" {
			return 409, errors.New("login or mail already exists")
		}
		log.Error(op+": ", result.Error)
		return 500, errors.New("server error")
	}
	return 201, nil
}

func (s *Storage) GetAccount(ctx context.Context, id int) (*entity.AccountDb, int, error) {
	const op = "psql.GetAccount"
	log := s.log.With(slog.String("op", op))

	var account entity.AccountDb
	result := s.db.WithContext(ctx).Where("id = ?", id).First(&account)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, 404, storage.ErrNotFound
		}
		log.Error(op+": ", result.Error)
		return nil, 500, errors.New("server error")
	}
	return &account, 200, nil
}

func (s *Storage) GetAccountByLogin(ctx context.Context, login string) (*entity.AccountDb, int, error) {
	const op = "psql.GetAccountByLogin"
	log := s.log.With(slog.String("op", op))

	var account entity.AccountDb
	result := s.db.WithContext(ctx).Where("login = ?", login).First(&account)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, 404, storage.ErrNotFound
		}
		log.Error(op+": ", result.Error)
		return nil, 500, errors.New("server error")
	}
	return &account, 200, nil
}

func (s *Storage) GetAccountByEmail(ctx context.Context, email string) (*entity.AccountDb, int, error) {
	const op = "psql.GetAccountByEmail"
	log := s.log.With(slog.String("op", op))

	var account entity.AccountDb
	result := s.db.WithContext(ctx).Where("mail = ?", email).First(&account)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, 404, storage.ErrNotFound
		}
		log.Error(op+": ", result.Error)
		return nil, 500, errors.New("server error")
	}
	return &account, 200, nil
}

func (s *Storage) PutAccount(ctx context.Context, account entity.AccountDb) (int, error) {
	const op = "psql.PutAccount"
	log := s.log.With(slog.String("op", op))

	if err := s.db.WithContext(ctx).Save(&account).Error; err != nil {
		log.Error(op+": ", err)
		return 500, errors.New("server error")
	}
	return 200, nil
}

func (s *Storage) CreateRefreshToken(ctx context.Context, token entity.RefreshToken) error {
	const op = "psql.CreateRefreshToken"
	log := s.log.With(slog.String("op", op))

	if err := s.db.WithContext(ctx).Create(&token).Error; err != nil {
		log.Error(op+": ", err)
		return errors.New("server error")
	}
	return nil
}

func (s *Storage) GetRefreshToken(ctx context.Context, tokenId string) (*entity.RefreshToken, error) {
	const op = "psql.GetRefreshToken"
	log := s.log.With(slog.String("op", op))

	var refreshToken entity.RefreshToken
	result := s.db.WithContext(ctx).Where("token_id = ?", tokenId).First(&refreshToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, storage.ErrNotFound
		}
		log.Error(op+": ", result.Error)
		return nil, errors.New("server error")
	}
	return &refreshToken, nil
}

func (s *Storage) DeleteRefreshToken(ctx context.Context, userId int) (int, error) {
	const op = "psql.DeleteRefreshToken"
	log := s.log.With(slog.String("op", op))

	result := s.db.WithContext(ctx).Where("user_id = ?", userId).Delete(&entity.RefreshToken{})
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Error(op+": ", result.Error)
		return 500, errors.New("server error")
	}
	return 200, nil
}

func (s *Storage) DeleteRefreshTokenByID(ctx context.Context, tokenId string) error {
	const op = "psql.DeleteRefreshTokenByID"
	log := s.log.With(slog.String("op", op))

	result := s.db.WithContext(ctx).Where("token_id = ?", tokenId).Delete(&entity.RefreshToken{})
	if result.Error != nil {
		log.Error(op, "error", result.Error)
		return errors.New("server error")
	}
	return nil
}

func (s *Storage) CreateOfficialRequest(ctx context.Context, id int) (int, error) {
	const op = "psql.CreateOfficialRequest"
	log := s.log.With(slog.String("op", op))

	if err := s.db.WithContext(ctx).Save(&entity.OfficialRequest{
		UserID:    id,
		CreatedAt: time.Now(),
		Comment:   nil,
	}).Error; err != nil {
		log.Error(op+": ", err)
		return 500, errors.New("server error")
	}
	return 201, nil
}

func (s *Storage) Subscribe(ctx context.Context, id, follower int) (int, error) {
	const op = "psql.Subscribe"
	log := s.log.With(slog.String("op", op))

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&entity.Subscribers{
			AccountId:    id,
			SubscriberId: follower,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.AccountDb{}).
			Where("id = ?", id).
			UpdateColumn("followers", gorm.Expr("followers + ?", 1)).Error
	})
	if err != nil {
		log.Error(op+": ", err)
		return 500, errors.New("server error")
	}
	return 200, nil
}

func (s *Storage) Unsubscribe(ctx context.Context, id, follower int) (int, error) {
	const op = "psql.Unsubscribe"
	log := s.log.With(slog.String("op", op))

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("account_id = ? AND subscriber_id = ?", id, follower).Delete(&entity.Subscribers{})
		if result.Error != nil {
			log.Error(op+": ", result.Error)
			return errors.New("server error")
		}
		if result.RowsAffected == 0 {
			return errors.New("not following")
		}
		return tx.Model(&entity.AccountDb{}).
			Where("id = ?", id).
			UpdateColumn("followers", gorm.Expr("followers - ?", 1)).Error
	})
	if err != nil {
		log.Error(op+": ", err)
		return 500, errors.New("server error")
	}
	return 200, nil
}
