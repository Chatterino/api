package cache

import (
	"net/http"
	"time"

	"github.com/Chatterino/api/internal/logger"
)

var (
	log logger.Logger
)

func SetLogger(newLog logger.Logger) {
	log = newLog
}

type Cache interface {
	Get(key string, r *http.Request) ([]byte, error)

	// GetOnly returns the cached value, and doesn't try to load it if it doesn't exist
	GetOnly(key string) []byte
}

type Loader func(key string, r *http.Request) ([]byte, time.Duration, error)

var NoSpecialDur time.Duration

var NewDefaultCache = NewPostgreSQLCache
