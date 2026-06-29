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
	"net/mail"
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

func NewAuthService(storage *psql.Storage, rStorage *redis.RedisDb, log *slog.Logger, mailConf *config.MailData, jwt *jwt.Jwt, s3 *s3.S3) *AuthService {
	if log == nil {
		log = slog.Default()
	}
	return &AuthService{
		Storage:  storage,
		RStorage: rStorage,
		log:      log,
		MailConf: mailConf,
		Jwt:      jwt,
		S3:       s3,
	}
}

func (a *AuthService) Registration(ctx context.Context, req *account_dto.RegReq) (int, error) {
	const op = "services.Registration"

	log := a.log.With(slog.String("op", op))

	if req.Login == "" || req.Password == "" || req.Mail == "" {
		return 400, errors.New("validation error")
	}

	if len(req.Password) < 7 {
		return 400, errors.New("validation error")
	}
	_, err := mail.ParseAddress(req.Mail)
	if err != nil {
		return 400, errors.New("validation error")
	}

	account := &entity.AccountDb{
		Mail:        req.Mail,
		Login:       req.Login,
		Password:    req.Password,
		Official:    false,
		Description: req.Description,
		Rating:      100,
		Deleted:     false,
		Active:      false,
		Followers:   0,
	}

	_, code, err := a.Storage.GetAccountByLogin(ctx, req.Login)
	if err == nil {
		return 409, errors.New("login already taken")
	}
	if code == 500 {
		return 500, errors.New("server error")
	}
	_, code, err = a.Storage.GetAccountByEmail(ctx, req.Mail)
	if err == nil {
		return 409, errors.New("mail already taken")
	}
	if code == 500 {
		return 500, errors.New("server error")
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Error generating hash password")
		return 500, errors.New("internal server error")

	}
	account.Password = string(hashPass)

	return a.Storage.CreateAccount(ctx, account)
}

func (a *AuthService) SendCode(ctx context.Context, email string, theme, description string) (int, error) {
	const op = "services.SendCode"
	log := a.log.With(slog.String("op", op))
	code, err := a.RStorage.Get(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			code = fmt.Sprintf("%06d", rand.Intn(1000000))
		} else {
			log.Error("Error getting code: " + err.Error())
			return 500, errors.New("server error")
		}
	}
	err = mail2.SendMail(*a.MailConf, email, fmt.Sprintf("%s: %s", description, code), theme)
	if err != nil {
		log.Error("Error sending code: " + err.Error())
		return 500, errors.New("server error")
	}
	err = a.RStorage.Put(ctx, email, code, time.Minute*5)
	if err != nil {
		log.Error("Error putting code: " + err.Error())
		return 500, errors.New("server error")
	}

	return 200, nil

}

func (a *AuthService) ConfirmCode(ctx context.Context, request account_dto.ConfirmCodeRequest) (int, string, string, error) {
	const op = "services.ConfirmCode"
	log := a.log.With(slog.String("op", op))

	codeToCheck, err := a.RStorage.Get(ctx, request.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return 400, "", "", errors.New("code not found or expired")
		}
		log.Error("Error getting code: " + err.Error())
		return 500, "", "", errors.New("server error")
	}
	if codeToCheck != request.Code {
		return 400, "", "", errors.New("invalid or expired code")
	}
	account, code, err := a.Storage.GetAccountByEmail(ctx, request.Email)
	if err != nil {
		return code, "", "", err
	}
	account.Active = true
	code, err = a.Storage.PutAccount(ctx, *account)
	if err != nil {
		return code, "", "", err
	}

	token, refreshToken, err := a.Jwt.TokensGenerate(ctx, account.Id)
	if err != nil {
		log.Error("Error generating token: " + err.Error())
		return 500, "", "", errors.New("server error")
	}
	_ = a.RStorage.Delete(ctx, request.Email)
	return 200, token, refreshToken, nil
}

func (a *AuthService) Authenticate(ctx context.Context, auth account_dto.AuthRequest) (string, string, int, error) {
	account, code, err := a.Storage.GetAccountByLogin(ctx, auth.Login)
	if err != nil {
		return "", "", code, err
	}

	if !account.Active {
		return "", "", 403, errors.New("account is not confirmed")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(auth.Password)); err != nil {
		return "", "", 401, errors.New("invalid login or password")
	}
	token, refreshToken, err := a.Jwt.TokensGenerate(ctx, account.Id)
	if err != nil {
		return "", "", 500, errors.New("server error")
	}
	return token, refreshToken, 200, nil
}

func (a *AuthService) ChangeData(ctx context.Context, id int, account account_dto.ChangeDataRequest) (int, error) {
	accountDb, code, err := a.Storage.GetAccount(ctx, id)
	if err != nil {
		return code, err
	}

	if account.File != nil {
		url, err := a.S3.Upload(account.File, "user-avatar/")
		if err != nil {
			a.log.Error("Error uploading file: " + err.Error())
			return 500, errors.New("server error")
		}
		accountDb.Avatar = url
	}

	accountDb.Description = account.Description

	code, err = a.Storage.PutAccount(ctx, *accountDb)
	if err != nil {
		return code, err
	}

	return 200, nil
}
func (a *AuthService) GetAccount(ctx context.Context, login string, id int) (int, *entity.Account, error) {
	var account *entity.AccountDb
	var err error
	var code int
	if id != -1 {
		account, code, err = a.Storage.GetAccount(ctx, id)
		if err != nil {
			return code, nil, err
		}
	} else if login != "" {
		account, code, err = a.Storage.GetAccountByLogin(ctx, login)
		if err != nil {
			return code, nil, err
		}
	} else {
		return 400, nil, errors.New("validation error")
	}

	if !account.Active || account.Deleted {
		return 404, nil, errors.New("account not found")
	}
	return 200, repo.FromDbUser(account), nil
}

func (a *AuthService) DeleteAccount(ctx context.Context, id int) (int, error) {
	account, code, err := a.Storage.GetAccount(ctx, id)
	if err != nil {
		return code, err
	}
	if account.Deleted {
		return 404, errors.New("account not found")
	}
	account.Deleted = true
	code, err = a.Storage.PutAccount(ctx, *account)
	if err != nil {
		return code, err
	}
	return 200, nil
}

func (a *AuthService) Follow(ctx context.Context, followerId, followingId int) (int, error) {
	if followerId == followingId {
		return 400, errors.New("cannot follow yourself")
	}
	return a.Storage.Subscribe(ctx, followingId, followerId)
}

func (a *AuthService) Unfollow(ctx context.Context, followerId, followingId int) (int, error) {
	return a.Storage.Unsubscribe(ctx, followingId, followerId)
}

func (a *AuthService) ConfirmChangePassword(ctx context.Context, mail string, code string, newPass string) (int, error) {
	const op = "AuthService.ConfirmChangePassword"

	log := a.log.With(slog.String("op", op))
	if newPass == "" || len(newPass) < 7 {
		return 400, errors.New("validation error")
	}

	account, responseCode, err := a.Storage.GetAccountByEmail(ctx, mail)
	if err != nil {
		return responseCode, err
	}
	codeToCheck, err := a.RStorage.Get(ctx, account.Mail)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return 400, errors.New("expired or invalid code")
		}
		log.Error("Error getting code: " + err.Error())
		return 500, errors.New("server error")
	}
	if codeToCheck != code {
		return 400, errors.New("invalid or expired code")
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Error generating password: " + err.Error())
		return 500, errors.New("server error")
	}
	account.Password = string(hashPass)

	return a.Storage.PutAccount(ctx, *account)
}

func (a *AuthService) ConfirmChangeMail(ctx context.Context, id int, newMail string, code string, password string) (int, error) {
	const op = "AuthService.ConfirmChangeMail"
	log := a.log.With(slog.String("op", op))

	account, responseCode, err := a.Storage.GetAccount(ctx, id)
	if err != nil {
		return responseCode, err
	}
	codeToCheck, err := a.RStorage.Get(ctx, account.Mail)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return 400, errors.New("expired or invalid code")
		}
		log.Error("Error getting code: " + err.Error())
		return 500, errors.New("server error")
	}
	if codeToCheck != code {
		return 400, errors.New("invalid or expired code")
	}
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil {
		return 401, errors.New("invalid password")
	}
	account.Mail = newMail
	return a.Storage.PutAccount(ctx, *account)
}

func (a *AuthService) RequestOfficial(ctx context.Context, id int) (int, error) {
	return a.Storage.CreateOfficialRequest(ctx, id)
}

func (a *AuthService) DeleteSessions(ctx context.Context, id int) (int, error) {
	return a.Storage.DeleteRefreshToken(ctx, id)
}
