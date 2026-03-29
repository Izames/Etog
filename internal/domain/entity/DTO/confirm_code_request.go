package DTO

type ConfirmCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}
