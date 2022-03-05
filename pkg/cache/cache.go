package cache

import (
	"context"
	"net/http"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string, r *http.Request) ([]byte, error)

	// GetOnly returns the cached value, and doesn't try to load it if it doesn't exist
	GetOnly(ctx context.Context, key string) []byte
}

type Loader interface {
	Load(ctx context.Context, key string, r *http.Request) ([]byte, time.Duration, error)
}

// type Loader func(ctx context.Context, key string, r *http.Request) ([]byte, time.Duration, error)

var NoSpecialDur time.Duration

var NewDefaultCache = NewPostgreSQLCache
