package lib

import (
	"context"
	"fmt"
	"time"

	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/errors"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

// RedisCache implements Cache interface using Redis
type RedisCache struct {
	cache  *cache.Cache
	client *redis.Client
	prefix string
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(config Config, logger Logger) *RedisCache {
	addr := config.Cache.Addr()

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       constants.RedisMainDB,
		Password: config.Cache.Password,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		logger.Zap.Fatalf("Failed to connect to Redis[%s]: %v", addr, err)
	}

	logger.Zap.Info("Redis cache connection established")
	return &RedisCache{
		client: client,
		prefix: config.Cache.KeyPrefix,
		cache: cache.New(&cache.Options{
			Redis:      client,
			LocalCache: cache.NewTinyLFU(1000, time.Minute),
		}),
	}
}

func (r *RedisCache) wrapperKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", r.prefix, key)
}

// Set stores a value with expiration
func (r *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	return r.cache.Set(&cache.Item{
		Ctx:            context.TODO(),
		Key:            r.wrapperKey(key),
		Value:          value,
		TTL:            expiration,
		SkipLocalCache: true,
	})
}

// Get retrieves a value by key
func (r *RedisCache) Get(key string, value interface{}) error {
	err := r.cache.Get(context.TODO(), r.wrapperKey(key), value)
	if err == cache.ErrCacheMiss {
		return errors.RedisKeyNoExist
	}
	return err
}

// Delete removes keys from cache
func (r *RedisCache) Delete(keys ...string) (bool, error) {
	wrapperKeys := make([]string, len(keys))
	for i, key := range keys {
		wrapperKeys[i] = r.wrapperKey(key)
	}

	cmd := r.client.Del(context.TODO(), wrapperKeys...)
	if err := cmd.Err(); err != nil {
		return false, err
	}

	return cmd.Val() > 0, nil
}

// Check verifies if keys exist
func (r *RedisCache) Check(keys ...string) (bool, error) {
	wrapperKeys := make([]string, len(keys))
	for i, key := range keys {
		wrapperKeys[i] = r.wrapperKey(key)
	}

	cmd := r.client.Exists(context.TODO(), wrapperKeys...)
	if err := cmd.Err(); err != nil {
		return false, err
	}
	return cmd.Val() > 0, nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// HSet sets a hash field
func (r *RedisCache) HSet(key, field string, value interface{}) error {
	return r.client.HSet(context.TODO(), r.wrapperKey(key), field, value).Err()
}

// HGet gets a hash field
func (r *RedisCache) HGet(key, field string, value interface{}) error {
	result, err := r.client.HGet(context.TODO(), r.wrapperKey(key), field).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.RedisKeyNoExist
		}
		return err
	}

	// For string values, directly assign
	if v, ok := value.(*string); ok {
		*v = result
		return nil
	}

	return nil
}

// HMSet sets multiple hash fields
func (r *RedisCache) HMSet(key string, values map[string]interface{}) error {
	return r.client.HMSet(context.TODO(), r.wrapperKey(key), values).Err()
}

// HDel deletes hash fields
func (r *RedisCache) HDel(key string, fields ...string) error {
	return r.client.HDel(context.TODO(), r.wrapperKey(key), fields...).Err()
}

// GetClient returns the underlying Redis client (for advanced operations)
func (r *RedisCache) GetClient() *redis.Client {
	return r.client
}
