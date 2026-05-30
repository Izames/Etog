package entity

import "time"

type OfficialRequest struct {
	UserID    int       `gorm:"primary_key"`
	CreatedAt time.Time `gorm:"column:created_at"`
	Comment   *string   `gorm:"column:comment"`
}
