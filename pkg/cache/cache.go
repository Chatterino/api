package cache

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	pCache "github.com/patrickmn/go-cache"
)

var kvCache *pCache.Cache

func init() {
	kvCache = pCache.New(30*time.Minute, 10*time.Minute)
}

type Loader func(key string, r *http.Request) (interface{}, time.Duration, error)

var NoSpecialDur time.Duration

type Cache struct {
	loader Loader

	requestsMutex sync.Mutex
	requests      map[string][]chan interface{}

	cacheDuration time.Duration

	prefix string
}

func (c *Cache) load(key string, r *http.Request) {
	value, overrideDuration, err := c.loader(key, r)

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

func (c *Cache) Get(key string, r *http.Request) (value interface{}) {
	var found bool
	cacheKey := c.prefix + ":" + key

	// If key is in cache, return value
	if value, found = kvCache.Get(cacheKey); found {
		return
	}

	responseChannel := make(chan interface{})

	c.requestsMutex.Lock()

	c.requests[key] = append(c.requests[key], responseChannel)

	first := len(c.requests[key]) == 1

	c.requestsMutex.Unlock()

	if first {
		go c.load(key, r)
	}

	value = <-responseChannel

	// If key is not in cache, sign up as a listener and ensure loader is only called once
	// Wait for loader to complete, then return value from loader
	return
}

func New(prefix string, loader Loader, cacheDuration time.Duration) *Cache {
	return &Cache{
		prefix:        prefix,
		loader:        loader,
		requests:      make(map[string][]chan interface{}),
		cacheDuration: cacheDuration,
	}
}
