package account_dto

type ConfirmEmailRequest struct {
	NewMail  string `json:"new_mail"`
	Code     string `json:"code"`
	Password string `json:"password"`
}
