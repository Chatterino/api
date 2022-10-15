package cache

import (
	"context"
	"net/http"
	"time"
)

type Response struct {
	Payload     []byte
	StatusCode  int
	ContentType string
}

type Cache interface {
	Get(ctx context.Context, key string, r *http.Request) (*Response, error)

	// GetOnly returns the cached value, and doesn't try to load it if it doesn't exist
	GetOnly(ctx context.Context, key string) *Response

	RegisterDependent(ctx context.Context, dependent DependentCache)

	// Commits dependent values that belong to the key
	commitDependents(ctx context.Context, key string) error

	// Rolls back uncommmited dependent values that belong to the key
	rollbackDependents(ctx context.Context, key string) error
}

type Loader interface {
	Load(ctx context.Context, key string, r *http.Request) ([]byte, *int, *string, time.Duration, error)
}

type DependentCache interface {
	// Returns (value, content type, error)
	Get(ctx context.Context, key string) ([]byte, string, error)

	Insert(ctx context.Context, key string, parentKey string, value []byte, contentType string) error

	commit(ctx context.Context, parentKey string) error
	rollback(ctx context.Context, parentKey string) error
}

var NoSpecialDur time.Duration

var NewDefaultCache = NewPostgreSQLCache

var (
	defaultStatusCode  int    = 200
	defaultContentType string = "application/json"
)
