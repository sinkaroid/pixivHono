package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
}

var (
	GlobalCache  Cache
	ErrCacheMiss = errors.New("cache: miss")
)

type MemoryCacheItem struct {
	Value     []byte
	ExpiresAt time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]MemoryCacheItem
}

func NewMemoryCache() *MemoryCache {
	mc := &MemoryCache{
		items: make(map[string]MemoryCacheItem),
	}
	// Start clean ticker every 1 minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			mc.clean()
		}
	}()
	return mc
}

func (mc *MemoryCache) clean() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	now := time.Now()
	for k, item := range mc.items {
		if now.After(item.ExpiresAt) {
			delete(mc.items, k)
		}
	}
}

func (mc *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	item, ok := mc.items[key]
	if !ok || time.Now().After(item.ExpiresAt) {
		return nil, ErrCacheMiss
	}
	return item.Value, nil
}

func (mc *MemoryCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.items[key] = MemoryCacheItem{
		Value:     val,
		ExpiresAt: time.Now().Add(ttl),
	}
	return nil
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(url string) (*RedisCache, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	// Test ping
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}
	return &RedisCache{client: client}, nil
}

func (rc *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := rc.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}
	return val, nil
}

func (rc *RedisCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return rc.client.Set(ctx, key, val, ttl).Err()
}

func Init(redisURL string) {
	if redisURL != "" {
		rc, err := NewRedisCache(redisURL)
		if err == nil {
			GlobalCache = rc
			fmt.Println("[search-cache] connected to Redis")
			return
		}
		fmt.Printf("[search-cache] error connecting to Redis: %v. Falling back to in-memory.\n", err)
	}
	GlobalCache = NewMemoryCache()
	fmt.Println("[search-cache] using in-memory cache")
}

func GetSearchCacheKey(query string, page int) string {
	// Equivalent to Keyv namespace + search_key
	// Keyv namespace "pixiv-search" will prefix keys with "pixiv-search:"
	return fmt.Sprintf("pixiv-search:q:%s:p:%d", stringsToLower(stringsTrimSpace(query)), page)
}

func stringsToLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func stringsTrimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && isSpace(s[start]) {
		start++
	}
	for end > start && isSpace(s[end-1]) {
		end--
	}
	return s[start:end]
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
}
