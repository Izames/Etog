package jwt

import (
	"Etog/storage/psql"
	"fmt"
	"testing"
	"time"

	jwt2 "github.com/golang-jwt/jwt/v5"
)

func setup() *Jwt {
	storage := psql.New("postgres://postgres:0@localhost:5433/etog?sslmode=disable", nil)
	return NewJwtLib("secret", storage)
}

func TestJwt_CreateAccessToken(t *testing.T) {
	jwt := setup()
	setUserId := 67

	token, err := jwt.CreateAccessToken(setUserId)

	if err != nil {
		t.Fatal(err)
	}

	keyFunc := func(_ *jwt2.Token) (interface{}, error) {
		return []byte(jwt.JwtKey), nil
	}

	getGoken, err := jwt2.Parse(token, keyFunc)
	if err != nil {
		t.Fatal(fmt.Errorf("error parsing token: %v", err))
	}

	if !getGoken.Valid {
		t.Fatal(fmt.Errorf("invalid token"))
	}

	claims, ok := getGoken.Claims.(jwt2.MapClaims)
	if !ok {
		t.Fatal(fmt.Errorf("invalid token"))
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		t.Fatal(fmt.Errorf("invalid token"))
	}
	userId, ok := claims["sub"].(float64)
	if !ok {
		t.Fatal(fmt.Errorf("invalid token"))
	}
	if setUserId != int(userId) {
		t.Fatal(fmt.Errorf("wrong user id get. want %d, got %f", setUserId, userId))
	}

	expTime := time.Unix(int64(expFloat), 0)

	// 2. Проверяем TTL (например 15 минут)
	expectedTTL := time.Hour
	actualTTL := time.Until(expTime)

	// допускаем небольшую погрешность (например 2 секунды)
	delta := time.Second * 2

	if actualTTL < expectedTTL-delta || actualTTL > expectedTTL+delta {
		t.Fatalf("wrong exp time. got %v, want ~%v", actualTTL, expectedTTL)
	}
}

func TestJwt_CreateRefreshToken(t *testing.T) {
	jwt := setup()
	setUserId := 67
	token, err := jwt.CreateRefreshToken(setUserId)
	if err != nil {
		t.Fatal(err)
	}
	keyFunc := func(_ *jwt2.Token) (interface{}, error) {
		return []byte(jwt.JwtKey), nil
	}
	getGoken, err := jwt2.Parse(token, keyFunc)
	if err != nil {
		t.Fatal(fmt.Errorf("error parsing token: %v", err))
	}
	if !getGoken.Valid {
		t.Fatal(fmt.Errorf("invalid token"))
	}
	claims, ok := getGoken.Claims.(jwt2.MapClaims)
	if !ok {
		t.Fatal(fmt.Errorf("invalid token"))
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		t.Fatal(fmt.Errorf("invalid token"))
	}
	expTime := time.Unix(int64(expFloat), 0)
	expectedTTL := time.Hour * 24 * 365
	actualTTL := time.Until(expTime)
	delta := time.Second * 2
	if actualTTL < expectedTTL-delta || actualTTL > expectedTTL+delta {
		t.Fatalf("wrong exp time. got %v, want ~%v", actualTTL, expectedTTL)
	}
	sub := claims["sub"].(float64)
	if int(sub) != setUserId {
		t.Fatal(fmt.Errorf("wrong user id get. want %d, got %f", setUserId, sub))
	}

}

func TestJwt_CheckAccessToken(t *testing.T) {
	jwt := setup()
	setUserId := 67
	token, err := jwt.CreateAccessToken(setUserId)
	if err != nil {
		t.Fatal(err)
	}
	if err = jwt.CheckAccessToken(token); err != nil {
		t.Fatal(err)
	}
}

func TestJwt_RefreshToken(t *testing.T) {
	jwt := setup()
	setUserId := 67
	token, err := jwt.CreateRefreshToken(setUserId)
	if err != nil {
		t.Fatal(err)
	}
	err, newToken := jwt.RefreshToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if err = jwt.CheckAccessToken(newToken); err != nil {
		t.Fatal(err)
	}
}

func TestJwt_DeleteRefreshToken(t *testing.T) {
	jwt := setup()
	setUserId := 67
	token, err := jwt.CreateRefreshToken(setUserId)
	if err != nil {
		t.Fatal(err)
	}
	err = jwt.DeleteRefreshToken(setUserId)
	if err != nil {
		t.Fatal(err)
	}
	err, _ = jwt.RefreshToken(token)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJwt_TokensGenerate(t *testing.T) {
	jwtLib := setup()
	setUserId := 67

	accessToken, refreshToken, err := jwtLib.TokensGenerate(setUserId)
	if err != nil {
		t.Fatal(err)
	}

	keyFunc := func(_ *jwt2.Token) (interface{}, error) {
		return []byte(jwtLib.JwtKey), nil
	}

	// --- ACCESS TOKEN ---
	accessParsed, err := jwt2.Parse(accessToken, keyFunc)
	if err != nil {
		t.Fatalf("error parsing access token: %v", err)
	}
	if !accessParsed.Valid {
		t.Fatal("access token invalid")
	}

	accessClaims, ok := accessParsed.Claims.(jwt2.MapClaims)
	if !ok {
		t.Fatal("invalid access claims")
	}

	subFloat, ok := accessClaims["sub"].(float64)
	if !ok {
		t.Fatal("invalid sub in access token")
	}

	if int(subFloat) != setUserId {
		t.Fatalf("wrong user id in access token. want %d, got %d", setUserId, int(subFloat))
	}

	// --- REFRESH TOKEN ---
	refreshParsed, err := jwt2.Parse(refreshToken, keyFunc)
	if err != nil {
		t.Fatalf("error parsing refresh token: %v", err)
	}
	if !refreshParsed.Valid {
		t.Fatal("refresh token invalid")
	}

	refreshClaims, ok := refreshParsed.Claims.(jwt2.MapClaims)
	if !ok {
		t.Fatal("invalid refresh claims")
	}

	// sub
	subFloat, ok = refreshClaims["sub"].(float64)
	if !ok {
		t.Fatal("invalid sub in refresh token")
	}

	if int(subFloat) != setUserId {
		t.Fatalf("wrong user id in refresh token. want %d, got %d", setUserId, int(subFloat))
	}

	// typ
	typ, ok := refreshClaims["typ"].(string)
	if !ok || typ != "refresh" {
		t.Fatal("invalid typ in refresh token")
	}

	// jti
	jti, ok := refreshClaims["jti"].(string)
	if !ok || jti == "" {
		t.Fatal("missing jti in refresh token")
	}
}
