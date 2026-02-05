package entity

type subscribers struct {
	AccountId    int `gorm:"column:account_id"`
	SubscriberId int `gorm:"column:subscriber_id"`
}
