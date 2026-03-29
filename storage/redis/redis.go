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

type redisDb struct {
	rdb  *redis.Client
	slog *slog.Logger
}

func NewRedisDb(config config.RedisDb, slog *slog.Logger) *redisDb {
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
	return &redisDb{rdb: db, slog: slog}
}

func (rd *redisDb) Put(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	if err := rd.rdb.Set(ctx, key, value, exp).Err(); err != nil {
		return err
	}
	return nil
}

func (rd *redisDb) Get(ctx context.Context, key string) (interface{}, error) {
	value, err := rd.rdb.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return nil, storage.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return value, nil
}
