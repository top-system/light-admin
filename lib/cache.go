package lib

import (
	"time"
)

// Cache defines the interface for cache operations
// Supports both Redis and Memory cache implementations
type Cache interface {
	// Set stores a value with expiration
	Set(key string, value interface{}, expiration time.Duration) error

	// Get retrieves a value by key
	Get(key string, value interface{}) error

	// Delete removes keys from cache
	Delete(keys ...string) (bool, error)

	// Check verifies if keys exist
	Check(keys ...string) (bool, error)

	// Close closes the cache connection
	Close() error

	// Hash operations
	HSet(key, field string, value interface{}) error
	HGet(key, field string, value interface{}) error
	HMSet(key string, values map[string]interface{}) error
	HDel(key string, fields ...string) error
}

// NewCache creates a cache instance based on configuration
// If type is "redis", returns RedisCache; otherwise returns MemoryCache
func NewCache(config Config, logger Logger) Cache {
	if config.Cache.IsRedis() {
		return NewRedisCache(config, logger)
	}
	return NewMemoryCache(config, logger)
}
