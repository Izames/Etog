package entity

type Contact struct {
	Id       int    `gorm:"primary_key;column:id"`
	Phone    string `gorm:"column:phone"`
	Mail     string `gorm:"column:mail"`
	Telegram string `gorm:"column:telegram"`
	Deleted  bool   `gorm:"column:deleted"`
}
