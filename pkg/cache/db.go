package cache

import (
	"context"
	"net/http"
	"time"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	cacheHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "db_cache_hits_total",
			Help: "Number of DB cache hits",
		},
	)
	cacheMisses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "db_cache_misses_total",
			Help: "Number of DB cache misses",
		},
	)
	clearedEntries = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "db_cache_cleared_entries_total",
			Help: "Number of cache entries cleared",
		},
	)
)

func init() {
	prometheus.MustRegister(cacheHits)
	prometheus.MustRegister(cacheMisses)
	prometheus.MustRegister(clearedEntries)
}

type PostgreSQLCache struct {
	loader Loader

	cacheDuration time.Duration

	prefix string

	pool db.Pool
}

var (
	// TODO: Make the "internal error" tooltip an actual tooltip
	tooltipInternalError = []byte("internal error")
)

func clearOldTooltips(ctx context.Context, pool db.Pool) (pgconn.CommandTag, error) {
	const query = "DELETE FROM cache WHERE now() > cached_until;"
	return pool.Exec(ctx, query)
}

func startTooltipClearer(ctx context.Context, pool db.Pool) {
}

func (c *PostgreSQLCache) load(ctx context.Context, key string, r *http.Request) ([]byte, error) {
	log := logger.FromContext(ctx)

	valueBytes, overrideDuration, err := c.loader.Load(ctx, key, r)

	var dur = c.cacheDuration
	if overrideDuration != 0 {
		dur = overrideDuration
	}

	if err != nil {
		return nil, err
	}

	cacheKey := c.prefix + ":" + key
	if _, err := c.pool.Exec(ctx, "INSERT INTO cache (key, value, cached_until) VALUES ($1, $2, $3)", cacheKey, valueBytes, time.Now().Add(dur)); err != nil {
		log.Errorw("Error inserting tooltip into cache",
			"prefix", c.prefix,
			"key", key,
			"error", err,
		)
	}
	return valueBytes, nil
}

func (c *PostgreSQLCache) loadFromDatabase(ctx context.Context, cacheKey string) ([]byte, error) {
	var value []byte
	err := c.pool.QueryRow(ctx, "SELECT value FROM cache WHERE key=$1", cacheKey).Scan(&value)
	if err == nil {
		return value, nil
	}

	if err != pgx.ErrNoRows {
		return nil, err
	}

	return nil, nil
}

func (c *PostgreSQLCache) Get(ctx context.Context, key string, r *http.Request) ([]byte, error) {
	log := logger.FromContext(ctx)
	cacheKey := c.prefix + ":" + key

	value, err := c.loadFromDatabase(ctx, cacheKey)
	if err != nil {
		log.Warnw("Unhandled sql error", "error", err)
		return tooltipInternalError, err
	} else if value != nil {
		cacheHits.Inc()
		log.Debugw("DB Get cache hit", "prefix", c.prefix, "key", key)
		return value, nil
	}

	cacheMisses.Inc()
	log.Debugw("DB Get cache miss", "prefix", c.prefix, "key", key)
	return c.load(ctx, key, r)
}

func (c *PostgreSQLCache) GetOnly(ctx context.Context, key string) []byte {
	log := logger.FromContext(ctx)
	cacheKey := c.prefix + ":" + key

	value, err := c.loadFromDatabase(ctx, cacheKey)
	if err != nil {
		log.Warnw("Unhandled sql error", "error", err)
		return nil
	} else if value != nil {
		cacheHits.Inc()
		log.Debugw("DB GetOnly cache hit", "prefix", c.prefix, "key", key)
		return value
	}

	cacheMisses.Inc()
	log.Debugw("DB GetOnly cache miss", "prefix", c.prefix, "key", key)
	return nil
}

func StartCacheClearer(ctx context.Context, pool db.Pool) {
	log := logger.FromContext(ctx)

	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if ct, err := clearOldTooltips(ctx, pool); err != nil {
				log.Errorw("Error clearing old tooltips")
			} else {
				clearedEntries.Add(float64(ct.RowsAffected()))
				log.Debugw("Cleared old tooltips", "rowsAffected", ct.RowsAffected())
			}
		}
	}
}

func NewPostgreSQLCache(ctx context.Context, cfg config.APIConfig, pool db.Pool, prefix string, loader Loader, cacheDuration time.Duration) *PostgreSQLCache {
	// Create connection pool if it's not already initialized
	return &PostgreSQLCache{
		prefix:        prefix,
		loader:        loader,
		cacheDuration: cacheDuration,
		pool:          pool,
	}
}
