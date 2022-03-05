package cache

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
	pCache "github.com/patrickmn/go-cache"
)

var kvCache *pCache.Cache

func init() {
	kvCache = pCache.New(30*time.Minute, 10*time.Minute)
}

type MemoryCache struct {
	loader Loader

	requestsMutex sync.Mutex
	requests      map[string][]chan []byte

	cacheDuration time.Duration

	prefix string
}

func (c *MemoryCache) load(ctx context.Context, key string, r *http.Request) {

	value, overrideDuration, err := c.loader.Load(ctx, key, r)

	var dur = c.cacheDuration
	if overrideDuration != 0 {
		dur = overrideDuration
	}

	// Cache it
	if err == nil {
		cacheKey := c.prefix + ":" + key
		kvCache.Set(cacheKey, value, dur)
	} else {
		fmt.Println("Error when some load function was called:", err)
	}

	c.requestsMutex.Lock()
	for _, ch := range c.requests[key] {
		ch <- value
	}
	delete(c.requests, key)
	c.requestsMutex.Unlock()
}

func (c *MemoryCache) Get(ctx context.Context, key string, r *http.Request) []byte {
	log := logger.FromContext(ctx)
	cacheKey := c.prefix + ":" + key

	// If key is in cache, return value
	if value, found := kvCache.Get(cacheKey); found && value != nil {
		log.Debugw("Memory Get cache hit", "prefix", c.prefix, "key", key)
		return value.([]byte)
	}

	responseChannel := make(chan []byte)

	c.requestsMutex.Lock()

	c.requests[key] = append(c.requests[key], responseChannel)

	first := len(c.requests[key]) == 1

	c.requestsMutex.Unlock()

	if first {
		log.Debugw("Memory Get cache miss", "prefix", c.prefix, "key", key)
		go c.load(ctx, key, r)
	}

	// If key is not in cache, sign up as a listener and ensure loader is only called once
	// Wait for loader to complete, then return value from loader
	return <-responseChannel
}

func (c *MemoryCache) GetOnly(ctx context.Context, key string) []byte {
	log := logger.FromContext(ctx)
	cacheKey := c.prefix + ":" + key

	if value, _ := kvCache.Get(cacheKey); value != nil {
		log.Debugw("Memory GetOnly cache hit", "prefix", c.prefix, "key", key)
		return value.([]byte)
	}

	log.Debugw("Memory GetOnly cache miss", "prefix", c.prefix, "key", key)
	return nil
}

func NewMemoryCache(cfg config.APIConfig, prefix string, loader Loader, cacheDuration time.Duration) *MemoryCache {
	return &MemoryCache{
		prefix:        prefix,
		loader:        loader,
		requests:      make(map[string][]chan []byte),
		cacheDuration: cacheDuration,
	}
}
