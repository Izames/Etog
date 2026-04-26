package handlers

import (
	"Etog/internal/config"
	"Etog/internal/domain/entity"
	"Etog/internal/domain/entity/DTO"
	"Etog/internal/http-server/services"
	"log/slog"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	log      *slog.Logger
	service  *services.AuthService
	config   config.Config
	MailConf config.MailData
}

func NewAccountHandler(log *slog.Logger, service *services.AuthService) AccountHandler {
	log.Info("NewAccountHandler run successfully\n")
	return AccountHandler{
		log:     log,
		service: service,
	}
}

func (a *AccountHandler) Registration(ctx *gin.Context) {
	var account entity.Account

	if ctx.ContentType() != "application/json" {
		ctx.JSON(400, gin.H{
			"error": "Content-Type must be 'application/json'",
		})
		return
	}

	if err := ctx.ShouldBindJSON(&account); err != nil {
		ctx.JSON(400, gin.H{
			"error": "Invalid JSON: " + err.Error(),
		})
		return
	}

	ctx.JSON(201, gin.H{
		"message": "Account created",
	})
}

func (a *AccountHandler) GetCode(ctx *gin.Context) {
	var email string
	err := ctx.ShouldBindJSON(&email)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}
	err, code := a.service.SendCode(ctx, email)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
	}

	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) ConfirmCode(ctx *gin.Context) {
	var confirmCode DTO.ConfirmCodeRequest
	if err := ctx.ShouldBindJSON(&confirmCode); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON: " + err.Error()})
	}

	err, code, token, refreshToken := a.service.ConfirmCode(ctx, confirmCode)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(code, gin.H{
		"token":         token,
		"refresh_token": refreshToken,
	})

}

//func (a *AccountHandler) Authenticate(ctx *gin.Context) {
//	var DTOAccount DTO.AuthRequest
//	if err := ctx.ShouldBindJSON(&DTOAccount); err != nil {
//		ctx.JSON(400, gin.H{"error": "Invalid JSON: " + err.Error()})
//		return
//	}
//
//	account, err := a.service.GetAccountByLogin(DTOAccount.Login)
//	if err != nil {
//		if errors.Is(err, storage.ErrNotFound) {
//			ctx.JSON(400, gin.H{
//				"error": "Account not found",
//			})
//			return
//		}
//		ctx.JSON(500, gin.H{"error": "Internal server error"})
//	}
//
//	if err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(DTOAccount.Password)); err != nil {
//		ctx.JSON(400, gin.H{
//			"error": "Password is incorrect",
//		})
//		return
//	}
//	if !account.Active {
//		ctx.JSON(401, gin.H{
//			"error": "Email not checked",
//		})
//		return
//	}
//	token, refreshToken, err := a.service.TokensGenerate(account.Id)
//	if err != nil {
//		ctx.JSON(500, gin.H{"error": "Internal server error"})
//		return
//	}
//	ctx.JSON(200, gin.H{
//		"token":         token,
//		"refresh_token": refreshToken,
//	})
//}

//сделать изменения аккаунта с аватаркой
//сделать удаление аккаунта
//сделать выход из всех сессий аккаунта
//сделать запрос на галочку аккаунта
//сделать подписку на кого-то
//сделать отписку  на кого-то
//сделать обновление почты/пароля
