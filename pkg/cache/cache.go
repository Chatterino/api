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
}

type Loader interface {
	Load(ctx context.Context, key string, r *http.Request) ([]byte, *int, *string, time.Duration, error)
}

var NoSpecialDur time.Duration

var NewDefaultCache = NewPostgreSQLCache

var defaultStatusCode int = 200
var defaultContentType string = "application/json"
