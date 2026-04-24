package psql

import (
	"Etog/internal/domain/entity"
	"Etog/storage"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

// сделать систему ошибок для определения проблем с базой данных, или просто не удалось найти

func New(storagePath string, log *slog.Logger) *Storage {
	if log == nil {
		log = slog.Default()
	}
	db, err := gorm.Open(postgres.Open(storagePath), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	log.Info("PostgreSQL run successfully\n")
	return &Storage{db: db}
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

func (s *Storage) CreateAccount(account entity.Account) error {
	result := s.db.Create(&account)
	if result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) {
			if pgErr.Code == "23505" {
				return storage.ErrAlreadyExists
			}
		}
		return result.Error
	}
	return nil
}

func (s *Storage) GetAccount(id int) (*entity.Account, error) {
	var account entity.Account
	result := s.db.Where("id = ?", id).First(&account)
	if result.Error != nil {
		if errors.Is(gorm.ErrRecordNotFound, result.Error) {
			return nil, storage.ErrNotFound
		}
		return nil, result.Error
	}
	return &account, nil
}

func (s *Storage) GetAccountByLogin(login string) (*entity.Account, error) {
	var account entity.Account
	result := s.db.Where("login = ?", login).First(&account)
	if result.Error != nil {
		if errors.Is(gorm.ErrRecordNotFound, result.Error) {
			return nil, storage.ErrNotFound
		}
		return nil, result.Error
	}
	return &account, nil
}

func (s *Storage) GetAccountByEmail(email string) (*entity.Account, error) {
	var account entity.Account
	result := s.db.Where("email = ?", email).First(&account)
	if result.Error != nil {
		if errors.Is(gorm.ErrRecordNotFound, result.Error) {
			return nil, storage.ErrNotFound
		}
		return nil, result.Error
	}
	return &account, nil
}

func (s *Storage) PutAccount(account entity.Account) error {
	result := s.db.Save(&account)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *Storage) CreateRefreshToken(token entity.RefreshToken) error {
	result := s.db.Create(&token)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *Storage) GetRefreshToken(tokenId string) (*entity.RefreshToken, error) {
	var refreshToken entity.RefreshToken
	result := s.db.Where("id = ?", tokenId).First(&refreshToken)
	if result.Error != nil {
		return nil, result.Error
	}
	return &refreshToken, nil
}

func (s *Storage) DeleteRefreshToken(userId int) error {
	result := s.db.Where("user_id = ?", userId).Delete(&entity.RefreshToken{})
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	return nil
}
