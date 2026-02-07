package lib

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/top-system/light-admin/errors"
)

// cacheItem represents a cached item with expiration
type cacheItem struct {
	Value      []byte
	Expiration int64 // Unix timestamp in nanoseconds, 0 means no expiration
}

// isExpired checks if the item has expired
func (item cacheItem) isExpired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

// hashItem represents a hash map stored in cache
type hashItem struct {
	Fields     map[string][]byte
	Expiration int64
}

// MemoryCache implements Cache interface using in-memory storage
type MemoryCache struct {
	items     sync.Map
	hashItems sync.Map
	hashMu    sync.Mutex // protects concurrent map read/write inside hashItems
	prefix    string
	logger    Logger
	stopCh    chan struct{}
}

// NewMemoryCache creates a new memory cache instance
func NewMemoryCache(config Config, logger Logger) *MemoryCache {
	mc := &MemoryCache{
		prefix: config.Cache.KeyPrefix,
		logger: logger,
		stopCh: make(chan struct{}),
	}

	// Start cleanup goroutine
	go mc.cleanupLoop()

	logger.Zap.Info("Memory cache initialized")
	return mc
}

// cleanupLoop periodically removes expired items
func (m *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopCh:
			return
		}
	}
}

// cleanup removes expired items from cache
func (m *MemoryCache) cleanup() {
	now := time.Now().UnixNano()

	// Cleanup regular items
	m.items.Range(func(key, value interface{}) bool {
		item := value.(cacheItem)
		if item.Expiration > 0 && now > item.Expiration {
			m.items.Delete(key)
		}
		return true
	})

	// Cleanup hash items
	m.hashItems.Range(func(key, value interface{}) bool {
		item := value.(hashItem)
		if item.Expiration > 0 && now > item.Expiration {
			m.hashItems.Delete(key)
		}
		return true
	})
}

func (m *MemoryCache) wrapperKey(key string) string {
	if m.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", m.prefix, key)
}

// Set stores a value with expiration
func (m *MemoryCache) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	var exp int64
	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}

	m.items.Store(m.wrapperKey(key), cacheItem{
		Value:      data,
		Expiration: exp,
	})

	return nil
}

// Get retrieves a value by key
func (m *MemoryCache) Get(key string, value interface{}) error {
	v, ok := m.items.Load(m.wrapperKey(key))
	if !ok {
		return errors.RedisKeyNoExist
	}

	item := v.(cacheItem)
	if item.isExpired() {
		m.items.Delete(m.wrapperKey(key))
		return errors.RedisKeyNoExist
	}

	return json.Unmarshal(item.Value, value)
}

// Delete removes keys from cache
func (m *MemoryCache) Delete(keys ...string) (bool, error) {
	deleted := false
	for _, key := range keys {
		wKey := m.wrapperKey(key)
		if _, ok := m.items.Load(wKey); ok {
			m.items.Delete(wKey)
			deleted = true
		}
		if _, ok := m.hashItems.Load(wKey); ok {
			m.hashItems.Delete(wKey)
			deleted = true
		}
	}
	return deleted, nil
}

// Check verifies if keys exist
func (m *MemoryCache) Check(keys ...string) (bool, error) {
	for _, key := range keys {
		wKey := m.wrapperKey(key)

		// Check regular items
		if v, ok := m.items.Load(wKey); ok {
			item := v.(cacheItem)
			if !item.isExpired() {
				return true, nil
			}
			m.items.Delete(wKey)
		}

		// Check hash items
		if v, ok := m.hashItems.Load(wKey); ok {
			item := v.(hashItem)
			if item.Expiration == 0 || time.Now().UnixNano() <= item.Expiration {
				return true, nil
			}
			m.hashItems.Delete(wKey)
		}
	}
	return false, nil
}

// Close stops the cleanup goroutine
func (m *MemoryCache) Close() error {
	close(m.stopCh)
	return nil
}

// HSet sets a hash field
func (m *MemoryCache) HSet(key, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	wKey := m.wrapperKey(key)

	m.hashMu.Lock()
	defer m.hashMu.Unlock()

	v, ok := m.hashItems.Load(wKey)
	var item hashItem
	if ok {
		old := v.(hashItem)
		// Deep copy Fields to avoid sharing the underlying map
		newFields := make(map[string][]byte, len(old.Fields)+1)
		for k, val := range old.Fields {
			newFields[k] = val
		}
		item = hashItem{Fields: newFields, Expiration: old.Expiration}
	} else {
		item = hashItem{Fields: make(map[string][]byte)}
	}
	item.Fields[field] = data

	m.hashItems.Store(wKey, item)
	return nil
}

// HGet gets a hash field
func (m *MemoryCache) HGet(key, field string, value interface{}) error {
	wKey := m.wrapperKey(key)

	m.hashMu.Lock()
	v, ok := m.hashItems.Load(wKey)
	if !ok {
		m.hashMu.Unlock()
		return errors.RedisKeyNoExist
	}

	item := v.(hashItem)
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		m.hashItems.Delete(wKey)
		m.hashMu.Unlock()
		return errors.RedisKeyNoExist
	}

	data, ok := item.Fields[field]
	m.hashMu.Unlock()

	if !ok {
		return errors.RedisKeyNoExist
	}

	// Handle string pointer specially
	if strPtr, ok := value.(*string); ok {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			// If unmarshal fails, try direct assignment
			*strPtr = string(data)
			return nil
		}
		*strPtr = str
		return nil
	}

	return json.Unmarshal(data, value)
}

// HMSet sets multiple hash fields
func (m *MemoryCache) HMSet(key string, values map[string]interface{}) error {
	wKey := m.wrapperKey(key)

	// Pre-marshal all values before taking the lock
	marshalled := make(map[string][]byte, len(values))
	for field, value := range values {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		marshalled[field] = data
	}

	m.hashMu.Lock()
	defer m.hashMu.Unlock()

	v, ok := m.hashItems.Load(wKey)
	var item hashItem
	if ok {
		old := v.(hashItem)
		// Deep copy Fields to avoid sharing the underlying map
		newFields := make(map[string][]byte, len(old.Fields)+len(marshalled))
		for k, val := range old.Fields {
			newFields[k] = val
		}
		item = hashItem{Fields: newFields, Expiration: old.Expiration}
	} else {
		item = hashItem{Fields: make(map[string][]byte, len(marshalled))}
	}

	for field, data := range marshalled {
		item.Fields[field] = data
	}

	m.hashItems.Store(wKey, item)
	return nil
}

// HDel deletes hash fields
func (m *MemoryCache) HDel(key string, fields ...string) error {
	wKey := m.wrapperKey(key)

	m.hashMu.Lock()
	defer m.hashMu.Unlock()

	v, ok := m.hashItems.Load(wKey)
	if !ok {
		return nil
	}

	old := v.(hashItem)
	// Deep copy Fields to avoid sharing the underlying map
	newFields := make(map[string][]byte, len(old.Fields))
	for k, val := range old.Fields {
		newFields[k] = val
	}

	for _, field := range fields {
		delete(newFields, field)
	}

	if len(newFields) == 0 {
		m.hashItems.Delete(wKey)
	} else {
		m.hashItems.Store(wKey, hashItem{Fields: newFields, Expiration: old.Expiration})
	}

	return nil
}
