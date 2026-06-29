package account_dto

type ChangePassRequest struct {
	NewPassword string `json:"new_password"`
	Code        string `json:"code"`
	Mail        string `json:"mail"`
}
