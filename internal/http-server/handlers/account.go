package handlers

import (
	"Etog/internal/config"
	"Etog/internal/domain/entity"
	"Etog/internal/domain/entity/DTO"
	"Etog/internal/lib/mail"
	"Etog/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AccountHandler struct {
	log      *slog.Logger
	service  AuthService
	config   config.Config
	MailConf config.MailData
}

type AuthService interface {
	CreateAccessToken(userId int) (string, error)
	CreateAccount(account entity.Account) error
	Put(ctx context.Context, key string, value interface{}, exp time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
	GetAccountByEmail(email string) (*entity.Account, error)
	GetAccountByLogin(login string) (*entity.Account, error)
	PutAccount(account entity.Account) error
	CreateRefreshToken(userId int, config config.Config) (string, error)
	TokensGenerate(userId int) (string, string, error)
}

func NewAccountHandler(log *slog.Logger, service AuthService) AccountHandler {
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

	if account.Login == "" || account.Password == "" || account.Mail == "" {
		ctx.JSON(400, gin.H{
			"error": "Please provide login, password and mail",
		})
		return
	}

	account.Official = false
	account.Rating = 100
	account.Active = false
	hashPass, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(400, gin.H{
			"error": "Failed to hash password: " + err.Error(),
		})
		return
	}
	account.Password = string(hashPass)

	err = a.service.CreateAccount(account)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			ctx.JSON(409, gin.H{
				"error": "Account already exists",
			})
		} else {
			ctx.JSON(500, gin.H{
				"error": "Internal server error",
			})
		}
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

	code, err := a.service.Get(ctx, email)
	if errors.Is(err, storage.ErrNotFound) {
		code1 := rand.Intn(10)
		code2 := rand.Intn(10)
		code3 := rand.Intn(10)
		code4 := rand.Intn(10)
		code5 := rand.Intn(10)
		code6 := rand.Intn(10)
		code = strconv.Itoa(code1) + strconv.Itoa(code2) + strconv.Itoa(code3) + strconv.Itoa(code4) + strconv.Itoa(code5) + strconv.Itoa(code6)
		err = a.service.Put(ctx, email, code, time.Minute*5)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
	} else if err != nil {
		ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if err = mail.SendMail(a.MailConf, email, fmt.Sprintf("ваш код для подтверждения почты: %s", code), "confirmation code"); err != nil {
		ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	ctx.JSON(200, gin.H{})
}

func (a *AccountHandler) ConfirmCode(ctx *gin.Context) {
	var confirmCode DTO.ConfirmCodeRequest
	var account *entity.Account
	if err := ctx.ShouldBindJSON(&confirmCode); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON: " + err.Error()})
	}

	code, err := a.service.Get(ctx, confirmCode.Email)
	if errors.Is(err, storage.ErrNotFound) {
		ctx.JSON(400, gin.H{
			"error": "code expired",
		})
		return
	} else if err != nil {
		ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if confirmCode.Code != code {
		ctx.JSON(400, gin.H{
			"error": "wrong code",
		})
		return
	}

	account, err = a.service.GetAccountByEmail(confirmCode.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			ctx.JSON(400, gin.H{
				"error": "Account not found",
			})
			return
		}
		ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	account.Active = true
	err = a.service.PutAccount(*account)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	token, refreshToken, err := a.service.TokensGenerate(account.Id)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(200, gin.H{
		"token":         token,
		"refresh_token": refreshToken,
	})

}

func (a *AccountHandler) Authenticate(ctx *gin.Context) {
	var DTOAccount DTO.AuthRequest
	if err := ctx.ShouldBindJSON(&DTOAccount); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	account, err := a.service.GetAccountByLogin(DTOAccount.Login)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			ctx.JSON(400, gin.H{
				"error": "Account not found",
			})
			return
		}
		ctx.JSON(500, gin.H{"error": "Internal server error"})
	}

	if err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(DTOAccount.Password)); err != nil {
		ctx.JSON(400, gin.H{
			"error": "Password is incorrect",
		})
		return
	}
	if !account.Active {
		ctx.JSON(401, gin.H{
			"error": "Email not checked",
		})
		return
	}
	token, refreshToken, err := a.service.TokensGenerate(account.Id)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	ctx.JSON(200, gin.H{
		"token":         token,
		"refresh_token": refreshToken,
	})
}

//сделать изменения аккаунта с аватаркой
//сделать удаление аккаунта
//сделать выход из всех сессий аккаунта
//сделать запрос на галочку аккаунта
//сделать подписку на кого-то
//сделать отписку  на кого-то
//сделать обновление почты/пароля
