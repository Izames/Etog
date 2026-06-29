package entity

import "time"

type RefreshToken struct {
	TokenId  string    `gorm:"primary_key; column:token_id;"`
	UserId   int       `gorm:"column:user_id"`
	ExpireAt time.Time `gorm:"column:expire_at"`
}

func (RefreshToken) TableName() string {
	return "refresh"
}
