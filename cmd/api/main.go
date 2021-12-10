package main

import (
	"log"
	"net/http"
	"net/url"
	"time"

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

func mountRouter(r *chi.Mux, cfg config.APIConfig) *chi.Mux {
	if cfg.BaseURL == "" {
		log.Printf("Listening on %s (Prefix=%s, BaseURL=%s)\n", cfg.BindAddress, prefix, cfg.BaseURL)
		return r
	}

	// figure out prefix from address
	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		log.Fatal("Scheme must be included in base url")
	}

	prefix = u.Path
	ur := chi.NewRouter()
	ur.Mount(prefix, r)

	log.Printf("Listening on %s (Prefix=%s, BaseURL=%s)\n", cfg.BindAddress, prefix, cfg.BaseURL)

	return ur
}

func listen(bind string, router *chi.Mux) {
	srv := &http.Server{
		Handler:      router,
		Addr:         bind,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func main() {
	cfg := config.New()

	resolver.InitializeStaticResponses(cfg)
	thumbnail.InitializeConfig(cfg)

	router := chi.NewRouter()

	// Strip trailing slashes from API requests
	router.Use(middleware.StripSlashes)

	var helixUsernameCache *cache.Cache

	helixClient, helixUsernameCache, err := twitchapiclient.New(cfg)
	if err != nil {
		log.Printf("[Twitch] %s\n", err.Error())
	}

	twitchemotes.Initialize(router, helixClient, helixUsernameCache)
	handleHealth(router)
	defaultresolver.Initialize(router, cfg, helixClient)

	listen(cfg.BindAddress, mountRouter(router, cfg))
}
