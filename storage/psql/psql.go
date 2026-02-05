package psql

import (
	"Etog/internal/domain/entity"
	"errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

// сделать систему ошибок для определения проблем с базой данных, или просто не удалось найти

func New(storagePath string) *Storage {
	db, err := gorm.Open(postgres.Open(storagePath), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return &Storage{db: db}
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
