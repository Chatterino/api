package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	chiprometheus "github.com/766b/chi-prometheus"
	"github.com/Chatterino/api/internal/caches/twitchusernamecache"
	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/internal/migration"
	defaultresolver "github.com/Chatterino/api/internal/resolvers/default"
	"github.com/Chatterino/api/internal/routes/twitchemotes"
	"github.com/Chatterino/api/internal/twitchapiclient"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/thumbnail"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var (
	startTime = time.Now()

	prefix string
)

func mountRouter(r *chi.Mux, cfg config.APIConfig, log logger.Logger) *chi.Mux {
	if cfg.BaseURL == "" {
		log.Debugw("Listening", "host", cfg.BindAddress, "prefix", prefix, "baseURL", cfg.BaseURL)
		return r
	}

	// figure out prefix from address
	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		log.Fatalw("Unable to parse base URL", "baseURL", cfg.BaseURL, "error", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		log.Fatal("Scheme must be included in base url")
	}

	prefix = u.Path
	ur := chi.NewRouter()
	ur.Mount(prefix, r)

	log.Debugw("Listening", "host", cfg.BindAddress, "prefix", prefix, "baseURL", cfg.BaseURL)

	return ur
}

func listen(ctx context.Context, bind string, router *chi.Mux, log logger.Logger) {
	srv := &http.Server{
		Handler:      router,
		Addr:         bind,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	log.Fatal(srv.ListenAndServe())
}

func runMigrations(ctx context.Context, pool db.Pool) {
	log := logger.FromContext(ctx)

	if oldVersion, newVersion, err := migration.Run(ctx, pool); err != nil {
		log.Fatalw("Error running database migrations",
			"error", err,
		)
	} else {
		if newVersion != oldVersion {
			log.Infow("Ran database migrations",
				"oldVersion", oldVersion,
				"newVersion", newVersion,
			)
		}
	}
}

func main() {
	cfg := config.New()

	var atomicLogLevel zap.AtomicLevel
	var err error

	if atomicLogLevel, err = zap.ParseAtomicLevel(cfg.LogLevel); err != nil {
		fmt.Printf("Invalid log level supplied (%s), defaulting to info\n", err)
		atomicLogLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	log := logger.New(atomicLogLevel, cfg.LogDevelopment)
	defer log.Sync()

	// attach logger to context
	ctx := logger.OnContext(context.Background(), log)

	pool, err := db.NewPool(ctx, cfg.DSN)
	if err != nil {
		log.Fatalw("Error initializing DB pool",
			"error", err,
		)
	}

	runMigrations(ctx, pool)

	go cache.StartCacheClearer(ctx, pool)

	resolver.InitializeStaticResponses(ctx, cfg)
	thumbnail.InitializeConfig(cfg)
	defer thumbnail.Shutdown()

	router := chi.NewRouter()
	prometheusMiddleware := chiprometheus.NewMiddleware("c2api")
	router.Use(prometheusMiddleware)

	// Strip trailing slashes from API requests
	router.Use(StripSlashes)

	var helixUsernameCache cache.Cache

	helixClient, err := twitchapiclient.New(ctx, cfg)
	if err != nil {
		log.Warnw("Error initializing Twitch API client",
			"error", err,
		)
	} else {
		helixUsernameCache = twitchusernamecache.New(ctx, cfg, pool, helixClient)
	}

	if cfg.EnablePrometheus {
		// Host a prometheus metrics instance on cfg.PrometheusBindAddress (127.0.0.1:9382 by default)
		listenPrometheus(cfg)
	}

	twitchemotes.Initialize(ctx, cfg, pool, router, helixClient, helixUsernameCache)
	handleRoot(router)
	handleHealth(router)
	handleLegal(router)
	defaultresolver.Initialize(ctx, cfg, pool, router, helixClient)

	listen(ctx, cfg.BindAddress, mountRouter(router, cfg, log), log)
}

// StripSlashes strips slashes at the end of a request.
// The StripSlashes middleware provided in chi has a bug, so a custom solution has to be used
// TODO: can be switched to chi middleware,
// if bug described in https://github.com/Chatterino/api/pull/422 is fixed
func StripSlashes(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 1 && path[len(path)-1] == '/' {
			r.URL.Path = path[:len(path)-1]
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
