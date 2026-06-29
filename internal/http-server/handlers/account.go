package handlers

import (
	"Etog/internal/config"
	"Etog/internal/domain/entity/DTO/account_dto"
	"Etog/internal/http-server/services"
	"log/slog"
	"strconv"

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
	var account account_dto.RegReq

	if err := ctx.ShouldBindJSON(&account); err != nil {
		ctx.JSON(400, gin.H{
			"error": "Validation error",
		})
		return
	}

	code, err := a.service.Registration(ctx, &account)
	if err != nil {
		ctx.JSON(code, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.JSON(201, gin.H{
		"message": "Account created, confirmation code sent to email",
	})
}

func (a *AccountHandler) SendCode(ctx *gin.Context) {
	type mail struct {
		Mail string `json:"mail"`
	}
	email := mail{}
	err := ctx.ShouldBindJSON(&email)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	code, err := a.service.SendCode(ctx, email.Mail, "confirmation code", "ваш код для подтверждения почты")
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) ConfirmCode(ctx *gin.Context) {
	var confirmCode account_dto.ConfirmCodeRequest
	if err := ctx.ShouldBindJSON(&confirmCode); err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}

	code, token, refreshToken, err := a.service.ConfirmCode(ctx, confirmCode)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(code, gin.H{
		"token":         token,
		"refresh_token": refreshToken,
	})

}

func (a *AccountHandler) Authenticate(ctx *gin.Context) {
	var DTOAccount account_dto.AuthRequest
	if err := ctx.ShouldBindJSON(&DTOAccount); err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	token, refreshToken, code, err := a.service.Authenticate(ctx, DTOAccount)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{
		"token":         token,
		"refresh_token": refreshToken,
	})
}

func (a *AccountHandler) ChangeData(ctx *gin.Context) {
	var DTOAccount account_dto.ChangeDataRequest
	id := ctx.GetInt("userId")
	if err := ctx.ShouldBind(&DTOAccount); err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	fileHeader, err := ctx.FormFile("avatar")
	if err == nil {
		DTOAccount.Avatar = fileHeader.Filename
		DTOAccount.File = fileHeader
	}
	code, err := a.service.ChangeData(ctx, id, DTOAccount)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) GetAccountByLoginOrId(ctx *gin.Context) {
	query := ctx.Param("query")
	id, err := strconv.Atoi(query)
	if err != nil {
		code, account, err := a.service.GetAccount(ctx, query, -1)
		if err != nil {
			ctx.JSON(code, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(code, *account)
	}
	code, account, err := a.service.GetAccount(ctx, query, id)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(code, *account)
}

func (a *AccountHandler) DeleteAccount(ctx *gin.Context) {
	id := ctx.GetInt("userId")
	code, err := a.service.DeleteAccount(ctx, id)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{})
}

//func (a *AccountHandler) RequestOfficial(ctx *gin.Context) {
//	id, err := strconv.Atoi(ctx.Param("id"))
//	if err != nil {
//		ctx.JSON(400, gin.H{"error": "Invalid ID: " + err.Error()})
//		return
//	}
//	if err = a.service.RequestOfficial(ctx, id); err != nil {
//		ctx.JSON(500, gin.H{"error": "internal server error"})
//		return
//	}
//	ctx.JSON(200, gin.H{})
//}

func (a *AccountHandler) Subscribe(ctx *gin.Context) {
	follower := ctx.GetInt("userId")
	followingStr := ctx.Param("userId")
	followingId, err := strconv.Atoi(followingStr)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	code, err := a.service.Follow(ctx, follower, followingId)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) Unsubscribe(ctx *gin.Context) {
	follower := ctx.GetInt("userId")
	followingStr := ctx.Param("userId")
	followingId, err := strconv.Atoi(followingStr)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	code, err := a.service.Unfollow(ctx, follower, followingId)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) ChangePassword(ctx *gin.Context) {
	type mail struct {
		Mail string `json:"mail"`
	}
	var email mail
	if err := ctx.ShouldBindJSON(&email); err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	code, err := a.service.SendCode(ctx, email.Mail, "change password code", "код для смены пароля")
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) ConfirmChangePassword(ctx *gin.Context) {
	var confirmPassReq account_dto.ChangePassRequest
	if err := ctx.ShouldBind(&confirmPassReq); err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	code, err := a.service.ConfirmChangePassword(ctx, confirmPassReq.Mail, confirmPassReq.Code, confirmPassReq.NewPassword)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) ChangeMail(ctx *gin.Context) {
	type mail struct {
		Mail string `json:"mail"`
	}
	var email mail
	if err := ctx.ShouldBindJSON(&email); err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	code, err := a.service.SendCode(ctx, email.Mail, "change email", "ваш код для смены почты")
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) ConfirmMail(ctx *gin.Context) {
	id := ctx.GetInt("userId")
	var confirmPassReq account_dto.ConfirmEmailRequest
	if err := ctx.ShouldBind(&confirmPassReq); err != nil {
		ctx.JSON(400, gin.H{"error": "validation error"})
		return
	}
	code, err := a.service.ConfirmChangeMail(ctx, id, confirmPassReq.NewMail, confirmPassReq.Code, confirmPassReq.Password)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) RequestOfficial(ctx *gin.Context) {
	id := ctx.GetInt("userId")
	code, err := a.service.RequestOfficial(ctx, id)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) DeleteSessions(ctx *gin.Context) {
	id := ctx.GetInt("userId")
	code, err := a.service.DeleteSessions(ctx, id)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{})
}
