package entity

type Account struct {
	Id          int    `gorm:"primary_key;column:id"`
	Mail        string `gorm:"column:mail"`
	Login       string `gorm:"column:login"`
	Password    string `gorm:"column:password"`
	Avatar      string `gorm:"column:avatar"`
	Official    bool   `gorm:"column:official"`
	Description string `gorm:"column:description"`
	Rating      int    `gorm:"column:rating"`
	Deleted     bool   `gorm:"column:deleted"`
	Active      bool   `gorm:"column:active"`
}

func (Account) TableName() string {
	return "account"
}
