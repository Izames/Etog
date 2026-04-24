package jwt

import (
	"Etog/internal/domain/entity"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Jwt struct {
}

func newJwtLib() *Jwt {
	return &Jwt{}
}

func (j *Jwt) CreateAccessToken(userId int) (string, error) {
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

func (j *Jwt) CreateRefreshToken(userId int, config config.Config) (string, error) {
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

	return tokenID
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
	refreshToken, err := a.CreateRefreshToken(userId, *a.Conf)
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}
