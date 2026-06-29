package services

import (
	"Etog/internal/config"
	"Etog/internal/domain/entity"
	"Etog/internal/domain/entity/DTO/account_dto"
	jwt2 "Etog/internal/lib/jwt"
	"Etog/storage"
	"Etog/storage/psql"
	"Etog/storage/redis"
	s4 "Etog/storage/s3"
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
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
	mail := &config.MailData{
		From:     "EventsTogether@yandex.com",
		Host:     "smtp.yandex.ru",
		Port:     "587",
		Password: "cjzlbkpxhiirnxrj",
	}
	s3Client, err := s4.NewS3(&config.S3{
		Region:   "ru-msk",
		Endpoint: "https://hb.vkcs.cloud",
		Access:   "sriyxkxEFpqD9hh2yB9fmJ",
		Secret:   "7RpNAZqDtUqeRK4ixN5ges9AvVH5CAYEnHArxt6gaSu2",
		Bucket:   "etog",
	})
	if err != nil {
		panic(err)
	}
	return NewAuthService(storage, rStorage, nil, mail, jwt, s3Client), storage, rStorage
}

func ginCtx() *gin.Context {
	ctx, _ := gin.CreateTestContext(nil)
	return ctx
}

func createTestAccount(t *testing.T, service *AuthService) *entity.AccountDb {
	t.Helper()
	acc := &account_dto.RegReq{
		Mail:        fmt.Sprintf("test_%s@mail.ru", uuid.New().String()),
		Login:       fmt.Sprintf("test_%s", uuid.New().String()),
		Password:    "password123",
		Description: "test account_dto",
	}
	ctx := context.Background()
	code, err := service.Registration(ctx, acc)
	if err != nil || code != 201 {
		t.Fatalf("failed to create account: %v %d", err, code)
	}
	created, _, err := service.Storage.GetAccountByEmail(ctx, acc.Mail)
	if err != nil {
		t.Fatal(err)
	}
	created.Active = true
	_, err = service.Storage.PutAccount(ctx, *created)
	if err != nil {
		t.Fatalf("failed to activate account: %v", err)
	}
	return created
}

func createTestFile(t *testing.T) *multipart.FileHeader {
	t.Helper()
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	part, err := writer.CreateFormFile("file", "avatar.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = part.Write([]byte("fake image content")); err != nil {
		t.Fatal(err)
	}
	err = writer.Close()
	if err != nil {
		return nil
	}
	req := httptest.NewRequest(http.MethodPost, "/upload", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err = req.ParseMultipartForm(10 << 20); err != nil {
		t.Fatal(err)
	}
	return req.MultipartForm.File["file"][0]
}

// --- Registration ---

func TestAuthService_Registration(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	req := &account_dto.RegReq{
		Mail:     fmt.Sprintf("test_%s@mail.ru", uuid.New().String()),
		Login:    fmt.Sprintf("test_%s", uuid.New().String()),
		Password: "password123",
	}

	code, err := service.Registration(ctx, req)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 201 {
		t.Fatalf("expected 201, got %d", code)
	}
	t.Log("func [Registration] is OK")
}

func TestAuthService_Registration_EmptyLogin(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	req := &account_dto.RegReq{
		Mail:     fmt.Sprintf("test_%s@mail.ru", uuid.New().String()),
		Login:    "",
		Password: "password123",
	}

	code, err := service.Registration(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
	t.Log("func [Registration EmptyLogin] is OK")
}

func TestAuthService_Registration_EmptyPassword(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	req := &account_dto.RegReq{
		Mail:     fmt.Sprintf("test_%s@mail.ru", uuid.New().String()),
		Login:    fmt.Sprintf("test_%s", uuid.New().String()),
		Password: "",
	}

	code, err := service.Registration(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
	t.Log("func [Registration EmptyPassword] is OK")
}

func TestAuthService_Registration_EmptyMail(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	req := &account_dto.RegReq{
		Mail:     "",
		Login:    fmt.Sprintf("test_%s", uuid.New().String()),
		Password: "password123",
	}

	code, err := service.Registration(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
	t.Log("func [Registration EmptyMail] is OK")
}

func TestAuthService_Registration_ShortPassword(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	req := &account_dto.RegReq{
		Mail:     fmt.Sprintf("test_%s@mail.ru", uuid.New().String()),
		Login:    fmt.Sprintf("test_%s", uuid.New().String()),
		Password: "123",
	}

	code, err := service.Registration(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
	t.Log("func [Registration ShortPassword] is OK")
}

func TestAuthService_Registration_WrongMail(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	req := &account_dto.RegReq{
		Mail:     "not-an-email",
		Login:    fmt.Sprintf("test_%s", uuid.New().String()),
		Password: "password123",
	}

	code, err := service.Registration(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
	t.Log("func [Registration WrongMail] is OK")
}

func TestAuthService_Registration_LoginTaken(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	created := createTestAccount(t, service)

	req := &account_dto.RegReq{
		Mail:     fmt.Sprintf("test_%s@mail.ru", uuid.New().String()),
		Login:    created.Login,
		Password: "password123",
	}

	code, err := service.Registration(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 409 {
		t.Fatalf("expected 409, got %d", code)
	}
	t.Log("func [Registration LoginTaken] is OK")
}

func TestAuthService_Registration_MailTaken(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	created := createTestAccount(t, service)

	req := &account_dto.RegReq{
		Mail:     created.Mail,
		Login:    fmt.Sprintf("test_%s", uuid.New().String()),
		Password: "password123",
	}

	code, err := service.Registration(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 409 {
		t.Fatalf("expected 409, got %d", code)
	}
	t.Log("func [Registration MailTaken] is OK")
}

func TestAuthService_Registration_TwoAccounts(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	req1 := &account_dto.RegReq{
		Mail:     fmt.Sprintf("test_%s@mail.ru", uuid.New().String()),
		Login:    fmt.Sprintf("test_%s", uuid.New().String()),
		Password: "password123",
	}
	req2 := &account_dto.RegReq{
		Mail:     fmt.Sprintf("test_%s@mail.ru", uuid.New().String()),
		Login:    fmt.Sprintf("test_%s", uuid.New().String()),
		Password: "password123",
	}

	code, err := service.Registration(ctx, req1)
	if err != nil || code != 201 {
		t.Fatalf("first account failed: %v %d", err, code)
	}

	code, err = service.Registration(ctx, req2)
	if err != nil || code != 201 {
		t.Fatalf("second account failed: %v %d", err, code)
	}
	t.Log("func [Registration TwoAccounts] is OK")
}

// --- SendCode ---

// success: код отправлен → 200
func TestAuthService_SendCode(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	code, err := service.SendCode(ctx, acc.Mail, "send code test", "смена почты тест")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	t.Log("func [SendCode] is OK")
}

// success: код сохранён в redis и он 6-значный
func TestAuthService_SendCode_SavedInRedis(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	code, err := service.SendCode(ctx, acc.Mail, "send code test", "тест отправки кода")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}

	saved, err := rStorage.Get(ctx, acc.Mail)
	if err != nil {
		t.Fatalf("expected code in redis, got error: %v", err)
	}
	if len(saved) != 6 {
		t.Fatalf("expected 6-digit code, got: %s", saved)
	}
	t.Log("func [SendCode SavedInRedis] is OK")
}

// success: повторный вызов переотправляет существующий код
func TestAuthService_SendCode_Resend(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	code, err := service.SendCode(ctx, acc.Mail, "send code test", "тестирование повторной отправки кода")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	t.Log("func [SendCode Resend] is OK")
}

// --- ConfirmCode ---

// success: верный код → 200 + валидные токены
func TestAuthService_ConfirmCode(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	code, token, refreshToken, err := service.ConfirmCode(ctx, account_dto.ConfirmCodeRequest{
		Email: acc.Mail,
		Code:  "123456",
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	if token == "" {
		t.Fatal("expected token, got empty string")
	}
	if refreshToken == "" {
		t.Fatal("expected refresh token, got empty string")
	}
	t.Log("func [ConfirmCode] is OK")
}

// success: access и refresh токены валидны и содержат верный sub
func TestAuthService_ConfirmCode_TokensValid(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	_, token, refreshToken, err := service.ConfirmCode(ctx, account_dto.ConfirmCodeRequest{
		Email: acc.Mail,
		Code:  "123456",
	})
	if err != nil {
		t.Fatal(err)
	}

	keyFunc := func(_ *jwt3.Token) (interface{}, error) {
		return []byte("secret"), nil
	}

	parsed, err := jwt3.Parse(token, keyFunc)
	if err != nil || !parsed.Valid {
		t.Fatalf("access token invalid: %v", err)
	}
	claims := parsed.Claims.(jwt3.MapClaims)
	sub, ok := claims["sub"].(float64)
	if !ok || int(sub) != acc.Id {
		t.Fatalf("invalid sub: %v", claims["sub"])
	}

	refreshParsed, err := jwt3.Parse(refreshToken, keyFunc)
	if err != nil || !refreshParsed.Valid {
		t.Fatalf("refresh token invalid: %v", err)
	}
	refreshClaims := refreshParsed.Claims.(jwt3.MapClaims)
	if refreshClaims["typ"] != "refresh" {
		t.Fatal("not a refresh token")
	}
	jti, ok := refreshClaims["jti"].(string)
	if !ok || jti == "" {
		t.Fatal("invalid jti")
	}
	subRefresh, ok := refreshClaims["sub"].(float64)
	if !ok || int(subRefresh) != acc.Id {
		t.Fatalf("invalid refresh sub: %v", refreshClaims["sub"])
	}
	t.Log("func [ConfirmCode TokensValid] is OK")
}

// success: refresh token можно обменять на новый access token
func TestAuthService_ConfirmCode_RefreshWorks(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	_, _, refreshToken, err := service.ConfirmCode(ctx, account_dto.ConfirmCodeRequest{
		Email: acc.Mail,
		Code:  "123456",
	})
	if err != nil {
		t.Fatal(err)
	}

	newToken, err := service.Jwt.RefreshToken(ctx, refreshToken)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if newToken == "" {
		t.Fatal("expected new token, got empty string")
	}
	t.Log("func [ConfirmCode RefreshWorks] is OK")
}

// fail: неверный код → 400
func TestAuthService_ConfirmCode_WrongCode(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	code, _, _, err := service.ConfirmCode(ctx, account_dto.ConfirmCodeRequest{
		Email: acc.Mail,
		Code:  "000000",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
	t.Log("func [ConfirmCode WrongCode] is OK")
}

// fail: код не запрашивался / истёк → 400
func TestAuthService_ConfirmCode_NoCode(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	code, _, _, err := service.ConfirmCode(ctx, account_dto.ConfirmCodeRequest{
		Email: acc.Mail,
		Code:  "123456",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
	t.Log("func [ConfirmCode NoCode] is OK")
}

// success: после подтверждения код удаляется из redis
func TestAuthService_ConfirmCode_CodeDeletedAfterConfirm(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	_, _, _, err = service.ConfirmCode(ctx, account_dto.ConfirmCodeRequest{
		Email: acc.Mail,
		Code:  "123456",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = rStorage.Get(ctx, acc.Mail)
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("expected code to be deleted from redis, got: %v", err)
	}
	t.Log("func [ConfirmCode CodeDeletedAfterConfirm] is OK")
}

// --- Authenticate ---

// success: верные данные → 200 + токены
func TestAuthService_Authenticate(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	token, refreshToken, code, err := service.Authenticate(ctx, account_dto.AuthRequest{
		Login:    acc.Login,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	if token == "" {
		t.Fatal("expected access token, got empty string")
	}
	if refreshToken == "" {
		t.Fatal("expected refresh token, got empty string")
	}
	t.Log("func [Authenticate] is OK")
}

// success: токены валидны и содержат верный sub
func TestAuthService_Authenticate_TokensValid(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	token, refreshToken, _, err := service.Authenticate(ctx, account_dto.AuthRequest{
		Login:    acc.Login,
		Password: "password123",
	})
	if err != nil {
		t.Fatal(err)
	}

	keyFunc := func(_ *jwt3.Token) (interface{}, error) { return []byte("secret"), nil }

	parsed, err := jwt3.Parse(token, keyFunc)
	if err != nil || !parsed.Valid {
		t.Fatalf("access token invalid: %v", err)
	}
	claims := parsed.Claims.(jwt3.MapClaims)
	sub, ok := claims["sub"].(float64)
	if !ok || int(sub) != acc.Id {
		t.Fatalf("expected sub=%d, got: %v", acc.Id, claims["sub"])
	}

	parsedRefresh, err := jwt3.Parse(refreshToken, keyFunc)
	if err != nil || !parsedRefresh.Valid {
		t.Fatalf("refresh token invalid: %v", err)
	}
	refreshClaims := parsedRefresh.Claims.(jwt3.MapClaims)
	if refreshClaims["typ"] != "refresh" {
		t.Fatal("expected typ=refresh")
	}
	t.Log("func [Authenticate TokensValid] is OK")
}

// fail: неверный пароль → 401
func TestAuthService_Authenticate_WrongPassword(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	_, _, code, err := service.Authenticate(ctx, account_dto.AuthRequest{
		Login:    acc.Login,
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 401 {
		t.Fatalf("expected 401, got %d", code)
	}
	t.Log("func [Authenticate WrongPassword] is OK")
}

// fail: логин не существует → 404
func TestAuthService_Authenticate_UnknownLogin(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	_, _, code, err := service.Authenticate(ctx, account_dto.AuthRequest{
		Login:    fmt.Sprintf("ghost_%s", uuid.New().String()),
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 404 {
		t.Fatalf("expected 404, got %d", code)
	}
	t.Log("func [Authenticate UnknownLogin] is OK")
}

// fail: аккаунт не подтверждён → 403
func TestAuthService_Authenticate_NotActive(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	// создаём аккаунт но НЕ активируем его
	req := &account_dto.RegReq{
		Mail:     fmt.Sprintf("test_%s@mail.ru", uuid.New().String()),
		Login:    fmt.Sprintf("test_%s", uuid.New().String()),
		Password: "password123",
	}
	code, err := service.Registration(ctx, req)
	if err != nil || code != 201 {
		t.Fatalf("setup failed: %v %d", err, code)
	}

	_, _, code, err = service.Authenticate(ctx, account_dto.AuthRequest{
		Login:    req.Login,
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 403 {
		t.Fatalf("expected 403, got %d", code)
	}
	t.Log("func [Authenticate NotActive] is OK")
}

// fail: пустой логин → 404
func TestAuthService_Authenticate_EmptyLogin(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	_, _, code, err := service.Authenticate(ctx, account_dto.AuthRequest{
		Login:    "",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code == 200 {
		t.Fatal("expected non-200, got 200")
	}
	t.Log("func [Authenticate EmptyLogin] is OK")
}

// fail: пустой пароль → 401
func TestAuthService_Authenticate_EmptyPassword(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	_, _, code, err := service.Authenticate(ctx, account_dto.AuthRequest{
		Login:    acc.Login,
		Password: "",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 401 {
		t.Fatalf("expected 401, got %d", code)
	}
	t.Log("func [Authenticate EmptyPassword] is OK")
}

// --- ChangeData ---

// success: описание обновляется
func TestAuthService_ChangeData_Description(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	code, err := service.ChangeData(ctx, acc.Id, account_dto.ChangeDataRequest{
		Description: "new description",
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	t.Log("func [ChangeData Description] is OK")
}

// success: описание реально сохранилось в бд
func TestAuthService_ChangeData_PersistsInDb(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	code, err := service.ChangeData(ctx, acc.Id, account_dto.ChangeDataRequest{
		Description: "persisted description",
	})
	if err != nil || code != 200 {
		t.Fatalf("change failed: %v %d", err, code)
	}

	_, updated, err := service.GetAccount(ctx, acc.Login, -1)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Description != "persisted description" {
		t.Fatalf("expected description='persisted description', got: %s", updated.Description)
	}
	t.Log("func [ChangeData PersistsInDb] is OK")
}

// fail: несуществующий id → 404
func TestAuthService_ChangeData_UnknownId(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	code, err := service.ChangeData(ctx, 999999999, account_dto.ChangeDataRequest{
		Description: "some description",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 404 {
		t.Fatalf("expected 404, got %d", code)
	}
	t.Log("func [ChangeData UnknownId] is OK")
}

// fail: изменение чужого аккаунта невозможно через jwt (проверяем что id из токена)
func TestAuthService_ChangeData_CannotChangeOther(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc1 := createTestAccount(t, service)
	acc2 := createTestAccount(t, service)

	// меняем acc2 используя id acc1 — данные должны применяться только к себе
	code, err := service.ChangeData(ctx, acc1.Id, account_dto.ChangeDataRequest{
		Description: "hacked",
	})
	if err != nil || code != 200 {
		t.Fatalf("unexpected: %v %d", err, code)
	}

	// проверяем что acc2 не изменился
	_, acc2After, err := service.GetAccount(ctx, acc2.Login, -1)
	if err != nil {
		t.Fatal(err)
	}
	if acc2After.Description == "hacked" {
		t.Fatal("acc2 was changed — id must come from JWT, not from request")
	}
	t.Log("func [ChangeData CannotChangeOther] is OK")
}

// --- GetAccount ---

// success: получение аккаунта по логину → 200
func TestAuthService_GetAccount_ByLogin(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	code, account, err := service.GetAccount(ctx, acc.Login, -1)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	if account == nil {
		t.Fatal("expected account, got nil")
	}
	if account.Login != acc.Login {
		t.Fatalf("expected login=%s, got: %s", acc.Login, account.Login)
	}
	t.Log("func [GetAccount ByLogin] is OK")
}

// success: получение аккаунта по id → 200
func TestAuthService_GetAccount_ById(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	code, account, err := service.GetAccount(ctx, "", acc.Id)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	if account == nil {
		t.Fatal("expected account, got nil")
	}
	if account.Login != acc.Login {
		t.Fatalf("expected login=%s, got: %s", acc.Login, account.Login)
	}
	t.Log("func [GetAccount ById] is OK")
}

// success: возвращаемые данные соответствуют аккаунту
func TestAuthService_GetAccount_DataCorrect(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	_, account, err := service.GetAccount(ctx, acc.Login, -1)
	if err != nil {
		t.Fatal(err)
	}

	if account.Login != acc.Login {
		t.Fatalf("login mismatch: expected %s, got %s", acc.Login, account.Login)
	}
	if account.Id != acc.Id {
		t.Fatalf("id mismatch: expected %d, got %d", acc.Id, account.Id)
	}
	t.Log("func [GetAccount DataCorrect] is OK")
}

// fail: логин не существует → 404
func TestAuthService_GetAccount_UnknownLogin(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	code, _, err := service.GetAccount(ctx, fmt.Sprintf("ghost_%s", uuid.New().String()), -1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 404 {
		t.Fatalf("expected 404, got %d", code)
	}
	t.Log("func [GetAccount UnknownLogin] is OK")
}

// fail: id не существует → 404
func TestAuthService_GetAccount_UnknownId(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	code, _, err := service.GetAccount(ctx, "", 999999999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 404 {
		t.Fatalf("expected 404, got %d", code)
	}
	t.Log("func [GetAccount UnknownId] is OK")
}

// fail: ни логин ни id не переданы → 400
func TestAuthService_GetAccount_NoParams(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	code, _, err := service.GetAccount(ctx, "", -1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
	t.Log("func [GetAccount NoParams] is OK")
}

// fail: удалённый аккаунт не возвращается → 404
func TestAuthService_GetAccount_Deleted(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	// логически удаляем аккаунт
	acc.Deleted = true
	_, err := service.Storage.PutAccount(ctx, *acc)
	if err != nil {
		t.Fatalf("failed to delete account: %v", err)
	}

	code, _, err := service.GetAccount(ctx, acc.Login, -1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 404 {
		t.Fatalf("expected 404, got %d", code)
	}
	t.Log("func [GetAccount Deleted] is OK")
}

// fail: неактивный аккаунт не возвращается → 404
func TestAuthService_GetAccount_NotActive(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	// создаём но не активируем
	req := &account_dto.RegReq{
		Mail:     fmt.Sprintf("test_%s@mail.ru", uuid.New().String()),
		Login:    fmt.Sprintf("test_%s", uuid.New().String()),
		Password: "password123",
	}
	code, err := service.Registration(ctx, req)
	if err != nil || code != 201 {
		t.Fatalf("setup failed: %v %d", err, code)
	}

	code, _, err = service.GetAccount(ctx, req.Login, -1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 404 {
		t.Fatalf("expected 404, got %d", code)
	}
	t.Log("func [GetAccount NotActive] is OK")
}

// --- DeleteAccount ---

// success: аккаунт помечается как удалённый → 200
func TestAuthService_DeleteAccount(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	code, err := service.DeleteAccount(ctx, acc.Id)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	t.Log("func [DeleteAccount] is OK")
}

// success: удалённый аккаунт недоступен через GetAccount → 404
func TestAuthService_DeleteAccount_NotAccessible(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	code, err := service.DeleteAccount(ctx, acc.Id)
	if err != nil || code != 200 {
		t.Fatalf("delete failed: %v %d", err, code)
	}

	code, _, err = service.GetAccount(ctx, acc.Login, -1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 404 {
		t.Fatalf("expected 404, got %d", code)
	}
	t.Log("func [DeleteAccount NotAccessible] is OK")
}

// success: флаг deleted реально выставлен в бд
func TestAuthService_DeleteAccount_FlagInDb(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	code, err := service.DeleteAccount(ctx, acc.Id)
	if err != nil || code != 200 {
		t.Fatalf("delete failed: %v %d", err, code)
	}

	raw, _, err := service.Storage.GetAccount(ctx, acc.Id)
	if err != nil {
		t.Fatalf("expected to find account in db, got: %v", err)
	}
	if !raw.Deleted {
		t.Fatal("expected Deleted=true in db, got false")
	}
	t.Log("func [DeleteAccount FlagInDb] is OK")
}

// fail: несуществующий id → 404
func TestAuthService_DeleteAccount_UnknownId(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	code, err := service.DeleteAccount(ctx, 999999999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 404 {
		t.Fatalf("expected 404, got %d", code)
	}
	t.Log("func [DeleteAccount UnknownId] is OK")
}

// fail: повторное удаление уже удалённого аккаунта → 404
func TestAuthService_DeleteAccount_AlreadyDeleted(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	code, err := service.DeleteAccount(ctx, acc.Id)
	if err != nil || code != 200 {
		t.Fatalf("first delete failed: %v %d", err, code)
	}

	code, err = service.DeleteAccount(ctx, acc.Id)
	if err == nil {
		t.Fatal("expected error on second delete, got nil")
	}
	if code != 404 {
		t.Fatalf("expected 404, got %d", code)
	}
	t.Log("func [DeleteAccount AlreadyDeleted] is OK")
}

// --- Follow / Unfollow ---

// success: подписка → 200
func TestAuthService_Follow(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc1 := createTestAccount(t, service)
	acc2 := createTestAccount(t, service)

	code, err := service.Follow(ctx, acc1.Id, acc2.Id)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	t.Log("func [Follow] is OK")
}

// success: после подписки счётчик подписчиков увеличился
func TestAuthService_Follow_CounterIncremented(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc1 := createTestAccount(t, service)
	acc2 := createTestAccount(t, service)

	before, _, err := service.Storage.GetAccount(ctx, acc2.Id)
	if err != nil {
		t.Fatal(err)
	}

	code, err := service.Follow(ctx, acc1.Id, acc2.Id)
	if err != nil || code != 200 {
		t.Fatalf("follow failed: %v %d", err, code)
	}

	after, _, err := service.Storage.GetAccount(ctx, acc2.Id)
	if err != nil {
		t.Fatal(err)
	}
	if after.Followers != before.Followers+1 {
		t.Fatalf("expected followers=%d, got %d", before.Followers+1, after.Followers)
	}
	t.Log("func [Follow CounterIncremented] is OK")
}

// fail: повторная подписка → ошибка
func TestAuthService_Follow_AlreadyFollowing(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc1 := createTestAccount(t, service)
	acc2 := createTestAccount(t, service)

	code, err := service.Follow(ctx, acc1.Id, acc2.Id)
	if err != nil || code != 200 {
		t.Fatalf("first follow failed: %v %d", err, code)
	}

	code, err = service.Follow(ctx, acc1.Id, acc2.Id)
	if err == nil {
		t.Fatal("expected error on duplicate follow, got nil")
	}
	if code == 200 {
		t.Fatal("expected non-200, got 200")
	}
	t.Log("func [Follow AlreadyFollowing] is OK")
}

// fail: подписка на самого себя → 400
func TestAuthService_Follow_Self(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	code, err := service.Follow(ctx, acc.Id, acc.Id)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
	t.Log("func [Follow Self] is OK")
}

// success: отписка → 200
func TestAuthService_Unfollow(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc1 := createTestAccount(t, service)
	acc2 := createTestAccount(t, service)

	code, err := service.Follow(ctx, acc1.Id, acc2.Id)
	if err != nil || code != 200 {
		t.Fatalf("follow failed: %v %d", err, code)
	}

	code, err = service.Unfollow(ctx, acc1.Id, acc2.Id)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	t.Log("func [Unfollow] is OK")
}

// success: после отписки счётчик подписчиков уменьшился
func TestAuthService_Unfollow_CounterDecremented(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc1 := createTestAccount(t, service)
	acc2 := createTestAccount(t, service)

	code, err := service.Follow(ctx, acc1.Id, acc2.Id)
	if err != nil || code != 200 {
		t.Fatalf("follow failed: %v %d", err, code)
	}

	before, _, err := service.Storage.GetAccount(ctx, acc2.Id)
	if err != nil {
		t.Fatal(err)
	}

	code, err = service.Unfollow(ctx, acc1.Id, acc2.Id)
	if err != nil || code != 200 {
		t.Fatalf("unfollow failed: %v %d", err, code)
	}

	after, _, err := service.Storage.GetAccount(ctx, acc2.Id)
	if err != nil {
		t.Fatal(err)
	}
	if after.Followers != before.Followers-1 {
		t.Fatalf("expected followers=%d, got %d", before.Followers-1, after.Followers)
	}
	t.Log("func [Unfollow CounterDecremented] is OK")
}

// fail: отписка без подписки → ошибка
func TestAuthService_Unfollow_NotFollowing(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc1 := createTestAccount(t, service)
	acc2 := createTestAccount(t, service)

	code, err := service.Unfollow(ctx, acc1.Id, acc2.Id)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code == 200 {
		t.Fatal("expected non-200, got 200")
	}
	t.Log("func [Unfollow NotFollowing] is OK")
}

// --- ConfirmChangePassword ---

// success: пароль успешно изменён → 200
func TestAuthService_ConfirmChangePassword(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	code, err := service.ConfirmChangePassword(ctx, acc.Mail, "123456", "newpassword123")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	t.Log("func [ConfirmChangePassword] is OK")
}

// success: после смены пароля можно войти с новым паролем
func TestAuthService_ConfirmChangePassword_LoginWithNewPass(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	code, err := service.ConfirmChangePassword(ctx, acc.Mail, "123456", "newpassword123")
	if err != nil || code != 200 {
		t.Fatalf("change failed: %v %d", err, code)
	}

	_, _, code, err = service.Authenticate(ctx, account_dto.AuthRequest{
		Login:    acc.Login,
		Password: "newpassword123",
	})
	if err != nil {
		t.Fatalf("expected login with new password to work, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	t.Log("func [ConfirmChangePassword LoginWithNewPass] is OK")
}

// success: после смены пароля старый пароль не работает
func TestAuthService_ConfirmChangePassword_OldPassRejected(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	code, err := service.ConfirmChangePassword(ctx, acc.Mail, "123456", "newpassword123")
	if err != nil || code != 200 {
		t.Fatalf("change failed: %v %d", err, code)
	}

	_, _, code, err = service.Authenticate(ctx, account_dto.AuthRequest{
		Login:    acc.Login,
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error with old password, got nil")
	}
	if code != 401 {
		t.Fatalf("expected 401, got %d", code)
	}
	t.Log("func [ConfirmChangePassword OldPassRejected] is OK")
}

// fail: неверный код → 400
func TestAuthService_ConfirmChangePassword_WrongCode(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	code, err := service.ConfirmChangePassword(ctx, acc.Mail, "000000", "newpassword123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
	t.Log("func [ConfirmChangePassword WrongCode] is OK")
}

// fail: код истёк / не запрашивался → 400
func TestAuthService_ConfirmChangePassword_NoCode(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	code, err := service.ConfirmChangePassword(ctx, acc.Mail, "123456", "newpassword123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
	t.Log("func [ConfirmChangePassword NoCode] is OK")
}

// fail: несуществующий email → 404
func TestAuthService_ConfirmChangePassword_UnknownMail(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	code, err := service.ConfirmChangePassword(ctx, "ghost@mail.ru", "123456", "newpassword123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 404 {
		t.Fatalf("expected 404, got %d", code)
	}
	t.Log("func [ConfirmChangePassword UnknownMail] is OK")
}

// --- ConfirmChangeMail ---

// success: почта успешно изменена → 200
func TestAuthService_ConfirmChangeMail(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)
	newMail := fmt.Sprintf("new_%s@mail.ru", uuid.New().String())

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	code, err := service.ConfirmChangeMail(ctx, acc.Id, newMail, "123456", "password123")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	t.Log("func [ConfirmChangeMail] is OK")
}

// success: новая почта реально сохранилась в бд
func TestAuthService_ConfirmChangeMail_PersistsInDb(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)
	newMail := fmt.Sprintf("new_%s@mail.ru", uuid.New().String())

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	code, err := service.ConfirmChangeMail(ctx, acc.Id, newMail, "123456", "password123")
	if err != nil || code != 200 {
		t.Fatalf("change failed: %v %d", err, code)
	}

	updated, _, err := service.Storage.GetAccount(ctx, acc.Id)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Mail != newMail {
		t.Fatalf("expected mail=%s, got: %s", newMail, updated.Mail)
	}
	t.Log("func [ConfirmChangeMail PersistsInDb] is OK")
}

// fail: неверный код → 400
func TestAuthService_ConfirmChangeMail_WrongCode(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)
	newMail := fmt.Sprintf("new_%s@mail.ru", uuid.New().String())

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	code, err := service.ConfirmChangeMail(ctx, acc.Id, newMail, "000000", "password123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
	t.Log("func [ConfirmChangeMail WrongCode] is OK")
}

// fail: неверный пароль → 401
func TestAuthService_ConfirmChangeMail_WrongPassword(t *testing.T) {
	service, _, rStorage := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)
	newMail := fmt.Sprintf("new_%s@mail.ru", uuid.New().String())

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	code, err := service.ConfirmChangeMail(ctx, acc.Id, newMail, "123456", "wrongpassword")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 401 {
		t.Fatalf("expected 401, got %d", code)
	}
	t.Log("func [ConfirmChangeMail WrongPassword] is OK")
}

// fail: код не запрашивался → 400
func TestAuthService_ConfirmChangeMail_NoCode(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)
	newMail := fmt.Sprintf("new_%s@mail.ru", uuid.New().String())

	code, err := service.ConfirmChangeMail(ctx, acc.Id, newMail, "123456", "password123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
	t.Log("func [ConfirmChangeMail NoCode] is OK")
}

// --- RequestOfficial ---

// success: заявка на верификацию отправлена → 200
func TestAuthService_RequestOfficial(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	code, err := service.RequestOfficial(ctx, acc.Id)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 201 {
		t.Fatalf("expected 201, got %d", code)
	}
	t.Log("func [RequestOfficial] is OK")
}

// --- DeleteSessions ---

// success: сессии удалены → 200
func TestAuthService_DeleteSessions(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	// создаём токены
	_, _, err := service.Jwt.TokensGenerate(ctx, acc.Id)
	if err != nil {
		t.Fatal(err)
	}

	code, err := service.DeleteSessions(ctx, acc.Id)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	t.Log("func [DeleteSessions] is OK")
}

// success: после удаления сессий refresh токен не работает
func TestAuthService_DeleteSessions_RefreshInvalidated(t *testing.T) {
	service, _, _ := setup()
	ctx := context.Background()

	acc := createTestAccount(t, service)

	_, refreshToken, err := service.Jwt.TokensGenerate(ctx, acc.Id)
	if err != nil {
		t.Fatal(err)
	}

	code, err := service.DeleteSessions(ctx, acc.Id)
	if err != nil || code != 200 {
		t.Fatalf("delete sessions failed: %v %d", err, code)
	}

	_, err = service.Jwt.RefreshToken(ctx, refreshToken)
	if err == nil {
		t.Fatal("expected refresh to fail after session delete, got nil")
	}
	t.Log("func [DeleteSessions RefreshInvalidated] is OK")
}
