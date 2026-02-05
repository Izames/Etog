package entity

type RefreshToken struct {
	TokenId []byte `gorm:"primary_key; column:token_id;"`
	UserId  int    `gorm:"column:user_id"`
}
