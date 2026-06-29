package redis

import (
	"Etog/internal/config"
	"Etog/storage"
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisDb struct {
	rdb *redis.Client
	log *slog.Logger
}

func NewRedisDb(config config.RedisDb, log *slog.Logger) *RedisDb {
	if log == nil {
		log = slog.Default()
	}
	index, err := strconv.Atoi(config.Index)
	if err != nil {
		panic(err)
	}
	db := redis.NewClient(&redis.Options{
		Addr:     config.Host + ":" + config.Port,
		DB:       index,
		Username: config.User,
		Password: config.Password,
	})
	if err = db.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	slog.Info("Redis db was created successfully")
	return &RedisDb{rdb: db, log: log}
}

func (rd *RedisDb) Put(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	const op = "redis.Put"

	if key == "" {
		return errors.New("key is empty")
	}

	if err := rd.rdb.Set(ctx, key, value, exp).Err(); err != nil {
		rd.log.Error(op+": ", err)
		return errors.New("server error")
	}
	return nil
}

func (rd *RedisDb) Get(ctx context.Context, key string) (string, error) {
	const op = "redis.Get"

	if key == "" {
		return "", errors.New("key is empty")
	}

	value, err := rd.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", storage.ErrNotFound
	}
	if err != nil {
		rd.log.Error(op+": ", err)
		return "", errors.New("server error")
	}

	return value, nil
}

func (rd *RedisDb) Delete(ctx context.Context, key string) error {
	const op = "redis.Delete"

	if key == "" {
		return errors.New("key is empty")
	}

	if err := rd.rdb.Del(ctx, key).Err(); err != nil {
		rd.log.Error(op+": ", err)
		return errors.New("server error")
	}
	return nil
}
