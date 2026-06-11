package redis

import (
	"Etog/internal/config"
	"Etog/storage"
	"context"
	"testing"
	"time"
)

func setupRedis() *RedisDb {
	cfg := config.RedisDb{
		Host:     "localhost",
		Port:     "8091",
		Index:    "0",
		User:     "redisuser",
		Password: "0",
	}
	return NewRedisDb(cfg, nil)
}

func Test_NewRedisDb(t *testing.T) {
	rd := setupRedis()
	if rd == nil {
		t.Errorf("RedisDb is nil")
		return
	}
	t.Log("func [NewRedisDb] is OK")
}

func Test_PutGet(t *testing.T) {
	rd := setupRedis()
	if rd == nil {
		t.Errorf("RedisDb is nil")
		return
	}

	ctx := context.Background()
	key := "test_key"
	value := "test_value"

	err := rd.Put(ctx, key, value, 1*time.Minute)
	if err != nil {
		t.Errorf("ошибка Put: %v", err)
		return
	}

	got, err := rd.Get(ctx, key)
	if err != nil {
		t.Errorf("ошибка Get: %v", err)
		return
	}
	if got != value {
		t.Errorf("значение не совпадает: got %v, want %v", got, value)
		return
	}

	// Cleanup
	rd.Delete(ctx, key)
	t.Log("func [Put/Get] is OK")
}

func Test_GetNotFound(t *testing.T) {
	rd := setupRedis()
	if rd == nil {
		t.Errorf("RedisDb is nil")
		return
	}

	ctx := context.Background()

	_, err := rd.Get(ctx, "nonexistent_key")
	if err == nil {
		t.Errorf("ожидалась ошибка, но её нет")
		return
	}
	if err != storage.ErrNotFound {
		t.Errorf("ожидалась ErrNotFound, got: %v", err)
		return
	}
	t.Log("func [Get NotFound] is OK")
}

func Test_Delete(t *testing.T) {
	rd := setupRedis()
	if rd == nil {
		t.Errorf("RedisDb is nil")
		return
	}

	ctx := context.Background()
	key := "delete_test_key"

	err := rd.Put(ctx, key, "some_value", 1*time.Minute)
	if err != nil {
		t.Errorf("ошибка Put: %v", err)
		return
	}

	err = rd.Delete(ctx, key)
	if err != nil {
		t.Errorf("ошибка Delete: %v", err)
		return
	}

	_, err = rd.Get(ctx, key)
	if err == nil {
		t.Errorf("ключ должен быть удалён")
		return
	}
	if err != storage.ErrNotFound {
		t.Errorf("ожидалась ErrNotFound после удаления, got: %v", err)
		return
	}
	t.Log("func [Delete] is OK")
}

func Test_PutExpiration(t *testing.T) {
	rd := setupRedis()
	if rd == nil {
		t.Errorf("RedisDb is nil")
		return
	}

	ctx := context.Background()
	key := "expire_test_key"

	err := rd.Put(ctx, key, "expire_value", 1*time.Second)
	if err != nil {
		t.Errorf("ошибка Put: %v", err)
		return
	}

	// Ключ должен быть доступен сразу
	_, err = rd.Get(ctx, key)
	if err != nil {
		t.Errorf("ключ должен существовать: %v", err)
		return
	}

	// Ждём истечения TTL
	time.Sleep(2 * time.Second)

	_, err = rd.Get(ctx, key)
	if err == nil {
		t.Errorf("ключ должен был истечь")
		return
	}
	if err != storage.ErrNotFound {
		t.Errorf("ожидалась ErrNotFound после истечения TTL, got: %v", err)
		return
	}
	t.Log("func [Put Expiration] is OK")
}
