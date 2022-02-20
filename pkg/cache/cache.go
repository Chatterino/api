package cache

import (
	"net/http"
	"time"
)

type Cache interface {
	Get(key string, r *http.Request) interface{}
}

type Loader func(key string, r *http.Request) (interface{}, time.Duration, error)

var NoSpecialDur time.Duration

var New = NewMemoryCache
