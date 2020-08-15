package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type loader func(key string, r *http.Request) (interface{}, error, time.Duration)

var noSpecialDur time.Duration

type loadingCache struct {
	loader loader

	requestsMutex sync.Mutex
	requests      map[string][]chan interface{}

	cacheDuration time.Duration

	prefix string
}

func (c *loadingCache) load(key string, r *http.Request) {
	value, err, overrideDuration := c.loader(key, r)

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

func (c *loadingCache) Get(key string, r *http.Request) (value interface{}) {
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

func newLoadingCache(prefix string, loader loader, cacheDuration time.Duration) *loadingCache {
	return &loadingCache{
		prefix:        prefix,
		loader:        loader,
		requests:      make(map[string][]chan interface{}),
		cacheDuration: cacheDuration,
	}
}
