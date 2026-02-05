package entity

type JoinedOnEvent struct {
	EventId int `gorm:"column:event_id"`
	UserId  int `gorm:"column:user_id"`
}
