package services

import (
	"Etog/internal/domain/entity"
	"Etog/storage"
	"Etog/storage/psql"
	"Etog/storage/redis"
	"errors"
	"log/slog"
	"net/mail"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Storage  *psql.Storage
	RStorage *redis.RedisDb
	log      *slog.Logger
}

func NewAuthService(storage *psql.Storage, rStorage *redis.RedisDb, log *slog.Logger) *AuthService {
	if log == nil {
		log = slog.Default()
	}
	return &AuthService{
		Storage:  storage,
		RStorage: rStorage,
		log:      log,
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

//func (a *AuthService) GetAccountByEmail(email string) (*entity.Account, error) {
//	return a.Storage.GetAccountByEmail(email)
//}
//
//func (a *AuthService) Put(ctx context.Context, key string, value interface{}, exp time.Duration) error {
//	return a.RStorage.Put(ctx, key, value, exp)
//}
//func (a *AuthService) Get(ctx context.Context, key string) (interface{}, error) {
//	return a.RStorage.Get(ctx, key)
//}
//
//func (a *AuthService) PutAccount(account entity.Account) error {
//	return a.Storage.PutAccount(account)
//}
//
//func (a *AuthService) GetAccountByLogin(login string) (*entity.Account, error) {
//	return a.Storage.GetAccountByLogin(login)
//}
