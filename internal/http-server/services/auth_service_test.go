package services

import (
	"Etog/internal/config"
	"Etog/internal/domain/entity"
	"Etog/internal/domain/entity/DTO"
	jwt2 "Etog/internal/lib/jwt"
	"Etog/storage/psql"
	"Etog/storage/redis"
	"context"
	"fmt"
	"testing"
	"time"

	jwt3 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func setup() (*AuthService, *psql.Storage, *redis.RedisDb) {
	storage := psql.New("postgres://postgres:0@localhost:5433/etog?sslmode=disable", nil)
	rStorage := redis.NewRedisDb(config.RedisDb{
		Host:     "localhost",
		Port:     "8091",
		Password: "0",
		User:     "redisuser",
		Index:    "0",
	}, nil)
	jwt := jwt2.NewJwtLib("secret", storage)

	return NewAuthService(storage, rStorage, nil, nil, jwt), storage, rStorage
}

func createTestAccount(t *testing.T, service *AuthService) *entity.Account {
	acc := &entity.Account{
		Mail:        fmt.Sprintf("test_%s@mail.ru", uuid.New().String()),
		Login:       fmt.Sprintf("test_%s", uuid.New().String()),
		Password:    "password123",
		Description: "test account",
	}

	err, code := service.Registration(acc)
	if err != nil || code != 200 {
		t.Fatalf("failed to create account: %v %d", err, code)
	}

	created, err := service.Storage.GetAccountByEmail(acc.Mail)
	if err != nil {
		t.Fatal(err)
	}

	return created
}

func TestAuthService_Registration(t *testing.T) {
	service, _, _ := setup()

	acc := &entity.Account{
		Mail:        fmt.Sprintf("user%s@mail.ru", uuid.New().String()),
		Login:       fmt.Sprintf("user_%s", uuid.New().String()),
		Password:    "password123",
		Description: "test",
	}

	err, code := service.Registration(acc)
	if err != nil {
		t.Fatal(err)
	}
	if code != 200 {
		t.Fatalf("wanted 200 got %d", code)
	}
}

func TestAuthService_Registration_ExpectFailOnSameEmail(t *testing.T) {
	service, _, _ := setup()

	mail := fmt.Sprintf("same_%d@mail.ru", time.Now().UnixNano())

	acc := &entity.Account{
		Mail:     mail,
		Login:    fmt.Sprintf("user_%s", uuid.New().String()),
		Password: "password123",
	}

	service.Registration(acc)
	err, code := service.Registration(acc)

	if err == nil || code != 409 {
		t.Fatalf("wanted 409 but got %d", code)
	}
}

func TestAuthService_Registration_FailOnShortPass(t *testing.T) {
	service, _, _ := setup()

	err, code := service.Registration(&entity.Account{
		Mail:     "test@mail.ru",
		Login:    "123456",
		Password: "123",
	})

	if err == nil || code != 400 {
		t.Fatalf("wanted 400 got %d", code)
	}
}

func TestAuthService_Registration_FailOnWrongMail(t *testing.T) {
	service, _, _ := setup()

	err, code := service.Registration(&entity.Account{
		Mail:     "wrong_mail",
		Login:    "123456",
		Password: "password123",
	})

	if err == nil || code != 400 {
		t.Fatalf("wanted 400 got %d", code)
	}
}

func TestAuthService_SendCode_And_Confirm(t *testing.T) {
	service, _, rStorage := setup()

	acc := createTestAccount(t, service)

	ctx := context.Background()

	// вручную кладем код (так как SendCode не сохраняет)
	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	err, code, token, refreshToken := service.ConfirmCode(ctx, DTO.ConfirmCodeRequest{
		Email: acc.Mail,
		Code:  "123456",
	})

	if err != nil {
		t.Fatal(err)
	}

	if code != 200 {
		t.Fatalf("wanted 200 got %d", code)
	}

	keyFunc := func(_ *jwt3.Token) (interface{}, error) {
		return []byte("secret"), nil
	}

	// 🔹 ACCESS TOKEN
	parsed, err := jwt3.Parse(token, keyFunc)
	if err != nil {
		t.Fatal(err)
	}

	if !parsed.Valid {
		t.Fatal("access token invalid")
	}

	claims := parsed.Claims.(jwt3.MapClaims)

	sub, ok := claims["sub"].(float64)
	if !ok || int(sub) != acc.Id {
		t.Fatalf("invalid sub: %v", claims["sub"])
	}

	refreshParsed, err := jwt3.Parse(refreshToken, keyFunc)
	if err != nil {
		t.Fatal(err)
	}

	if !refreshParsed.Valid {
		t.Fatal("refresh token invalid")
	}

	refreshClaims := refreshParsed.Claims.(jwt3.MapClaims)

	if refreshClaims["typ"] != "refresh" {
		t.Fatal("not refresh token")
	}

	jti, ok := refreshClaims["jti"].(string)
	if !ok || jti == "" {
		t.Fatal("invalid jti")
	}

	subRefresh, ok := refreshClaims["sub"].(float64)
	if !ok || int(subRefresh) != acc.Id {
		t.Fatal("invalid refresh sub")
	}

	err, tokenID := service.Jwt.RefreshToken(refreshToken)
	if err != nil {
		t.Fatal(err)
	}

	if tokenID == "" {
		t.Fatal("empty token id")
	}
}
