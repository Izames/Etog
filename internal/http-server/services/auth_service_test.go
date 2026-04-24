package services

import (
	"Etog/internal/config"
	"Etog/internal/domain/entity"
	"Etog/storage/psql"
	"Etog/storage/redis"
	"fmt"
	"testing"
)

func TestAuthService_Registration(t *testing.T) {
	storage := psql.New("postgres://postgres:0@localhost:5433/etog?sslmode=disable", nil)
	rStorage := redis.NewRedisDb(config.RedisDb{
		Host:     "localhost",
		Port:     "8091",
		Password: "0",
		User:     "redisuser",
		Index:    "0",
	}, nil)
	Authservice := NewAuthService(storage, rStorage, nil)
	err, code := Authservice.Registration(&entity.Account{
		Mail:        "some@mail.ru",
		Login:       "12345678",
		Password:    "eweewewe",
		Description: "my macdac account",
	})
	if err != nil {
		t.Fatal(err)
	}
	if code != 200 {
		t.Fatal(fmt.Sprintf("wanted 200 got %d", code))
	}
}

func TestAuthService_Registration_ExpectFailOnSameEmail(t *testing.T) {
	storage := psql.New("postgres://postgres:0@localhost:5433/etog?sslmode=disable", nil)
	rStorage := redis.NewRedisDb(config.RedisDb{
		Host:     "localhost",
		Port:     "8091",
		Password: "0",
		User:     "redisuser",
		Index:    "0",
	}, nil)
	Authservice := NewAuthService(storage, rStorage, nil)
	err, code := Authservice.Registration(&entity.Account{
		Mail:        "some@mail.ru",
		Login:       "123456",
		Password:    "eweewewe",
		Description: "my macdac account",
	})
	if err == nil || code != 409 {
		t.Fatal(fmt.Sprintf("wanted error but got %d", code))
		return
	}

}

func TestAuthService_Registration_FailOnShortPass(t *testing.T) {
	storage := psql.New("postgres://postgres:0@localhost:5433/etog?sslmode=disable", nil)
	rStorage := redis.NewRedisDb(config.RedisDb{
		Host:     "localhost",
		Port:     "8091",
		Password: "0",
		User:     "redisuser",
		Index:    "0",
	}, nil)
	Authservice := NewAuthService(storage, rStorage, nil)
	err, code := Authservice.Registration(&entity.Account{
		Mail:        "some",
		Login:       "123456",
		Password:    "ewe",
		Description: "my macdac account",
	})
	if err == nil || code != 400 {
		t.Fatal("wanted 400 got ", code)
	}
}

func TestAuthService_Registration_FailOnWrongMail(t *testing.T) {
	storage := psql.New("postgres://postgres:0@localhost:5433/etog?sslmode=disable", nil)
	rStorage := redis.NewRedisDb(config.RedisDb{
		Host:     "localhost",
		Port:     "8091",
		Password: "0",
		User:     "redisuser",
		Index:    "0",
	}, nil)
	Authservice := NewAuthService(storage, rStorage, nil)
	err, code := Authservice.Registration(&entity.Account{
		Mail:        "some",
		Login:       "123456",
		Password:    "ewsdfdsfe",
		Description: "my macdac account",
	})
	if err == nil || code != 400 {
		t.Fatal("wanted 400 got ", code)
	}
}
