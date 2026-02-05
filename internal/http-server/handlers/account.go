package handlers

import (
	"Etog/internal/domain/entity"
	"Etog/internal/http-server/services"
	"log/slog"
)

type AccountHandler struct {
	log     *slog.Logger
	service services.AuthService
}

func (a *AccountHandler) Authentification(account *entity.Account) error {

}
