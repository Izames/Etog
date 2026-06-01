package entity

type AccountDb struct {
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
	Followers   int    `gorm:"column:followers"`
}

func (AccountDb) TableName() string {
	return "account_dto"
}

type Account struct {
	Login       string `gorm:"column:login"`
	Avatar      string `gorm:"column:avatar"`
	Official    bool   `gorm:"column:official"`
	Description string `gorm:"column:description"`
	Rating      int    `gorm:"column:rating"`
	Followers   int    `gorm:"column:followers"`
}
