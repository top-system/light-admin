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

type Redis struct {
	cache  *cache.Cache
	client *redis.Client
	prefix string
}

// NewRedis creates a new redis client instance
func NewRedis(config Config, logger Logger) Redis {
	addr := config.Redis.Addr()

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       constants.RedisMainDB,
		Password: config.Redis.Password,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		logger.Zap.Fatalf("Error to open redis[%s] connection: %v", addr, err)
	}

	logger.Zap.Info("Redis connection established")
	return Redis{
		client: client,
		prefix: config.Redis.KeyPrefix,
		cache: cache.New(&cache.Options{
			Redis:      client,
			LocalCache: cache.NewTinyLFU(1000, time.Minute),
		}),
	}
}

func (a Redis) wrapperKey(key string) string {
	return fmt.Sprintf("%s:%s", a.prefix, key)
}

func (a Redis) Set(key string, value interface{}, expiration time.Duration) error {
	return a.cache.Set(&cache.Item{
		Ctx:            context.TODO(),
		Key:            a.wrapperKey(key),
		Value:          value,
		TTL:            expiration,
		SkipLocalCache: true,
	})
}

func (a Redis) Get(key string, value interface{}) error {
	err := a.cache.Get(context.TODO(), a.wrapperKey(key), value)
	if err == cache.ErrCacheMiss {
		err = errors.RedisKeyNoExist
	}

	return err
}

func (a Redis) Delete(keys ...string) (bool, error) {
	wrapperKeys := make([]string, len(keys))
	for index, key := range keys {
		wrapperKeys[index] = a.wrapperKey(key)
	}

	cmd := a.client.Del(context.TODO(), wrapperKeys...)
	if err := cmd.Err(); err != nil {
		return false, err
	}

	return cmd.Val() > 0, nil
}

func (a Redis) Check(keys ...string) (bool, error) {
	wrapperKeys := make([]string, len(keys))
	for index, key := range keys {
		wrapperKeys[index] = a.wrapperKey(key)
	}

	cmd := a.client.Exists(context.TODO(), wrapperKeys...)
	if err := cmd.Err(); err != nil {
		return false, err
	}
	return cmd.Val() > 0, nil
}

func (a Redis) Close() error {
	return a.client.Close()
}

func (a Redis) GetClient() *redis.Client {
	return a.client
}

// HSet sets a hash field
func (a Redis) HSet(key, field string, value interface{}) error {
	return a.client.HSet(context.TODO(), a.wrapperKey(key), field, value).Err()
}

// HGet gets a hash field
func (a Redis) HGet(key, field string, value interface{}) error {
	result, err := a.client.HGet(context.TODO(), a.wrapperKey(key), field).Result()
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
func (a Redis) HMSet(key string, values map[string]interface{}) error {
	return a.client.HMSet(context.TODO(), a.wrapperKey(key), values).Err()
}

// HDel deletes hash fields
func (a Redis) HDel(key string, fields ...string) error {
	return a.client.HDel(context.TODO(), a.wrapperKey(key), fields...).Err()
}
