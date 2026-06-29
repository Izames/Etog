package jwt

import (
	"Etog/storage/psql"
	"context"
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func setupJwt() (*Jwt, *psql.Storage) {
	storage := psql.New("postgres://postgres:0@localhost:5433/etog?sslmode=disable", nil)
	return NewJwtLib("secret", storage), storage
}

// --- CreateAccessToken ---

// success: токен создаётся и валиден
func TestJwt_CreateAccessToken(t *testing.T) {
	j, _ := setupJwt()

	token, err := j.CreateAccessToken(1)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if token == "" {
		t.Fatal("expected token, got empty string")
	}
	t.Log("func [CreateAccessToken] is OK")
}

// success: токен содержит верный sub
func TestJwt_CreateAccessToken_ClaimsValid(t *testing.T) {
	j, _ := setupJwt()

	token, err := j.CreateAccessToken(42)
	if err != nil {
		t.Fatal(err)
	}

	keyFunc := func(_ *jwt.Token) (interface{}, error) { return []byte("secret"), nil }
	parsed, err := jwt.Parse(token, keyFunc)
	if err != nil || !parsed.Valid {
		t.Fatalf("token invalid: %v", err)
	}

	claims := parsed.Claims.(jwt.MapClaims)
	sub, ok := claims["sub"].(float64)
	if !ok || int(sub) != 42 {
		t.Fatalf("expected sub=42, got: %v", claims["sub"])
	}
	t.Log("func [CreateAccessToken ClaimsValid] is OK")
}

// --- CreateRefreshToken ---

// success: refresh токен создаётся и содержит верные claims
func TestJwt_CreateRefreshToken(t *testing.T) {
	j, _ := setupJwt()
	ctx := context.Background()

	token, err := j.CreateRefreshToken(ctx, 1)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if token == "" {
		t.Fatal("expected token, got empty string")
	}

	keyFunc := func(_ *jwt.Token) (interface{}, error) { return []byte("secret"), nil }
	parsed, err := jwt.Parse(token, keyFunc)
	if err != nil || !parsed.Valid {
		t.Fatalf("token invalid: %v", err)
	}

	claims := parsed.Claims.(jwt.MapClaims)
	if claims["typ"] != tokenTypeRefresh {
		t.Fatalf("expected typ=refresh, got: %v", claims["typ"])
	}
	if claims["jti"] == "" {
		t.Fatal("expected jti, got empty")
	}
	t.Log("func [CreateRefreshToken] is OK")
}

// --- CheckAccessToken ---

// success: валидный токен проходит проверку
func TestJwt_CheckAccessToken(t *testing.T) {
	j, _ := setupJwt()

	token, err := j.CreateAccessToken(1)
	if err != nil {
		t.Fatal(err)
	}

	if err = j.CheckAccessToken(token); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	t.Log("func [CheckAccessToken] is OK")
}

// fail: поддельный токен не проходит проверку
func TestJwt_CheckAccessToken_Fake(t *testing.T) {
	j, _ := setupJwt()

	if err := j.CheckAccessToken("fake.token.string"); err == nil {
		t.Fatal("expected error, got nil")
	}
	t.Log("func [CheckAccessToken Fake] is OK")
}

// fail: токен подписан другим ключом
func TestJwt_CheckAccessToken_WrongKey(t *testing.T) {
	other := NewJwtLib("other-secret", nil)

	token, err := other.CreateAccessToken(1)
	if err != nil {
		t.Fatal(err)
	}

	j, _ := setupJwt()
	if err = j.CheckAccessToken(token); err == nil {
		t.Fatal("expected error, got nil")
	}
	t.Log("func [CheckAccessToken WrongKey] is OK")
}

// --- RefreshToken ---

// success: refresh токен обменивается на новый access токен
func TestJwt_RefreshToken(t *testing.T) {
	j, _ := setupJwt()
	ctx := context.Background()

	refreshToken, err := j.CreateRefreshToken(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	newToken, err := j.RefreshToken(ctx, refreshToken)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if newToken == "" {
		t.Fatal("expected new access token, got empty string")
	}
	t.Log("func [RefreshToken] is OK")
}

// fail: access токен нельзя использовать как refresh
func TestJwt_RefreshToken_AccessTokenRejected(t *testing.T) {
	j, _ := setupJwt()
	ctx := context.Background()

	accessToken, err := j.CreateAccessToken(1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = j.RefreshToken(ctx, accessToken)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidTokenType) {
		t.Fatalf("expected ErrInvalidTokenType, got: %v", err)
	}
	t.Log("func [RefreshToken AccessTokenRejected] is OK")
}

// fail: поддельный refresh токен отклоняется
func TestJwt_RefreshToken_Fake(t *testing.T) {
	j, _ := setupJwt()
	ctx := context.Background()

	_, err := j.RefreshToken(ctx, "fake.token.string")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	t.Log("func [RefreshToken Fake] is OK")
}

// --- TokensGenerate ---

// success: генерируются оба токена
func TestJwt_TokensGenerate(t *testing.T) {
	j, _ := setupJwt()
	ctx := context.Background()

	access, refresh, err := j.TokensGenerate(ctx, 1)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if access == "" {
		t.Fatal("expected access token, got empty string")
	}
	if refresh == "" {
		t.Fatal("expected refresh token, got empty string")
	}
	t.Log("func [TokensGenerate] is OK")
}

// success: access и refresh токены различаются по типу
func TestJwt_TokensGenerate_TypesDiffer(t *testing.T) {
	j, _ := setupJwt()
	ctx := context.Background()

	access, refresh, err := j.TokensGenerate(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	keyFunc := func(_ *jwt.Token) (interface{}, error) { return []byte("secret"), nil }

	parsedAccess, _ := jwt.Parse(access, keyFunc)
	accessClaims := parsedAccess.Claims.(jwt.MapClaims)
	if accessClaims["typ"] == tokenTypeRefresh {
		t.Fatal("access token should not have typ=refresh")
	}

	parsedRefresh, _ := jwt.Parse(refresh, keyFunc)
	refreshClaims := parsedRefresh.Claims.(jwt.MapClaims)
	if refreshClaims["typ"] != tokenTypeRefresh {
		t.Fatal("refresh token should have typ=refresh")
	}
	t.Log("func [TokensGenerate TypesDiffer] is OK")
}
