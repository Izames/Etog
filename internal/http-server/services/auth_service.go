package services

import (
	"Etog/internal/config"
	"Etog/internal/domain/entity"
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Storage  Storage
	RStorage RStorage
	Conf     config.Config
}

type Storage interface {
	CreateRefreshToken(token entity.RefreshToken) error
	GetRefreshToken(tokenId string) (*entity.RefreshToken, error)
	DeleteRefreshToken(userId int) error
	CreateAccount(account entity.Account) error
	GetAccountByEmail(email string) (*entity.Account, error)
	PutAccount(account entity.Account) error
	GetAccountByLogin(login string) (*entity.Account, error)
}

type RStorage interface {
	Put(ctx context.Context, key string, value interface{}, exp time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
}

func NewAuthService(storage Storage, conf config.Config, rStorage RStorage) *AuthService {
	return &AuthService{
		Storage:  storage,
		Conf:     conf,
		RStorage: rStorage,
	}
}

func (a *AuthService) CreateAccessToken(userId int) (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		"sub": userId,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.Conf.JWTKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (a *AuthService) CreateRefreshToken(userId int, config config.Config) (string, error) {
	GenTokenID := uuid.New().String()
	tokenID, err := bcrypt.GenerateFromPassword([]byte(GenTokenID), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	refreshClaims := jwt.MapClaims{
		"exp": time.Now().Add(24 * 365 * time.Hour).Unix(),
		"sub": userId,
		"jti": tokenID,
		"typ": "refresh",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(config.JWTKey))
	if err != nil {
		return "", err
	}

	if err = a.Storage.CreateRefreshToken(entity.RefreshToken{
		TokenId: tokenID,
		UserId:  userId,
	}); err != nil {
		return "", err
	}

	return refreshTokenString, nil
}

func (a *AuthService) CheckAccessToken(tokenString string) error {
	keyFunc := func() jwt.Keyfunc {
		return func(_ *jwt.Token) (interface{}, error) {
			return a.Conf.JWTKey, nil
		}
	}
	token, err := jwt.Parse(tokenString, keyFunc())
	if err != nil {
		return err
	}
	if !token.Valid {
		return fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid token")
	}
	exp, ok := claims["exp"].(int64)
	if !ok {
		return fmt.Errorf("invalid token")
	}
	if time.Now().After(time.Unix(exp, 0)) {
		return fmt.Errorf("invalid token")
	}
	return nil
}

func (a *AuthService) RefreshToken(refreshTokenString string) error {
	keyFunc := func() jwt.Keyfunc {
		return func(_ *jwt.Token) (interface{}, error) {
			return a.Conf.JWTKey, nil
		}
	}
	token, err := jwt.Parse(refreshTokenString, keyFunc(), jwt.WithExpirationRequired())
	if err != nil {
		return err
	}
	if !token.Valid {
		return fmt.Errorf("invalid refresh token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid refresh token")
	}
	exp, ok := claims["exp"].(int64)
	if !ok {
		return fmt.Errorf("invalid refresh token")
	}
	if time.Now().After(time.Unix(exp, 0)) {
		return fmt.Errorf("invalid refresh token")
	}
	tokenId, ok := claims["jti"].(string)
	if !ok {
		return fmt.Errorf("invalid refresh token")
	}
	_, err = a.Storage.GetRefreshToken(tokenId)
	if err != nil {
		return fmt.Errorf("invalid refresh token")
	}
	return nil
}

func (a *AuthService) DeleteRefreshToken(userId int) error {
	err := a.Storage.DeleteRefreshToken(userId)
	if err != nil {
		return err
	}
	return nil
}

func (a *AuthService) TokensGenerate(userId int) (string, string, error) {
	token, err := a.CreateAccessToken(userId)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := a.CreateRefreshToken(userId, a.Conf)
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

func (a *AuthService) CreateAccount(account entity.Account) error {
	return a.Storage.CreateAccount(account)
}

func (a *AuthService) GetAccountByEmail(email string) (*entity.Account, error) {
	return a.Storage.GetAccountByEmail(email)
}

func (a *AuthService) Put(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	return a.RStorage.Put(ctx, key, value, exp)
}
func (a *AuthService) Get(ctx context.Context, key string) (interface{}, error) {
	return a.RStorage.Get(ctx, key)
}

func (a *AuthService) PutAccount(account entity.Account) error {
	return a.Storage.PutAccount(account)
}

func (a *AuthService) GetAccountByLogin(login string) (*entity.Account, error) {
	return a.Storage.GetAccountByLogin(login)
}
