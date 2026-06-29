package jwt

import (
	"Etog/internal/domain/entity"
	"Etog/storage"
	"Etog/storage/psql"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	accessTokenTTL  = time.Hour
	refreshTokenTTL = 24 * 365 * time.Hour

	claimSub = "sub"
	claimExp = "exp"
	claimJTI = "jti"
	claimTyp = "typ"

	tokenTypeRefresh = "refresh"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidTokenType = errors.New("invalid token type")
)

type Jwt struct {
	JwtKey  string
	Storage *psql.Storage
	parser  *jwt.Parser
}

func NewJwtLib(jwtKey string, storage *psql.Storage) *Jwt {
	return &Jwt{
		JwtKey:  jwtKey,
		Storage: storage,
		parser:  jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()})),
	}
}

func (j *Jwt) keyFunc(_ *jwt.Token) (interface{}, error) {
	return []byte(j.JwtKey), nil
}

func (j *Jwt) CreateAccessToken(userId int) (string, error) {
	claims := jwt.MapClaims{
		claimExp: time.Now().Add(accessTokenTTL).Unix(),
		claimSub: userId,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(j.JwtKey))
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return tokenString, nil
}

func (j *Jwt) CreateRefreshToken(ctx context.Context, userId int) (string, error) {
	tokenID := uuid.New().String()

	refreshClaims := jwt.MapClaims{
		claimExp: time.Now().Add(refreshTokenTTL).Unix(),
		claimSub: userId,
		claimJTI: tokenID,
		claimTyp: tokenTypeRefresh,
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	refreshTokenString, err := refreshToken.SignedString([]byte(j.JwtKey))
	if err != nil {
		return "", fmt.Errorf("sign refresh token: %w", err)
	}

	if err := j.Storage.CreateRefreshToken(ctx, entity.RefreshToken{
		UserId:   userId,
		TokenId:  tokenID,
		ExpireAt: time.Now().Add(refreshTokenTTL),
	}); err != nil {
		return "", fmt.Errorf("store refresh token: %w", err)
	}

	return refreshTokenString, nil
}

func (j *Jwt) parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := j.parser.Parse(tokenString, j.keyFunc)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func extractUserId(claims jwt.MapClaims) (int, error) {
	userIdF, ok := claims[claimSub].(float64)
	if !ok {
		return 0, ErrInvalidToken
	}
	return int(userIdF), nil
}

func (j *Jwt) CheckAccessToken(tokenString string) error {
	_, err := j.parseToken(tokenString)
	return err
}

func (j *Jwt) RefreshToken(ctx context.Context, refreshTokenString string) (string, error) {
	claims, err := j.parseToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	typ, ok := claims[claimTyp].(string)
	if !ok || typ != tokenTypeRefresh {
		return "", ErrInvalidTokenType
	}

	tokenId, ok := claims[claimJTI].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	userId, err := extractUserId(claims)
	if err != nil {
		return "", err
	}

	if _, err := j.Storage.GetRefreshToken(ctx, tokenId); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", ErrInvalidToken
		}
		return "", fmt.Errorf("get refresh token: %w", err)
	}

	newAccessToken, err := j.CreateAccessToken(userId)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

func (j *Jwt) TokensGenerate(ctx context.Context, userId int) (string, string, error) {
	accessToken, err := j.CreateAccessToken(userId)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := j.CreateRefreshToken(ctx, userId)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (j *Jwt) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token is required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		claims, err := j.parseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token is invalid or expired"})
			c.Abort()
			return
		}

		if typ, ok := claims[claimTyp].(string); ok && typ == tokenTypeRefresh {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
			c.Abort()
			return
		}

		userId, err := extractUserId(claims)
		if err != nil || userId == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("userId", userId)
		c.Next()
	}
}
