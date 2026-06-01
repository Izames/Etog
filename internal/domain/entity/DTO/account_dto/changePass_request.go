package account_dto

type ChangePassRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
