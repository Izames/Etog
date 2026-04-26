package jwt

import (
	"Etog/internal/domain/entity"
	"Etog/storage/psql"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Jwt struct {
	JwtKey  string
	Storage *psql.Storage
}

func NewJwtLib(JwtKey string, storage *psql.Storage) *Jwt {
	return &Jwt{JwtKey: JwtKey, Storage: storage}
}

func (j *Jwt) CreateAccessToken(userId int) (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		"sub": userId,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(j.JwtKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *Jwt) CreateRefreshToken(userId int) (string, error) {
	GenTokenID := uuid.New().String()

	tokenID, err := bcrypt.GenerateFromPassword([]byte(GenTokenID), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	refreshClaims := jwt.MapClaims{
		"exp": time.Now().Add(24 * 365 * time.Hour).Unix(),
		"sub": userId,
		"jti": string(tokenID),
		"typ": "refresh",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	refreshTokenString, err := refreshToken.SignedString([]byte(j.JwtKey))
	if err != nil {
		return "", err
	}
	err = j.Storage.CreateRefreshToken(entity.RefreshToken{
		UserId:   userId,
		TokenId:  tokenID,
		ExpireAt: time.Now().Add(time.Hour * 24 * 365),
	})

	return refreshTokenString, nil
}

func (j *Jwt) CheckAccessToken(tokenString string) error {
	keyFunc := func(_ *jwt.Token) (interface{}, error) {
		return []byte(j.JwtKey), nil
	}

	token, err := jwt.Parse(tokenString, keyFunc)
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

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("invalid token")
	}

	exp := int64(expFloat)

	if time.Now().After(time.Unix(exp, 0)) {
		return fmt.Errorf("invalid token")
	}

	return nil
}

// return new access token
func (j *Jwt) RefreshToken(refreshTokenString string) (error, string) {
	keyFunc := func(_ *jwt.Token) (interface{}, error) {
		return []byte(j.JwtKey), nil
	}

	token, err := jwt.Parse(refreshTokenString, keyFunc, jwt.WithExpirationRequired())
	if err != nil {
		return err, ""
	}

	if !token.Valid {
		return fmt.Errorf("invalid refresh token"), ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid refresh token"), ""
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("invalid refresh token"), ""
	}

	exp := int64(expFloat)

	if time.Now().After(time.Unix(exp, 0)) {
		return fmt.Errorf("invalid refresh token"), ""
	}

	tokenId, ok := claims["jti"].(string)
	if !ok {
		return fmt.Errorf("invalid refresh token"), ""
	}

	_, err = j.Storage.GetRefreshToken(tokenId)
	if err != nil {
		return fmt.Errorf("invalid refresh token"), ""
	}
	newToken, err := j.CreateAccessToken(int(claims["sub"].(float64)))

	return nil, newToken
}

func (j *Jwt) DeleteRefreshToken(userId int) error {
	err := j.Storage.DeleteRefreshToken(userId)
	if err != nil {
		return err
	}
	return nil
}

func (j *Jwt) TokensGenerate(userId int) (string, string, error) {
	token, err := j.CreateAccessToken(userId)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := j.CreateRefreshToken(userId)
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}
