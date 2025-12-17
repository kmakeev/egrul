package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/egrul-system/services/api-gateway/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Cache общий интерфейс кэша
type Cache interface {
	// Get читает значение по ключу и декодирует в dest.
	// Возвращает found=false, если ключ отсутствует.
	Get(ctx context.Context, key string, dest interface{}) (found bool, err error)
	// Set записывает значение с TTL.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	// Delete удаляет ключ.
	Delete(ctx context.Context, key string) error
	// Close закрывает клиент.
	Close() error
}

// RedisCache реализация Cache на базе Redis.
type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
}

// DefaultTTL время жизни записей в кэше по умолчанию.
const DefaultTTL = 5 * time.Minute

// NewRedisCache создает новый Redis-кэш на основе конфигурации.
func NewRedisCache(cfg config.RedisConfig, logger *zap.Logger) *RedisCache {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &RedisCache{
		client: client,
		logger: logger.Named("redis_cache"),
	}
}

// Get реализует Cache.Get.
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	if c == nil || c.client == nil {
		return false, nil
	}

	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		c.logger.Warn("redis get failed", zap.String("key", key), zap.Error(err))
		return false, err
	}

	if err := json.Unmarshal(data, dest); err != nil {
		c.logger.Warn("redis unmarshal failed", zap.String("key", key), zap.Error(err))
		return false, err
	}

	return true, nil
}

// Set реализует Cache.Set.
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if ttl <= 0 {
		ttl = DefaultTTL
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		c.logger.Warn("redis set failed", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

// Delete реализует Cache.Delete.
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if c == nil || c.client == nil {
		return nil
	}
	if err := c.client.Del(ctx, key).Err(); err != nil {
		c.logger.Warn("redis delete failed", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

// Close закрывает соединение с Redis.
func (c *RedisCache) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}


