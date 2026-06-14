package services

import (
	"Etog/internal/config"
	"Etog/internal/domain/entity"
	"Etog/internal/domain/entity/DTO/account_dto"
	jwt2 "Etog/internal/lib/jwt"
	"Etog/storage/psql"
	"Etog/storage/redis"
	"Etog/storage/s3"
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwt3 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

func ginCtx() *gin.Context {
	ctx, _ := gin.CreateTestContext(nil)
	return ctx
}

func createTestAccount(t *testing.T, service *AuthService) *entity.AccountDb {
	t.Helper()
	acc := &entity.AccountDb{
		Mail:        fmt.Sprintf("test_%s@mail.ru", uuid.New().String()),
		Login:       fmt.Sprintf("test_%s", uuid.New().String()),
		Password:    "password123",
		Description: "test account_dto",
	}
	err, code := service.Registration(acc)
	if err != nil || code != 200 {
		t.Fatalf("failed to create account_dto: %v %d", err, code)
	}
	created, err := service.Storage.GetAccountByEmail(acc.Mail)
	if err != nil {
		t.Fatal(err)
	}
	created.Active = true
	if err = service.Storage.PutAccount(*created); err != nil {
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

	acc := &entity.AccountDb{
		Mail:     fmt.Sprintf("user%s@mail.ru", uuid.New().String()),
		Login:    fmt.Sprintf("user_%s", uuid.New().String()),
		Password: "password123",
	}

	err, code := service.Registration(acc)
	if err != nil {
		t.Fatal(err)
	}
	if code != 200 {
		t.Fatalf("wanted 200 got %d", code)
	}
	t.Log("func [Registration] is OK")
}

func TestAuthService_Registration_SameEmail(t *testing.T) {
	service, _, _ := setup()

	mail := fmt.Sprintf("same_%s@mail.ru", uuid.New().String())
	acc := &entity.AccountDb{
		Mail:     mail,
		Login:    fmt.Sprintf("user_%s", uuid.New().String()),
		Password: "password123",
	}

	err, code := service.Registration(acc)
	if err != nil || code != 200 {
		t.Fatalf("first registration failed: %v %d", err, code)
	}

	acc.Login = fmt.Sprintf("user_%s", uuid.New().String())
	err, code = service.Registration(acc)
	if err == nil || code != 409 {
		t.Fatalf("wanted 409 got %d", code)
	}
	t.Log("func [Registration SameEmail] is OK")
}

func TestAuthService_Registration_ShortPassword(t *testing.T) {
	service, _, _ := setup()

	err, code := service.Registration(&entity.AccountDb{
		Mail:     "test@mail.ru",
		Login:    "123456",
		Password: "123",
	})
	if err == nil || code != 400 {
		t.Fatalf("wanted 400 got %d", code)
	}
	t.Log("func [Registration ShortPassword] is OK")
}

func TestAuthService_Registration_WrongMail(t *testing.T) {
	service, _, _ := setup()

	err, code := service.Registration(&entity.AccountDb{
		Mail:     "wrong_mail",
		Login:    "123456",
		Password: "password123",
	})
	if err == nil || code != 400 {
		t.Fatalf("wanted 400 got %d", code)
	}
	t.Log("func [Registration WrongMail] is OK")
}

// --- SendCode & ConfirmCode ---

func TestAuthService_SendCode_And_Confirm(t *testing.T) {
	service, _, rStorage := setup()

	acc := createTestAccount(t, service)
	ctx := context.Background()

	err := rStorage.Put(ctx, acc.Mail, "123456", time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}

	err, code, token, refreshToken := service.ConfirmCode(ctx, account_dto.ConfirmCodeRequest{
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

	// Access token
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

	// Refresh token
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

	err, newToken := service.Jwt.RefreshToken(refreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if newToken == "" {
		t.Fatal("empty new token")
	}
	t.Log("func [SendCode & ConfirmCode] is OK")
}

// --- Authenticate ---

func TestAuthService_Authenticate_Success(t *testing.T) {
	service, db, _ := setup()

	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	account := entity.AccountDb{
		Mail:     uuid.New().String() + "@mail.ru",
		Login:    uuid.New().String() + "_login",
		Password: string(hash),
		Active:   true,
	}
	if err := db.PutAccount(account); err != nil {
		t.Fatal(err)
	}

	token, refresh, err, code := service.Authenticate(account_dto.AuthRequest{
		Login:    account.Login,
		Password: password,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	if token == "" || refresh == "" {
		t.Fatal("tokens should not be empty")
	}
	t.Log("func [Authenticate Success] is OK")
}

func TestAuthService_Authenticate_NotFound(t *testing.T) {
	service, _, _ := setup()

	_, _, err, code := service.Authenticate(account_dto.AuthRequest{
		Login:    "no_such_user_" + uuid.New().String(),
		Password: "1234",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if code != 404 {
		t.Fatalf("expected 404, got %d", code)
	}
	t.Log("func [Authenticate NotFound] is OK")
}

func TestAuthService_Authenticate_WrongPassword(t *testing.T) {
	service, db, _ := setup()

	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
	account := entity.AccountDb{
		Mail:     uuid.New().String() + "@mail.ru",
		Login:    uuid.New().String() + "_login",
		Password: string(hash),
		Active:   true,
	}
	if err := db.PutAccount(account); err != nil {
		t.Fatal(err)
	}

	_, _, err, code := service.Authenticate(account_dto.AuthRequest{
		Login:    account.Login,
		Password: "wrongpass",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
	t.Log("func [Authenticate WrongPassword] is OK")
}

func TestAuthService_Authenticate_NotActive(t *testing.T) {
	service, db, _ := setup()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	account := entity.AccountDb{
		Mail:     uuid.New().String() + "@mail.ru",
		Login:    uuid.New().String() + "_login",
		Password: string(hash),
		Active:   false,
	}
	if err := db.PutAccount(account); err != nil {
		t.Fatal(err)
	}

	_, _, err, code := service.Authenticate(account_dto.AuthRequest{
		Login:    account.Login,
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if code != 401 {
		t.Fatalf("expected 401, got %d", code)
	}
	t.Log("func [Authenticate NotActive] is OK")
}

// --- ChangeData ---

func TestAuthService_ChangeData_Success(t *testing.T) {
	service, _, _ := setup()
	service.S3, _ = s3.NewS3(&config.S3{
		Endpoint: "https://hb.vkcs.cloud",
		Region:   "ru-msk",
		Bucket:   "etog",
		Access:   "sriyxkxEFpqD9hh2yB9fmJ",
		Secret:   "7RpNAZqDtUqeRK4ixN5ges9AvVH5CAYEnHArxt6gaSu2",
	})

	acc := createTestAccount(t, service)
	ctx := context.WithValue(context.Background(), "UserId", acc.Id)

	err, code := service.ChangeData(account_dto.ChangeDataRequest{
		Description: "new description",
		File:        createTestFile(t),
	}, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200 got %d", code)
	}

	updated, errDb := service.Storage.GetAccount(acc.Id)
	if errDb != nil {
		t.Fatal(errDb)
	}
	if updated.Description != "new description" {
		t.Fatalf("expected updated description got %s", updated.Description)
	}
	t.Log("func [ChangeData Success] is OK")
}

func TestAuthService_ChangeData_UploadError(t *testing.T) {
	service, _, _ := setup()
	service.S3, _ = s3.NewS3(&config.S3{
		Endpoint: "https://hb.vkcs.cloud",
		Region:   "ru-msk",
		Bucket:   "etog",
		Access:   "sriyxkxEFpqD9hh2yB9fmJ",
		Secret:   "7RpNAZqDtUqeRK4ixN5ges9AvVH5CAYEnHArxt6gaSu2",
	})

	acc := createTestAccount(t, service)
	ctx := context.WithValue(context.Background(), "UserId", acc.Id)

	err, code := service.ChangeData(account_dto.ChangeDataRequest{
		Description: "updated",
		File:        &multipart.FileHeader{},
	}, ctx)
	if err == nil {
		t.Fatal("expected error")
	}
	if code != 500 {
		t.Fatalf("expected 500 got %d", code)
	}
	t.Log("func [ChangeData UploadError] is OK")
}

// --- GetAccount ---

func TestAuthService_GetAccount_ById(t *testing.T) {
	service, _, _ := setup()
	acc := createTestAccount(t, service)

	code, account, err := service.GetAccount(ginCtx(), "", acc.Id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != http.StatusOK {
		t.Fatalf("expected 200 got %d", code)
	}
	if account == nil {
		t.Fatal("account is nil")
	}
	t.Log("func [GetAccount ById] is OK")
}

func TestAuthService_GetAccount_ByLogin(t *testing.T) {
	service, _, _ := setup()
	acc := createTestAccount(t, service)

	code, account, err := service.GetAccount(ginCtx(), acc.Login, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != http.StatusOK {
		t.Fatalf("expected 200 got %d", code)
	}
	if account == nil {
		t.Fatal("account is nil")
	}
	if account.Login != acc.Login {
		t.Fatalf("expected login %s got %s", acc.Login, account.Login)
	}
	t.Log("func [GetAccount ByLogin] is OK")
}

func TestAuthService_GetAccount_NotFound(t *testing.T) {
	service, _, _ := setup()

	code, account, err := service.GetAccount(ginCtx(), "", 999999)
	if err == nil {
		t.Fatal("expected error")
	}
	if code != http.StatusNotFound && code != http.StatusInternalServerError {
		t.Fatalf("unexpected status code %d", code)
	}
	if account != nil {
		t.Fatal("account should be nil")
	}
	t.Log("func [GetAccount NotFound] is OK")
}

func TestAuthService_GetAccount_WrongData(t *testing.T) {
	service, _, _ := setup()

	code, account, err := service.GetAccount(ginCtx(), "", -1)
	if err == nil {
		t.Fatal("expected error")
	}
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", code)
	}
	if account != nil {
		t.Fatal("account should be nil")
	}
	t.Log("func [GetAccount WrongData] is OK")
}

func TestAuthService_GetAccount_Deleted(t *testing.T) {
	service, db, _ := setup()
	acc := createTestAccount(t, service)

	acc.Deleted = true
	if err := db.PutAccount(*acc); err != nil {
		t.Fatal(err)
	}

	code, account, err := service.GetAccount(ginCtx(), "", acc.Id)
	if err == nil {
		t.Fatal("expected error")
	}
	if code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", code)
	}
	if account != nil {
		t.Fatal("account should be nil")
	}
	t.Log("func [GetAccount Deleted] is OK")
}

func TestAuthService_GetAccount_NotActive(t *testing.T) {
	service, db, _ := setup()
	acc := createTestAccount(t, service)

	acc.Active = false
	if err := db.PutAccount(*acc); err != nil {
		t.Fatal(err)
	}

	code, account, err := service.GetAccount(ginCtx(), "", acc.Id)
	if err == nil {
		t.Fatal("expected error")
	}
	if code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", code)
	}
	if account != nil {
		t.Fatal("account should be nil")
	}
	t.Log("func [GetAccount NotActive] is OK")
}

// --- DeleteAccount ---

func TestAuthService_DeleteAccount_Success(t *testing.T) {
	service, _, _ := setup()
	acc := createTestAccount(t, service)

	code, err := service.DeleteAccount(ginCtx(), acc.Id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != http.StatusOK {
		t.Fatalf("expected 200 got %d", code)
	}

	updated, errDb := service.Storage.GetAccount(acc.Id)
	if errDb != nil {
		t.Fatal(errDb)
	}
	if !updated.Deleted {
		t.Fatal("account should be deleted")
	}
	t.Log("func [DeleteAccount Success] is OK")
}

func TestAuthService_DeleteAccount_NotFound(t *testing.T) {
	service, _, _ := setup()

	code, err := service.DeleteAccount(ginCtx(), 999999)
	if err == nil {
		t.Fatal("expected error")
	}
	if code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", code)
	}
	t.Log("func [DeleteAccount NotFound] is OK")
}

// --- RequestOfficial ---

func TestAuthService_RequestOfficial(t *testing.T) {
	service, _, _ := setup()
	acc := createTestAccount(t, service)

	err := service.RequestOfficial(ginCtx(), acc.Id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Log("func [RequestOfficial] is OK")
}

// --- Subscribe / Unsubscribe ---

func TestAuthService_SubscribeUnsubscribe(t *testing.T) {
	service, _, _ := setup()

	account := createTestAccount(t, service)
	follower := createTestAccount(t, service)

	// Subscribe
	err := service.Subscribe(ginCtx(), account.Id, follower.Id)
	if err != nil {
		t.Fatalf("subscribe error: %v", err)
	}

	updated, errDb := service.Storage.GetAccount(account.Id)
	if errDb != nil {
		t.Fatal(errDb)
	}
	if updated.Followers != account.Followers+1 {
		t.Fatalf("followers не увеличился: got %d, want %d", updated.Followers, account.Followers+1)
	}
	t.Log("func [Subscribe] is OK")

	// Unsubscribe
	err = service.Unsubscribe(ginCtx(), account.Id, follower.Id)
	if err != nil {
		t.Fatalf("unsubscribe error: %v", err)
	}

	updated, errDb = service.Storage.GetAccount(account.Id)
	if errDb != nil {
		t.Fatal(errDb)
	}
	if updated.Followers != account.Followers {
		t.Fatalf("followers не уменьшился: got %d, want %d", updated.Followers, account.Followers)
	}
	t.Log("func [Unsubscribe] is OK")
}

// --- ChangePassword ---

func TestAuthService_ChangePassword_Success(t *testing.T) {
	service, _, _ := setup()
	acc := createTestAccount(t, service)

	// createTestAccount создаёт с паролем "password123", но Registration хэширует его
	// поэтому меняем через ChangePasword зная исходный пароль
	err := service.ChangePasword(ginCtx(), acc.Id, "password123", "newpass456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Log("func [ChangePassword Success] is OK")
}

func TestAuthService_ChangePassword_WrongOldPassword(t *testing.T) {
	service, _, _ := setup()
	acc := createTestAccount(t, service)

	err := service.ChangePasword(ginCtx(), acc.Id, "wrongpass", "newpass456")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "Old password incorrect" {
		t.Fatalf("wrong error: got %v", err)
	}
	t.Log("func [ChangePassword WrongOldPassword] is OK")
}

func TestAuthService_ChangePassword_AccountNotFound(t *testing.T) {
	service, _, _ := setup()

	err := service.ChangePasword(ginCtx(), 999999, "any", "newpass")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "internal server error" {
		t.Fatalf("wrong error: got %v", err)
	}
	t.Log("func [ChangePassword AccountNotFound] is OK")
}
