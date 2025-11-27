package entity

import (
	"time"

	"gorm.io/datatypes"
)

type MockEvent struct {
	Id          int            `gorm:"primaryKey;column:id"`
	Organizer   string         `gorm:"column:organizer"`
	Name        string         `gorm:"column:name"`
	Price       float64        `gorm:"column:price"`
	Address     string         `gorm:"column:address"`
	Type        string         `gorm:"column:type"`
	Date        time.Time      `gorm:"column:date"`
	Description string         `gorm:"column:description"`
	Passed      string         `gorm:"column:passed"`
	MaxPeople   int            `gorm:"column:max_people"` // Исправлено: с большой буквы
	Media       datatypes.JSON `gorm:"column:media;type:jsonb"`
	Phone       string         `gorm:"column:phone"`
	Mail        string         `gorm:"column:mail"`
	Telegram    string         `gorm:"column:telegram"`
	Deleted     bool           `gorm:"column:deleted"`
}

// Указываем точное имя таблицы
func (MockEvent) TableName() string {
	return "mock_event"
}
