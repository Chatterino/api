package main

import (
	"time"

	cache "github.com/patrickmn/go-cache"
)

var kvCache *cache.Cache

func init() {
	kvCache = cache.New(30*time.Minute, 10*time.Minute)
}
