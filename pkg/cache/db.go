package cache

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
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

type wrappedResponse struct {
	response *Response
	err      error
}

func init() {
	prometheus.MustRegister(cacheHits)
	prometheus.MustRegister(cacheMisses)
	prometheus.MustRegister(clearedEntries)
}

type PostgreSQLCache struct {
	loader Loader

	cacheDuration time.Duration

	keyProvider KeyProvider

	pool db.Pool

	dependentCaches []DependentCache

	requestsMutex sync.Mutex
	requests      map[string][]chan wrappedResponse
}

// TODO: Make the "internal error" tooltip an actual tooltip
var tooltipInternalError = []byte("internal error")

// Returns the number of deleted tooltip entries
func clearOldTooltips(ctx context.Context, pool db.Pool) (int, error) {
	log := logger.FromContext(ctx)

	const query = "DELETE FROM cache WHERE now() > cached_until RETURNING key;"

	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Errorw("Error deleting old tooltips from cache",
			"error", err,
		)
		return -1, err
	}

	// Remember the deleted keys: they may be parent keys of dependent values
	var deletedParents []string
	for rows.Next() {
		var curKey string
		err := rows.Scan(&curKey)
		if err != nil {
			log.Warnw("Could not scan a deleted key")
			continue
		}
		deletedParents = append(deletedParents, curKey)
	}

	_, err = pool.Exec(ctx, "DELETE FROM dependent_values WHERE parent_key = ANY($1)", deletedParents)
	if err != nil {
		log.Errorw("Error deleting dependent values",
			"err", err,
		)
		return -1, err
	}

	return len(deletedParents), nil
}

// This is meant as a safety measure in case a DependentCache user does not properly manage parent
// keys.
func clearExpiredDependentValues(ctx context.Context, pool db.Pool) error {
	log := logger.FromContext(ctx)

	_, err := pool.Exec(ctx, "DELETE FROM dependent_values WHERE now() > expiration_timestamp")
	if err != nil {
		log.Errorw("Error clearing expired dependent values",
			"err", err,
		)
		return err
	}

	return nil
}

func startTooltipClearer(ctx context.Context, pool db.Pool) {
}

func (c *PostgreSQLCache) load(ctx context.Context, key string, r *http.Request) (*Response, error) {
	log := logger.FromContext(ctx)

	payload, statusCode, contentType, overrideDuration, err := c.loader.Load(ctx, key, r)
	// If the parent cannot be inserted into the cache, rollback the dependents
	defer c.rollbackDependents(ctx, key)

	if statusCode == nil {
		log.Warnw("Missing status code, setting to 200 default")
		statusCode = &defaultStatusCode
	}
	if contentType == nil {
		log.Warnw("Missing content type, setting to application/json default")
		contentType = &defaultContentType
	}

	dur := c.cacheDuration
	if overrideDuration != 0 {
		dur = overrideDuration
	}

	if err != nil {
		return nil, err
	}

	cacheKey := c.keyProvider.CacheKey(ctx, key)
	if _, err := c.pool.Exec(ctx, "INSERT INTO cache (key, value, http_status_code, http_content_type, cached_until) VALUES ($1, $2, $3, $4, $5)", cacheKey, payload, *statusCode, *contentType, time.Now().Add(dur)); err != nil {
		log.Errorw("Error inserting tooltip into cache",
			"cacheKey", cacheKey,
			"key", key,
			"error", err,
		)
	}
	// Parent entry was inserted correctly, commit the dependents to prevent them from being rolled
	// back
	c.commitDependents(ctx, key)

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
	cacheKey := c.keyProvider.CacheKey(ctx, key)

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
		log.Debugw("DB Get cache hit", "cacheKey", cacheKey)
		return cacheResponse, nil
	}

	// If key is not in cache, sign up as a listener and ensure loader is only called once
	cacheMisses.Inc()
	log.Debugw("DB Get cache miss", "cacheKey", cacheKey)
	responseChannel := make(chan wrappedResponse)

	c.requestsMutex.Lock()

	c.requests[key] = append(c.requests[key], responseChannel)

	first := len(c.requests[key]) == 1

	c.requestsMutex.Unlock()

	if first {
		go func() {
			response, err := c.load(ctx, key, r)

			r := wrappedResponse{
				response,
				err,
			}
			c.requestsMutex.Lock()
			for _, ch := range c.requests[key] {
				ch <- r
			}
			delete(c.requests, key)
			c.requestsMutex.Unlock()
		}()
	}

	// Wait for loader to complete, then return value from loader
	response := <-responseChannel
	return response.response, response.err
}

func (c *PostgreSQLCache) GetOnly(ctx context.Context, key string) *Response {
	log := logger.FromContext(ctx)
	cacheKey := c.keyProvider.CacheKey(ctx, key)

	value, err := c.loadFromDatabase(ctx, cacheKey)
	if err != nil {
		log.Warnw("Unhandled sql error", "error", err)
		return nil
	} else if value != nil {
		cacheHits.Inc()
		log.Debugw("DB GetOnly cache hit", "cacheKey", cacheKey)
		return value
	}

	cacheMisses.Inc()
	log.Debugw("DB GetOnly cache miss", "cacheKey", cacheKey)
	return nil
}

func StartCacheClearer(ctx context.Context, pool db.Pool) {
	log := logger.FromContext(ctx)

	tooltipTicker := time.NewTicker(1 * time.Minute)
	dependentValuesTicker := time.NewTicker(12 * time.Hour)
	for {
		select {
		case <-ctx.Done():
			return

		case <-tooltipTicker.C:
			if numDeleted, err := clearOldTooltips(ctx, pool); err != nil {
				log.Errorw("Error clearing old tooltips")
			} else {
				clearedEntries.Add(float64(numDeleted))
				log.Debugw("Cleared old tooltips", "rowsAffected", numDeleted)
			}

		case <-dependentValuesTicker.C:
			if err := clearExpiredDependentValues(ctx, pool); err != nil {
				log.Errorw("Error clearing expired dependent values")
			}
		}
	}
}

func (c *PostgreSQLCache) RegisterDependent(ctx context.Context, dependent DependentCache) {
	c.dependentCaches = append(c.dependentCaches, dependent)
}

func (c *PostgreSQLCache) commitDependents(ctx context.Context, key string) error {
	parentKey := c.keyProvider.CacheKey(ctx, key)

	// XXX: If we knew whether all dependent caches were PostgreSQLDependentCaches, this could be
	//      optimized (since all commit queries will be the same). But there might be other
	//      DependentCache implementations and thus we must delegate to each registered one.
	for _, dependent := range c.dependentCaches {
		err := dependent.commit(ctx, parentKey)
		if err != nil {
			continue
		}
	}

	return nil
}

func (c *PostgreSQLCache) rollbackDependents(ctx context.Context, key string) error {
	parentKey := c.keyProvider.CacheKey(ctx, key)

	// XXX: If we knew whether all dependent caches were PostgreSQLDependentCaches, this could be
	//      optimized (since all rollback queries will be the same). But there might be other
	//      DependentCache implementations and thus we must delegate to each registered one.
	for _, dependent := range c.dependentCaches {
		err := dependent.rollback(ctx, parentKey)
		if err != nil {
			continue
		}
	}

	return nil
}

func NewPostgreSQLCache(ctx context.Context, cfg config.APIConfig, pool db.Pool, keyProvider KeyProvider, loader Loader, cacheDuration time.Duration) *PostgreSQLCache {
	// Create connection pool if it's not already initialized
	return &PostgreSQLCache{
		keyProvider:   keyProvider,
		loader:        loader,
		cacheDuration: cacheDuration,
		pool:          pool,
		requests:      make(map[string][]chan wrappedResponse),
	}
}

var _ Cache = (*PostgreSQLCache)(nil)

type PostgreSQLDependentCache struct {
	keyProvider KeyProvider

	pool db.Pool
}

// The time after which dependent values will get cleaned up regardless of whether the parent key
// exists or not. This is done as a fail-safe in case of improper parent key management.
var dependentExpirationDuration = 24 * time.Hour

func (c *PostgreSQLDependentCache) loadFromDatabase(ctx context.Context, cacheKey string) (*Response, error) {
	var response Response
	err := c.pool.QueryRow(
		ctx,
		"SELECT value, http_status_code, http_content_type FROM cache WHERE key=$1",
		cacheKey,
	).Scan(&response.Payload, &response.StatusCode, &response.ContentType)
	if err == nil {
		return &response, nil
	}

	if err != pgx.ErrNoRows {
		return nil, err
	}

	return nil, nil
}

func (c *PostgreSQLDependentCache) Get(ctx context.Context, key string) ([]byte, string, error) {
	log := logger.FromContext(ctx)

	cacheKey := c.keyProvider.CacheKey(ctx, key)
	var value []byte
	var contentType string

	err := c.pool.QueryRow(
		ctx,
		"SELECT value, http_content_type FROM dependent_values WHERE key=$1",
		cacheKey,
	).Scan(&value, &contentType)
	if err != nil {
		if err != pgx.ErrNoRows {
			// An actual error
			log.Warnw("Unhandled sql error", "error", err)
			return nil, "", err
		}

		// Cache entry didn't exist
		return nil, "", nil
	}

	return value, contentType, nil
}

func (c *PostgreSQLDependentCache) Insert(
	ctx context.Context, key string, parentKey string, value []byte, contentType string,
) error {
	log := logger.FromContext(ctx)

	cacheKey := c.keyProvider.CacheKey(ctx, key)
	if _, err := c.pool.Exec(
		ctx,
		"INSERT INTO dependent_values (key, parent_key, value, http_content_type, "+
			"expiration_timestamp) VALUES ($1, $2, $3, $4, $5)",
		cacheKey, parentKey, value, contentType, time.Now().Add(dependentExpirationDuration),
	); err != nil {
		log.Errorw("Error inserting dependent value",
			"cacheKey", cacheKey,
			"parentKey", parentKey,
			"error", err,
		)
		return err
	}

	return nil
}

func (c *PostgreSQLDependentCache) commit(ctx context.Context, parentKey string) error {
	log := logger.FromContext(ctx)

	_, err := c.pool.Exec(
		ctx,
		"UPDATE dependent_values SET committed = TRUE WHERE parent_key = $1 AND NOT committed",
		parentKey,
	)
	if err != nil {
		log.Errorw("Error committing dependent values",
			"parentKey", parentKey,
			"err", err,
		)
		return err
	}

	return nil
}

func (c *PostgreSQLDependentCache) rollback(ctx context.Context, parentKey string) error {
	log := logger.FromContext(ctx)

	_, err := c.pool.Exec(
		ctx,
		"DELETE FROM dependent_values WHERE parent_key = $1 AND NOT committed",
		parentKey,
	)
	if err != nil {
		log.Errorw("Error rolling back dependent values",
			"err", err,
		)
		return err
	}

	return nil
}

func NewPostgreSQLDependentCache(
	ctx context.Context, cfg config.APIConfig, pool db.Pool, keyProvider KeyProvider,
) *PostgreSQLDependentCache {
	return &PostgreSQLDependentCache{
		keyProvider: keyProvider,
		pool:        pool,
	}
}

var _ DependentCache = (*PostgreSQLDependentCache)(nil)
