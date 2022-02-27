package cache

import (
	"context"
	"net/http"
	"time"

	"github.com/Chatterino/api/internal/migration"
	"github.com/Chatterino/api/pkg/config"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
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
}

var (
	// connection pool
	pool *pgxpool.Pool

	// TODO: Make the "internal error" tooltip an actual tooltip
	tooltipInternalError = []byte("internal error")
)

func clearOldTooltips(ctx context.Context) (pgconn.CommandTag, error) {
	const query = "DELETE FROM cache WHERE now() > cached_until;"
	return pool.Exec(ctx, query)
}

func startTooltipClearer(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if ct, err := clearOldTooltips(ctx); err != nil {
				log.Errorw("Error clearing old tooltips")
			} else {
				clearedEntries.Add(float64(ct.RowsAffected()))
				log.Debugw("Cleared old tooltips", "rowsAffected", ct.RowsAffected())
			}
		}
	}
}

func (c *PostgreSQLCache) load(key string, r *http.Request) ([]byte, error) {
	valueBytes, overrideDuration, err := c.loader(key, r)

	var dur = c.cacheDuration
	if overrideDuration != 0 {
		dur = overrideDuration
	}

	if err != nil {
		return nil, err
	}

	cacheKey := c.prefix + ":" + key
	if _, err := pool.Exec(context.Background(), "INSERT INTO cache (key, value, cached_until) VALUES ($1, $2, $3)", cacheKey, valueBytes, time.Now().Add(dur)); err != nil {
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
	err := pool.QueryRow(ctx, "SELECT value FROM cache WHERE key=$1", cacheKey).Scan(&value)
	if err == nil {
		return value, nil
	}

	if err != pgx.ErrNoRows {
		return nil, err
	}

	return nil, nil
}

func (c *PostgreSQLCache) Get(key string, r *http.Request) ([]byte, error) {
	cacheKey := c.prefix + ":" + key
	ctx := context.Background()

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
	return c.load(key, r)
}

func (c *PostgreSQLCache) GetOnly(key string) []byte {
	cacheKey := c.prefix + ":" + key
	ctx := context.Background()

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

func initPool(ctx context.Context, dsn string) {
	if pool != nil {
		// connection pool already initialized
		return
	}

	var err error

	log.Debugw("Initialize pool")
	pool, err = pgxpool.Connect(ctx, dsn)

	if err != nil {
		log.Fatalw("Error connecting to pool", "dsn", dsn, "error", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalw("Error pinging pool", "dsn", dsn, "error", err)
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalw("Error acquiring connection from pool",
			"dsn", dsn,
			"error", err,
		)
	}
	defer conn.Release()

	if oldVersion, newVersion, err := migration.Run(ctx, conn.Conn()); err != nil {
		log.Fatalw("Error running database migrations",
			"dsn", dsn,
			"error", err,
		)
	} else {
		log.Infow("Ran database migrations",
			"oldVersion", oldVersion,
			"newVersion", newVersion,
		)
	}

	go startTooltipClearer(ctx)

	// TODO: We currently don't close the connection pool
}

func NewPostgreSQLCache(cfg config.APIConfig, prefix string, loader Loader, cacheDuration time.Duration) *PostgreSQLCache {
	ctx := context.Background()
	initPool(ctx, cfg.DSN)

	// Create connection pool if it's not already initialized
	return &PostgreSQLCache{
		prefix:        prefix,
		loader:        loader,
		cacheDuration: cacheDuration,
	}
}
