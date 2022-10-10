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

func (c *PostgreSQLCache) load(ctx context.Context, key string, r *http.Request) (*Response, error) {
	log := logger.FromContext(ctx)

	payload, statusCode, contentType, overrideDuration, err := c.loader.Load(ctx, key, r)

	if statusCode == nil {
		log.Warnw("Missing status code, setting to 200 default")
		statusCode = &defaultStatusCode
	}
	if contentType == nil {
		log.Warnw("Missing content type, setting to application/json default")
		contentType = &defaultContentType
	}

	var dur = c.cacheDuration
	if overrideDuration != 0 {
		dur = overrideDuration
	}

	if err != nil {
		return nil, err
	}

	cacheKey := c.prefix + ":" + key
	if _, err := c.pool.Exec(ctx, "INSERT INTO cache (key, value, http_status_code, http_content_type, cached_until) VALUES ($1, $2, $3, $4, $5)", cacheKey, payload, *statusCode, *contentType, time.Now().Add(dur)); err != nil {
		log.Errorw("Error inserting tooltip into cache",
			"prefix", c.prefix,
			"key", key,
			"error", err,
		)
	}

	return &Response{
		Payload:     payload,
		StatusCode:  *statusCode,
		ContentType: *contentType,
	}, nil
}

func (c *PostgreSQLCache) loadFromDatabase(ctx context.Context, cacheKey string) (*Response, error) {
	var response Response
	err := c.pool.QueryRow(ctx, "SELECT value, http_status_code, http_content_type FROM cache WHERE key=$1", cacheKey).Scan(&response.Payload, &response.StatusCode, &response.ContentType)
	if err == nil {
		return &response, nil
	}

	if err != pgx.ErrNoRows {
		return nil, err
	}

	return nil, nil
}

func (c *PostgreSQLCache) Get(ctx context.Context, key string, r *http.Request) (*Response, error) {
	log := logger.FromContext(ctx)
	cacheKey := c.prefix + ":" + key

	cacheResponse, err := c.loadFromDatabase(ctx, cacheKey)
	if err != nil {
		log.Warnw("Unhandled sql error", "error", err)
		tooltipInternalError := Response{
			Payload:     []byte(`{"status":500,"message":"Internal server error (PSQL) loading thumbnail"}`),
			StatusCode:  500,
			ContentType: "application/json",
		}
		return &tooltipInternalError, err
	} else if cacheResponse != nil {
		cacheHits.Inc()
		log.Debugw("DB Get cache hit", "prefix", c.prefix, "key", key)
		return cacheResponse, nil
	}

	cacheMisses.Inc()
	log.Debugw("DB Get cache miss", "prefix", c.prefix, "key", key)
	return c.load(ctx, key, r)
}

func (c *PostgreSQLCache) GetOnly(ctx context.Context, key string) *Response {
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

var _ Cache = (*PostgreSQLCache)(nil)
