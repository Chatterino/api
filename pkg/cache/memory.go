package cache

import (
	"context"
	"errors"
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
	requests      map[string][]chan Response

	cacheDuration time.Duration

	keyProvider KeyProvider
}

func (*MemoryCache) RegisterDependent(ctx context.Context, dependent DependentCache) {
	// Required for Cache interface
	panic("DependentCache management is unimplemented for MemoryCache")
}

func (*MemoryCache) commitDependents(ctx context.Context, key string) error {
	// Required for Cache interface
	panic("DependentCache management is unimplemented for MemoryCache")
}

func (*MemoryCache) rollbackDependents(ctx context.Context, key string) error {
	// Required for Cache interface
	panic("DependentCache management is unimplemented for MemoryCache")
}

func (c *MemoryCache) load(ctx context.Context, key string, r *http.Request) {
	log := logger.FromContext(ctx)

	payload, statusCode, contentType, overrideDuration, err := c.loader.Load(ctx, key, r)

	if statusCode == nil {
		log.Warnw("Missing status code, setting to 200 default")
		statusCode = &defaultStatusCode
	}
	if contentType == nil {
		log.Warnw("Missing content type, setting to application/json default")
		contentType = &defaultContentType
	}

	dur := c.cacheDuration
	if overrideDuration != 0 {
		dur = overrideDuration
	}

	response := Response{
		Payload:     payload,
		StatusCode:  *statusCode,
		ContentType: *contentType,
	}

	// Cache it
	if err == nil {
		cacheKey := c.keyProvider.CacheKey(ctx, key)
		kvCache.Set(cacheKey, response, dur)
	} else {
		fmt.Println("Error when some load function was called:", err)
	}

	c.requestsMutex.Lock()
	for _, ch := range c.requests[key] {
		ch <- response
	}
	delete(c.requests, key)
	c.requestsMutex.Unlock()
}

func (c *MemoryCache) Get(ctx context.Context, key string, r *http.Request) (*Response, error) {
	log := logger.FromContext(ctx)
	cacheKey := c.keyProvider.CacheKey(ctx, key)

	// If key is in cache, return value
	if value, found := kvCache.Get(cacheKey); found && value != nil {
		log.Debugw("Memory Get cache hit", "cacheKey", cacheKey)
		if response, ok := value.(Response); ok {
			return &response, nil
		}

		return nil, errors.New("error getting stuff from kvcache")
	}

	responseChannel := make(chan Response)

	c.requestsMutex.Lock()

	c.requests[key] = append(c.requests[key], responseChannel)

	first := len(c.requests[key]) == 1

	c.requestsMutex.Unlock()

	if first {
		log.Debugw("Memory Get cache miss", "cacheKey", cacheKey)
		go c.load(ctx, key, r)
	}

	// If key is not in cache, sign up as a listener and ensure loader is only called once
	// Wait for loader to complete, then return value from loader
	response := <-responseChannel
	return &response, nil
}

func (c *MemoryCache) GetOnly(ctx context.Context, key string) *Response {
	log := logger.FromContext(ctx)
	cacheKey := c.keyProvider.CacheKey(ctx, key)

	if value, _ := kvCache.Get(cacheKey); value != nil {
		log.Debugw("Memory GetOnly cache hit", "cacheKey", cacheKey)
		if response, ok := value.(Response); ok {
			return &response
		}

		log.Debugw("Memory GetOnly cache type mismatch", "cacheKey", cacheKey)
		return nil
	}

	log.Debugw("Memory GetOnly cache miss", "cacheKey", cacheKey)
	return nil
}

func NewMemoryCache(cfg config.APIConfig, keyProvider KeyProvider, loader Loader, cacheDuration time.Duration) *MemoryCache {
	return &MemoryCache{
		keyProvider:   keyProvider,
		loader:        loader,
		requests:      make(map[string][]chan Response),
		cacheDuration: cacheDuration,
	}
}

var _ Cache = (*MemoryCache)(nil)
