package account_dto

type RegReq struct {
	Mail        string `gorm:"column:mail"`
	Login       string `gorm:"column:login"`
	Password    string `gorm:"column:password"`
	Description string `gorm:"column:description"`
}
