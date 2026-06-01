package services

import (
	"Etog/internal/config"
	"Etog/internal/domain/entity"
	"Etog/internal/domain/entity/DTO/account_dto"
	"Etog/internal/domain/repo"
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
	"net/http"
	"net/mail"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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

func (a *AuthService) Registration(account *entity.AccountDb) (error, int) {
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
			a.log.Error("AccountDb already exists " + account.Mail)
			return errors.New("account_dto already exists"), 409
		} else {
			a.log.Error("Error creating account_dto: " + err.Error())
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

func (a *AuthService) ConfirmCode(ctx context.Context, request account_dto.ConfirmCodeRequest) (error, int, string, string) {
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
		log.Error("Error getting account_dto: " + err.Error())
		return errors.New("internal server error"), 500, "", ""
	}
	account.Active = true
	err = a.Storage.PutAccount(*account)
	if err != nil {
		log.Error("Error putting account_dto: " + err.Error())
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

func (a *AuthService) Authenticate(auth account_dto.AuthRequest) (string, string, error, int) {
	account, err := a.Storage.GetAccountByLogin(auth.Login)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", "", errors.New("account_dto not found"), 404
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

func (a *AuthService) ChangeData(account account_dto.ChangeDataRequest, ctx context.Context) (error, int) {
	UserId := ctx.Value("UserId").(int)

	accountDb, err := a.Storage.GetAccount(UserId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			a.log.Error("AccountDb not found")
			return errors.New("account_dto not found"), 404
		}
		a.log.Error("Error getting account_dto: " + err.Error())
		return errors.New("internal server error"), 500
	}

	url, err := a.S3.Upload(account.File, "user-avatar/")

	if err != nil {
		a.log.Error("Error uploading file: " + err.Error())
		return errors.New("internal server error"), 500
	}

	accountDb.Avatar = url

	accountDb.Description = account.Description

	err = a.Storage.PutAccount(*accountDb)
	if err != nil {
		a.log.Error("Error putting account_dto: " + err.Error())
	}

	return nil, 200
}

func (a *AuthService) GetAccount(ctx *gin.Context, login string, id int) (int, *entity.Account, error) {
	var account *entity.AccountDb
	var err error
	if id != -1 {
		account, err = a.Storage.GetAccount(id)
		if err != nil {
			a.log.Error("Error getting account_dto: " + err.Error())
			return http.StatusInternalServerError, nil, errors.New("internal server error")
		}
	} else if login != "" {
		account, err = a.Storage.GetAccountByLogin(login)
		if err != nil {
			a.log.Error("Error getting account_dto: " + err.Error())
			return http.StatusInternalServerError, nil, errors.New("internal server error")
		}
	} else {
		return 400, nil, errors.New("wrong data")
	}
	if account == nil {
		return http.StatusNotFound, nil, errors.New("account_dto not found")
	}
	if !account.Active || account.Deleted {
		return http.StatusNotFound, nil, errors.New("account_dto not found")
	}
	return http.StatusOK, repo.FromDbUser(account), nil
}

func (a *AuthService) DeleteAccount(ctx *gin.Context, id int) (int, error) {
	account, err := a.Storage.GetAccount(id)
	if err != nil {
		a.log.Error("Error getting account_dto: " + err.Error())
		return http.StatusInternalServerError, errors.New("internal server error")
	}
	account.Deleted = true
	if err = a.Storage.PutAccount(*account); err != nil {
		a.log.Error("Error putting account_dto: " + err.Error())
		return http.StatusInternalServerError, errors.New("internal server error")
	}
	return http.StatusOK, nil
}

func (a *AuthService) RequestOfficial(ctx *gin.Context, id int) error {
	return a.Storage.CreateOfficialRequest(id)
}

func (a *AuthService) Subscribe(ctx *gin.Context, following int, follower int) error {
	if err := a.Storage.Subscribe(following, follower); err != nil {
		a.log.Error("Error subscribing to follower: " + err.Error())
		return errors.New("internal server error")
	}
	return nil
}

func (a *AuthService) Unsubscribe(ctx *gin.Context, following int, follower int) error {
	if err := a.Storage.Unsubscribe(following, follower); err != nil {
		a.log.Error("Error unsubscribing from follower: " + err.Error())
		return errors.New("internal server error")
	}
	return nil
}

func (a *AuthService) ChangePasword(ctx *gin.Context, id int, oldPassword string, newPassword string) error {
	account, err := a.Storage.GetAccount(id)
	if err != nil {
		a.log.Error("Error getting account_dto: " + err.Error())
		return errors.New("internal server error")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(oldPassword)); err != nil {
		a.log.Error("Old password incorrect")
		return errors.New("Old password incorrect")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("Error hashing password: " + err.Error())
		return errors.New("internal server error")
	}
	account.Password = string(hash)
	if err = a.Storage.PutAccount(*account); err != nil {
		a.log.Error("Error putting account_dto: " + err.Error())
		return errors.New("internal server error")
	}
	return nil
}
