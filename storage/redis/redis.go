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
	rdb  *redis.Client
	slog *slog.Logger
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
	slog.Info("Redis db was created successfully\n")
	return &RedisDb{rdb: db, slog: log}
}

func (rd *RedisDb) Put(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	if err := rd.rdb.Set(ctx, key, value, exp).Err(); err != nil {
		return err
	}
	return nil
}

func (rd *RedisDb) Get(ctx context.Context, key string) (interface{}, error) {
	value, err := rd.rdb.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return nil, storage.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return value, nil
}

func (rd *RedisDb) Delete(ctx context.Context, key string) error {
	if err := rd.rdb.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}
