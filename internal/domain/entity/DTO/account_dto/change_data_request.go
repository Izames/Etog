package account_dto

import "mime/multipart"

type ChangeDataRequest struct {
	Avatar      string                `json:"avatar"`
	Description string                `json:"description"`
	File        *multipart.FileHeader `json:"file"`
}
