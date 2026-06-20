package handlers

import (
	"Etog/internal/config"
	"Etog/internal/domain/entity/DTO/account_dto"
	"Etog/internal/http-server/services"
	"log/slog"
	"net/http"
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

	err, code := a.service.Registration(&account)
	if err != nil {
		ctx.JSON(code, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.JSON(201, gin.H{
		"message": "AccountDb created",
	})
}

func (a *AccountHandler) GetCode(ctx *gin.Context) {
	type mail struct {
		Mail string `json:"mail"`
	}
	email := mail{}
	err := ctx.ShouldBindJSON(&email)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}
	err, code := a.service.SendCode(ctx, email.Mail)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) ConfirmCode(ctx *gin.Context) {
	var confirmCode account_dto.ConfirmCodeRequest
	if err := ctx.ShouldBindJSON(&confirmCode); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
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

func (a *AccountHandler) Authenticate(ctx *gin.Context) {
	var DTOAccount account_dto.AuthRequest
	if err := ctx.ShouldBindJSON(&DTOAccount); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}
	token, refreshToken, err, code := a.service.Authenticate(DTOAccount)
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
	if err := ctx.ShouldBindJSON(&DTOAccount); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}
	fileHeader, err := ctx.FormFile("avatar")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	DTOAccount.Avatar = fileHeader.Filename
	err, code := a.service.ChangeData(DTOAccount, ctx)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) GetAccountByLogin(ctx *gin.Context) {
	type GetRequest struct {
		Login string `json:"login"`
		Id    int    `json:"id"`
	}
	var getRequest GetRequest
	if err := ctx.ShouldBindJSON(&getRequest); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	code, account, err := a.service.GetAccount(ctx, getRequest.Login, getRequest.Id)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err})
		return
	}
	ctx.JSON(code, *account)
}

func (a *AccountHandler) DeleteAccount(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid ID: " + err.Error()})
		return
	}
	code, err := a.service.DeleteAccount(ctx, id)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err})
		return
	}
	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) RequestOfficial(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid ID: " + err.Error()})
		return
	}
	if err = a.service.RequestOfficial(ctx, id); err != nil {
		ctx.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) Subscribe(ctx *gin.Context) {
	follower, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid token: " + err.Error()})
		return
	}
	type SubscribeRequest struct {
		Following int `json:"following"`
	}
	var subscribeRequest SubscribeRequest
	if err = ctx.ShouldBindJSON(&subscribeRequest); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}
	if err = a.service.Subscribe(ctx, subscribeRequest.Following, follower); err != nil {
		ctx.JSON(400, gin.H{"error": "internal server error"})
		return
	}
	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) Unsubscribe(ctx *gin.Context) {
	follower, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid ID: " + err.Error()})
		return
	}
	type UnsubscribeRequest struct {
		Following int `json:"following"`
	}
	var unsubscribeRequest UnsubscribeRequest
	if err = ctx.ShouldBindJSON(&unsubscribeRequest); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}
	if err = a.service.Unsubscribe(ctx, follower, unsubscribeRequest.Following); err != nil {
		ctx.JSON(400, gin.H{"error": "internal server error"})
		return
	}
	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) ChangePassword(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid ID: " + err.Error()})
		return
	}
	var reqDto account_dto.ChangePassRequest
	if err = ctx.ShouldBindJSON(&reqDto); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}
	if err = a.service.ChangePasword(ctx, id, reqDto.OldPassword, reqDto.NewPassword); err != nil {
		ctx.JSON(400, gin.H{"error": "internal server error"})
		return
	}
	ctx.JSON(200, gin.H{})
}

//func (a *AccountHandler) ChangeMail(ctx *gin.Context) {
//	type GetRequest struct {
//		NewMail string `json:"mail"`
//		Code    string `json:"code"`
//	}
//	id, err := strconv.Atoi(ctx.Param("id"))
//	if err != nil {
//		ctx.JSON(400, gin.H{"error": "Invalid ID: " + err.Error()})
//		return
//
//	}
//	var getRequest GetRequest
//	if err = ctx.ShouldBindJSON(&getRequest); err != nil {
//		ctx.JSON(400, gin.H{"error": "Invalid JSON: " + err.Error()})
//		return
//	}
//
//}

//восстановить пароль
