package cache

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Chatterino/api/pkg/config"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// TODO: For hot cache, check out https://github.com/go-chi/stampede

type PostgreSQLCache struct {
	loader Loader

	requestsMutex sync.Mutex
	requests      map[string][]chan interface{}

	cacheDuration time.Duration

	prefix string
}

var (
	// connection pool
	pool *pgxpool.Pool

	// TODO: Make the "internal error" tooltip an actual tooltip
	tooltipInternalError = []byte("internal error")
)

func clearOldTooltips(ctx context.Context, conn *pgx.Conn) error {
	const query = "DELETE FROM tooltips WHERE now() > cached_until;"
	_, err := conn.Exec(ctx, query)
	return err
}

func startTooltipClearer() {
	go func() {

	}()
}

func (c *PostgreSQLCache) load(key string, r *http.Request) {
	fmt.Println("Load", key)
	value, overrideDuration, err := c.loader(key, r)

	var dur = c.cacheDuration
	if overrideDuration != 0 {
		dur = overrideDuration
	}

	// Cache it
	if err == nil {
		cacheKey := c.prefix + ":" + key
		_, err := pool.Query(context.Background(), "INSERT INTO tooltips (url, tooltip, cached_until) VALUES ($1, $2, $3)", cacheKey, value, time.Now().Add(dur))
		if err != nil {
			fmt.Println("Error inserting tooltip into cache:", err)
		}
		// kvCache.Set(cacheKey, value, dur)
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

func (c *PostgreSQLCache) Get(key string, r *http.Request) (value interface{}) {
	cacheKey := c.prefix + ":" + key
	ctx := context.Background()

	var tooltip string
	err := pool.QueryRow(ctx, "SELECT tooltip FROM tooltips WHERE url=$1", cacheKey).Scan(&tooltip)
	if err == nil {
		return []byte(tooltip)
	}

	if err != pgx.ErrNoRows {
		fmt.Println("Unhandled sql error:", err)
		return tooltipInternalError
	}

	// Tooltip for this key not found

	fmt.Println(tooltip)

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

func initPool(ctx context.Context, dsn string) {
	if pool != nil {
		// connection pool already initialized
		return
	}

	var err error

	pool, err = pgxpool.Connect(ctx, dsn)

	if err != nil {
		fmt.Println("Error connecting to pool:", err)
	}

	if err := pool.Ping(ctx); err != nil {
		fmt.Println("Error pinging pool:", err)
	}

	// TODO: We currently don't close the connection pool
}

func NewPostgreSQLCache(cfg config.APIConfig, prefix string, loader Loader, cacheDuration time.Duration) *PostgreSQLCache {
	ctx := context.Background()
	initPool(ctx, cfg.DSN)

	// Create connection pool if it's not already initialized
	return &PostgreSQLCache{
		prefix:        prefix,
		loader:        loader,
		requests:      make(map[string][]chan interface{}),
		cacheDuration: cacheDuration,
	}
}
