package services

import (
	"Etog/internal/config"
	"Etog/internal/domain/entity"
	"Etog/internal/domain/entity/DTO"
	"Etog/internal/lib/jwt"
	mail2 "Etog/internal/lib/mail"
	"Etog/storage"
	"Etog/storage/psql"
	"Etog/storage/redis"
	"Etog/storage/s3"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"mime/multipart"
	"net/mail"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Storage  *psql.Storage
	RStorage *redis.RedisDb
	log      *slog.Logger
	MailConf *config.MailData
	Jwt      *jwt.Jwt
	S3       *s3.S3
}

func NewAuthService(storage *psql.Storage, rStorage *redis.RedisDb, log *slog.Logger, mailConf *config.MailData, jwt *jwt.Jwt) *AuthService {
	if log == nil {
		log = slog.Default()
	}
	return &AuthService{
		Storage:  storage,
		RStorage: rStorage,
		log:      log,
		MailConf: mailConf,
		Jwt:      jwt,
	}
}

func (a *AuthService) Registration(account *entity.Account) (error, int) {
	const op = "services.Registration"

	log := a.log.With(slog.String("op", op))

	if account.Login == "" || account.Password == "" || account.Mail == "" {
		return errors.New("please provide login, password and mail"), 400
	}

	if len(account.Password) < 7 {
		return errors.New("password must be at least 8 characters"), 400
	}
	_, err := mail.ParseAddress(account.Mail)
	if err != nil {
		return errors.New("invalid mail address"), 400
	}

	account.Official = false
	account.Rating = 100
	account.Active = false
	hashPass, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Error generating hash password")
		return errors.New("internal server error"), 500

	}
	account.Password = string(hashPass)

	err = a.Storage.CreateAccount(*account)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			a.log.Error("Account already exists " + account.Mail)
			return errors.New("account already exists"), 409
		} else {
			a.log.Error("Error creating account: " + err.Error())
			return errors.New("internal server error"), 500
		}
	}
	return nil, 200
}

func (a *AuthService) SendCode(ctx context.Context, email string) (error, int) {
	const op = "services.SendCode"
	log := a.log.With(slog.String("op", op))
	code, err := a.RStorage.Get(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			code1 := rand.Intn(10)
			code2 := rand.Intn(10)
			code3 := rand.Intn(10)
			code4 := rand.Intn(10)
			code5 := rand.Intn(10)
			code6 := rand.Intn(10)
			code = strconv.Itoa(code1) + strconv.Itoa(code2) + strconv.Itoa(code3) + strconv.Itoa(code4) + strconv.Itoa(code5) + strconv.Itoa(code6)
		} else {
			log.Error("Error getting code: " + err.Error())
			return errors.New("internal server error"), 500
		}
	}
	err = mail2.SendMail(*a.MailConf, email, fmt.Sprintf("ваш код для подтверждения почты: %s", code), "confirmation code")
	if err != nil {
		log.Error("Error sending code: " + err.Error())
		return errors.New("internal server error"), 500
	}
	err = a.RStorage.Put(ctx, email, code, time.Minute*5)
	if err != nil {
		log.Error("Error putting code: " + err.Error())
		return errors.New("internal server error"), 500
	}

	return nil, 200

}

func (a *AuthService) ConfirmCode(ctx context.Context, request DTO.ConfirmCodeRequest) (error, int, string, string) {
	const op = "services.ConfirmCode"
	log := a.log.With(slog.String("op", op))

	codeToCheck, err := a.RStorage.Get(ctx, request.Email)
	if err != nil {
		log.Error("Error getting code: " + err.Error())
		return errors.New("internal server error"), 500, "", ""
	}
	if codeToCheck != request.Code {
		return errors.New("invalid code"), 400, "", ""
	}
	account, err := a.Storage.GetAccountByEmail(request.Email)
	if err != nil {
		log.Error("Error getting account: " + err.Error())
		return errors.New("internal server error"), 500, "", ""
	}
	account.Active = true
	err = a.Storage.PutAccount(*account)
	if err != nil {
		log.Error("Error putting account: " + err.Error())
		return errors.New("internal server error"), 500, "", ""
	}

	token, refreshToken, err := a.Jwt.TokensGenerate(account.Id)
	if err != nil {
		log.Error("Error generating token: " + err.Error())
		return errors.New("internal server error"), 500, "", ""
	}
	_ = a.RStorage.Delete(ctx, request.Email)
	return nil, 200, token, refreshToken
}

func (a *AuthService) Authenticate(auth DTO.AuthRequest) (string, string, error, int) {
	account, err := a.Storage.GetAccountByLogin(auth.Login)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", "", errors.New("account not found"), 404
		}
		return "", "", errors.New("internal server error"), 500
	}

	if err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(auth.Password)); err != nil {
		return "", "", errors.New("password incorrect"), 400
	}
	if !account.Active {
		return "", "", errors.New("email is not active"), 401
	}
	token, refreshToken, err := a.Jwt.TokensGenerate(account.Id)
	if err != nil {
		return "", "", err, 500
	}
	return token, refreshToken, nil, 200
}

func (a *AuthService) ChangeData(account DTO.ChangeDataRequest, ctx context.Context, fileHeader *multipart.FileHeader) (interface{}, interface{}) {
	UserId := ctx.Value("UserId").(int)

	accountDb, err := a.Storage.GetAccount(UserId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			a.log.Error("Account not found")
			return errors.New("account not found"), 404
		}
		a.log.Error("Error getting account: " + err.Error())
		return errors.New("internal server error"), 500
	}

	url, err := a.S3.Upload(fileHeader, "user-avatar/")

	if err != nil {
		a.log.Error("Error uploading file: " + err.Error())
		return errors.New("internal server error"), 500
	}

	accountDb.Avatar = url

	accountDb.Description = account.Description

	return accountDb, account
}
