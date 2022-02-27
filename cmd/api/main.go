package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/internal/logger"
	defaultresolver "github.com/Chatterino/api/internal/resolvers/default"
	"github.com/Chatterino/api/internal/routes/twitchemotes"
	"github.com/Chatterino/api/internal/twitchapiclient"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/thumbnail"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

func listen(bind string, router *chi.Mux, log logger.Logger) {
	srv := &http.Server{
		Handler:      router,
		Addr:         bind,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func main() {
	log := logger.New()
	defer log.Sync()

	cache.SetLogger(log)
	resolver.SetLogger(log)

	cfg := config.New(log)

	resolver.InitializeStaticResponses(cfg)
	thumbnail.InitializeConfig(cfg)

	router := chi.NewRouter()

	// Strip trailing slashes from API requests
	router.Use(middleware.StripSlashes)

	var helixUsernameCache cache.Cache

	helixClient, helixUsernameCache, err := twitchapiclient.New(cfg)
	if err != nil {
		log.Warnw("Error initializing Twitch API client", "error", err)
	}

	if cfg.EnablePrometheus {
		// Host a prometheus metrics instance on cfg.PrometheusBindAddress (127.0.0.1:9382 by default)
		listenPrometheus(cfg)
	}

	twitchemotes.Initialize(cfg, router, helixClient, helixUsernameCache)
	handleRoot(router)
	handleHealth(router)
	handleLegal(router)
	defaultresolver.Initialize(router, cfg, helixClient)

	listen(cfg.BindAddress, mountRouter(router, cfg, log), log)
}
